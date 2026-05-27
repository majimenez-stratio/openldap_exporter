package openldap_exporter

import (
	"os"
	"testing"
)

func TestProcessTLSoptions_LDAPdefaults(t *testing.T) {
	cfg := NewLDAPConfig()
	if err := cfg.ProcessTLSoptions("ldap://myhost", false, false); err != nil {
		t.Fatal(err)
	}
	if cfg.Scheme != SchemeLDAP {
		t.Errorf("scheme: got %q, want %q", cfg.Scheme, SchemeLDAP)
	}
	if cfg.Host != "myhost" {
		t.Errorf("host: got %q, want %q", cfg.Host, "myhost")
	}
	if cfg.Port != "389" {
		t.Errorf("port: got %q, want %q", cfg.Port, "389")
	}
	if cfg.Addr != "myhost:389" {
		t.Errorf("addr: got %q, want %q", cfg.Addr, "myhost:389")
	}
	if cfg.UseTLS {
		t.Error("UseTLS should be false for ldap://")
	}
	if cfg.UseStartTLS {
		t.Error("UseStartTLS should be false when not requested")
	}
}

func TestProcessTLSoptions_LDAPSdefaults(t *testing.T) {
	cfg := NewLDAPConfig()
	if err := cfg.ProcessTLSoptions("ldaps://secure.host", false, false); err != nil {
		t.Fatal(err)
	}
	if cfg.Scheme != SchemeLDAPS {
		t.Errorf("scheme: got %q, want %q", cfg.Scheme, SchemeLDAPS)
	}
	if cfg.Port != "636" {
		t.Errorf("port: got %q, want %q", cfg.Port, "636")
	}
	if !cfg.UseTLS {
		t.Error("UseTLS should be true for ldaps://")
	}
}

func TestProcessTLSoptions_ExplicitPort(t *testing.T) {
	cfg := NewLDAPConfig()
	if err := cfg.ProcessTLSoptions("ldap://myhost:1389", false, false); err != nil {
		t.Fatal(err)
	}
	if cfg.Addr != "myhost:1389" {
		t.Errorf("addr: got %q, want %q", cfg.Addr, "myhost:1389")
	}
}

func TestProcessTLSoptions_StartTLS(t *testing.T) {
	cfg := NewLDAPConfig()
	if err := cfg.ProcessTLSoptions("ldap://myhost", true, false); err != nil {
		t.Fatal(err)
	}
	if !cfg.UseStartTLS {
		t.Error("UseStartTLS should be true")
	}
	if cfg.UseTLS {
		t.Error("UseTLS should be false when using StartTLS")
	}
}

// StartTLS is ignored when already using ldaps://
func TestProcessTLSoptions_StartTLSIgnoredForLDAPS(t *testing.T) {
	cfg := NewLDAPConfig()
	if err := cfg.ProcessTLSoptions("ldaps://myhost", true, false); err != nil {
		t.Fatal(err)
	}
	if cfg.UseStartTLS {
		t.Error("UseStartTLS must be false when TLS is already active")
	}
}

func TestProcessTLSoptions_InsecureSkipVerify(t *testing.T) {
	cfg := NewLDAPConfig()
	if err := cfg.ProcessTLSoptions("ldaps://myhost", false, true); err != nil {
		t.Fatal(err)
	}
	if !cfg.TLSConfig.InsecureSkipVerify {
		t.Error("InsecureSkipVerify should be true")
	}
}

func TestProcessTLSoptions_LDAPIscheme(t *testing.T) {
	cfg := NewLDAPConfig()
	if err := cfg.ProcessTLSoptions("ldapi:///var/run/ldapi", false, false); err != nil {
		t.Fatal(err)
	}
	if cfg.Scheme != SchemeLDAPI {
		t.Errorf("scheme: got %q, want %q", cfg.Scheme, SchemeLDAPI)
	}
	if cfg.Protocol != "unix" {
		t.Errorf("protocol: got %q, want %q", cfg.Protocol, "unix")
	}
}

func TestProcessTLSoptions_UnknownScheme(t *testing.T) {
	cfg := NewLDAPConfig()
	err := cfg.ProcessTLSoptions("ftp://myhost", false, false)
	if err == nil {
		t.Error("expected error for unknown scheme, got nil")
	}
}

func TestLoadCACert_NonExistentFile(t *testing.T) {
	cfg := NewLDAPConfig()
	err := cfg.LoadCACert("/nonexistent/path/ca.pem")
	if err == nil {
		t.Error("expected error for missing CA file, got nil")
	}
}

func TestLoadCACert_InvalidPEM(t *testing.T) {
	f, err := os.CreateTemp("", "ca-*.pem")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString("not a valid PEM certificate")
	f.Close()

	cfg := NewLDAPConfig()
	if err := cfg.LoadCACert(f.Name()); err == nil {
		t.Error("expected error for invalid PEM, got nil")
	}
}

func TestLoadCACert_ValidPEM(t *testing.T) {
	// Generate a real self-signed certificate so AppendCertsFromPEM succeeds.
	certPEM := generateSelfSignedCert(t)

	f, err := os.CreateTemp("", "ca-*.pem")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.Write(certPEM)
	f.Close()

	cfg := NewLDAPConfig()
	if err := cfg.LoadCACert(f.Name()); err != nil {
		t.Errorf("unexpected error with valid PEM: %v", err)
	}
	if cfg.TLSConfig.RootCAs == nil {
		t.Error("RootCAs should be set after loading a valid CA cert")
	}
}
