package dao_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDao(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dao Suite")
}
