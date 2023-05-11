/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package service

import (
	"errors"
	"time"

	"github.com/TencentBlueKing/gopkg/collection/set"
	"github.com/TencentBlueKing/gopkg/errorx"
	"github.com/TencentBlueKing/gopkg/stringx"
	"github.com/jmoiron/sqlx"

	"iam/pkg/database"
	"iam/pkg/database/dao"
	"iam/pkg/service/types"
)

//go:generate mockgen -source=$GOFILE -destination=./mock/$GOFILE -package=mock

// PolicyVersion ...
const PolicyVersion = "1"

// PolicySVC ...
const PolicySVC = "PolicySVC"

const (
	expressionTypeCustom   int64 = 0 // 自定义的expression类型
	expressionTypeTemplate int64 = 1 // 模板的expression类型

	expressionTypeUnreferenced int64 = -1 // 未被引用的模板expression类型

	expressionPKActionWithoutResource = -1 // 操作不关联资源时的expression pk

	// PolicyTemplateIDCustom template id for custom policy
	PolicyTemplateIDCustom int64 = 0
)

var errPolicy = errors.New("policy data error")

// PolicyService ...
type PolicyService interface {
	// for auth

	ListAuthBySubjectAction(subjectPKs []int64, actionPK int64) ([]types.AuthPolicy, error)
	ListExpressionByPKs(pks []int64) ([]types.AuthExpression, error)

	// for saas

	ListThinBySubjectActionTemplate(subjectPK int64, actionPKs []int64, templateID int64) ([]types.ThinPolicy, error)
	ListThinBySubjectTemplateBeforeExpiredAt(subjectPK int64, templateID, expiredAt int64) ([]types.ThinPolicy, error)

	AlterCustomPolicies(subjectPK int64, createPolicies, updatePolicies []types.Policy, deletePolicyIDs []int64,
		actionPKWithResourceTypeSet *set.Int64Set) (map[int64][]int64, error)
	AlterCustomPoliciesWithTx(
		tx *sqlx.Tx,
		subjectPK int64,
		createPolicies, updatePolicies []types.Policy,
		deletePolicyIDs []int64,
		actionPKWithResourceTypeSet *set.Int64Set,
	) (map[int64][]int64, error)

	DeleteByPKs(subjectPK int64, pks []int64) error

	DeleteByActionPKWithTx(tx *sqlx.Tx, actionPK int64) error

	CreateAndDeleteTemplatePoliciesWithTx(
		tx *sqlx.Tx,
		subjectPK, templateID int64,
		createPolicies []types.Policy,
		deletePolicyIDs []int64,
		actionPKWithResourceTypeSet *set.Int64Set,
	) error
	UpdateTemplatePoliciesWithTx(
		tx *sqlx.Tx,
		subjectPK int64,
		policies []types.Policy,
		actionPKWithResourceTypeSet *set.Int64Set,
	) error

	// for pap
	BulkDeleteBySubjectPKsWithTx(tx *sqlx.Tx, pks []int64) error

	// for model update

	HasAnyByActionPK(actionPK int64) (bool, error)

	// for expression clean task
	DeleteUnreferencedExpressions() error
}

type policyService struct {
	manager          dao.PolicyManager
	expressionManger dao.ExpressionManager
}

// NewPolicyService ...
func NewPolicyService() PolicyService {
	return &policyService{
		manager:          dao.NewPolicyManager(),
		expressionManger: dao.NewExpressionManager(),
	}
}

// ListAuthBySubjectAction ...
func (s *policyService) ListAuthBySubjectAction(subjectPKs []int64, actionPK int64) ([]types.AuthPolicy, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PolicySVC, "ListAuthBySubjectAction")
	nowUnix := time.Now().Unix()
	daoPolicies, err := s.manager.ListAuthBySubjectAction(subjectPKs, actionPK, nowUnix)
	if err != nil {
		return nil, errorWrapf(
			err, "manager.ListAuthBySubjectAction subjectPKs=`%+v`, actionPK=`%d`, expiredAt=`%d`",
			subjectPKs, actionPK, nowUnix,
		)
	}

	policies := make([]types.AuthPolicy, 0, len(daoPolicies))
	for _, p := range daoPolicies {
		policies = append(policies, types.AuthPolicy{
			PK:           p.PK,
			SubjectPK:    p.SubjectPK,
			ExpressionPK: p.ExpressionPK,
			ExpiredAt:    p.ExpiredAt,
		})
	}
	return policies, nil
}

