package auth

import "testing"

func TestValidateHost(t *testing.T) {
	if err := ValidateHost("https://use.virtualtext.app"); err != nil {
		t.Fatal(err)
	}
	if err := ValidateHost("http://127.0.0.1:3000"); err != nil {
		t.Fatal(err)
	}
	if err := ValidateHost("http://evil.example"); err == nil {
		t.Fatal("expected http remote host to fail")
	}
}
