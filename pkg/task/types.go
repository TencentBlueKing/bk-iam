/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package task

import (
	"context"
	"strconv"

	jsoniter "github.com/json-iterator/go"

	"iam/pkg/service/types"
)

// GroupAlterEventProducer ...
type GroupAlterEventProducer interface {
	Publish(types.GroupAlterEvent) error
}

// GroupAlterEventConsumer
type GroupAlterEventConsumer interface {
	Run(ctx context.Context)
	Handle(GroupAlterMessage) error
}

// GroupAlterMessage ...
type GroupAlterMessage struct {
	GroupPK   int64 `json:"group_pk"`
	ActionPK  int64 `json:"action_pk"`
	SubjectPK int64 `json:"subject_pk"`

	// TODO 增加 op/ts 字段
}

// UniqueID ...
func (m *GroupAlterMessage) UniqueID() string {
	return strconv.FormatInt(
		m.GroupPK,
		10,
	) + ":" + strconv.FormatInt(
		m.ActionPK,
		10,
	) + ":" + strconv.FormatInt(
		m.SubjectPK,
		10,
	)
}

// String ...
func (m *GroupAlterMessage) String() (string, error) {
	return jsoniter.MarshalToString(m)
}

// NewGroupAlterMessageFromString ...
func NewGroupAlterMessageFromString(s string) (m GroupAlterMessage, err error) {
	if err := jsoniter.UnmarshalFromString(s, &m); err != nil {
		return m, err
	}
	return m, nil
}
