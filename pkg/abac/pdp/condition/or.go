package condition

import (
	"fmt"
	"strings"

	"iam/pkg/abac/pdp/types"
)

// OrCondition 逻辑OR
type OrCondition struct {
	content []Condition
}

func NewOrCondition(content []Condition) Condition {
	return &OrCondition{content: content}
}

func newOrCondition(key string, values []interface{}) (Condition, error) {
	if key != "content" {
		return nil, fmt.Errorf("or condition not support key %s", key)
	}

	conditions := make([]Condition, 0, len(values))
	var (
		condition Condition
		err       error
	)

	for _, v := range values {
		condition, err = newConditionFromInterface(v)
		if err != nil {
			return nil, fmt.Errorf("or condition parser error: %w", err)
		}

		conditions = append(conditions, condition)
	}

	return &OrCondition{content: conditions}, nil
}

// GetName 名称
func (c *OrCondition) GetName() string {
	return "OR"
}

// Eval 求值
func (c *OrCondition) Eval(ctx types.AttributeGetter) bool {
	for _, condition := range c.content {
		if condition.Eval(ctx) {
			return true
		}
	}
	return false
}

func (c *OrCondition) PartialEval(ctx types.AttributeGetter) (bool, Condition) {
	remainContent := make([]Condition, 0, len(c.content))
	for _, condition := range c.content {
		if condition.GetName() == "AND" || condition.GetName() == "OR" {
			ok, ci := condition.(LogicalCondition).PartialEval(ctx)
			if ok {
				// NOTE: here, return any condition?
				return true, NewAnyCondition()
			} else {
				remainContent = append(remainContent, ci)
			}
		} else {
			key := condition.GetKeys()[0]
			dotIdx := strings.LastIndexByte(key, '.')
			if dotIdx == -1 {
				panic("should contains dot in key")
			}
			_type := key[:dotIdx]

			if !ctx.HasKey(_type) {
				remainContent = append(remainContent, condition)
			}
		}
	}

	return true, NewOrCondition(remainContent)
}

func (c *OrCondition) Translate() (map[string]interface{}, error) {
	content := make([]interface{}, 0, len(c.content))
	for _, c := range c.content {
		ct, err := c.Translate()
		if err != nil {
			return nil, err
		}
		content = append(content, ct)
	}

	return map[string]interface{}{
		"op":      "OR",
		"content": content,
	}, nil

}

// GetKeys 返回嵌套条件中所有包含的属性key
func (c *OrCondition) GetKeys() []string {
	keys := make([]string, 0, len(c.content))
	for _, condition := range c.content {
		keys = append(keys, condition.GetKeys()...)
	}
	return keys
}
