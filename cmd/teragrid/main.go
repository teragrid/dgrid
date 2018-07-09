package main

import (
	"os"
	"path/filepath"

	"github.com/teragrid/teralibs/cli"

	cmd "github.com/teragrid/teragrid/cmd/teragrid/commands"
	cfg "github.com/teragrid/teragrid/config"
	nm "github.com/teragrid/teragrid/node"
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
		cmd.ResetPrivValidatorCmd,
		cmd.ShowValidatorCmd,
		cmd.TestnetFilesCmd,
		cmd.ShowNodeIDCmd,
		cmd.GenNodeKeyCmd,
		cmd.VersionCmd)

	//	// NOTE:
	//	// Users wishing to:
	//	//	* Use an external signer for their validators
	//	//	* Supply an in-proc asura app
	//	//	* Supply a genesis doc file from another source
	//	//	* Provide their own DB implementation
	//	// can copy this file and use something other than the
	//	// DefaultNewNode function
	nodeFunc := nm.DefaultNewNode

	// Create & start node
	rootCmd.AddCommand(cmd.NewRunNodeCmd(nodeFunc))

	cmd := cli.PrepareBaseCmd(rootCmd, "TM", os.ExpandEnv(filepath.Join("$HOME", cfg.DefaultTendermintDir)))
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
