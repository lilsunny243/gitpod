// Copyright (c) 2020 Gitpod GmbH. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License.AGPL.txt in the project root for license information.

package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gitpod-io/gitpod/gitpod-cli/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	rootCmdName = "gp"
)

type GpError struct {
	Err       error
	Message   string
	OutCome   string
	ErrorCode string
	ExitCode  int
	Slience   bool
}

func (e GpError) Error() string {
	if e.Slience {
		return ""
	}
	ret := e.Message
	if ret != "" && e.Err != nil {
		ret += ": "
	}
	if e.Err != nil {
		ret += e.Err.Error()
	}
	return ret
}

func GetCommandName(path string) []string {
	return strings.Fields(strings.TrimSpace(strings.TrimPrefix(path, rootCmdName)))
}

var rootCmd = &cobra.Command{
	Use:           rootCmdName,
	SilenceErrors: true,
	Short:         "Command line interface for Gitpod",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		cmd.SilenceUsage = true
		cmdName := GetCommandName(cmd.CommandPath())
		usedFlags := []string{}
		flags := cmd.Flags()
		flags.VisitAll(func(flag *pflag.Flag) {
			if flag.Changed {
				usedFlags = append(usedFlags, flag.Name)
			}
		})
		utils.AnalyticsEvent.Data.Command = cmdName
		utils.AnalyticsEvent.Data.Flags = usedFlags
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

	err := rootCmd.Execute()
	exitCode := 0
	utils.AnalyticsEvent.Data.Outcome = utils.Outcome_Success
	utils.AnalyticsEvent.Data.Duration = time.Since(utils.AnalyticsEvent.StartTime).Milliseconds()

	if err != nil {
		utils.AnalyticsEvent.Data.Outcome = utils.Outcome_SystemErr
		exitCode = 1
		if gpErr, ok := err.(GpError); ok {
			if gpErr.ExitCode != 0 {
				exitCode = gpErr.ExitCode
			}
			if gpErr.OutCome != "" {
				utils.AnalyticsEvent.Data.Outcome = gpErr.OutCome
			}
			if gpErr.ErrorCode != "" {
				utils.AnalyticsEvent.Data.ErrorCode = gpErr.ErrorCode
			}
		}
		if utils.AnalyticsEvent.Data.ErrorCode == "" {
			switch utils.AnalyticsEvent.Data.Outcome {
			case utils.Outcome_UserErr:
				utils.AnalyticsEvent.Data.ErrorCode = utils.UserErrorCode
			case utils.Outcome_SystemErr:
				utils.AnalyticsEvent.Data.ErrorCode = utils.SystemErrorCode
			}
		}
	}

	sendAnalytics()

	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(exitCode)
	}
}

func sendAnalytics() {
	if len(utils.AnalyticsEvent.Data.Command) == 0 {
		return
	}
	data, err := utils.AnalyticsEvent.ExportToJson()
	if err != nil {
		return
	}
	sendAnalytics := exec.Command(
		"/proc/self/exe",
		"send-analytics",
		"--data",
		data,
	)
	sendAnalytics.Stdout = ioutil.Discard
	sendAnalytics.Stderr = ioutil.Discard

	// fire and release
	_ = sendAnalytics.Start()
	if sendAnalytics.Process != nil {
		_ = sendAnalytics.Process.Release()
	}
}
