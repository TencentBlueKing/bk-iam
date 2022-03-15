package cacheimpls_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCacheimpls(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cacheimpls Suite")
}
