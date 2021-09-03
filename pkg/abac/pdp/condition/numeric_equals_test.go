package condition

import (
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("NumericEquals", func() {
	var c *NumericEqualsCondition
	BeforeEach(func() {
		c = &NumericEqualsCondition{
			baseCondition{
				Key:   "ok",
				Value: []interface{}{1, 2},
			},
		}
	})

	It("new", func() {
		condition, err := newNumericEqualsCondition("ok", []interface{}{1, 2})
		assert.NoError(GinkgoT(), err)
		assert.NotNil(GinkgoT(), condition)
	})

	It("GetName", func() {
		assert.Equal(GinkgoT(), "NumericEquals", c.GetName())
	})
	Context("Eval", func() {
		It("true", func() {
			assert.True(GinkgoT(), c.Eval(ctx(1)))
			assert.True(GinkgoT(), c.Eval(ctx(2)))
		})

		It("false", func() {
			assert.False(GinkgoT(), c.Eval(ctx(3)))
		})

		It("attr list", func() {
			assert.True(GinkgoT(), c.Eval(listCtx{2, 3}))
			assert.False(GinkgoT(), c.Eval(listCtx{3, 4}))
		})

	})

	Describe("Translate", func() {
		It("fail, empty value", func() {
			c1, err := newNumericEqualsCondition("key", []interface{}{})
			assert.NoError(GinkgoT(), err)

			_, err = c1.Translate()
			assert.Error(GinkgoT(), err)
			assert.Equal(GinkgoT(), errMustNotEmpty, err)
		})

		It("ok, eq", func() {
			expected := map[string]interface{}{
				"op":    "eq",
				"field": "key",
				"value": 1,
			}
			c, err := newNumericEqualsCondition("key", []interface{}{1})
			assert.NoError(GinkgoT(), err)

			c1, err := c.Translate()
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), expected, c1)
		})

		It("ok, in", func() {
			expected := map[string]interface{}{
				"op":    "in",
				"field": "key",
				"value": []interface{}{1, 2},
			}
			c, err := newNumericEqualsCondition("key", []interface{}{1, 2})
			assert.NoError(GinkgoT(), err)

			c1, err := c.Translate()
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), expected, c1)

		})
	})

})