// ListExpressionByPKs ...
func (s *policyService) ListExpressionByPKs(pks []int64) ([]types.AuthExpression, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PolicySVC, "ListExpressionByPKs")
	daoExpressions, err := s.expressionManger.ListAuthByPKs(pks)
	if err != nil {
		return nil, errorWrapf(err, "expressionManger.ListAuthByPKs pks=`%+v`", pks)
	}

	expressions := make([]types.AuthExpression, 0, len(daoExpressions))
	for _, e := range daoExpressions {
		expressions = append(expressions, types.AuthExpression{
			PK:         e.PK,
			Expression: e.Expression,
			Signature:  e.Signature,
		})
	}
	return expressions, nil
}

func (s *policyService) convertToThinPolicies(daoPolicies []dao.Policy) []types.ThinPolicy {
	thinPolicies := make([]types.ThinPolicy, 0, len(daoPolicies))
	for _, p := range daoPolicies {
		thinPolicies = append(thinPolicies, types.ThinPolicy{
			Version:   PolicyVersion,
			ID:        p.PK,
			ActionPK:  p.ActionPK,
			ExpiredAt: p.ExpiredAt,
		})
	}
	return thinPolicies
}

// ListThinBySubjectActionTemplate ...
func (s *policyService) ListThinBySubjectActionTemplate(
	subjectPK int64,
	actionPKs []int64,
	templateID int64,
) ([]types.ThinPolicy, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PolicySVC, "ListThinBySubjectSystemTemplate")
	daoPolicies, err := s.manager.ListBySubjectActionTemplate(subjectPK, actionPKs, templateID)
	if err != nil {
		return nil, errorWrapf(
			err, "manager.ListBySubjectActionTemplate subjectPK=`%d`, actionPKs=`%+v`, templateID=`%d`",
			subjectPK, actionPKs, templateID)
	}

	return s.convertToThinPolicies(daoPolicies), nil
}

// ListThinBySubjectTemplateBeforeExpiredAt ...
func (s *policyService) ListThinBySubjectTemplateBeforeExpiredAt(
	subjectPK int64,
	templateID,
	expiredAt int64,
) ([]types.ThinPolicy, error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PolicySVC, "ListBySubjectTemplateBeforeExpiredAt")
	daoPolicies, err := s.manager.ListBySubjectTemplateBeforeExpiredAt(subjectPK, templateID, expiredAt)
	if err != nil {
		return nil, errorWrapf(
			err, "manager.ListBySubjectTemplateBeforeExpiredAt subjectPK=`%d`, templateID=`%d`, expiredAt=`%d`",
			subjectPK, templateID, expiredAt)
	}

	return s.convertToThinPolicies(daoPolicies), nil
}

// AlterCustomPolicies subject custom alter policies
func (s *policyService) AlterCustomPolicies(
	subjectPK int64,
	createPolicies, updatePolicies []types.Policy,
	deletePolicyIDs []int64,
	actionPKWithResourceTypeSet *set.Int64Set,
) (updatedActionPKExpressionPKs map[int64][]int64, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PolicySVC, "AlterCustomPolicies")

	// 使用事务
	tx, err := database.GenerateDefaultDBTx()
	defer database.RollBackWithLog(tx)
	if err != nil {
		err = errorWrapf(err, "define tx fail")
		return
	}

	updatedActionPKExpressionPKs, err = s.AlterCustomPoliciesWithTx(
		tx, subjectPK,
		createPolicies, updatePolicies, deletePolicyIDs,
		actionPKWithResourceTypeSet,
	)
	if err != nil {
		err = errorWrapf(err, "s.AlterCustomPoliciesWithTx subjectPK=`%d` fail", subjectPK)
		return
	}

	err = tx.Commit()
	return updatedActionPKExpressionPKs, err
}

