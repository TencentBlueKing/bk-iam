package condition

import "iam/pkg/abac/pdp/types"

// AnyCondition 任意条件
type AnyCondition struct {
	baseCondition
}

//nolint:unparam
func newAnyCondition(key string, values []interface{}) (Condition, error) {
	return &AnyCondition{
		baseCondition: baseCondition{
			Key:   key,
			Value: values,
		},
	}, nil
}

func NewAnyCondition() Condition {
	return &AnyCondition{
		baseCondition: baseCondition{
			Key:   "Any",
			Value: []interface{}{},
		},
	}
}

// GetName 名称
func (c *AnyCondition) GetName() string {
	return "Any"
}

// Eval 求值
func (c *AnyCondition) Eval(ctx types.AttributeGetter) bool {
	return true
}

func (c *AnyCondition) Translate() (map[string]interface{}, error) {
	return map[string]interface{}{
		"op":    "any",
		"field": c.Key,
		"value": c.Value,
	}, nil
}

// GetKeys 属性key
func (c *AnyCondition) GetKeys() []string {
	return []string{}
}
