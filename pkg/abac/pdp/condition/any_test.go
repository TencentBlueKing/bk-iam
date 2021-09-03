package condition

import (
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("Any", func() {
	var c *AnyCondition
	BeforeEach(func() {
		c = &AnyCondition{
			baseCondition{
				Key:   "ok",
				Value: []interface{}{"a", "b"},
			},
		}
	})

	It("new", func() {
		condition, err := newAnyCondition("ok", []interface{}{"a", "b"})
		assert.NoError(GinkgoT(), err)
		assert.NotNil(GinkgoT(), condition)
	})

	It("GetName", func() {
		assert.Equal(GinkgoT(), "Any", c.GetName())
	})

	It("GetKeys", func() {
		assert.Empty(GinkgoT(), c.GetKeys())
	})

	It("Eval", func() {
		assert.True(GinkgoT(), c.Eval(ctx(1)))
		assert.True(GinkgoT(), c.Eval(boolCtx(false)))
		assert.True(GinkgoT(), c.Eval(listCtx{1, 2}))
		assert.True(GinkgoT(), c.Eval(errCtx(1)))

	})

	//Describe("anyTranslate", func() {
	//	It("ok", func() {
	//		expected := ExprCell{
	//			"op":    "any",
	//			"field": "key",
	//			"value": []interface{}{"a"},
	//		}
	//		ec, err := anyTranslate("key", []interface{}{"a"})
	//		assert.NoError(GinkgoT(), err)
	//		assert.Equal(GinkgoT(), expected, ec)
	//	})
	//})
	//
})
