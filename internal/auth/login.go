package auth

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

func RandomState() string {
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}

func ValidateHost(host string) error {
	u, err := url.Parse(host)
	if err != nil || u.Host == "" {
		return fmt.Errorf("invalid host")
	}
	if u.Scheme == "https" {
		return nil
	}
	if u.Scheme == "http" && (u.Hostname() == "127.0.0.1" || u.Hostname() == "localhost") {
		return nil
	}
	return fmt.Errorf("host must be https")
}

func Login(host, state string, timeout time.Duration) (string, error) {
	if err := ValidateHost(host); err != nil {
		return "", err
	}
	code, err := waitForCode(host, state, timeout)
	if err != nil {
		return "", err
	}
	return exchangeCode(host, code, state)
}

func waitForCode(host, state string, timeout time.Duration) (string, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", err
	}
	defer ln.Close()
	port := ln.Addr().(*net.TCPAddr).Port
	redirect := fmt.Sprintf("http://127.0.0.1:%d/callback", port)
	authorize := fmt.Sprintf("%s/cli/authorize?redirect_uri=%s&state=%s",
		strings.TrimRight(host, "/"), url.QueryEscape(redirect), url.QueryEscape(state))

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if host := r.Host; host != "" && !strings.HasPrefix(host, "127.0.0.1") && !strings.HasPrefix(host, "localhost") {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if r.URL.Query().Get("token") != "" {
			http.Error(w, "token in URL is not accepted", http.StatusBadRequest)
			errCh <- fmt.Errorf("server returned a token in the URL")
			return
		}
		if r.URL.Query().Get("state") != state {
			http.Error(w, "state mismatch", http.StatusBadRequest)
			errCh <- fmt.Errorf("state mismatch")
			return
		}
		if errMsg := r.URL.Query().Get("error"); errMsg != "" {
			fmt.Fprint(w, "Authorization cancelled. You can close this window.")
			errCh <- fmt.Errorf("authorization %s", errMsg)
			return
		}
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "missing code", http.StatusBadRequest)
			errCh <- fmt.Errorf("missing code")
			return
		}
		fmt.Fprint(w, "VirtualText CLI is authorized. You can close this window.")
		codeCh <- code
	})
	srv := &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	go srv.Serve(ln)
	defer srv.Close()

	if err := OpenBrowser(authorize); err != nil {
		fmt.Fprintf(os.Stderr, "Open this URL to authorize:\n%s\n", authorize)
	}

	select {
	case code := <-codeCh:
		return code, nil
	case err := <-errCh:
		return "", err
	case <-time.After(timeout):
		return "", fmt.Errorf("timed out waiting for browser authorization")
	}
}

func exchangeCode(host, code, state string) (string, error) {
	body, _ := json.Marshal(map[string]string{"code": code, "state": state})
	req, err := http.NewRequest(http.MethodPost, strings.TrimRight(host, "/")+"/cli/token", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	payload, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("token exchange http %d: %s", resp.StatusCode, strings.TrimSpace(string(payload)))
	}
	var out struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(payload, &out); err != nil {
		return "", err
	}
	if out.Token == "" {
		return "", fmt.Errorf("token exchange returned empty token")
	}
	return out.Token, nil
}

func OpenBrowser(rawURL string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", rawURL)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", rawURL)
	default:
		cmd = exec.Command("xdg-open", rawURL)
	}
	return cmd.Start()
}
