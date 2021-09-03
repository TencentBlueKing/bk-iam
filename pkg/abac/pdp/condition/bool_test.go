package condition

import (
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("Bool", func() {
	var c *BoolCondition
	BeforeEach(func() {
		c = &BoolCondition{
			baseCondition{
				Key:   "ok",
				Value: []interface{}{true},
			},
		}
	})

	It("new", func() {
		condition, err := newBoolCondition("ok", []interface{}{true})
		assert.NoError(GinkgoT(), err)
		assert.NotNil(GinkgoT(), condition)
	})

	It("GetName", func() {
		assert.Equal(GinkgoT(), "Bool", c.GetName())
	})

	Context("Eval", func() {
		It("errCtx", func() {
			allowed := c.Eval(errCtx(1))
			assert.False(GinkgoT(), allowed)
		})

		It("true", func() {
			assert.True(GinkgoT(), c.Eval(boolCtx(true)))
		})

		It("false, multi attr values", func() {
			assert.False(GinkgoT(), c.Eval(listCtx{1, 2}))
		})

		It("false, attr value not bool", func() {
			assert.False(GinkgoT(), c.Eval(ctx(1)))
		})

		It("false, multi expr values", func() {
			c = &BoolCondition{
				baseCondition{
					Key:   "ok",
					Value: []interface{}{true, true},
				},
			}
			assert.False(GinkgoT(), c.Eval(boolCtx(true)))
		})

		It("fail, exprValue not bool", func() {
			c = &BoolCondition{
				baseCondition{
					Key:   "ok",
					Value: []interface{}{123},
				},
			}
			assert.False(GinkgoT(), c.Eval(boolCtx(true)))
		})

	})

	Describe("Translate", func() {
		It("not support multi value", func() {
			c1, err := newBoolCondition("key", []interface{}{true, false})
			assert.NoError(GinkgoT(), err)

			_, err = c1.Translate()
			assert.Contains(GinkgoT(), err.Error(), "bool not support multi value")
		})

		It("ok", func() {
			expected := map[string]interface{}{
				"op":    "eq",
				"field": "ok",
				"value": true,
			}
			ec, err := c.Translate()
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), expected, ec)
		})
	})

})
