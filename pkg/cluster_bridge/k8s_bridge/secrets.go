package k8s_bridge

import (
	"context"
	"time"

	k8scorev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (br *K8sBridge) CreateSecrets(name string, namespace string, certificate string, privateKey string, caCertificate string) error {

	//Generator Secret Object
	data := make(map[string]string)
	data[k8scorev1.TLSPrivateKeyKey] = privateKey
	data[k8scorev1.TLSCertKey] = certificate
	if caCertificate != "" {
		data["ca.crt"] = caCertificate
	}

	secret := &k8scorev1.Secret{
		TypeMeta: k8smetav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: k8smetav1.ObjectMeta{
			Name: name,
		},
		Type:       k8scorev1.SecretTypeTLS,
		StringData: data,
	}

	// Create Secret
	start := time.Now()
	_, err := br.clientset.CoreV1().Secrets(namespace).Create(context.TODO(), secret, k8smetav1.CreateOptions{})
	recordK8sWrite(k8sResourceSecret, kubernetesWriteVerbCreate, start, err)
	if err != nil {
		return err
	}

	return nil
}

func (br *K8sBridge) DeleteSecrets(name string, namespace string) error {

	// Delete Secret
	start := time.Now()
	err := br.clientset.CoreV1().Secrets(namespace).Delete(context.TODO(), name, k8smetav1.DeleteOptions{})
	recordK8sWrite(k8sResourceSecret, kubernetesWriteVerbDelete, start, err)
	if err != nil {
		return err
	}

	return nil

}

func (br *K8sBridge) SecretsExist(name string, namespace string) (bool, error) {

	// Secret Exist?
	re, err := br.clientset.CoreV1().Secrets(namespace).Get(context.TODO(), name, k8smetav1.GetOptions{})

	if err != nil {
		if k8serrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	if re != nil {
		return true, nil
	}

	return false, nil
}
