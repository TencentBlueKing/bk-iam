/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package dao

//func Test_resourceTypeManager_List(t *testing.T) {
//	database.RunWithMock(t, func(db *sqlx.DB, mock sqlmock.Sqlmock, t *testing.T) {
//		mockData := []interface{}{
//			ResourceType{
//				PK:     1,
//				System: "cmdb",
//				ID:     "host",
//			},
//			ResourceType{
//				PK:     1,
//				System: "cmdb",
//				ID:     "instance",
//			},
//		}
//
//		mockQuery := `^SELECT pk, system_id, id FROM resource_type WHERE pk IN`
//		mockRows := database.NewMockRows(mock, mockData...)
//		mock.ExpectQuery(mockQuery).WithArgs(1, 2, 3).WillReturnRows(mockRows)
//
//		manager := &resourceTypeManager{DB: db}
//		resourceTypes, err := manager.List([]int64{1, 2, 3})
//
//		assert.NoError(t, err)
//		assert.Equal(t, len(resourceTypes), 2)
//		assert.Equal(t, resourceTypes[0].ID, "host")
//	})
//}