func (s *policyService) AlterCustomPoliciesWithTx(
	tx *sqlx.Tx,
	subjectPK int64,
	createPolicies, updatePolicies []types.Policy,
	deletePolicyIDs []int64,
	actionPKWithResourceTypeSet *set.Int64Set,
) (updatedActionPKExpressionPKs map[int64][]int64, err error) {
	// 自定义权限每个policy对应一个expression
	// 创建policy的同时创建expression
	// 修改policy时直接修改关联的expression
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PolicySVC, "AlterCustomPoliciesWithTx")

	daoCreateExpressions := make([]dao.Expression, 0, len(createPolicies))
	daoCreatePolicies := make([]dao.Policy, 0, len(createPolicies))

	// 记录需要创建的policy与expression的索引关系
	type policyExpressionIndex struct {
		policyIndex     int
		expressionIndex int
	}
	policyExpressionIndexes := make([]policyExpressionIndex, 0, len(createPolicies))
	for _, p := range createPolicies {
		// 操作有关联资源类型
		if actionPKWithResourceTypeSet.Has(p.ActionPK) {
			daoCreateExpressions = append(daoCreateExpressions, dao.Expression{
				Type:       expressionTypeCustom,
				Expression: p.Expression,
				Signature:  stringx.MD5Hash(p.Expression), // 计算Hash
			})

			daoCreatePolicies = append(daoCreatePolicies, dao.Policy{
				SubjectPK: p.SubjectPK,
				ActionPK:  p.ActionPK,
				ExpiredAt: p.ExpiredAt,
			})

			policyExpressionIndexes = append(policyExpressionIndexes, policyExpressionIndex{
				policyIndex:     len(daoCreatePolicies) - 1,
				expressionIndex: len(daoCreateExpressions) - 1,
			})
		} else {
			// 无关联资源的自定义权限, expression 为 -1, 不创建expression对象
			daoCreatePolicies = append(daoCreatePolicies, dao.Policy{
				SubjectPK:    p.SubjectPK,
				ActionPK:     p.ActionPK,
				ExpressionPK: expressionPKActionWithoutResource,
				ExpiredAt:    p.ExpiredAt,
			})
		}
	}

	updatePolicyPKs := make([]int64, 0, len(updatePolicies))
	updatePolicyMap := make(map[int64]types.Policy, len(updatePolicyPKs))
	for _, p := range updatePolicies {
		updatePolicyPKs = append(updatePolicyPKs, p.ID)
		updatePolicyMap[p.ID] = p
	}

	daoForUpdatePolicies, err := s.manager.ListBySubjectPKAndPKs(subjectPK, updatePolicyPKs)
	if err != nil {
		err = errorWrapf(err, "manager.ListBySubjectPKAndPKs subjectPK=`%d`, pks=`%+v`", subjectPK, updatePolicyPKs)
		return
	}

	updatedActionPKExpressionPKs = make(map[int64][]int64)

	daoUpdateExpressions := make([]dao.Expression, 0, len(daoForUpdatePolicies))
	daoUpdatePolicies := make([]dao.Policy, 0, len(daoForUpdatePolicies))
	for _, p := range daoForUpdatePolicies {
		up := updatePolicyMap[p.PK]
		if up.ActionPK == p.ActionPK && p.TemplateID == 0 {
			daoUpdateExpressions = append(daoUpdateExpressions, dao.Expression{
				PK:         p.ExpressionPK,
				Type:       expressionTypeCustom,
				Expression: up.Expression,
				Signature:  stringx.MD5Hash(up.Expression),
			})

			// 更新过期时间
			if up.ExpiredAt > p.ExpiredAt {
				p.ExpiredAt = up.ExpiredAt
				daoUpdatePolicies = append(daoUpdatePolicies, p)
			}

			// collect the update actionPK/expressionPK, we need to know: {actionPK:[expr1PK, expr2PK] }
			updatedActionPKExpressionPKs[p.ActionPK] = append(updatedActionPKExpressionPKs[p.ActionPK], p.ExpressionPK)
		}
	}

	expressionPKs, err := s.expressionManger.BulkCreateWithTx(tx, daoCreateExpressions)
	if err != nil {
		err = errorWrapf(err, "expressionManger.BulkCreateWithTx expressions=`%+v`", daoCreateExpressions)
		return
	}
	// 根据已记录的索引关系填充policy的expressionPK
	for _, index := range policyExpressionIndexes {
		daoCreatePolicies[index.policyIndex].ExpressionPK = expressionPKs[index.expressionIndex]
	}

	err = s.manager.BulkCreateWithTx(tx, daoCreatePolicies)
	if err != nil {
		err = errorWrapf(err, "manager.BulkCreateWithTx policies=`%+v`", daoCreatePolicies)
		return
	}

	if len(daoUpdatePolicies) != 0 {
		err = s.manager.BulkUpdateExpiredAtWithTx(tx, daoUpdatePolicies)
		if err != nil {
			err = errorWrapf(err, "manager.BulkUpdateExpiredAtWithTx policies=`%+v`", daoUpdatePolicies)
			return
		}
	}

	err = s.expressionManger.BulkUpdateWithTx(tx, daoUpdateExpressions)
	if err != nil {
		err = errorWrapf(err, "expressionManger.BulkUpdateWithTx expressions=`%+v`", daoUpdateExpressions)
		return
	}

	err = s.deleteByPKsWithTx(tx, subjectPK, deletePolicyIDs)
	if err != nil {
		err = errorWrapf(err, "deleteByPKsWithTx subjectPK=`%d`, pks=`%+v`", subjectPK, deletePolicyIDs)
		return
	}

	return updatedActionPKExpressionPKs, err
}

