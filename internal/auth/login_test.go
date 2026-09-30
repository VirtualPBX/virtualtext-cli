package auth

import (
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"
)

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

func TestCallbackPageRemainsAvailable(t *testing.T) {
	const host = "https://use.virtualtext.app"
	lb, err := startLoopback(host, "state-1")
	if err != nil {
		t.Fatal(err)
	}
	defer lb.stop()

	body := getCallback(t, lb.callback+"?code=one-time&state=state-1")
	if !strings.Contains(body, "<!DOCTYPE html>") {
		t.Fatalf("expected an HTML document, got %s", body)
	}
	if !strings.Contains(body, "VirtualText CLI is authorized. You can close this window.") {
		t.Fatalf("missing success sentence: %s", body)
	}
	if !strings.Contains(body, `href="https://use.virtualtext.app"`) {
		t.Fatalf("missing return link: %s", body)
	}
	if strings.Contains(body, "one-time") {
		t.Fatalf("callback page included the authorization code: %s", body)
	}

	select {
	case code := <-lb.codeCh:
		if code != "one-time" {
			t.Fatalf("code = %q", code)
		}
	case <-time.After(time.Second):
		t.Fatal("code was not delivered")
	}

	conn, err := net.DialTimeout("tcp", lb.addr, time.Second)
	if err != nil {
		t.Fatalf("listener closed before the page was read: %v", err)
	}
	conn.Close()

	cancelled := getCallback(t, lb.callback+"?error=access_denied&state=state-1")
	if !strings.Contains(cancelled, "Authorization cancelled. You can close this window.") {
		t.Fatalf("missing cancel sentence: %s", cancelled)
	}
}

func getCallback(t *testing.T, rawURL string) string {
	t.Helper()
	resp, err := http.Get(rawURL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Fatalf("content-type %q", ct)
	}
	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return string(payload)
}
