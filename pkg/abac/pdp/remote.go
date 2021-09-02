/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package pdp

import (
	"iam/pkg/abac/pdp/loader"
	"iam/pkg/abac/pip"
	"iam/pkg/abac/types"
	"iam/pkg/abac/types/request"
	"iam/pkg/errorx"
)

func fillRemoteResourceAttrs(r *request.Request, policies []types.AuthPolicy) (err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PDP, "fillRemoteResourceAttrs")
	var attrs map[string]interface{}

	resources := r.GetRemoteResources()
	for _, resource := range resources {
		attrs, err = queryRemoteResourceAttrs(resource, policies)
		if err != nil {
			err = errorWrapf(err, "queryRemoteResourceAttrs resource=`%+v` fail", resource)
			return err
		}
		resource.Attribute = attrs
	}
	return nil
}

func queryRemoteResourceAttrs(
	resource *types.Resource,
	policies []types.AuthPolicy,
) (attrs map[string]interface{}, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PDPHelper, "queryRemoteResourceAttrs")

	var keys []string
	keys, err = loader.GetPoliciesAttrKeys(resource, policies)
	if err != nil {
		err = errorWrapf(err,
			"condition.GetPoliciesAttrKeys policies=`%+v`, resource=`%+v` fail",
			policies, resource)
		return
	}

	// 6. PIP查询依赖resource相关keys的属性
	attrs, err = pip.QueryRemoteResourceAttribute(resource.System, resource.Type, resource.ID, keys)
	if err != nil {
		err = errorWrapf(err,
			"pip.QueryRemoteResourceAttribute system=`%s`, resourceType=`%s`, resourceID=`%s`, keys=`%+v` fail",
			resource.System, resource.Type, resource.ID, keys)
		return
	}
	return
}

func queryExtResourceAttrs(
	resource *types.ExtResource,
	policies []types.AuthPolicy,
) (resources []map[string]interface{}, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PDPHelper, "queryExtResourceAttrs")

	keys, err := loader.GetPoliciesAttrKeys(&types.Resource{
		System: resource.System,
		Type:   resource.Type,
		ID:     resource.IDs[0],
	}, policies)
	if err != nil {
		err = errorWrapf(err,
			"condition.GetPoliciesAttrKeys policies=`%+v`, resource=`%+v` fail",
			policies, resource)
		return
	}

	// 6. PIP查询依赖resource相关keys的属性
	resources, err = pip.BatchQueryRemoteResourcesAttribute(resource.System, resource.Type, resource.IDs, keys)
	if err != nil {
		err = errorWrapf(err,
			"pip.BatchQueryRemoteResourcesAttribute system=`%s`, resourceType=`%s`, resourceIDs length=`%d`, keys=`%+v` fail",
			resource.System, resource.Type, len(resource.IDs), keys)
		return
	}
	return
}
