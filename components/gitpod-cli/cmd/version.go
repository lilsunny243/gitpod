// Copyright (c) 2022 Gitpod GmbH. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License.AGPL.txt in the project root for license information.

package cmd

import (
	_ "embed"
	"fmt"

	gitpod "github.com/gitpod-io/gitpod/gitpod-cli/pkg/gitpod"

	"github.com/spf13/cobra"
)

// urlCmd represents the url command
var versionCmd = &cobra.Command{
	Use:    "version",
	Hidden: false,
	Short:  "Prints the version of the CLI",
	Args:   cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		fmt.Println(gitpod.Version)
		return
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
