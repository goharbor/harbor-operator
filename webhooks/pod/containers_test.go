package pod_test

import (
	"testing"

	"github.com/containers/image/v5/docker/reference"
	"github.com/goharbor/harbor-operator/webhooks/pod"
	"github.com/stretchr/testify/require"
)

func Test_RegistryFromImageRef(t *testing.T) {
	type testcase struct {
		description      string
		imageRef         string
		expectedRegistry string
	}

	tests := []testcase{
		{
			description:      "image reference with hostname with port and image tag set",
			imageRef:         "somehost:443/public/busybox:latest",
			expectedRegistry: "somehost:443",
		},
		{
			description:      "image reference with hostname with port and no image tag set",
			imageRef:         "somehost:443/public/busybox",
			expectedRegistry: "somehost:443",
		},
		{
			description:      "image reference with hostname with port and image sha set",
			imageRef:         "somehost:443/public/busybox@sha256:7cc4b5aefd1d0cadf8d97d4350462ba51c694ebca145b08d7d41b41acc8db5aa",
			expectedRegistry: "somehost:443",
		},
		{
			description:      "image reference with url and image tag set",
			imageRef:         "example.com/busybox:latest",
			expectedRegistry: "example.com",
		},
		{
			description:      "image reference with url and no image tag set",
			imageRef:         "example.com/busybox",
			expectedRegistry: "example.com",
		},
		{
			description:      "image reference with url, project and no image tag set",
			imageRef:         "example.com/nginxinc/nginx-unprivileged",
			expectedRegistry: "example.com",
		},
		{
			description:      "image reference with url and image sha set",
			imageRef:         "example.com/busybox@sha256:7cc4b5aefd1d0cadf8d97d4350462ba51c694ebca145b08d7d41b41acc8db5aa",
			expectedRegistry: "example.com",
		},
		{
			description:      "bare image reference with image tag set",
			imageRef:         "busybox:latest",
			expectedRegistry: pod.BareRegistry,
		},
		{
			description:      "bare image reference with project and no image tag set",
			imageRef:         "nginxinc/nginx-unprivileged",
			expectedRegistry: pod.BareRegistry,
		},
		{
			description:      "bare image reference with and no image tag set",
			imageRef:         "busybox",
			expectedRegistry: pod.BareRegistry,
		},
		{
			description:      "bare image reference with image sha set",
			imageRef:         "busybox@sha256:7cc4b5aefd1d0cadf8d97d4350462ba51c694ebca145b08d7d41b41acc8db5aa",
			expectedRegistry: pod.BareRegistry,
		},
	}

	for _, testcase := range tests {
		output, err := pod.RegistryFromImageRef(testcase.imageRef)
		require.NoError(t, err, testcase.description)
		require.Equal(t, testcase.expectedRegistry, output, testcase.description)
	}
}

func Test_RegistryFromImageRef_EmptyErr(t *testing.T) {
	_, err := pod.RegistryFromImageRef("")
	require.EqualError(t, err, reference.ErrReferenceInvalidFormat.Error())
}

func Test_ReplaceRegistryInImageRef(t *testing.T) {
	type testcase struct {
		description string
		imageRef    string
		newRegistry string
		expectedRef string
	}

	tests := []testcase{
		{
			description: "image reference with hostname with port and image tag set",
			imageRef:    "somehost:443/public/busybox:1.32.0",
			newRegistry: "harbor.example.com/proxy-cache",
			expectedRef: "harbor.example.com/proxy-cache/public/busybox:1.32.0",
		},
		{
			description: "image reference with hostname with port and no image tag set",
			imageRef:    "somehost:443/public/busybox",
			newRegistry: "harbor.example.com/proxy-cache",
			expectedRef: "harbor.example.com/proxy-cache/public/busybox:latest",
		},
		{
			description: "image reference with hostname with port and image sha set",
			imageRef:    "somehost:443/public/busybox@sha256:7cc4b5aefd1d0cadf8d97d4350462ba51c694ebca145b08d7d41b41acc8db5aa",
			newRegistry: "harbor.example.com/proxy-cache",
			expectedRef: "harbor.example.com/proxy-cache/public/busybox@sha256:7cc4b5aefd1d0cadf8d97d4350462ba51c694ebca145b08d7d41b41acc8db5aa",
		},
		{
			description: "image reference with url and image tag set",
			imageRef:    "example.com/busybox:1.32.0",
			newRegistry: "harbor.example.com/proxy-cache",
			expectedRef: "harbor.example.com/proxy-cache/busybox:1.32.0",
		},
		{
			description: "image reference with url and no image tag set",
			imageRef:    "example.com/busybox",
			newRegistry: "harbor.example.com/proxy-cache",
			expectedRef: "harbor.example.com/proxy-cache/busybox:latest",
		},
		{
			description: "image reference with url and image sha set",
			imageRef:    "example.com/busybox@sha256:7cc4b5aefd1d0cadf8d97d4350462ba51c694ebca145b08d7d41b41acc8db5aa",
			newRegistry: "harbor.example.com/proxy-cache",
			expectedRef: "harbor.example.com/proxy-cache/busybox@sha256:7cc4b5aefd1d0cadf8d97d4350462ba51c694ebca145b08d7d41b41acc8db5aa",
		},
		{
			description: "bare image reference with image tag set",
			imageRef:    "busybox:1.32.0",
			newRegistry: "harbor.example.com/proxy-cache",
			expectedRef: "harbor.example.com/proxy-cache/library/busybox:1.32.0",
		},
		{
			description: "bare image reference with and no image tag set",
			imageRef:    "busybox",
			newRegistry: "harbor.example.com/proxy-cache",
			expectedRef: "harbor.example.com/proxy-cache/library/busybox:latest",
		},
		{
			description: "bare image reference with image sha set",
			imageRef:    "busybox@sha256:7cc4b5aefd1d0cadf8d97d4350462ba51c694ebca145b08d7d41b41acc8db5aa",
			newRegistry: "harbor.example.com/proxy-cache",
			expectedRef: "harbor.example.com/proxy-cache/library/busybox@sha256:7cc4b5aefd1d0cadf8d97d4350462ba51c694ebca145b08d7d41b41acc8db5aa",
		},
	}

	for _, testcase := range tests {
		output, err := pod.ReplaceRegistryInImageRef(testcase.imageRef, testcase.newRegistry)
		require.NoError(t, err)
		require.Equal(t, testcase.expectedRef, output, testcase.description)
	}
}
