package rest

import (
	"context"
	"fmt"

	goharborv1beta1 "github.com/goharbor/harbor-operator/apis/goharbor.io/v1beta1"
	"github.com/goharbor/harbor-operator/pkg/rest/model"
	v2 "github.com/goharbor/harbor-operator/pkg/rest/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateHarborV2Client(ctx context.Context, client client.Client, hsc *goharborv1beta1.HarborServerConfiguration) (*v2.Client, error) {
	server, err := createHarborServer(ctx, client, hsc)
	if err != nil {
		return nil, err
	}

	return v2.NewWithServer(server)
}

// Check if the server configuration is valid.
// That is checking if the admin password secret object is valid.
func createHarborServer(ctx context.Context, client client.Client, hsc *goharborv1beta1.HarborServerConfiguration) (*model.HarborServer, error) {
	// construct accessCreds from Secret
	secretNSedName := types.NamespacedName{
		Namespace: hsc.Spec.AccessCredential.Namespace,
		Name:      hsc.Spec.AccessCredential.AccessSecretRef,
	}

	username, password, err := createAccessCredsFromSecret(ctx, client, secretNSedName)
	if err != nil {
		return nil, err
	}

	// put server config into client
	return model.NewHarborServer(hsc.Spec.ServerURL, username, password, hsc.Spec.Insecure), nil
}

func createAccessCredsFromSecret(ctx context.Context, client client.Client, secretNSedName types.NamespacedName) (string, string, error) {
	accessSecret := &corev1.Secret{}
	if err := client.Get(ctx, secretNSedName, accessSecret); err != nil {
		// No matter what errors (including not found) occurred, the server configuration is invalid
		return "", "", fmt.Errorf("get access secret error: %w", err)
	}

	username, password, err := model.GetCredential(accessSecret)
	if err != nil {
		return "", "", fmt.Errorf("get credential error: %w", err)
	}

	return username, password, nil
}
