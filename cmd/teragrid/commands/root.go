package commands

import (
	"os"
	//"os/user"
	"path/filepath"
	//"runtime"
	//"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	cfg "github.com/teragrid/teragrid/config"
	"github.com/teragrid/teralibs/cli"
	tmflags "github.com/teragrid/teralibs/cli/flags"
	"github.com/teragrid/teralibs/log"
)

var (
	mainConfig = cfg.DefaultConfig()
	//mainConfig *cfg.Config
	logger = log.NewTMLogger(log.NewSyncWriter(os.Stdout))
)

func init() {
	registerFlagsRootCmd(RootCmd)
}

func registerFlagsRootCmd(cmd *cobra.Command) {
	cmd.PersistentFlags().String("log_level", mainConfig.LogLevel, "Log level")
	cmd.PersistentFlags().StringP("config", "c", "", "Alternate configuration file to read. Defaults to $HOME/.tendermint/")

	//viper.BindPFlag("ConfigFileName", cmd.PersistentFlags().Lookup("config"))
	//viper.BindPFlag("Home", cmd.PersistentFlags().Lookup("home"))
}

// ParseConfig retrieves the default environment configuration,
// sets up the Teragrid root and ensures that the root exists
func ParseConfig() (*cfg.Config, error) {

	var conf *cfg.Config

	rootDir := viper.GetString("home")
	chains := viper.GetStringSlice("chains")

	if len(chains) > 0 {
		hasDefault := false
		chainConfigs := make([]*cfg.ChainConfig, len(chains))
		for idx, item := range chains {
			if item == "default" {
				hasDefault = true
			}
			chainConfigs[idx] = cfg.DefaultChainConfig(item)
		}
		if !hasDefault {
			//chainConfigs = append(chainConfigs, cfg.DefaultChainConfig("default"))
		}
		conf = &cfg.Config{
			RootDir:      "",
			LogLevel:     cfg.DefaultPackageLogLevels(),
			ChainConfigs: chainConfigs,
		}
	} else {
		conf = cfg.DefaultConfig()
	}
	if rootDir == "" {
		panic("Error")
	}
	conf.SetRoot(rootDir)

	for _, chain := range conf.ChainConfigs {
		var chainConfig = filepath.Join(chain.RootDir, "config", "config.toml")
		//fmt.Println("chainConfig " + chainConfig)

		var chainViper = viper.New()
		//chainViper.SetConfigType("json")
		chainViper.SetConfigFile(chainConfig)
		err := chainViper.ReadInConfig()
		if err == nil {
			err = chainViper.Unmarshal(chain)
			if err != nil {
				panic(err)
			}
		}
	}

	cfg.EnsureRoot(conf.RootDir, conf)
	return conf, nil
}

// RootCmd is the root command for Teragrid core.
var RootCmd = &cobra.Command{
	Use:   "tendermint",
	Short: "Teragrid Core (BFT Consensus) in Go",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
		if cmd.Name() == VersionCmd.Name() {
			return nil
		}
		mainConfig, err = ParseConfig()
		if err != nil {
			return err
		}
		logger, err = tmflags.ParseLogLevel(mainConfig.LogLevel, logger, cfg.DefaultLogLevel())
		if err != nil {
			return err
		}
		if viper.GetBool(cli.TraceFlag) {
			logger = log.NewTracingLogger(logger)
		}
		logger = logger.With("module", "main")
		return nil
	},
}
