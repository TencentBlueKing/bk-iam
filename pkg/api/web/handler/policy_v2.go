package handler

import (
	"github.com/TencentBlueKing/gopkg/errorx"
	"github.com/gin-gonic/gin"

	"iam/pkg/abac/pap"
	"iam/pkg/abac/types"
	"iam/pkg/service"
	svctypes "iam/pkg/service/types"
	"iam/pkg/util"
)

// AlterPoliciesV2 godoc
// @Summary Alter policies v2/变更用户自定义申请策略
// @Description alter policies v2
// @ID v2-api-web-alter-policies
// @Tags web
// @Accept json
// @Produce json
// @Param system_id path string true "system id"
// @Param body body policiesAlterSerializerV2 true "create and update policies"
// @Success 200 {object} util.Response
// @Header 200 {string} X-Request-Id "the request id"
// @Security AppCode
// @Security AppSecret
// @Router /api/v2/web/systems/{system_id}/policies [post]
func AlterPoliciesV2(c *gin.Context) {
	var body policiesAlterSerializerV2
	if err := c.ShouldBindJSON(&body); err != nil {
		util.BadRequestErrorJSONResponse(c, util.ValidationErrorMessage(err))
		return
	}

	// TODO: 虽然SaaSy已经保证了参数正确性，但后台得看看还需要做哪些校验

	systemID := c.Param("system_id")

	subject := types.Subject{
		Type:      body.Subject.Type,
		ID:        body.Subject.ID,
		Attribute: types.NewSubjectAttribute(),
	}

	createPolicies := make([]types.Policy, 0, len(body.CreatePolicies))
	for _, p := range body.CreatePolicies {
		createPolicies = append(
			createPolicies,
			convertToInternalTypesPolicy(systemID, subject, 0, service.PolicyTemplateIDCustom, p),
		)
	}

	updatePolicies := make([]types.Policy, 0, len(body.UpdatePolicies))
	for _, p := range body.UpdatePolicies {
		updatePolicies = append(
			updatePolicies,
			convertToInternalTypesPolicy(systemID, subject, p.ID, service.PolicyTemplateIDCustom, p.policy),
		)
	}

	resourceChangedActions := make([]types.ResourceChangedAction, 0, len(body.ResourceActions))
	for _, ra := range body.ResourceActions {
		resourceChangedActions = append(resourceChangedActions, types.ResourceChangedAction{
			Resource: types.ThinResourceNode{
				System: ra.Resource.SystemID,
				Type:   ra.Resource.Type,
				ID:     ra.Resource.ID,
			},
			CreatedActionIDs: ra.CreatedActionIDs,
			DeletedActionIDs: ra.DeletedActionIDs,
		})
	}

	ctl := pap.NewPolicyControllerV2()
	err := ctl.Alter(
		systemID, body.Subject.Type, body.Subject.ID, body.TemplateID,
		createPolicies, updatePolicies, body.DeletePolicyIDs,
		resourceChangedActions,
		svctypes.ConvertToAuthTypeInt(body.GroupAuthType),
	)
	if err != nil {
		err = errorx.Wrapf(
			err,
			"Handler",
			"AlterPoliciesV2",
			"systemID=`%s`, subjectType=`%s`, subjectID=`%s`, templateID=`%d`, "+
				"createPolicies=`%+v`, updatePolicies=`%+v` deletePolicyIDs=`%+v`, "+
				"resourceActions=`%+v groupAuthType=`%s``",
			systemID, body.Subject.Type, body.Subject.ID, body.TemplateID,
			createPolicies, updatePolicies, body.DeletePolicyIDs,
			resourceChangedActions, body.GroupAuthType,
		)
		util.SystemErrorJSONResponse(c, err)
		return
	}

	util.SuccessJSONResponse(c, "ok", gin.H{})
}
