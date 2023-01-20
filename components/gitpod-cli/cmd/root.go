// Copyright (c) 2020 Gitpod GmbH. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License.AGPL.txt in the project root for license information.

package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gitpod-io/gitpod/gitpod-cli/pkg/supervisor"
	"github.com/gitpod-io/gitpod/gitpod-cli/pkg/utils"
	"github.com/spf13/cobra"
)

type contextKey int

const (
	ctxKeyAnalytics        contextKey = iota
	ctxKeySupervisorClient contextKey = iota
	ctxKeyError            contextKey = iota
	rootCmdName                       = "gp"
)

var skipAnalytics = false

func GetCommandName(path string) []string {
	return strings.Fields(strings.TrimSpace(strings.TrimPrefix(path, rootCmdName)))
}

var rootCmd = &cobra.Command{
	Use:   rootCmdName,
	Short: "Command line interface for Gitpod",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if os.Args[0] != "gp" {
			// skip analytics when running in development mode
			skipAnalytics = true
		}

		if cmd.Name() == "send-analytics" {
			// skip itself, otherwise we'd end up in a loop
			skipAnalytics = true
		}

		ctx := context.Background()
		supervisorClient, err := supervisor.New(ctx)
		if err != nil {
			utils.LogError(ctx, err, "Could not initialize supervisor client", supervisorClient)
			return
		}
		supervisorClientCtx := context.WithValue(cmd.Context(), ctxKeySupervisorClient, supervisorClient)
		cmd.SetContext(supervisorClientCtx)

		if skipAnalytics {
			return
		}

		cmdName := GetCommandName(cmd.CommandPath())

		event := utils.NewAnalyticsEvent(ctx, supervisorClient, &utils.TrackCommandUsageParams{
			Command: cmdName,
		})

		analyticsCtx := context.WithValue(cmd.Context(), ctxKeyAnalytics, event)
		cmd.SetContext(analyticsCtx)
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()

		supervisorClient := ctx.Value(ctxKeySupervisorClient).(*supervisor.SupervisorClient)
		defer supervisorClient.Close()

		if skipAnalytics {
			return
		}

		event := ctx.Value(ctxKeyAnalytics).(*utils.AnalyticsEvent)

		cmdErr := ctx.Value(ctxKeyError)
		if cmdErr != nil {
			utils.LogError(ctx, cmdErr.(error), "gp cli error", supervisorClient)
			event.Set("Outcome", utils.Outcome_SystemErr)
		} else {
			event.Set("Outcome", utils.Outcome_Success)
		}

		sendAnalytics := exec.Command(
			rootCmdName,
			"send-analytics",
			"--data",
			event.ExportToJson(ctx),
		)

		sendAnalytics.Stdout = os.Stdout
		sendAnalytics.Stderr = os.Stderr
		sendAnalytics.Stdin = os.Stdin

		// sendAnalytics.Stdout = ioutil.Discard
		// sendAnalytics.Stderr = ioutil.Discard
		// sendAnalytics.Stdin = os.Stdin

		// fire and forget
		_ = sendAnalytics.Start()

		if cmdErr != nil {
			os.Exit(1)
		}
	},
}

var noColor bool

// Execute runs the root command
func Execute() {
	entrypoint := strings.TrimPrefix(filepath.Base(os.Args[0]), "gp-")
	for _, c := range rootCmd.Commands() {
		if c.Name() == entrypoint {
			// we can't call subcommands directly (they just call their parents - thanks cobra),
			// so instead we have to manipulate the os.Args
			os.Args = append([]string{os.Args[0], entrypoint}, os.Args[1:]...)
			break
		}
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