func (s *policyService) deleteByPKsWithTx(tx *sqlx.Tx, subjectPK int64, pks []int64) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PolicySVC, "deleteByPKsWithTx")
	deletePolicies, err := s.manager.ListBySubjectPKAndPKs(subjectPK, pks)
	if err != nil {
		return errorWrapf(err, "manager.ListBySubjectPKAndPKs subjectPK=`%d`, pks=`%+v`", subjectPK, pks)
	}

	policyPKs := make([]int64, 0, len(deletePolicies))
	expressionPKs := make([]int64, 0, len(deletePolicies))
	for _, p := range deletePolicies {
		if p.TemplateID == 0 {
			expressionPKs = append(expressionPKs, p.ExpressionPK)
			policyPKs = append(policyPKs, p.PK)
		}
	}

	_, err = s.expressionManger.BulkDeleteByPKsWithTx(tx, expressionPKs)
	if err != nil {
		return errorWrapf(err, "expressionManger.BulkDeleteByPKsWithTx pks=`%+v`", expressionPKs)
	}

	_, err = s.manager.BulkDeleteByTemplatePKsWithTx(tx, subjectPK, PolicyTemplateIDCustom, policyPKs)
	if err != nil {
		return errorWrapf(err, "manager.BulkDeleteByPKsWithTx subjectPK=`%d`, pks=`%+v`", subjectPK, policyPKs)
	}
	return err
}

// DeleteByPKs ...
func (s *policyService) DeleteByPKs(subjectPK int64, pks []int64) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PolicySVC, "DeleteByPKs")

	tx, err := database.GenerateDefaultDBTx()
	if err != nil {
		return errorWrapf(err, "define tx fail")
	}
	defer database.RollBackWithLog(tx)

	err = s.deleteByPKsWithTx(tx, subjectPK, pks)
	if err != nil {
		return errorWrapf(err, "deleteByPKsWithTx subjectPK=`%d`, pks=`%+v`", subjectPK, pks)
	}

	err = tx.Commit()
	if err != nil {
		return errorWrapf(err, "tx.Commit fail")
	}
	return err
}

// HasAnyByActionPK ...
func (s *policyService) HasAnyByActionPK(actionPK int64) (bool, error) {
	return s.manager.HasAnyByActionPK(actionPK)
}

