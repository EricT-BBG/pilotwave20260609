package authenticator

import (
	"fmt"
	"log"

	//	"strings"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io/ioutil"

	"github.com/spf13/viper"
	"golang.org/x/crypto/pkcs12"

	//	"gopkg.in/ldap.v3"
	"github.com/go-ldap/ldap/v3"
)

type ADConnector struct {
	host       string
	port       int
	enabledTLS bool
	username   string
	password   string
}

func NewADConnector() *ADConnector {
	return &ADConnector{}
}

func (adClient *ADConnector) Init() {
	adClient.host = viper.GetString("active_directory.host")
	adClient.port = viper.GetInt("active_directory.port")
	adClient.enabledTLS = viper.GetBool("active_directory.tls")
	adClient.username = viper.GetString("active_directory.username")
	adClient.password = viper.GetString("active_directory.password")
}

func (adClient *ADConnector) LoadCert() (*tls.Config, error) {

	// Load client cert
	cert, err := tls.LoadX509KeyPair(viper.GetString("active_directory.cert_file"), viper.GetString("active_directory.key_file"))
	if err != nil {
		return nil, err
	}

	// Load CA cert
	caCert, err := ioutil.ReadFile(viper.GetString("active_directory.ca_file"))
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Setup config
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
		ServerName:   viper.GetString("active_directory.domain"),
	}

	tlsConfig.BuildNameToCertificate()

	return tlsConfig, nil
}

func (adClient *ADConnector) LoadPFX(pfxFile string, password string) (*tls.Certificate, error) {

	// Loading file
	log.Printf("Loading p12 file... %s", pfxFile)
	pfxData, err := ioutil.ReadFile(pfxFile)
	if err != nil {
		return nil, err
	}

	// Read blocks of PFX data
	blocks, err := pkcs12.ToPEM(pfxData, password)
	if err != nil {
		return nil, err
	}

	var pemData []byte
	for _, b := range blocks {
		pemData = append(pemData, pem.EncodeToMemory(b)...)
	}

	// then use PEM data for tls to construct tls certificate:
	cert, err := tls.X509KeyPair(pemData, pemData)
	if err != nil {
		return nil, err
	}
	/*
		//_, certificate, err := pkcs12.Decode(pfxData, viper.GetString("active_directory.p12_password"))
		_, cert, err := pkcs12.Decode(pfxData, password)
		if err != nil {
			return nil, err
		}
	*/
	return &cert, nil
}

func (adClient *ADConnector) Dial() (*ldap.Conn, error) {

	host := fmt.Sprintf("%s:%d", adClient.host, adClient.port)
	log.Printf("Connect to AD ... %s", host)

	var l *ldap.Conn

	// Reconnect with TLS
	if adClient.enabledTLS {
		/*
			//err = l.StartTLS(&tls.Config{InsecureSkipVerify: true})
			tlsConfig, err := adClient.LoadCert()
			if err != nil {
				return nil, err
			}

			err = l.StartTLS(tlsConfig)
			if err != nil {
				return nil, err
			}
		*/
		tlsConfig, err := adClient.LoadCert()
		if err != nil {
			return nil, err
		}

		l, err = ldap.DialTLS("tcp", host, tlsConfig)
		if err != nil {
			return nil, err
		}
	} else {

		conn, err := ldap.Dial("tcp", host)
		if err != nil {
			return nil, err
		}

		l = conn
	}

	log.Printf("Authenticating %s ...", adClient.username)

	err := l.Bind(adClient.username, adClient.password)
	if err != nil {
		return nil, err
	}

	return l, nil
}

func (adClient *ADConnector) Exists(username string) (bool, error) {

	conn, err := adClient.Dial()
	if err != nil {
		return false, err
	}

	defer conn.Close()

	// Search for the given username
	log.Printf("login: %s", username)
	searchRequest := ldap.NewSearchRequest(
		//		strings.Join(dcStr, ","),
		viper.GetString("active_directory.base"),
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		//fmt.Sprintf("(&(objectClass=organizationalPerson)(uid=%s))", username),
		fmt.Sprintf("(mail=%s)", username),
		[]string{"dn", "ou", "mail", "uid"},
		nil,
	)

	log.Printf("Searching user ... %s", username)

	sr, err := conn.Search(searchRequest)
	if err != nil {
		return false, err
	}

	if len(sr.Entries) == 0 {
		return false, errors.New("User does not exist")
	}

	if len(sr.Entries) > 1 {
		return false, errors.New("too many entries returned")
	}

	// Rebind as the read only user for any further queries
	err = conn.Bind(adClient.username, adClient.password)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (adClient *ADConnector) Verify(username string, password string) (bool, error) {

	conn, err := adClient.Dial()
	if err != nil {
		return false, err
	}

	defer conn.Close()

	// Search for the given username
	log.Printf("login: %s", username)
	searchRequest := ldap.NewSearchRequest(
		//		strings.Join(dcStr, ","),
		viper.GetString("active_directory.base"),
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		//fmt.Sprintf("(&(objectClass=organizationalPerson)(uid=%s))", username),
		fmt.Sprintf("(mail=%s)", username),
		[]string{"dn", "ou", "mail", "uid"},
		nil,
	)

	log.Printf("Searching user ... %s", username)

	sr, err := conn.Search(searchRequest)
	if err != nil {
		return false, err
	}

	if len(sr.Entries) == 0 {
		return false, errors.New("User does not exist")
	}

	if len(sr.Entries) > 1 {
		return false, errors.New("too many entries returned")
	}

	userdn := sr.Entries[0].DN

	log.Printf("Verifying username ... %s", userdn)

	// Bind as the user to verify their password
	err = conn.Bind(userdn, password)
	if err != nil {
		return false, err
	}

	// Rebind as the read only user for any further queries
	err = conn.Bind(adClient.username, adClient.password)
	if err != nil {
		return false, err
	}

	return true, nil
}
