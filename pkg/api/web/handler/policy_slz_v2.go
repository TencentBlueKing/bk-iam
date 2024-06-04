package handler

// 变更策略的 body
type policiesAlterSerializerV2 struct {
	Subject    subject `json:"subject"     binding:"required"`
	TemplateID int64   `json:"template_id" binding:"omitempty"`

	CreatePolicies  []policy       `json:"create_policies"   binding:"required"`
	UpdatePolicies  []updatePolicy `json:"update_policies"   binding:"required"`
	DeletePolicyIDs []int64        `json:"delete_policy_ids" binding:"required"`

	ResourceActions []resourceAction `json:"resource_actions" binding:"required"`

	GroupAuthType string `json:"group_auth_type" binding:"required,oneof=all rbac abac none"`
}

type resourceSerializer struct {
	SystemID string `json:"system_id" binding:"required"`
	Type     string `json:"type"      binding:"required"`
	ID       string `json:"id"        binding:"required"`
}

type resourceAction struct {
	Resource         resourceSerializer `json:"resource"           binding:"required"`
	CreatedActionIDs []string           `json:"created_action_ids" binding:"required"`
	DeletedActionIDs []string           `json:"deleted_action_ids" binding:"required"`
}
