package condition

import (
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("And", func() {
	wantAndCondition := &AndCondition{
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

	var c *AndCondition
	BeforeEach(func() {
		c1, _ := newStringEqualsCondition("k1", []interface{}{"a", "b"})
		c2, _ := newNumericEqualsCondition("k1", []interface{}{"b", "c"})
		c = &AndCondition{
			[]Condition{
				c1,
				c2,
			},
		}
	})

	Describe("New", func() {
		It("New ok", func() {
			c := NewAndCondition([]Condition{})
			assert.NotNil(GinkgoT(), c)
			assert.Equal(GinkgoT(), "AND", c.GetName())
		})

		It("wrong key", func() {
			_, err := newAndCondition("wrong", []interface{}{"abc"})
			assert.Error(GinkgoT(), err)

		})

		It("ok", func() {
			data := []interface{}{
				map[string]interface{}{"StringEquals": map[string]interface{}{"system": []interface{}{"linux"}}},
				map[string]interface{}{"StringPrefix": map[string]interface{}{"path": []interface{}{"/biz,1/"}}},
			}
			c, err := newAndCondition("content", data)
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), wantAndCondition, c)
		})

		It("fail", func() {
			data := []interface{}{
				map[string]interface{}{"StringEquals": map[string]interface{}{"system": []interface{}{"linux"}}},
				1,
			}
			_, err := newAndCondition("content", data)
			assert.Error(GinkgoT(), err)
		})
	})

	It("GetName", func() {
		assert.Equal(GinkgoT(), "AND", c.GetName())
	})

	It("Eval", func() {
		assert.True(GinkgoT(), c.Eval(strCtx("b")))

		assert.False(GinkgoT(), c.Eval(strCtx("a")))
		assert.False(GinkgoT(), c.Eval(strCtx("c")))

		assert.False(GinkgoT(), c.Eval(strCtx("d")))
	})

	It("GetKeys", func() {
		oc := AndCondition{
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

	Describe("andTranslate", func() {
		It("ok, empty", func() {
			want := map[string]interface{}{
				"op":      "AND",
				"content": []map[string]interface{}{},
			}
			c, err := newAndCondition("content", []interface{}{})
			assert.NoError(GinkgoT(), err)
			ec, err := c.Translate()
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), want, ec)
		})

		It("fail, wrong value", func() {
			_, err := newAndCondition("content", []interface{}{123})
			assert.Error(GinkgoT(), err)
		})

		It("fail, new error", func() {
			_, err := newAndCondition("content", []interface{}{
				map[string]interface{}{
					"NoSupportOP": "",
				},
			})
			assert.Error(GinkgoT(), err)
		})

		It("fail, translate error", func() {
			bc, err := newBoolCondition("test", []interface{}{true, false})
			assert.NoError(GinkgoT(), err)
			c := NewAndCondition([]Condition{
				bc,
			})
			_, err = c.Translate()
			assert.Error(GinkgoT(), err)
		})

		It("ok", func() {
			want := map[string]interface{}{
				"op": "AND",
				"content": []map[string]interface{}{
					{
						"op":    "eq",
						"field": "number",
						"value": 1,
					},
					{
						"op":    "in",
						"field": "os",
						"value": []interface{}{"linux", "windows"},
					},
				},
			}
			value := []interface{}{
				map[string]interface{}{
					"NumericEquals": map[string]interface{}{
						"number": []interface{}{1},
					},
				},
				map[string]interface{}{
					"StringEquals": map[string]interface{}{
						"os": []interface{}{"linux", "windows"},
					},
				},
			}
			c, err := newAndCondition("content", value)
			assert.NoError(GinkgoT(), err)
			ec, err := c.Translate()
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), want, ec)
		})

		It("fail, singleTranslate fail", func() {
			value := []interface{}{
				map[string]interface{}{
					"NumericEquals": map[string]interface{}{
						"number": []interface{}{1},
					},
				},
				map[string]interface{}{
					"NotExists": map[string]interface{}{
						"os": []interface{}{"linux", "windows"},
					},
				},
			}
			_, err := newAndCondition("content", value)
			assert.Error(GinkgoT(), err)
		})

	})

	// TODO: PartialEval

})
