package prefixed_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestLogrusPrefixedFormatter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "LogrusPrefixedFormatter Suite")
}
