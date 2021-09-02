package types

import (
	"strings"

	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("Object", func() {
	It("NewObjectSet", func() {
		os := NewObjectSet()
		assert.NotNil(GinkgoT(), os)
	})

	Describe("ObjectSet eval", func() {

		var o ObjectSetInterface
		BeforeEach(func() {
			o = NewObjectSet()
		})

		It("Set", func() {
			o.Set("book", map[string]interface{}{"price": 10})
			assert.Equal(GinkgoT(), o.Size(), 1)
		})

		It("Get", func() {
			_, exists := o.Get("book")
			assert.False(GinkgoT(), exists)

			o.Set("book", map[string]interface{}{"price": 10})
			attrs, exists := o.Get("book")
			assert.True(GinkgoT(), exists)
			assert.Len(GinkgoT(), attrs, 1)
		})

		It("Has", func() {
			has := o.Has("book")
			assert.False(GinkgoT(), has)

			o.Set("book", map[string]interface{}{"price": 10})

			has = o.Has("book")
			assert.True(GinkgoT(), has)
		})

		It("Del", func() {
			o.Del("book")

			o.Set("book", map[string]interface{}{"price": 10})
			has := o.Has("book")
			assert.True(GinkgoT(), has)

			o.Del("book")
			has = o.Has("book")
			assert.False(GinkgoT(), has)
		})

		It("Size", func() {
			assert.Equal(GinkgoT(), o.Size(), 0)

			o.Set("book", map[string]interface{}{"price": 10})
			assert.Equal(GinkgoT(), o.Size(), 1)
		})

		Describe("GetAttribute", func() {
			BeforeEach(func() {
				o.Set("book", map[string]interface{}{"price": 10})
			})

			It("invalid key", func() {
				value := o.GetAttribute("aaa")
				assert.Nil(GinkgoT(), value)
			})

			It("missing type", func() {
				value := o.GetAttribute("pen.price")
				assert.Nil(GinkgoT(), value)
			})

			It("missing attribute", func() {
				value := o.GetAttribute("book.size")
				assert.Nil(GinkgoT(), value)
			})

			It("hit", func() {
				value := o.GetAttribute("book.price")
				assert.NotNil(GinkgoT(), value)
			})
		})

		Describe("GetAttribute strings.Split and strings.IndexByte", func() {
			It("should be equal", func() {
				s := "biz.id"

				parts := strings.Split(s, ".")

				idx := strings.IndexByte(s, '.')
				assert.Equal(GinkgoT(), parts[0], s[:idx])
				assert.Equal(GinkgoT(), parts[1], s[idx+1:])

			})

		})

	})

})
