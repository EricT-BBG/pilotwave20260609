package cluster_bridge

import "testing"

func TestIsSystemNamespaceName(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{name: "kube-system", want: true},
		{name: "kube-public", want: true},
		{name: "openshift-monitoring", want: true},
		{name: "istio-system", want: true},
		{name: "default", want: false},
		{name: "monitoring", want: false},
		{name: "pilotwave", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSystemNamespaceName(tt.name); got != tt.want {
				t.Fatalf("IsSystemNamespaceName(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}
