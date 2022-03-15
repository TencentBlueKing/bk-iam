package condition

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("BaseLogicalCondition", func() {
	var c *baseLogicalCondition
	BeforeEach(func() {
		c1, _ := newStringEqualsCondition("k1", []interface{}{"a", "b"})
		c2, _ := newNumericEqualsCondition("k2", []interface{}{"b", "c"})
		c = &baseLogicalCondition{
			content: []Condition{
				c1,
				c2,
			},
		}
	})

	It("GetKeys", func() {
		keys := c.GetKeys()
		assert.Len(GinkgoT(), keys, 2)
		assert.Contains(GinkgoT(), keys, "k1")
		assert.Contains(GinkgoT(), keys, "k2")
	})

	Describe("HasKey", func() {
		It("ok", func() {
			ok1 := c.HasKey(func(key string) bool {
				return key == "k1"
			})
			assert.True(GinkgoT(), ok1)
		})

		It("not ok", func() {
			ok2 := c.HasKey(func(key string) bool {
				return key == "k3"
			})
			assert.False(GinkgoT(), ok2)
		})
	})

	Describe("GetFirstMatchKeyValues", func() {
		It("ok", func() {
			v, ok := c.GetFirstMatchKeyValues(func(key string) bool {
				return key == "k1"
			})
			assert.True(GinkgoT(), ok)
			assert.Equal(GinkgoT(), []interface{}{"a", "b"}, v)
		})

		It("not ok", func() {
			_, ok := c.GetFirstMatchKeyValues(func(key string) bool {
				return key == "k3"
			})
			assert.False(GinkgoT(), ok)
		})

	})
})
