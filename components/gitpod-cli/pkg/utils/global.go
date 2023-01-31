// Copyright (c) 2023 Gitpod GmbH. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License.AGPL.txt in the project root for license information.

package utils

import (
	"time"

	"github.com/gitpod-io/gitpod/common-go/analytics"
)

var AnalyticsEvent *analyticsEvent = &analyticsEvent{
	StartTime: time.Now(),
	w:         analytics.NewFromEnvironment(),
	Data: &TrackCommandUsageParams{
		Timestamp: time.Now().UnixMilli(),
	},
}
