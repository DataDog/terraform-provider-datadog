package split

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSplit(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Split Suite")
}
