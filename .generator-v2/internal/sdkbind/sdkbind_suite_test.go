package sdkbind

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSdkbind(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Sdkbind Suite")
}
