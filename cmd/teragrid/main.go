package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"

	"github.com/teragrid/teralibs/cli"

	cmd "github.com/teragrid/teragrid/cmd/teragrid/commands"
	cfg "github.com/teragrid/teragrid/config"
	nm "github.com/teragrid/teragrid/node"
)

func main() {

	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	//viper.AddConfigPath(os.U)
	err := viper.ReadInConfig()
	if err != nil {
		//panic(err)
		fmt.Println("Config not found")
	} else {
		chains := viper.GetStringSlice("chains")
		fmt.Println("ChainSize:", len(chains))
		for idx, item := range chains {
			fmt.Println("Chain", idx, item)
		}
	}

	/*
		//err2 := viper.Unmarshal(&cfgIn)
		//	if err2 != nil {
		//		cfg := config.DefaultConfig()
		//		fmt.Println("ConfigSize:", len(cfg.ChainConfigs))
		//		viper.SetDefault("LogLevel", cfg.LogLevel)
		//		viper.SetDefault("Chains", cfg.ChainConfigs)
		//		viper.SetConfigType("json")
		//		viper.WriteConfig()
		//		return
		//	}
		//	cfgIn.ChainConfigs = []viper.GetStringMap("chains")
		//	cfgIn.LogLevel = viper.GetString("LogLevel")
		//	fmt.Println("LogLevel:", cfgIn.LogLevel)
		//	fmt.Println("ConfigSize:", len(cfgIn.ChainConfigs))
		return
	*/
	rootCmd := cmd.RootCmd
	rootCmd.AddCommand(
		//		cmd.GenValidatorCmd,
		cmd.InitFilesCmd,
		//		cmd.ProbeUpnpCmd,
		//		cmd.LiteCmd,
		//		cmd.ReplayCmd,
		//		cmd.ReplayConsoleCmd,
		//		cmd.ResetAllCmd,
		//		cmd.ResetPrivValidatorCmd,
		//		cmd.ShowValidatorCmd,
		cmd.TestnetFilesCmd,
		//		cmd.ShowNodeIDCmd,
		//		cmd.GenNodeKeyCmd,
		cmd.VersionCmd)

	// NOTE:
	// Users wishing to:
	//	* Use an external signer for their validators
	//	* Supply an in-proc asura app
	//	* Supply a genesis doc file from another source
	//	* Provide their own DB implementation
	// can copy this file and use something other than the
	// DefaultNewNode function
	nodeFunc := nm.DefaultNewNode

	// Create & start node
	rootCmd.AddCommand(cmd.NewRunNodeCmd(nodeFunc))

	cmd := cli.PrepareBaseCmd(rootCmd, "TM", os.ExpandEnv(filepath.Join("$HOME", cfg.DefaultteragridDir)))
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
