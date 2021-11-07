package rest

import (
	"context"
	"fmt"
	"net/url"

	goharborv1beta1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	"github.com/goharbor/harbor-operator/pkg/rest/legacy"
	"github.com/goharbor/harbor-operator/pkg/rest/model"
	v2 "github.com/goharbor/harbor-operator/pkg/rest/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateHarborClients(ctx context.Context, client client.Client, hsc *goharborv1beta1.HarborServerConfiguration) (*v2.Client, *legacy.Client, error) {
	server, err := createHarborServer(ctx, client, hsc)
	if err != nil {
		return nil, nil, err
	}

	return v2.NewWithServer(server), legacy.NewWithServer(server), nil
}

func CreateHarborV2Client(ctx context.Context, client client.Client, hsc *goharborv1beta1.HarborServerConfiguration) (*v2.Client, error) {
	server, err := createHarborServer(ctx, client, hsc)
	if err != nil {
		return nil, err
	}

	return v2.NewWithServer(server), nil
}

func CreateHarborLegacyClient(ctx context.Context, client client.Client, hsc *goharborv1beta1.HarborServerConfiguration) (*legacy.Client, error) {
	server, err := createHarborServer(ctx, client, hsc)
	if err != nil {
		return nil, err
	}

	return legacy.NewWithServer(server), nil
}

// Check if the server configuration is valid.
// That is checking if the admin password secret object is valid.
func createHarborServer(ctx context.Context, client client.Client, hsc *goharborv1beta1.HarborServerConfiguration) (*model.HarborServer, error) {
	// construct accessCreds from Secret
	secretNSedName := types.NamespacedName{
		Namespace: hsc.Spec.AccessCredential.Namespace,
		Name:      hsc.Spec.AccessCredential.AccessSecretRef,
	}

	cred, err := createAccessCredsFromSecret(ctx, client, secretNSedName)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(hsc.Spec.ServerURL)
	if err != nil {
		return nil, err
	}

	// put server config into client
	return model.NewHarborServer(u, cred, hsc.Spec.InSecure), nil
}

func createAccessCredsFromSecret(ctx context.Context, client client.Client, secretNSedName types.NamespacedName) (*model.AccessCred, error) {
	accessSecret := &corev1.Secret{}
	if err := client.Get(ctx, secretNSedName, accessSecret); err != nil {
		// No matter what errors (including not found) occurred, the server configuration is invalid
		return nil, fmt.Errorf("get access secret error: %w", err)
	}

	// convert secrets to AccessCred
	cred := &model.AccessCred{}
	if err := cred.FillIn(accessSecret); err != nil {
		return nil, fmt.Errorf("fill in secret error: %w", err)
	}

	return cred, nil
}
