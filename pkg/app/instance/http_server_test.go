package instance

import (
	"net"
	"strings"
	"testing"

	httpserver "git.brobridge.com/pilotwave/pilotwave/pkg/http_server/server"
	muxmanager "git.brobridge.com/pilotwave/pilotwave/pkg/mux_manager/manager"
	"github.com/spf13/viper"
)

func TestInitHTTPServerReturnsListenError(t *testing.T) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("listen on test port: %v", err)
	}
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port
	previousPort := viper.Get("service.port")
	defer viper.Set("service.port", previousPort)
	viper.Set("service.port", port)

	app := NewAppInstance()
	app.muxManager = muxmanager.NewMuxManager(app)
	app.httpServer = httpserver.NewServer(app)

	err = app.initHTTPServer()
	if err == nil {
		t.Fatal("expected initHTTPServer to return listen error for occupied port")
	}
	if !strings.Contains(err.Error(), "address already in use") {
		t.Fatalf("expected address-in-use error, got %v", err)
	}
}
