package commands

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	cfg "github.com/teragrid/teragrid/config"
	"github.com/teragrid/teragrid/p2p"
	"github.com/teragrid/teragrid/types"
	"github.com/teragrid/teragrid/types/priv_validator"
	cmn "github.com/teragrid/teralibs/common"
)

var (
	nValidators    int
	nNonValidators int
	outputDir      string
	nodeDirPrefix  string

	populatePersistentPeers bool
	hostnamePrefix          string
	startingIPAddress       string
	p2pPort                 int
)

const (
	nodeDirPerm = 0755
)

func init() {
	TestnetFilesCmd.Flags().IntVar(&nValidators, "v", 4,
		"Number of validators to initialize the testnet with")
	TestnetFilesCmd.Flags().IntVar(&nNonValidators, "n", 0,
		"Number of non-validators to initialize the testnet with")
	TestnetFilesCmd.Flags().StringVar(&outputDir, "o", "./mytestnet",
		"Directory to store initialization data for the testnet")
	TestnetFilesCmd.Flags().StringVar(&nodeDirPrefix, "node-dir-prefix", "node",
		"Prefix the directory name for each node with (node results in node0, node1, ...)")

	TestnetFilesCmd.Flags().BoolVar(&populatePersistentPeers, "populate-persistent-peers", true,
		"Update config of each node with the list of persistent peers build using either hostname-prefix or starting-ip-address")
	TestnetFilesCmd.Flags().StringVar(&hostnamePrefix, "hostname-prefix", "node",
		"Hostname prefix (node results in persistent peers list ID0@node0:26656, ID1@node1:26656, ...)")
	TestnetFilesCmd.Flags().StringVar(&startingIPAddress, "starting-ip-address", "",
		"Starting IP address (192.168.0.1 results in persistent peers list ID0@192.168.0.1:26656, ID1@192.168.0.2:26656, ...)")
	TestnetFilesCmd.Flags().IntVar(&p2pPort, "p2p-port", 26656,
		"P2P Port")
}

// TestnetFilesCmd allows initialisation of files for a Teragrid testnet.
var TestnetFilesCmd = &cobra.Command{
	Use:   "testnet",
	Short: "Initialize files for a Teragrid testnet",
	Long: `testnet will create "v" + "n" number of directories and populate each with
necessary files (private validator, genesis, config, etc.).

Note, strict routability for addresses is turned off in the config file.

Optionally, it will fill in persistent_peers list in config file using either hostnames or IPs.

Example:

	teragrid testnet --v 4 --o ./output --populate-persistent-peers --starting-ip-address 192.168.10.2
	`,
	RunE: testnetFiles,
}

func testnetFiles(cmd *cobra.Command, args []string) error {
	config := cfg.DefaultConfig()
	genVals := make([]types.GenesisValidator, nValidators)
	for _, chain := range config.ChainConfigs {
		for i := 0; i < nValidators; i++ {
			nodeDirName := cmn.Fmt("%s%d", nodeDirPrefix, i)
			nodeDir := filepath.Join(outputDir, nodeDirName)
			chain.SetRoot(nodeDir)

			err := os.MkdirAll(filepath.Join(nodeDir, "config"), nodeDirPerm)
			if err != nil {
				_ = os.RemoveAll(outputDir)
				return err
			}

			initFilesWithConfig(config)

			pvFile := filepath.Join(nodeDir, chain.BaseConfig.PrivValidator)
			pv := privval.LoadFilePV(pvFile)
			genVals[i] = types.GenesisValidator{
				PubKey: pv.GetPubKey(),
				Power:  1,
				Name:   nodeDirName,
			}
		}

		for i := 0; i < nNonValidators; i++ {
			nodeDir := filepath.Join(outputDir, cmn.Fmt("%s%d", nodeDirPrefix, i+nValidators))
			chain.SetRoot(nodeDir)

			err := os.MkdirAll(filepath.Join(nodeDir, "config"), nodeDirPerm)
			if err != nil {
				_ = os.RemoveAll(outputDir)
				return err
			}

			initFilesWithConfig(config)
		}

		// Generate genesis doc from generated validators
		genDoc := &types.GenesisDoc{
			GenesisTime: time.Now(),
			ChainID:     "chain-" + cmn.RandStr(6),
			Validators:  genVals,
		}

		// Write genesis file.
		for i := 0; i < nValidators+nNonValidators; i++ {
			nodeDir := filepath.Join(outputDir, cmn.Fmt("%s%d", nodeDirPrefix, i))
			if err := genDoc.SaveAs(filepath.Join(nodeDir, chain.BaseConfig.Genesis)); err != nil {
				_ = os.RemoveAll(outputDir)
				return err
			}
		}

		if populatePersistentPeers {
			err := populatePersistentPeersInConfigAndWriteIt(config)
			if err != nil {
				_ = os.RemoveAll(outputDir)
				return err
			}
		}
	}
	fmt.Printf("Successfully initialized %v node directories\n", nValidators+nNonValidators)
	return nil
}

func hostnameOrIP(i int) string {
	if startingIPAddress != "" {
		ip := net.ParseIP(startingIPAddress)
		ip = ip.To4()
		if ip == nil {
			fmt.Printf("%v: non ipv4 address\n", startingIPAddress)
			os.Exit(1)
		}

		for j := 0; j < i; j++ {
			ip[3]++
		}
		return ip.String()
	}

	return fmt.Sprintf("%s%d", hostnamePrefix, i)
}

func populatePersistentPeersInConfigAndWriteIt(config *cfg.Config) error {
	persistentPeers := make([]string, nValidators+nNonValidators)
	for _, chain := range config.ChainConfigs {
		for i := 0; i < nValidators+nNonValidators; i++ {
			nodeDir := filepath.Join(outputDir, cmn.Fmt("%s%d", nodeDirPrefix, i))
			chain.SetRoot(nodeDir)
			nodeKey, err := p2p.LoadNodeKey(chain.NodeKeyFile())
			if err != nil {
				return err
			}
			persistentPeers[i] = p2p.IDAddressString(nodeKey.ID(), fmt.Sprintf("%s:%d", hostnameOrIP(i), p2pPort))
		}
		persistentPeersList := strings.Join(persistentPeers, ",")

		for i := 0; i < nValidators+nNonValidators; i++ {
			nodeDir := filepath.Join(outputDir, cmn.Fmt("%s%d", nodeDirPrefix, i))
			chain.SetRoot(nodeDir)
			chain.P2P.PersistentPeers = persistentPeersList
			chain.P2P.AddrBookStrict = false

			// overwrite default config
			cfg.WriteConfigFile(filepath.Join(nodeDir, "config", "config.toml"), config)
		}
	}
	return nil
}