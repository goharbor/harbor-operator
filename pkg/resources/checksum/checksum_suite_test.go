package checksum_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestChecksum(t *testing.T) {
	t.Parallel()

	RegisterFailHandler(Fail)
	RunSpecs(t, "Checksum Suite")
}
