/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package evalctx

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"iam/pkg/abac/pdp/condition"
	pdptypes "iam/pkg/abac/pdp/types"
	"iam/pkg/abac/types"
	"iam/pkg/abac/types/request"
)

/*
PDP模块表达式求值
*/

// EvalContext 表达式求值上下文
// 只有一个Resource的信息
type EvalContext struct {
	*request.Request
	objSet pdptypes.ObjectSetInterface
}

// NewEvalContext new context
func NewEvalContext(req *request.Request) *EvalContext {
	objSet := pdptypes.NewObjectSet()

	for _, r := range req.Resources {
		// maybe nil here
		if r.Attribute == nil {
			r.Attribute = types.Attribute{}
		}
		// set id into attributes
		r.Attribute.Set("id", r.ID)

		/*
			兼容用户鉴权传入的 iam path 为以下2中情况, 统一处理成2段式
			1. 旧的格式 2段式: /resource_type_id,resource_id/
			2. RBAC的格式, 3段式: /system_id,resource_type_id,resource_id/
		*/
		iamPaths, ok := r.Attribute.Get(types.IamPath)
		if ok {
			iamPaths = standardizeIamPaths(iamPaths)
			r.Attribute.Set(types.IamPath, iamPaths)
		}

		// bk_job.script => attributes
		_type := r.System + "." + r.Type
		objSet.Set(_type, r.Attribute)
	}
	// TODO: 需要限制接入系统资源id字段不能配置为attribute; 因为会被覆盖
	return &EvalContext{
		Request: req,
		objSet:  objSet,
	}
}

func standardizeIamPaths(iamPaths interface{}) interface{} {
	paths := make([]string, 0, 2)
	switch vs := iamPaths.(type) {
	case []interface{}:
		for _, v := range vs {
			if s, ok := v.(string); ok {
				paths = append(paths, s)
			} else {
				return iamPaths
			}
		}
	case string:
		paths = append(paths, vs)
	default:
		return iamPaths
	}

	standardizedPaths := make([]interface{}, 0, len(paths))
	for _, path := range paths {
		if !strings.HasPrefix(path, "/") {
			standardizedPaths = append(standardizedPaths, path)
			continue
		}

		nodes := strings.Split(strings.Trim(path, "/"), "/")

		isValid := true
		var sb strings.Builder
		sb.WriteString("/")
		for _, node := range nodes {
			parts := strings.Split(node, ",")
			// 只要有一个节点不满足3段式格式, 则不做裁剪
			if len(parts) != 3 {
				isValid = false
				break
			}

			// 从3段式格式裁剪为2段式
			sb.WriteString(parts[1])
			sb.WriteString(",")
			sb.WriteString(parts[2])
			sb.WriteString("/")
		}

		// 不满足需要裁剪条件的路径, 不做处理, 保持为原始的格式
		if !isValid {
			standardizedPaths = append(standardizedPaths, path)
			continue
		}

		s := sb.String()
		// NOTE: 如果原始的字符串最后没有/, 则不加/
		if !strings.HasSuffix(path, "/") && len(s) > 1 {
			s = s[:len(s)-1]
		}

		standardizedPaths = append(standardizedPaths, s)
	}

	return standardizedPaths
}

// GetAttr 获取资源的属性值
func (c *EvalContext) GetAttr(name string) (interface{}, error) {
	// name should be {system}.{resource_type}.{attr_key}
	return c.objSet.GetAttribute(name), nil
}

func (c *EvalContext) HasResource(_type string) bool {
	// has {system}.{resource_type}
	return c.objSet.Has(_type)
}

func (c *EvalContext) SetEnv(envs map[string]interface{}) {
	c.objSet.Set(c.System+types.IamEnvSuffix, envs)
}

func (c *EvalContext) UnsetEnv() {
	c.objSet.Del(c.System + types.IamEnvSuffix)
}

func (c *EvalContext) HasEnv() bool {
	return c.objSet.Has(c.System + types.IamEnvSuffix)
}

func (c *EvalContext) InitEnvironments(cond condition.Condition, currentTime time.Time) error {
	// build envs
	c.UnsetEnv()

	hasEnvFunc := func(key string) bool {
		return strings.Contains(key, types.IamEnvSuffix)
	}
	// 限制: 一条策略, 有多个环境属性, 但是只能有一个tz(逻辑计算无法在同一个表达式中支持多个时区)
	hasEnvTzFunc := func(key string) bool {
		return strings.HasSuffix(key, types.IamEnvTzSuffix)
	}

	if cond.HasKey(hasEnvFunc) {
		// NOTE: 开启环境属性, 不一定会有tz, 而是 有配置时间相关环境属性, 一定会配置tz
		if tzValues, exists := cond.GetFirstMatchKeyValues(hasEnvTzFunc); exists {
			if len(tzValues) != 1 {
				return fmt.Errorf("pdp ctx initEnvironments got not tz in condition")
			}

			tz, ok := tzValues[0].(string)
			if !ok {
				return fmt.Errorf("pdp ctx initEnvironments got tz not a string")
			}

			envs, err := GenTimeEnvsFromCache(tz, currentTime)
			if err != nil {
				return fmt.Errorf("pdp gen envs fail, %w", err)
			}
			c.SetEnv(envs)
		}

		// NOTE: if got more envs, build it here before set

		// e.g.
		/*
				{
					// basic, should all have tz field
					"tz": "Asia/Shanghai",
					// now
					"hms": 172910,
					// later:
					"ts": 1638523704,
					"weekday": 3,
					"monthday": 29,
			        "month": 12,
				}
		*/
	}
	return nil
}

// GenTimeEnvsFromCache will return the same time-related envs if the tz and timestamp are same!
// NOTE: cache only if the envs is same for every request
//       if you will change the envs later(e.g. set some value from request, do not cache it!)
//       at that time, you should remove this func, use a new collection like sync.Pool
func GenTimeEnvsFromCache(tz string, currentTime time.Time) (map[string]interface{}, error) {
	key := tz + strconv.FormatInt(currentTime.Unix(), 10)

	cachedEnvs, ok := localTimeEnvsCache.Get(key)
	// hit
	if ok {
		return cachedEnvs.(map[string]interface{}), nil
	}
	// miss
	envs, err := genTimeEnvs(tz, currentTime)
	if err != nil {
		return nil, err
	}

	localTimeEnvsCache.SetDefault(key, envs)
	return envs, nil
}

func genTimeEnvs(tz string, currentTime time.Time) (map[string]interface{}, error) {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return nil, fmt.Errorf("pdp load policy timezone location fail, tz=%s,  %w", tz, err)
	}

	t := currentTime.In(loc)

	// hms means hour-minute-second, transfer 08:30:20 to 83020; 10:41:21 to 104121
	hms := int64(10000*t.Hour() + 100*t.Minute() + t.Second())

	envs := map[string]interface{}{
		"tz":  tz,
		"hms": hms,
		// "ts":  t.Unix(),
	}
	return envs, nil
}
