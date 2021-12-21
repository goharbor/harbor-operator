package model

import (
	gruntime "github.com/go-openapi/runtime"
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

// GetCredential put secret into AccessCred.
func GetCredential(secret *corev1.Secret) (string, string, error) {
	decodedAK, ok1 := secret.Data[accessKey]
	decodedAS, ok2 := secret.Data[accessSecret]

	if !(ok1 && ok2) {
		return "", "", errors.New("invalid access secret")
	}

	if len(decodedAK) == 0 || len(decodedAS) == 0 {
		return "", "", errors.New("access key and secret can't be empty")
	}

	return string(decodedAK), string(decodedAS), nil
}

// HarborServer contains connection data.
type HarborServer struct {
	ServerURL string
	Username  string
	Password  string
	Insecure  bool
}

// NewHarborServer returns harbor server with inputs.
func NewHarborServer(url, username, password string, insecure bool) *HarborServer {
	return &HarborServer{
		ServerURL: url,
		Username:  username,
		Password:  password,
		Insecure:  insecure,
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

// ClientV2 created based on the server data. Harbor V2 API.
func (h *HarborServer) ClientV2() (*HarborClientV2, error) {
	c := &hc.ClientSetConfig{
		URL:      h.ServerURL,
		Username: h.Username,
		Password: h.Password,
		Insecure: h.Insecure,
	}

	cs, err := hc.NewClientSet(c)
	if err != nil {
		return nil, err
	}

	return &HarborClientV2{
		Client: cs.V2(),
	}, nil
}
