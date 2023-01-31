// Copyright (c) 2022 Gitpod GmbH. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License.AGPL.txt in the project root for license information.

package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/gitpod-io/gitpod/gitpod-cli/pkg/gitpod"
	serverapi "github.com/gitpod-io/gitpod/gitpod-protocol"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/spf13/cobra"
)

// setTimeoutCmd sets the timeout of current workspace
var setTimeoutCmd = &cobra.Command{
	Use:   "set <duration>",
	Args:  cobra.ExactArgs(1),
	Short: "Set timeout of current workspace",
	Long: `Set timeout of current workspace.

Duration must be in the format of <n>m (minutes), <n>h (hours), or <n>d (days).
For example, 30m, 1h, 2d, etc.`,
	Example: `gitpod timeout set 1h`,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		wsInfo, err := gitpod.GetWSInfo(ctx)
		if err != nil {
			return
		}
		client, err := gitpod.ConnectToServer(ctx, wsInfo, []string{
			"function:setWorkspaceTimeout",
			"resource:workspace::" + wsInfo.WorkspaceId + "::get/update",
		})
		if err != nil {
			return
		}
		defer client.Close()
		duration, err := time.ParseDuration(args[0])
		if err != nil {
			return
		}
		if _, err = client.SetWorkspaceTimeout(ctx, wsInfo.WorkspaceId, duration); err != nil {
			if err, ok := err.(*jsonrpc2.Error); ok && err.Code == serverapi.PLAN_PROFESSIONAL_REQUIRED {
				return fmt.Errorf("Cannot extend workspace timeout for current plan, please upgrade your plan")
			}
			return
		}
		fmt.Printf("Workspace timeout has been set to %d minutes.\n", int(duration.Minutes()))
		return
	},
}

func init() {
	timeoutCmd.AddCommand(setTimeoutCmd)
}
