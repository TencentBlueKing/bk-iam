package condition

import (
	"strings"

	"iam/pkg/abac/pdp/types"
)

// StringPrefixCondition 字符串前缀匹配
type StringPrefixCondition struct {
	baseCondition
}

func newStringPrefixCondition(key string, values []interface{}) (Condition, error) {
	return &StringPrefixCondition{
		baseCondition: baseCondition{
			Key:   key,
			Value: values,
		},
	}, nil
}

// GetName 名称
func (c *StringPrefixCondition) GetName() string {
	return "StringPrefix"
}

// Eval 求值
func (c *StringPrefixCondition) Eval(ctx types.AttributeGetter) bool {
	return c.forOr(ctx, func(a, b interface{}) bool {
		aStr, ok := a.(string)
		if !ok {
			return false
		}

		bStr, ok := b.(string)
		if !ok {
			return false
		}

		// 支持表达式中最后一个节点为任意
		// /biz,1/set,*/ -> /biz,1/set,
		if c.Key == iamPath && strings.HasSuffix(bStr, ",*/") {
			bStr = bStr[0 : len(bStr)-2]
		}

		return strings.HasPrefix(aStr, bStr)
	})
}
func (c *StringPrefixCondition) Translate() (map[string]interface{}, error) {
	content := make([]map[string]interface{}, 0, len(c.Value))
	for _, v := range c.Value {
		content = append(content, map[string]interface{}{
			"op":    "starts_with",
			"field": c.Key,
			"value": v,
		})
	}

	switch len(content) {
	case 0:
		return nil, errMustNotEmpty
	case 1:
		return content[0], nil
	default:
		return map[string]interface{}{
			"op":      "OR",
			"content": content,
		}, nil
	}

}
