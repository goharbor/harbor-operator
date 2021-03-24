package k8s

import (
	"strconv"

	"github.com/mitchellh/hashstructure/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// HarborClusterLastAppliedHash contains the last applied hash.
	HarborClusterLastAppliedHash = "goharbor.io/last-applied-hash"
)

func SetLastAppliedHash(obj metav1.Object, value interface{}) error {
	hash, err := hashstructure.Hash(value, hashstructure.FormatV2, nil)
	if err != nil {
		return err
	}

	annotations := obj.GetAnnotations()

	if annotations == nil {
		annotations = make(map[string]string)
	}

	annotations[HarborClusterLastAppliedHash] = strconv.FormatUint(hash, 10)
	obj.SetAnnotations(annotations)

	return nil
}

// UpdateLastAppliedHash update the list-applied-hash annotation in to Object.
func UpdateLastAppliedHash(to, from metav1.Object) {
	if to == nil || from == nil {
		return
	}

	if from.GetAnnotations() == nil {
		return
	}

	annotations := to.GetAnnotations()

	if annotations == nil {
		annotations = make(map[string]string)
	}

	annotations[HarborClusterLastAppliedHash] = from.GetAnnotations()[HarborClusterLastAppliedHash]
	to.SetAnnotations(annotations)
}

func HashEquals(o1, o2 metav1.Object) bool {
	if o1 == nil || o2 == nil {
		return o1 == o2
	}

	return o1.GetAnnotations()[HarborClusterLastAppliedHash] ==
		o2.GetAnnotations()[HarborClusterLastAppliedHash]
}
