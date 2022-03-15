package evalctx_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestEvalctx(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Evalctx Suite")
}
