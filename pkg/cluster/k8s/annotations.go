package k8s

const (
	// HarborClusterLastAppliedHash contains the last applied hash.
	HarborClusterLastAppliedHash = "goharbor.io/last-applied-hash"
)

func GetLastAppliedHash(annotations map[string]string) string {
	if annotations == nil {
		return ""
	}
	return annotations[HarborClusterLastAppliedHash]
}
