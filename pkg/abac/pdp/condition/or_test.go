package condition

import (
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("Or", func() {
	wantOrCondition := &OrCondition{
		content: []Condition{
			&StringEqualsCondition{
				baseCondition: baseCondition{
					Key:   "system",
					Value: []interface{}{"linux"},
				},
			},
			&StringPrefixCondition{
				baseCondition: baseCondition{
					Key:   "path",
					Value: []interface{}{"/biz,1/"},
				},
			},
		},
	}

	var c *OrCondition
	BeforeEach(func() {
		c1, _ := newStringEqualsCondition("k1", []interface{}{"a", "b"})
		c2, _ := newNumericEqualsCondition("k2", []interface{}{123})
		c = &OrCondition{
			[]Condition{
				c1,
				c2,
			},
		}
	})

	Describe("New", func() {
		It("wrong key", func() {
			_, err := newOrCondition("wrong", []interface{}{"abc"})
			assert.Error(GinkgoT(), err)

		})

		It("ok", func() {
			data := []interface{}{
				map[string]interface{}{"StringEquals": map[string]interface{}{"system": []interface{}{"linux"}}},
				map[string]interface{}{"StringPrefix": map[string]interface{}{"path": []interface{}{"/biz,1/"}}},
			}
			c, err := newOrCondition("content", data)
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), wantOrCondition, c)
		})

		It("fail", func() {
			data := []interface{}{
				map[string]interface{}{"StringEquals": map[string]interface{}{"system": []interface{}{"linux"}}},
				1,
			}
			_, err := newOrCondition("content", data)
			assert.Error(GinkgoT(), err)
		})
	})

	It("GetName", func() {
		assert.Equal(GinkgoT(), "OR", c.GetName())
	})

	It("Eval", func() {
		assert.True(GinkgoT(), c.Eval(strCtx("a")))
		assert.True(GinkgoT(), c.Eval(strCtx("b")))
		assert.True(GinkgoT(), c.Eval(ctx(123)))

		assert.False(GinkgoT(), c.Eval(strCtx("c")))
		assert.False(GinkgoT(), c.Eval(ctx(456)))
	})

	It("GetKeys", func() {
		oc := OrCondition{
			content: []Condition{
				&NumericEqualsCondition{
					baseCondition{
						Key: "hello",
					},
				},
			},
		}

		keys := oc.GetKeys()
		assert.Len(GinkgoT(), keys, 1)
		assert.Equal(GinkgoT(), "hello", keys[0])
	})

	//Describe("orTranslate", func() {
	//	It("ok, empty", func() {
	//		want := ExprCell{
	//			"op":      "OR",
	//			"content": []interface{}{},
	//		}
	//		ec, err := orTranslate("host", []interface{}{})
	//
	//		assert.NoError(GinkgoT(), err)
	//		assert.Equal(GinkgoT(), want, ec)
	//	})
	//
	//	It("fail, wrong value", func() {
	//		_, err := orTranslate("host", []interface{}{123})
	//		assert.Error(GinkgoT(), err)
	//	})
	//
	//	It("fail, singleTranslate error", func() {
	//		_, err := orTranslate("host", []interface{}{
	//			map[string]interface{}{
	//				"NoSupportOP": "",
	//			},
	//		})
	//		assert.Error(GinkgoT(), err)
	//	})
	//
	//	It("ok", func() {
	//		want := ExprCell{
	//			"op": "OR",
	//			"content": []interface{}{
	//				ExprCell(map[string]interface{}{
	//					"op":    "eq",
	//					"field": "host.number",
	//					"value": 1,
	//				}),
	//				ExprCell(map[string]interface{}{
	//					"op":    "in",
	//					"field": "host.os",
	//					"value": []interface{}{"linux", "windows"},
	//				}),
	//			},
	//		}
	//		value := []interface{}{
	//			map[string]interface{}{
	//				"NumericEquals": map[string]interface{}{
	//					"number": []interface{}{1},
	//				},
	//			},
	//			map[string]interface{}{
	//				"StringEquals": map[string]interface{}{
	//					"os": []interface{}{"linux", "windows"},
	//				},
	//			},
	//		}
	//		ec, err := orTranslate("host", value)
	//		assert.NoError(GinkgoT(), err)
	//		assert.Equal(GinkgoT(), want, ec)
	//	})
	//
	//	It("fail, singleTranslate fail", func() {
	//		value := []interface{}{
	//			map[string]interface{}{
	//				"NumericEquals": map[string]interface{}{
	//					"number": []interface{}{1},
	//				},
	//			},
	//			map[string]interface{}{
	//				"NotExists": map[string]interface{}{
	//					"os": []interface{}{"linux", "windows"},
	//				},
	//			},
	//		}
	//		_, err := orTranslate("host", value)
	//		assert.Error(GinkgoT(), err)
	//	})
	//
	//})
	//
})
