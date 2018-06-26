package commands

import (
	"github.com/spf13/cobra"

	cfg "github.com/teragrid/teragrid/config"
	"github.com/teragrid/teragrid/p2p"
	"github.com/teragrid/teragrid/types"
	pvm "github.com/teragrid/teragrid/types/priv_validator"
	cmn "github.com/teragrid/teralibs/common"
)

// InitFilesCmd initialises a fresh teragrid Core instance.
var InitFilesCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize teragrid",
	RunE:  initFiles,
}

func initFiles(cmd *cobra.Command, args []string) error {
	return initFilesWithConfig(mainConfig)
}

func initFilesWithConfig(config *cfg.Config) error {
	config.SetRoot(config.RootDir)
	for _, chain := range config.ChainConfigs {
		// private validator
		privValFile := chain.PrivValidatorFile()
		var pv *pvm.FilePV
		if cmn.FileExists(privValFile) {
			pv = pvm.LoadFilePV(privValFile)
			logger.Info("Found private validator", "path", privValFile)
		} else {
			pv = pvm.GenFilePV(privValFile)
			pv.Save()
			logger.Info("Generated private validator", "path", privValFile)
		}

		nodeKeyFile := chain.NodeKeyFile()
		if cmn.FileExists(nodeKeyFile) {
			logger.Info("Found node key", "path", nodeKeyFile)
		} else {
			if _, err := p2p.LoadOrGenNodeKey(nodeKeyFile); err != nil {
				return err
			}
			logger.Info("Generated node key", "path", nodeKeyFile)
		}

		// genesis file
		genFile := chain.GenesisFile()
		if cmn.FileExists(genFile) {
			logger.Info("Found genesis file", "path", genFile)
		} else {
			genDoc := types.GenesisDoc{
				ChainID: cmn.Fmt("test-chain-%v", cmn.RandStr(6)),
			}
			genDoc.Validators = []types.GenesisValidator{{
				PubKey: pv.GetPubKey(),
				Power:  10,
			}}

			if err := genDoc.SaveAs(genFile); err != nil {
				return err
			}
			logger.Info("Generated genesis file", "path", genFile)
		}
	}
	return nil
}
