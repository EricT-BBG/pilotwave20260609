package api

import (
	"net/http"
	"strings"

	cluster_bridge "git.brobridge.com/pilotwave/pilotwave/pkg/cluster_bridge"
	"github.com/gin-gonic/gin"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

func isUpdateConflict(err error) bool {
	if err == nil {
		return false
	}

	if k8serrors.IsConflict(err) {
		return true
	}

	return strings.Contains(strings.ToLower(err.Error()), "resource version changed")
}

func writeClusterUnavailable(c *gin.Context, err error) bool {
	if !cluster_bridge.IsIstioUnavailable(err) {
		return false
	}

	c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
	c.Abort()
	return true
}
