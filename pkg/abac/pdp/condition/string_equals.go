package condition

import "iam/pkg/abac/pdp/types"

// StringEqualsCondition 字符串相等
type StringEqualsCondition struct {
	baseCondition
}

//nolint:unparam
func newStringEqualsCondition(key string, values []interface{}) (Condition, error) {
	return &StringEqualsCondition{
		baseCondition: baseCondition{
			Key:   key,
			Value: values,
		},
	}, nil
}

// GetName 名称
func (c *StringEqualsCondition) GetName() string {
	return "StringEquals"
}

// Eval 求值
func (c *StringEqualsCondition) Eval(ctx types.AttributeGetter) bool {
	return c.forOr(ctx, func(a, b interface{}) bool {
		return a == b
	})
}
func (c *StringEqualsCondition) Translate() (map[string]interface{}, error) {
	exprCell := map[string]interface{}{
		"field": c.Key,
	}

	switch len(c.Value) {
	case 0:
		return nil, errMustNotEmpty
	case 1:
		exprCell["op"] = "eq"
		exprCell["value"] = c.Value[0]
	default:
		exprCell["op"] = "in"
		exprCell["value"] = c.Value
	}
	return exprCell, nil

}
