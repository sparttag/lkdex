package main

import (
	"fmt"

	"github.com/lianxiangcloud/linkchain/types"
	nm "github.com/lianxiangcloud/lkdex/node"
	"github.com/spf13/cobra"
)

// AddWalletFlags exposes some common configuration options on the command-line
// These are exposed for convenience of commands embedding a node
func AddNodeFlags(cmd *cobra.Command) {
	// bind flags
	cmd.Flags().Bool("detach", config.BaseConfig.Detach, "Run as daemon")
	cmd.Flags().Int("max_concurrency", config.BaseConfig.MaxConcurrency, "Max number of threads to use for a parallel job")
	// cmd.Flags().String("pidfile", config.BaseConfig.Pidfile, "File path to write the daemon's PID to")
	cmd.Flags().String("log_level", config.BaseConfig.LogLevel, "0-4 or categories")
	cmd.Flags().String("home", config.BaseConfig.RootDir, "home")
	cmd.Flags().String("log_dir", config.BaseConfig.LogPath, "log_dir")
	cmd.Flags().Bool("test_net", config.BaseConfig.TestNet, "signparam will be set to 29154 if this flag is set")

	cmd.Flags().String("contract_addr", config.BaseConfig.ContractAddr, "dexcontract contract address")
	cmd.Flags().String("daemon.peer_rpc", config.Daemon.PeerRPC, "peer rpc url")
	cmd.Flags().String("daemon.peer_ws", config.Daemon.PeerWS, "peer ws url")
	cmd.Flags().String("wallet_daemon.peer_rpc", config.WalletDaemon.PeerRPC, "wallet rpc url")

	// rpc flags
	cmd.Flags().StringSlice("rpc.http_modules", config.RPC.HTTPModules, "API's offered over the HTTP-RPC interface")
	cmd.Flags().String("rpc.http_endpoint", config.RPC.HTTPEndpoint, "RPC listen address. Port required")
	cmd.Flags().String("rpc.ws_endpoint", config.RPC.WSEndpoint, " WS-RPC server listening address. Port required")
	cmd.Flags().StringSlice("rpc.ws_modules", config.RPC.WSModules, "API's offered over the WS-RPC interface")
	cmd.Flags().Bool("rpc.ws_expose_all", config.RPC.WSExposeAll, "Enable the WS-RPC server to expose all APIs")
	cmd.Flags().String("rpc.ipc_endpoint", config.RPC.IpcEndpoint, "Filename for IPC socket/pipe within the datadir (explicit paths escape it)")

}

// NewRunNodeCmd returns the command that allows the CLI to start a node.
// It can be used with a custom PrivValidator and in-process ABCI application.
func NewRunNodeCmd(nodeProvider nm.NodeProvider) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "node",
		Short: "Run the dex node",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := config.BaseConfig.SavePid()
			if err != nil {
				panic(err)
			}
			logger.Info("NewRunNodeCmd", "base", config.BaseConfig, "daemon", config.Daemon, "wallet", config.WalletDaemon, "rpc", config.RPC, "log", config.Log, "dexContractAddr", config.ContractAddr)
			fmt.Printf("conf:%v\n", *config)

			types.InitSignParam(config.TestNet)

			// Create & start node
			n, err := nodeProvider(config, logger.With("module", "node"))
			if err != nil {
				return fmt.Errorf("Failed to create node: %v", err)
			}

			if err := n.Start(); err != nil {
				return fmt.Errorf("Failed to start node: %v", err)
			}
			logger.Info("Started node", "nodeInfo", "n.Switch().NodeInfo()")

			// Trap signal, run forever.
			n.RunForever()

			return nil
		},
	}

	AddNodeFlags(cmd)
	return cmd
}
