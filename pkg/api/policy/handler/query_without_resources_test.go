/*
 * TencentBlueKing is pleased to support the open source community by making BlueKing-IAM available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 */

package handler

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestQueryByActionsWithoutResourcesRejectsResources(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		path       string
		body       string
		wantStatus int
		call       func(*gin.Context)
	}{
		{
			name: "v1", path: "/api/v1/policy/query_by_actions_without_resources", wantStatus: 200,
			body: `{"system":"bk_itsm","subject":{"type":"user","id":"admin"},` +
				`"actions":[{"id":"ticket_view"}],` +
				`"resources":[{"system":"bk_itsm","type":"ticket","id":"1","attribute":{}}]}`,
			call: BatchQueryByActionsWithoutResources,
		},
		{
			name: "v2", path: "/api/v2/policy/systems/bk_itsm/query_by_actions_without_resources/", wantStatus: 400,
			body: `{"subject":{"type":"user","id":"admin"},` +
				`"actions":[{"id":"ticket_view"}],` +
				`"resources":[{"system":"bk_itsm","type":"ticket","id":"1","attribute":{}}]}`,
			call: BatchQueryV2ByActionsWithoutResources,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest("POST", tt.path, bytes.NewBufferString(tt.body))
			ctx.Request.Header.Set("Content-Type", "application/json")

			tt.call(ctx)

			assert.Equal(t, tt.wantStatus, recorder.Code)
			assert.Contains(t, recorder.Body.String(), "resources must be empty")
		})
	}
}
