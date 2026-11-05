package providermod

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestProvidermod(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Providermod Suite")
}
