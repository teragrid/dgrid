package commands

import (
	"os"

	"github.com/spf13/cobra"

	pvm "github.com/teragrid/dgrid/core/types/validator"
	"github.com/teragrid/dgrid/pkg/log"
)

// ResetAllCmd removes the database of this dgrid core
// instance.
var ResetAllCmd = &cobra.Command{
	Use:   "unsafe_reset_all",
	Short: "(unsafe) Remove all the data and WAL, reset this node's validator",
	Run:   resetAll,
}

// ResetValidatorCmd resets the private validator files.
var ResetValidatorCmd = &cobra.Command{
	Use:   "unsafe_reset_validator",
	Short: "(unsafe) Reset this node's validator",
	Run:   resetValidator,
}

// ResetAll removes the validator files.
// Exported so other CLI tools can use it.
func ResetAll(dbDir, validatorFile string, logger log.Logger) {
	resetFilePV(validatorFile, logger)
	if err := os.RemoveAll(dbDir); err != nil {
		logger.Error("Error removing directory", "err", err)
		return
	}
	logger.Info("Removed all data", "dir", dbDir)
}

// XXX: this is totally unsafe.
// it's only suitable for testnets.
func resetAll(cmd *cobra.Command, args []string) {
	ResetAll(config.DBDir(), config.ValidatorFile(), logger)
}

// XXX: this is totally unsafe.
// it's only suitable for testnets.
func resetValidator(cmd *cobra.Command, args []string) {
	resetFilePV(config.ValidatorFile(), logger)
}

func resetFilePV(validatorFile string, logger log.Logger) {
	// Get Validator
	if _, err := os.Stat(validatorFile); err == nil {
		pv := pvm.LoadFilePV(validatorFile)
		pv.Reset()
		logger.Info("Reset Validator", "file", validatorFile)
	} else {
		pv := pvm.GenFilePV(validatorFile)
		pv.Save()
		logger.Info("Generated Validator", "file", validatorFile)
	}
}
