package condition

import (
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("StringEquals", func() {

	var c *StringEqualsCondition
	BeforeEach(func() {
		c = &StringEqualsCondition{
			baseCondition{
				Key:   "ok",
				Value: []interface{}{"a", "b"},
			},
		}
	})

	It("new", func() {
		condition, err := newStringEqualsCondition("ok", []interface{}{"a", "b"})
		assert.NoError(GinkgoT(), err)
		assert.NotNil(GinkgoT(), condition)
	})

	It("GetName", func() {
		assert.Equal(GinkgoT(), "StringEquals", c.GetName())
	})

	Context("Eval", func() {
		It("true", func() {
			assert.True(GinkgoT(), c.Eval(strCtx("a")))
			assert.True(GinkgoT(), c.Eval(strCtx("b")))
		})

		It("false", func() {
			assert.False(GinkgoT(), c.Eval(strCtx("c")))
		})

		It("attr list", func() {
			assert.True(GinkgoT(), c.Eval(listCtx{"a", "d"}))

			assert.False(GinkgoT(), c.Eval(listCtx{"e", "f"}))
		})

	})

	//Describe("stringEqualsTranslate", func() {
	//	It("fail, empty value", func() {
	//		_, err := stringEqualsTranslate("key", []interface{}{})
	//		assert.Error(GinkgoT(), err)
	//		assert.Equal(GinkgoT(), errMustNotEmpty, err)
	//	})
	//
	//	It("ok, single eq", func() {
	//		expected := ExprCell{
	//			"op":    "eq",
	//			"field": "key",
	//			"value": "a",
	//		}
	//		ec, err := stringEqualsTranslate("key", []interface{}{"a"})
	//		assert.NoError(GinkgoT(), err)
	//		assert.Equal(GinkgoT(), expected, ec)
	//
	//	})
	//
	//	It("ok, multiple in", func() {
	//		expected := ExprCell{
	//			"op":    "in",
	//			"field": "key",
	//			"value": []interface{}{"a", "b"},
	//		}
	//		ec, err := stringEqualsTranslate("key", []interface{}{"a", "b"})
	//		assert.NoError(GinkgoT(), err)
	//		assert.Equal(GinkgoT(), expected, ec)
	//
	//	})
	//})
	//

})