// generateSignatureExpressionPKMap generate signature expressionPK map if expression does not exist create it
func (s *policyService) generateSignatureExpressionPKMap(
	tx *sqlx.Tx, policies []types.Policy, actionPKWithResourceTypeSet *set.Int64Set,
) (signatureExpressionPKMap map[string]int64, err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PolicySVC, "generateSignatureExpressionPKMap")

	signatures := make([]string, 0, len(policies))
	for i := range policies {
		signature := stringx.MD5Hash(policies[i].Expression)
		signatures = append(signatures, signature)
	}

	// 查询可引用的expression
	expressions, err := s.expressionManger.ListDistinctBySignaturesType(signatures, expressionTypeTemplate)
	if err != nil {
		err = errorWrapf(err, "expressionManger.ListDistinctBySignaturesType signatures=`%+v` type=`%s`",
			signatures, expressionTypeTemplate)
		return nil, err
	}
	signatureExpressionPKMap = make(map[string]int64, len(expressions))
	existSignatures := set.NewStringSet()
	for _, e := range expressions {
		signatureExpressionPKMap[e.Signature] = e.PK
		existSignatures.Add(e.Signature)
	}

	daoExpressions := make([]dao.Expression, 0, len(policies))
	for i, signature := range signatures {
		// 已存在的signature, 不创建
		if existSignatures.Has(signature) {
			continue
		}

		policy := policies[i]
		// 如果操作没有关联资源类型, 不需要创建expression
		if !actionPKWithResourceTypeSet.Has(policy.ActionPK) {
			continue
		}

		daoExpressions = append(daoExpressions, dao.Expression{
			Type:       expressionTypeTemplate,
			Expression: policy.Expression,
			Signature:  signature, // 计算Hash
		})
		existSignatures.Add(signature)
	}

	// 创建引用不到的expression
	if len(daoExpressions) != 0 {
		var expressionPKs []int64
		expressionPKs, err = s.expressionManger.BulkCreateWithTx(tx, daoExpressions)
		if err != nil {
			err = errorWrapf(err, "expressionManger.BulkCreateWithTx expressions=`%+v`", daoExpressions)
			return
		}
		for i, e := range daoExpressions {
			signatureExpressionPKMap[e.Signature] = expressionPKs[i]
		}
	}

	return signatureExpressionPKMap, nil
}

// CreateAndDeleteTemplatePoliciesWithTx subject create and delete template policies
func (s *policyService) CreateAndDeleteTemplatePoliciesWithTx(
	tx *sqlx.Tx,
	subjectPK, templateID int64,
	createPolicies []types.Policy,
	deletePolicyIDs []int64,
	actionPKWithResourceTypeSet *set.Int64Set,
) (err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PolicySVC, "CreateAndDeleteTemplatePolicies")

	// 生成 signature -> expression pk map
	signatureExpressionPKMap, err := s.generateSignatureExpressionPKMap(
		tx, createPolicies, actionPKWithResourceTypeSet)
	if err != nil {
		err = errorWrapf(err, "generateSignatureExpressionPKMap fail")
		return
	}

	daoCreatePolicies := make([]dao.Policy, 0, len(createPolicies))
	for _, p := range createPolicies {
		signature := stringx.MD5Hash(p.Expression)
		// 操作有关联资源类型
		if actionPKWithResourceTypeSet.Has(p.ActionPK) {
			expressionPK, ok := signatureExpressionPKMap[signature]
			if !ok {
				err = errorWrapf(errPolicy, "generate policy expression error policy=`%+v`", p)
				return
			}

			daoCreatePolicies = append(daoCreatePolicies, dao.Policy{
				SubjectPK:    p.SubjectPK,
				ActionPK:     p.ActionPK,
				ExpiredAt:    p.ExpiredAt,
				ExpressionPK: expressionPK,
				TemplateID:   p.TemplateID,
			})
		} else {
			// 无关联资源的自定义权限, expression 为 -1, 不创建expression对象
			daoCreatePolicies = append(daoCreatePolicies, dao.Policy{
				SubjectPK:    p.SubjectPK,
				ActionPK:     p.ActionPK,
				ExpressionPK: expressionPKActionWithoutResource,
				ExpiredAt:    p.ExpiredAt,
				TemplateID:   p.TemplateID,
			})
		}
	}

	err = s.manager.BulkCreateWithTx(tx, daoCreatePolicies)
	if err != nil {
		err = errorWrapf(err, "manager.BulkCreateWithTx policies=`%+v`", daoCreatePolicies)
		return
	}

	_, err = s.manager.BulkDeleteByTemplatePKsWithTx(tx, subjectPK, templateID, deletePolicyIDs)
	if err != nil {
		err = errorWrapf(err, "deleteByPKsWithTx subjectPK=`%d`, pks=`%+v`", subjectPK, deletePolicyIDs)
		return
	}

	return err
}

