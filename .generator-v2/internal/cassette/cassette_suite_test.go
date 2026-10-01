package cassette

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCassette(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cassette Suite")
}
