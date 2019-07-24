package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	cell "github.com/teragrid/dgrid/cell"
)

// AddCellFlags exposes some common configuration options on the command-line
// These are exposed for convenience of commands embedding a dgrid cell
func AddCellFlags(cmd *cobra.Command) {
	// bind flags
	cmd.Flags().String("hostname", config.Hostname, "Cell Name")

	// bind flags
	cmd.Flags().String("consensus", config.Consensus, "Consensus Protocol")

	// priv val flags
	cmd.Flags().String("validator_laddr", config.ValidatorListenAddr, "Socket address to listen on for connections from external validator process")

	// cell flags
	cmd.Flags().Bool("fast_sync", config.FastSync, "Fast blockchain syncing")

	// asura flags
	cmd.Flags().String("proxy_app", config.ProxyApp, "Proxy app address, or 'nilapp' or 'kvstore' for local testing.")
	cmd.Flags().String("asura", config.Asura, "Specify asura transport (socket | grpc)")

	// rpc flags
	cmd.Flags().String("rpc.laddr", config.RPC.ListenAddress, "RPC listen address. Port required")
	cmd.Flags().String("rpc.grpc_laddr", config.RPC.GRPCListenAddress, "GRPC listen address (BroadcastTx only). Port required")
	cmd.Flags().Bool("rpc.unsafe", config.RPC.Unsafe, "Enabled unsafe rpc methods")

	// p2p flags
	cmd.Flags().String("p2p.laddr", config.P2P.ListenAddress, "Cell listen address. (0.0.0.0:0 means any interface, any port)")
	cmd.Flags().String("p2p.seeds", config.P2P.Seeds, "Comma-delimited ID@host:port seed cells")
	cmd.Flags().String("p2p.persistent_peers", config.P2P.PersistentPeers, "Comma-delimited ID@host:port persistent peers")
	cmd.Flags().Bool("p2p.skip_upnp", config.P2P.SkipUPNP, "Skip UPNP configuration")
	cmd.Flags().Bool("p2p.pex", config.P2P.PexReactor, "Enable/disable Peer-Exchange")
	cmd.Flags().Bool("p2p.seed_mode", config.P2P.SeedMode, "Enable/disable seed mode")
	cmd.Flags().String("p2p.private_peer_ids", config.P2P.PrivatePeerIDs, "Comma-delimited private peer IDs")

	// consensus flags
	cmd.Flags().Bool("consensus.create_empty_blocks", config.Consensus.CreateEmptyBlocks, "Set this to false to only produce blocks when there are txs or when the AppHash changes")
}

// NewRunCellCmd returns the command that allows the CLI to start a cell.
// It can be used with a custom Validator and in-process asura application.
func NewRunCellCmd(cellProvider cell.CellProvider) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cell",
		Short: "Run the dgrid cell",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create & start cell
			n, err := cellProvider(config, logger)
			if err != nil {
				return fmt.Errorf("Failed to create cell: %v", err)
			}

			if err := n.Start(); err != nil {
				return fmt.Errorf("Failed to start cell: %v", err)
			}
			logger.Info("Started cell", "cellInfo", n.Switch().CellInfo())

			// Trap signal, run forever.
			n.RunForever()

			return nil
		},
	}

	AddCellFlags(cmd)
	return cmd
}
