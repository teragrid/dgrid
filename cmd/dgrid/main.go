/*
Copyright Teragrid.Network 2019 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"os"
	"path/filepath"

	"github.com/teragrid/dgrid/pkg/cli"

	"github.com/teragrid/dgrid/cell"
	cmd "github.com/teragrid/dgrid/cmd/dgrid/commands"
	cfg "github.com/teragrid/dgrid/core/config"
)

func main() {
	rootCmd := cmd.RootCmd
	rootCmd.AddCommand(
		cmd.GenValidatorCmd,
		cmd.InitFilesCmd,
		cmd.ProbeUpnpCmd,
		cmd.LiteCmd,
		cmd.ReplayCmd,
		cmd.ReplayConsoleCmd,
		cmd.ResetAllCmd,
		cmd.ResetValidatorCmd,
		cmd.ShowValidatorCmd,
		cmd.TestnetFilesCmd,
		cmd.ShowNodeIDCmd,
		cmd.GenNodeKeyCmd,
		cmd.VersionCmd)

	// NOTE:
	// Users wishing to:
	//	* Use an external signer for their validators
	//	* Supply an in-proc asura app
	//	* Supply a genesis doc file from another source
	//	* Provide their own DB implementation
	// can copy this file and use something other than the
	// DefaultNewCell function
	cellFunc := cell.DefaultNewCell

	// Create & start node
	rootCmd.AddCommand(cmd.NewRunCellCmd(cellFunc))

	cmd := cli.PrepareBaseCmd(rootCmd, "TG", os.ExpandEnv(filepath.Join("$HOME", cfg.DefaultdgridDir)))
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
