package app

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"
)

// Only the MAX HTTP client uses the optional extra root. System trust is unchanged.
func newMAXClient() (*http.Client, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if path := os.Getenv("MAX_CA_FILE"); path != "" {
		pem, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read MAX_CA_FILE: %w", err)
		}
		roots, err := x509.SystemCertPool()
		if err != nil {
			roots = x509.NewCertPool()
		}
		if !roots.AppendCertsFromPEM(pem) {
			return nil, errors.New("MAX_CA_FILE must contain a PEM certificate")
		}
		transport.TLSClientConfig = &tls.Config{RootCAs: roots, MinVersion: tls.VersionTLS12}
	}
	return &http.Client{Timeout: 40 * time.Second, Transport: transport}, nil
}
func CheckBot() error {
	client, err := newMAXClient()
	if err != nil {
		return err
	}
	a := &App{token: os.Getenv("MAX_BOT_TOKEN"), client: client}
	var bot struct {
		ID       int64  `json:"user_id"`
		Name     string `json:"name"`
		Username string `json:"username"`
	}
	if err = a.maxRequest(context.Background(), "GET", "/me", nil, &bot); err != nil {
		return err
	}
	fmt.Printf("Bot: %s (@%s), ID: %d\n", bot.Name, bot.Username, bot.ID)
	return nil
}
