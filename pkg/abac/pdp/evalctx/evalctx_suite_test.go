package evalctx_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestEvalctx(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Evalctx Suite")
}
