// Copyright (c) 2020 Gitpod GmbH. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License.AGPL.txt in the project root for license information.

package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gitpod-io/gitpod/gitpod-cli/pkg/gitpod"
	protocol "github.com/gitpod-io/gitpod/gitpod-protocol"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/spf13/cobra"
)

const (
	ErrorCodeSnapshotNotFound = 404
	ErrorCodeSnapshotError    = 630
)

// snapshotCmd represents the snapshotCmd command
var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Take a snapshot of the current workspace",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		ctx, cancel := context.WithCancel(context.Background())
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-sigChan
			cancel()
		}()
		wsInfo, err := gitpod.GetWSInfo(ctx)
		if err != nil {
			return
		}
		client, err := gitpod.ConnectToServer(ctx, wsInfo, []string{
			"function:takeSnapshot",
			"function:waitForSnapshot",
			"resource:workspace::" + wsInfo.WorkspaceId + "::get/update",
		})
		if err != nil {
			return
		}
		defer client.Close()
		snapshotId, err := client.TakeSnapshot(ctx, &protocol.TakeSnapshotOptions{
			WorkspaceID: wsInfo.WorkspaceId,
			DontWait:    true,
		})
		if err != nil {
			return
		}
		for ctx.Err() == nil {
			err = client.WaitForSnapshot(ctx, snapshotId)
			if err != nil {
				var responseErr *jsonrpc2.Error
				if errors.As(err, &responseErr) && (responseErr.Code == ErrorCodeSnapshotNotFound || responseErr.Code == ErrorCodeSnapshotError) {
					return err
				}
				time.Sleep(time.Second * 3)
			} else {
				break
			}
		}
		url := fmt.Sprintf("%s/#snapshot/%s", wsInfo.GitpodHost, snapshotId)
		fmt.Println(url)
		return
	},
}

func init() {
	rootCmd.AddCommand(snapshotCmd)
}
