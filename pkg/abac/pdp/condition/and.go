package condition

import (
	"fmt"
	"strings"

	"iam/pkg/abac/pdp/types"
)

// AndCondition 逻辑AND
type AndCondition struct {
	content []Condition
}

func NewAndCondition(content []Condition) Condition {
	return &AndCondition{content: content}
}

func newAndCondition(key string, values []interface{}) (Condition, error) {
	if key != "content" {
		return nil, fmt.Errorf("and condition not support key %s", key)
	}

	conditions := make([]Condition, 0, len(values))
	var (
		condition Condition
		err       error
	)
	for _, v := range values {
		condition, err = newConditionFromInterface(v)
		if err != nil {
			return nil, fmt.Errorf("and condition parser error: %w", err)
		}

		conditions = append(conditions, condition)
	}

	return &AndCondition{content: conditions}, nil
}

// GetName 名称
func (c *AndCondition) GetName() string {
	return "AND"
}

// Eval 求值
func (c *AndCondition) Eval(ctx types.AttributeGetter) bool {
	for _, condition := range c.content {
		if !condition.Eval(ctx) {
			return false
		}
	}

	return true
}

func (c *AndCondition) PartialEval(ctx types.AttributeGetter) (bool, Condition) {
	// TODO: get from sync.Pool
	remainContent := make([]Condition, 0, len(c.content))
	for _, condition := range c.content {
		if condition.GetName() == "AND" || condition.GetName() == "OR" {
			ok, ci := condition.(LogicalCondition).PartialEval(ctx)
			if !ok {
				return false, nil
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

			if ctx.HasKey(_type) {
				// resource exists and eval fail, no remain content
				if !condition.Eval(ctx) {
					return false, nil
				}
			} else {
				remainContent = append(remainContent, condition)
			}
		}
	}

	return true, NewAndCondition(remainContent)
}

func (c *AndCondition) Translate() (map[string]interface{}, error) {
	content := make([]interface{}, 0, len(c.content))
	for _, c := range c.content {
		ct, err := c.Translate()
		if err != nil {
			return nil, err
		}
		content = append(content, ct)
	}

	return map[string]interface{}{
		"op":      "AND",
		"content": content,
	}, nil

}

// GetKeys 返回嵌套条件中所有包含的属性key
func (c *AndCondition) GetKeys() []string {
	keys := make([]string, 0, len(c.content))
	for _, condition := range c.content {
		keys = append(keys, condition.GetKeys()...)
	}
	return keys
}
