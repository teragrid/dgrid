package commands

import (
	"github.com/spf13/cobra"

	"github.com/teragrid/dgrid/core/blockchain/p2p"
	cfg "github.com/teragrid/dgrid/core/config"
	"github.com/teragrid/dgrid/core/consensus/validator"
	cmn "github.com/teragrid/dgrid/pkg/common"
	"github.com/teragrid/dgrid/core/types"
)

// InitFilesCmd initialises a fresh dgrid Core instance.
var InitFilesCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize dgrid",
	RunE:  initFiles,
}

func initFiles(cmd *cobra.Command, args []string) error {
	return initFilesWithConfig(config)
}

func initFilesWithConfig(config *cfg.Config) error {
	// private validator
	validatorFile := config.ValidatorFile()
	var fVal *validator.FilePV
	if cmn.FileExists(validatorFile) {
		fVal = validator.LoadFilePV(validatorFile)
		logger.Info("Found private validator", "path", validatorFile)
	} else {
		fVal = validator.GenFilePV(validatorFile)
		fVal.Save()
		logger.Info("Generated private validator", "path", validatorFile)
	}

	nodeKeyFile := config.NodeKeyFile()
	if cmn.FileExists(nodeKeyFile) {
		logger.Info("Found node key", "path", nodeKeyFile)
	} else {
		if _, err := p2p.LoadOrGenNodeKey(nodeKeyFile); err != nil {
			return err
		}
		logger.Info("Generated node key", "path", nodeKeyFile)
	}

	// genesis file
	genFile := config.GenesisFile()
	if cmn.FileExists(genFile) {
		logger.Info("Found genesis file", "path", genFile)
	} else {
		genDoc := types.GenesisDoc{
			LeagueID: cmn.Fmt("test-chain-%v", cmn.RandStr(6)),
		}
		genDoc.Validators = []types.GenesisValidator{{
			PubKey: fVal.GetPubKey(),
			Power:  10,
		}}

		if err := genDoc.SaveAs(genFile); err != nil {
			return err
		}
		logger.Info("Generated genesis file", "path", genFile)
	}

	return nil
}
