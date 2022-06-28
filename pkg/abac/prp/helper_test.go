/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package prp

import (
	"errors"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"iam/pkg/abac/types"
	"iam/pkg/cacheimpls"
	svctypes "iam/pkg/service/types"
)

var _ = Describe("Helper", func() {
	Describe("getEffectSubjectPKs", func() {
		var ctl *gomock.Controller
		var patches *gomonkey.Patches
		var s types.Subject
		BeforeEach(func() {
			ctl = gomock.NewController(GinkgoT())

			s = types.NewSubject()
		})
		AfterEach(func() {
			ctl.Finish()
			if patches != nil {
				patches.Reset()
			}
		})

		It("subject GetPK fail", func() {
			_, err := GetEffectGroupPKs("test", s)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "subject.Attribute.GetPK")
		})

		It("subject GetDepartmentPKs fail", func() {
			s.FillAttributes(123, []int64{1, 2, 3})
			s.Attribute.Delete(types.DeptAttrName)

			_, err := GetEffectGroupPKs("test", s)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "subject.GetDepartmentPKs")
		})

		It("cacheimpls.ListSubjectEffectGroups fail", func() {
			patches = gomonkey.ApplyFunc(cacheimpls.ListSystemSubjectEffectGroups,
				func(systemID string, pks []int64) ([]svctypes.ThinSubjectGroup, error) {
					return nil, errors.New("list subject_group fail")
				})
			s.FillAttributes(123, []int64{1, 2, 3})
			_, err := GetEffectGroupPKs("test", s)
			assert.Error(GinkgoT(), err)
			assert.Contains(GinkgoT(), err.Error(), "ListSubjectEffectGroups")
		})

		It("ok", func() {
			patches = gomonkey.ApplyFunc(cacheimpls.ListSystemSubjectEffectGroups,
				func(systemID string, pks []int64) ([]svctypes.ThinSubjectGroup, error) {
					return []svctypes.ThinSubjectGroup{
						{
							GroupPK:         4,
							PolicyExpiredAt: 0,
						},
						{
							GroupPK:         5,
							PolicyExpiredAt: time.Now().Add(1 * time.Minute).Unix(),
						},
						{
							GroupPK:         6,
							PolicyExpiredAt: time.Now().Add(1 * time.Minute).Unix(),
						},
						{
							GroupPK:         6,
							PolicyExpiredAt: time.Now().Add(1 * time.Minute).Unix(),
						},
						{
							GroupPK:         7,
							PolicyExpiredAt: time.Now().Add(1 * time.Minute).Unix(),
						},
					}, nil
				})

			s.FillAttributes(123, []int64{1, 2, 3})
			pks, err := GetEffectGroupPKs("test", s)
			assert.NoError(GinkgoT(), err)

			// all = user(123) +  groups(5,6,7)
			assert.ElementsMatch(GinkgoT(), []int64{5, 6, 7}, pks)
		})
	})
})
