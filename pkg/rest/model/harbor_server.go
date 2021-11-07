package model

import (
	"net/url"

	gruntime "github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	hc "github.com/goharbor/go-client/pkg/harbor"
	assistclient "github.com/goharbor/go-client/pkg/sdk/assist/client"
	v2client "github.com/goharbor/go-client/pkg/sdk/v2.0/client"
	legacyclient "github.com/goharbor/go-client/pkg/sdk/v2.0/legacy/client"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
)

const (
	accessKey    = "accessKey"
	accessSecret = "accessSecret"
)

// AccessCred contains credential data for accessing the harbor server.
type AccessCred struct {
	AccessKey    string
	AccessSecret string
}

// FillIn put secret into AccessCred.
func (ac *AccessCred) FillIn(secret *corev1.Secret) error {
	decodedAK, ok1 := secret.Data[accessKey]
	decodedAS, ok2 := secret.Data[accessSecret]

	if !(ok1 && ok2) {
		return errors.New("invalid access secret")
	}

	ac.AccessKey = string(decodedAK)
	ac.AccessSecret = string(decodedAS)

	return nil
}

// Validate validates wether the key and secret has correct format.
func (ac *AccessCred) Validate(secret *corev1.Secret) error {
	if len(ac.AccessKey) == 0 || len(ac.AccessSecret) == 0 {
		return errors.New("access key and secret can't be empty")
	}

	return nil
}

// HarborServer contains connection data.
type HarborServer struct {
	ServerURL  *url.URL
	AccessCred *AccessCred
	InSecure   bool
}

// NewHarborServer returns harbor server with inputs.
func NewHarborServer(serverURL *url.URL, accessCred *AccessCred, insecure bool) *HarborServer {
	return &HarborServer{
		ServerURL:  serverURL,
		AccessCred: accessCred,
		InSecure:   insecure,
	}
}

// HarborAssistClient keeps Harbor client.
type HarborAssistClient struct {
	Client *assistclient.HarborAPI
}

// HarborLegacyClient keeps Harbor client.
type HarborLegacyClient struct {
	Client *legacyclient.HarborAPI
}

// HarborClientV2 keeps Harbor client v2.
type HarborClientV2 struct {
	Client *v2client.HarborAPI
	Auth   gruntime.ClientAuthInfoWriter
}

// ClientLegacy created based on the server data.
func (h *HarborServer) ClientLegacy() *HarborLegacyClient {
	c := &hc.Config{
		URL:      h.ServerURL,
		AuthInfo: httptransport.BasicAuth(h.AccessCred.AccessKey, h.AccessCred.AccessSecret),
	}

	cs := hc.NewClientSet(c)

	return &HarborLegacyClient{
		Client: cs.Legacy(),
	}
}

// ClientV2 created based on the server data. Harbor V2 API.
func (h *HarborServer) ClientV2() *HarborClientV2 {

	c := &hc.Config{
		URL:      h.ServerURL,
		AuthInfo: httptransport.BasicAuth(h.AccessCred.AccessKey, h.AccessCred.AccessSecret),
	}

	cs := hc.NewClientSet(c)

	return &HarborClientV2{
		Client: cs.V2(),
	}
}
