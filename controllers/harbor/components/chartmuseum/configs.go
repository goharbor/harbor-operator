package chartmuseum

import (
	"context"
	"fmt"

	"github.com/alecthomas/units"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	containerregistryv1alpha1 "github.com/ovh/harbor-operator/api/v1alpha1"
	"github.com/ovh/harbor-operator/pkg/factories/application"
)

const (
	maxUploadSize = 20 * units.MiB
)

// https://github.com/goharbor/harbor/blob/master/make/photon/prepare/templates/chartserver/env.jinja

func (c *ChartMuseum) GetConfigMaps(ctx context.Context) []*corev1.ConfigMap {
	operatorName := application.GetName(ctx)
	harborName := c.harbor.Name

	return []*corev1.ConfigMap{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      c.harbor.NormalizeComponentName(containerregistryv1alpha1.ChartMuseumName),
				Namespace: c.harbor.Namespace,
				Labels: map[string]string{
					"app":      containerregistryv1alpha1.ChartMuseumName,
					"harbor":   harborName,
					"operator": operatorName,
				},
			},
			Data: map[string]string{
				"PORT":                       fmt.Sprintf("%d", port),
				"STORAGE":                    "local",
				"BASIC_AUTH_USER":            "chart_controller",
				"STORAGE_LOCAL_ROOTDIR":      "/mnt/chartmuseum",
				"DEPTH":                      "1",
				"DEBUG":                      "false",
				"LOG_JSON":                   "true",
				"DISABLE_METRICS":            "false",
				"DISABLE_API":                "false",
				"DISABLE_STATEFILES":         "false",
				"ALLOW_OVERWRITE":            "true",
				"CHART_URL":                  fmt.Sprintf("%s/chartrepo", c.harbor.Spec.PublicURL),
				"AUTH_ANONYMOUS_GET":         "false",
				"INDEX_LIMIT":                "0",
				"MAX_STORAGE_OBJECTS":        "0",
				"MAX_UPLOAD_SIZE":            fmt.Sprintf("%d", maxUploadSize),
				"CHART_POST_FORM_FIELD_NAME": "chart",
				"PROV_POST_FORM_FIELD_NAME":  "prov",
			},
		},
	}
}
