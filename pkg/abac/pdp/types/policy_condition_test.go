package types_test

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"iam/pkg/abac/pdp/types"
)

var _ = Describe("PolicyCondition", func() {
	Describe("PolicyCondition.ToNewPolicyCondition", func() {
		It("single, not any", func() {
			pc := types.PolicyCondition{
				"StringEquals": {
					"system": []interface{}{"linux"},
				},
			}
			want := types.PolicyCondition{
				"StringEquals": {
					"bk_cmdb.host.system": []interface{}{"linux"},
				},
			}
			npc, err := pc.ToNewPolicyCondition("bk_cmdb", "host")
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), want, npc)
		})

		It("single, any", func() {
			pc := types.PolicyCondition{
				"Any": {
					"": []interface{}{},
				},
			}
			npc, err := pc.ToNewPolicyCondition("bk_cmdb", "host")
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), pc, npc)
		})

		It("AND", func() {
			pc := types.PolicyCondition{
				"AND": map[string][]interface{}{
					"content": {
						map[string]interface{}{
							"StringEquals": map[string]interface{}{
								"system": []interface{}{"linux"},
							},
						},
						map[string]interface{}{
							"StringPrefix": map[string]interface{}{
								"path": []interface{}{"/biz,1/"},
							},
						},
					},
				},
			}

			want := types.PolicyCondition{
				"AND": map[string][]interface{}{
					"content": {
						types.PolicyCondition{
							"StringEquals": map[string][]interface{}{
								"bk_cmdb.host.system": {"linux"},
							},
						},
						types.PolicyCondition{
							"StringPrefix": map[string][]interface{}{
								"bk_cmdb.host.path": {"/biz,1/"},
							},
						},
					},
				},
			}
			npc, err := pc.ToNewPolicyCondition("bk_cmdb", "host")
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), want, npc)
		})

		It("AND, fail1", func() {
			pc := types.PolicyCondition{
				"AND": map[string][]interface{}{
					"content": {
						// here, will assert fail
						map[string]map[string]interface{}{
							"StringEquals": {
								"system": []interface{}{"linux"},
							},
						},
						map[string]map[string]interface{}{
							"StringPrefix": {
								"path": []interface{}{"/biz,1/"},
							},
						},
					},
				},
			}

			npc, err := pc.ToNewPolicyCondition("bk_cmdb", "host")
			assert.Error(GinkgoT(), err)
			assert.Nil(GinkgoT(), npc)
		})
	})

	Describe("ResourceExpression.ToNewPolicyCondition", func() {
		It("single, not any", func() {
			pc := types.PolicyCondition{
				"StringEquals": {
					"system": []interface{}{"linux"},
				},
			}
			re := types.ResourceExpression{
				System:     "bk_cmdb",
				Type:       "host",
				Expression: pc,
			}
			want := types.PolicyCondition{
				"StringEquals": {
					"bk_cmdb.host.system": []interface{}{"linux"},
				},
			}
			npc, err := re.ToNewPolicyCondition()
			assert.NoError(GinkgoT(), err)
			assert.Equal(GinkgoT(), want, npc)
		})
	})
})