// UpdateTemplatePoliciesWithTx subject update template policies
func (s *policyService) UpdateTemplatePoliciesWithTx(
	tx *sqlx.Tx,
	subjectPK int64,
	policies []types.Policy,
	actionPKWithResourceTypeSet *set.Int64Set,
) (err error) {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PolicySVC, "UpdateTemplatePoliciesWithTx")

	// 1. 查询要更新的policies数据
	policyPKs := make([]int64, 0, len(policies))
	for _, p := range policies {
		policyPKs = append(policyPKs, p.ID)
	}
	daoPolicies, err := s.manager.ListBySubjectPKAndPKs(subjectPK, policyPKs)
	if err != nil {
		err = errorWrapf(err, "manager.ListBySubjectPKAndPKs subjectPK=`%d`, pks=`%+v`", subjectPK, policyPKs)
		return
	}
	daoPolicyMap := make(map[int64]dao.Policy, len(daoPolicies))
	for _, p := range daoPolicies {
		daoPolicyMap[p.PK] = p
	}

	// 2. 生成 signature -> expression pk map
	signatureExpressionPKMap, err := s.generateSignatureExpressionPKMap(
		tx, policies, actionPKWithResourceTypeSet)
	if err != nil {
		err = errorWrapf(err, "generateSignatureExpressionPKMap fail")
		return
	}

	// 3. 生成需要更新的policies
	daoUpdatePolicies := make([]dao.Policy, 0, len(policies))
	for _, p := range policies {
		daoPolicy, ok := daoPolicyMap[p.ID]
		// policy不存在
		if !ok {
			err = errorWrapf(errPolicy, "policy not exists id=`%d`", p.ID)
			return
		}

		// policy数据不一致
		if p.ActionPK != daoPolicy.ActionPK || daoPolicy.TemplateID != p.TemplateID {
			err = errorWrapf(errPolicy, "policy action template error ID=`%d`, actionPK=`%d`, templateID=`%d`",
				p.ID, p.ActionPK, p.TemplateID)
			return
		}

		// 操作未关联资源类型, 不更新
		if daoPolicy.ExpressionPK == expressionPKActionWithoutResource {
			continue
		}

		signature := stringx.MD5Hash(p.Expression)
		daoPolicy.ExpressionPK, ok = signatureExpressionPKMap[signature]
		if !ok {
			err = errorWrapf(errPolicy, "generate policy expression error ID=`%d`", p.ID)
			return
		}

		daoUpdatePolicies = append(daoUpdatePolicies, daoPolicy)
	}

	// 4. 更新policy的expression pk引用
	err = s.manager.BulkUpdateExpressionPKWithTx(tx, daoUpdatePolicies)
	if err != nil {
		err = errorWrapf(err, "manager.BulkUpdateExpressionByPKWithTx policies=`%+v`", daoUpdatePolicies)
		return
	}

	return err
}

// DeleteByActionPKWithTx ...
func (s *policyService) DeleteByActionPKWithTx(tx *sqlx.Tx, actionPK int64) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PolicySVC, "DeleteByActionPK")

	// TODO: 需要改造，先查询数据量，然后再按照数量进行删除，同时需要考虑
	// 由于删除时可能数量较大，耗时长，锁行数据较多，影响鉴权，所以需要循环删除，限制每次删除的记录数，以及最多执行删除多少次
	rowLimit := int64(10000)
	maxAttempts := 100 // 相当于最多删除100万数据

	for i := 0; i < maxAttempts; i++ {
		rowsAffected, err := s.manager.DeleteByActionPKWithTx(tx, actionPK, rowLimit)
		if err != nil {
			return errorWrapf(err, "manager.DeleteByActionPKWithTx actionPK=`%d`", actionPK)
		}
		// 如果已经没有需要删除的了，就停止
		if rowsAffected == 0 {
			break
		}
	}

	return nil
}

