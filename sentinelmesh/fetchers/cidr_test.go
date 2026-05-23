package fetchers

import (
	"testing"
)

func TestExpandCIDR(t *testing.T) {
	tests := []struct {
		name    string
		cidr    string
		wantLen int
		wantErr bool
	}{
		{"single /32", "192.168.1.1/32", 1, false},
		{"small /30", "192.168.1.0/30", 2, false},
		{"/28", "192.168.1.0/28", 14, false},
		{"full /24", "192.168.1.0/24", 254, false},
		{"invalid", "not-a-cidr", 0, true},
		{"too large", "10.0.0.0/16", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ips, err := ExpandCIDR(tt.cidr)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(ips) != tt.wantLen {
				t.Errorf("got %d ips, want %d", len(ips), tt.wantLen)
			}
		})
	}
}

func TestExpandCIDRNoBroadcastNetwork(t *testing.T) {
	ips, err := ExpandCIDR("10.0.0.0/24")
	if err != nil {
		t.Fatal(err)
	}
	// Should not contain network (10.0.0.0) or broadcast (10.0.0.255)
	for _, ip := range ips {
		if ip == "10.0.0.0" || ip == "10.0.0.255" {
			t.Errorf("should not contain %s", ip)
		}
	}
	// First should be .1, last should be .254
	if ips[0] != "10.0.0.1" {
		t.Errorf("first = %s, want 10.0.0.1", ips[0])
	}
	if ips[len(ips)-1] != "10.0.0.254" {
		t.Errorf("last = %s, want 10.0.0.254", ips[len(ips)-1])
	}
}
