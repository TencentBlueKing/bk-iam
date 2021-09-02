package condition

import "iam/pkg/abac/pdp/types"

// NumericEqualsCondition Number相等
type NumericEqualsCondition struct {
	baseCondition
}

//nolint:unparam
func newNumericEqualsCondition(key string, values []interface{}) (Condition, error) {
	return &NumericEqualsCondition{
		baseCondition: baseCondition{
			Key:   key,
			Value: values,
		},
	}, nil
}

// GetName 名称
func (c *NumericEqualsCondition) GetName() string {
	return "NumericEquals"
}

// Eval 求值
func (c *NumericEqualsCondition) Eval(ctx types.AttributeGetter) bool {
	return c.forOr(ctx, func(a, b interface{}) bool {
		return a == b
	})
}

func (c *NumericEqualsCondition) Translate() (map[string]interface{}, error) {
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