// DeleteUnreferencedExpressions 删除未被引用的expression
func (s *policyService) DeleteUnreferencedExpressions() error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PolicySVC, "DeleteUnquotedExpression")
	updateAt := time.Now().Unix() - 24*60*60 // 取前一天的时间戳

	// 1. 更新被引用但是标记为未引用的expression
	err := s.expressionManger.ChangeReferencedExpressionTypeBeforeUpdateAt(
		expressionTypeUnreferenced, expressionTypeTemplate, updateAt)
	if err != nil {
		return errorWrapf(err, "expressionManger.ChangeReferencedExpressionTypeBeforeUpdateAt "+
			"fromType=`%d`, toType=`%d`, updateAt=`%d`",
			expressionTypeUnreferenced, expressionTypeTemplate, updateAt)
	}

	// 2. 删除标记未被引用的expression
	// 由于删除时可能数量较大，耗时长，锁行数据较多，影响鉴权，所以需要循环删除，限制每次删除的记录数，以及最多执行删除多少次
	rowLimit := int64(5000)
	maxAttempts := 100 // 相当于最多删除50万数据

	for i := 0; i < maxAttempts; i++ {
		rowsAffected, err := s.expressionManger.DeleteUnreferencedExpressionByTypeBeforeUpdateAt(
			expressionTypeUnreferenced,
			updateAt,
			rowLimit,
		)
		if err != nil {
			return errorWrapf(err, "expressionManger.DeleteByTypeBeforeUpdateAt type=`%d`, updateAt=`%d`",
				expressionTypeUnreferenced, updateAt)
		}

		// 如果已经没有需要删除的了，就停止
		if rowsAffected == 0 {
			break
		}
	}

	// 清理自定义权限的未被引用的expression
	for i := 0; i < maxAttempts; i++ {
		rowsAffected, err := s.expressionManger.DeleteUnreferencedExpressionByTypeBeforeUpdateAt(
			expressionTypeCustom,
			updateAt,
			rowLimit,
		)
		if err != nil {
			return errorWrapf(err, "expressionManger.DeleteByTypeBeforeUpdateAt type=`%d`, updateAt=`%d`",
				expressionTypeCustom, updateAt)
		}

		// 如果已经没有需要删除的了，就停止
		if rowsAffected == 0 {
			break
		}
	}

	// 3. 标记未被引用的expression
	err = s.expressionManger.ChangeUnreferencedExpressionType(expressionTypeTemplate, expressionTypeUnreferenced)
	if err != nil {
		return errorWrapf(err, "expressionManger.ChangeUnreferencedExpressionType fromType=`%d`, toType=`%d`",
			expressionTypeTemplate, expressionTypeUnreferenced)
	}

	return nil
}

// BulkDeleteBySubjectPKsWithTx ...
func (s *policyService) BulkDeleteBySubjectPKsWithTx(tx *sqlx.Tx, pks []int64) error {
	errorWrapf := errorx.NewLayerFunctionErrorWrapf(PolicySVC, "BulkDeleteBySubjectPKs")

	// 查询Policy里的Subject单独的Expression
	expressionPKs, err := s.manager.ListExpressionBySubjectsTemplate(pks, 0)
	if err != nil {
		return errorWrapf(err, "policyManager.ListExpressionBySubjectsTemplate subjectPKs=`%+v` fail", pks)
	}

	// 删除策略 policy
	err = s.manager.BulkDeleteBySubjectPKsWithTx(tx, pks)
	if err != nil {
		return errorWrapf(
			err, "policyManager.BulkDeleteBySubjectPKsWithTx subject_pks=`%+v` fail", pks)
	}

	// 删除策略对应的非来着权限模板的Expression
	_, err = s.expressionManger.BulkDeleteByPKsWithTx(tx, expressionPKs)
	if err != nil {
		return errorWrapf(
			err, "expressionManager.BulkDeleteByPKsWithTx pks=`%+v` fail", expressionPKs)
	}
	return nil
}
