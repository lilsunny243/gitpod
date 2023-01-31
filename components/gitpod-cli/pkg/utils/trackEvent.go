// Copyright (c) 2023 Gitpod GmbH. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License.AGPL.txt in the project root for license information.

package utils

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/gitpod-io/gitpod/common-go/analytics"
)

const (
	Outcome_Success   = "success"
	Outcome_UserErr   = "user_error"
	Outcome_SystemErr = "system_error"
)

const (
	// System
	SystemErrorCode = "system_error"
	UserErrorCode   = "user_error"

	// Rebuild
	RebuildErrorCode_ImageBuildFailed    = "rebuild_image_build_failed"
	RebuildErrorCode_DockerErr           = "rebuild_docker_err"
	RebuildErrorCode_DockerNotFound      = "rebuild_docker_not_found"
	RebuildErrorCode_DockerRunFailed     = "rebuild_docker_run_failed"
	RebuildErrorCode_MalformedGitpodYaml = "rebuild_malformed_gitpod_yaml"
	RebuildErrorCode_MissingGitpodYaml   = "rebuild_missing_gitpod_yaml"
	RebuildErrorCode_NoCustomImage       = "rebuild_no_custom_image"
)

type TrackCommandUsageParams struct {
	Command            []string `json:"command,omitempty"`
	Flags              []string `json:"flags,omitempty"`
	Duration           int64    `json:"duration,omitempty"`
	ErrorCode          string   `json:"errorCode,omitempty"`
	WorkspaceId        string   `json:"workspaceId,omitempty"`
	InstanceId         string   `json:"instanceId,omitempty"`
	Timestamp          int64    `json:"timestamp,omitempty"`
	ImageBuildDuration int64    `json:"imageBuildDuration,omitempty"`
	Outcome            string   `json:"outcome,omitempty"`
}

type analyticsEvent struct {
	Data      *TrackCommandUsageParams
	StartTime time.Time
	w         analytics.Writer
}

func NewAnalyticsEvent() *analyticsEvent {
	return &analyticsEvent{
		w: analytics.NewFromEnvironment(),
	}
}

func (e *analyticsEvent) ExportToJson() string {
	data, err := json.Marshal(e.Data)
	if err != nil {
		LogError(err, "error marshaling analytics data")
		os.Exit(1)
	}
	return string(data)
}

func (e *analyticsEvent) Send(ctx context.Context, userId string) {
	defer e.w.Close()

	data := make(map[string]interface{})
	jsonData, err := json.Marshal(e.Data)
	if err != nil {
		LogError(err, "Could not marshal event data")
		return
	}
	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		LogError(err, "Could not unmarshal event data")
		return
	}

	e.w.Track(analytics.TrackMessage{
		Identity:   analytics.Identity{UserID: userId},
		Event:      "gp_command",
		Properties: data,
	})
}
