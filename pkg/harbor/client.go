package harbor

import (
	"context"
	"crypto/tls"
	"net/http"
)

// Client defines common harbor client interface.
type Client interface {
	// ApplyConfiguration applies configuration to harbor instance.
	ApplyConfiguration(ctx context.Context, config []byte) error
}

// ClientOption wraps client options.
type ClientOption func(*Options)

// Options defines client options.
type Options struct {
	credential *credential
	httpClient *http.Client
}

type credential struct {
	username string
	password string
}

// client implements Client.
type client struct {
	url  string
	opts *Options
}

// NewClient constructs harbor client.
func NewClient(url string, opts ...ClientOption) Client {
	clientOpts := defaultOptions()

	for _, o := range opts {
		o(clientOpts)
	}

	return &client{
		url:  url,
		opts: clientOpts,
	}
}

// WithCredential injects credential.
func WithCredential(username, password string) ClientOption {
	return func(opts *Options) {
		cred := &credential{username: username, password: password}
		opts.credential = cred
	}
}

// WithHTTPClient injects http client.
func WithHTTPClient(ct *http.Client) ClientOption {
	return func(opts *Options) {
		opts.httpClient = ct
	}
}

// defaultOptions returns default options.
func defaultOptions() *Options {
	// skip cert verify
	ts := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	return &Options{httpClient: &http.Client{Transport: ts}}
}
