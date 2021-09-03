package condition

import (
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("StringPrefix", func() {

	var c *StringPrefixCondition
	BeforeEach(func() {
		c = &StringPrefixCondition{
			baseCondition{
				Key:   "ok",
				Value: []interface{}{"/biz,1/", "/biz,2/"},
			},
		}
	})

	It("new", func() {
		condition, err := newStringPrefixCondition("ok", []interface{}{"a", "b"})
		assert.NoError(GinkgoT(), err)
		assert.NotNil(GinkgoT(), condition)
	})

	It("GetName", func() {
		assert.Equal(GinkgoT(), "StringPrefix", c.GetName())
	})
	Context("Eval", func() {
		It("true", func() {
			assert.True(GinkgoT(), c.Eval(strCtx("/biz,1/set,2/")))
			assert.True(GinkgoT(), c.Eval(strCtx("/biz,2/set,3/")))
		})

		It("false", func() {
			assert.False(GinkgoT(), c.Eval(strCtx("c")))
		})

		It("attr list", func() {
			assert.True(GinkgoT(), c.Eval(listCtx{"/biz,1/set,2/", "d"}))

			assert.False(GinkgoT(), c.Eval(listCtx{"e", "f"}))
		})

		It("false, attr value not string", func() {
			assert.False(GinkgoT(), c.Eval(listCtx{1}))
		})

		It("false, expr value not string", func() {
			c = &StringPrefixCondition{
				baseCondition{
					Key:   "ok",
					Value: []interface{}{1},
				},
			}
			assert.False(GinkgoT(), c.Eval(strCtx("/biz,1/set,2/")))
		})

		It("_bk_iam_path_", func() {
			c = &StringPrefixCondition{
				baseCondition{
					Key:   iamPath,
					Value: []interface{}{"/biz,1/set,*/"},
				},
			}

			assert.True(GinkgoT(), c.Eval(strCtx("/biz,1/set,2/")))
			assert.False(GinkgoT(), c.Eval(strCtx("/biz,1/module,2/")))

		})
	})

	Describe("Translate", func() {
		It("fail, empty value", func() {
			c, err := newStringPrefixCondition("key", []interface{}{})
			assert.NoError(GinkgoT(), err)
			_, err = c.Translate()
			assert.Error(GinkgoT(), err)
			assert.Equal(GinkgoT(), errMustNotEmpty, err)
		})

		It("ok, single", func() {
			expected := map[string]interface{}{
				"op":    "starts_with",
				"field": "key",
				"value": "/biz,1/set,1/",
			}
			c, err := newStringPrefixCondition("key", []interface{}{"/biz,1/set,1/"})
			assert.NoError(GinkgoT(), err)
			ec, err := c.Translate()
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), expected, ec)
		})

		It("ok, multiple or", func() {
			expected := map[string]interface{}{
				"op": "OR",
				"content": []map[string]interface{}{
					{
						"op":    "starts_with",
						"field": "key",
						"value": "/biz,1/set,1/",
					},
					{
						"op":    "starts_with",
						"field": "key",
						"value": "/biz,2/set,2/",
					},
				},
			}

			c, err := newStringPrefixCondition("key", []interface{}{"/biz,1/set,1/", "/biz,2/set,2/"})
			assert.NoError(GinkgoT(), err)
			ec, err := c.Translate()
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), expected, ec)
		})
	})

})
