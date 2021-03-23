package k8s

import (
	"github.com/mitchellh/hashstructure/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
)

const (
	// HarborClusterLastAppliedHash contains the last applied hash.
	HarborClusterLastAppliedHash = "goharbor.io/last-applied-hash"
)

func SetLastAppliedHash(obj metav1.Object) error {
	hash, err := hashstructure.Hash(obj, hashstructure.FormatV2, nil)
	if err != nil {
		return err
	}
	obj.GetAnnotations()[HarborClusterLastAppliedHash] = strconv.FormatUint(hash, 10)
	return nil
}

func HashEquals(o1, o2 metav1.Object) bool {
	if o1 == nil || o2 == nil {
		return o1 == o2
	}
	return o1.GetAnnotations()[HarborClusterLastAppliedHash] ==
		o2.GetAnnotations()[HarborClusterLastAppliedHash]
}
