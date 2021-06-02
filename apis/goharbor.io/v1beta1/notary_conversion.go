package v1beta1

import "github.com/goharbor/harbor-operator/apis/goharbor.io/v1alpha3"

func Convert_v1alpha3_NotaryLoggingSpec_To_v1beta1_NotaryLoggingSpec(src *v1alpha3.NotaryLoggingSpec, dst *NotaryLoggingSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &NotaryLoggingSpec{}
	}

	dst.Level = src.Level

}

func Convert_v1alpha3_NotaryStorageSpec_To_v1beta1_NotaryStorageSpec(src *v1alpha3.NotaryStorageSpec, dst *NotaryStorageSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &NotaryStorageSpec{}
	}

	dst.Postgres = src.Postgres
}

func Convert_v1beta1_NotaryLoggingSpec_To_v1alpha3_NotaryLoggingSpec(src *NotaryLoggingSpec, dst *v1alpha3.NotaryLoggingSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &v1alpha3.NotaryLoggingSpec{}
	}

	dst.Level = src.Level

}

func Convert_v1beta1_NotaryStorageSpec_To_v1alpha3_NotaryStorageSpec(src *NotaryStorageSpec, dst *v1alpha3.NotaryStorageSpec) {
	if src == nil {
		return
	}

	if dst == nil {
		dst = &v1alpha3.NotaryStorageSpec{}
	}

	dst.Postgres = src.Postgres
}
