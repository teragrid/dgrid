package rpctest

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/teragrid/dgrid/pkg/log"

	asura "github.com/teragrid/dgrid/asura/types"

	cfg "github.com/teragrid/dgrid/core/config"
	cmn "github.com/teragrid/dgrid/pkg/common"
	nm "github.com/teragrid/dgrid/node"
	"github.com/teragrid/dgrid/core/blockchain/p2p"
	"github.com/teragrid/dgridcore/consensus/validator"
	"github.com/teragrid/dgrid/proxy"
	ctypes "github.com/teragrid/dgrid/rpc/core/types"
	core_grpc "github.com/teragrid/dgrid/rpc/grpc"
	rpcclient "github.com/teragrid/dgrid/rpc/lib/client"
)

var globalConfig *cfg.Config

func waitForRPC() {
	laddr := GetConfig().RPC.ListenAddress
	client := rpcclient.NewJSONRPCClient(laddr)
	ctypes.RegisterAmino(client.Codec())
	result := new(ctypes.ResultStatus)
	for {
		_, err := client.Call("status", map[string]interface{}{}, result)
		if err == nil {
			return
		} else {
			fmt.Println("error", err)
			time.Sleep(time.Millisecond)
		}
	}
}

func waitForGRPC() {
	client := GetGRPCClient()
	for {
		_, err := client.Ping(context.Background(), &core_grpc.RequestPing{})
		if err == nil {
			return
		}
	}
}

// f**ing long, but unique for each test
func makePathname() string {
	// get path
	p, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	// fmt.Println(p)
	sep := string(filepath.Separator)
	return strings.Replace(p, sep, "_", -1)
}

func randPort() int {
	port, err := cmn.GetFreePort()
	if err != nil {
		panic(err)
	}
	return port
}

func makeAddrs() (string, string, string) {
	return fmt.Sprintf("tcp://0.0.0.0:%d", randPort()),
		fmt.Sprintf("tcp://0.0.0.0:%d", randPort()),
		fmt.Sprintf("tcp://0.0.0.0:%d", randPort())
}

// GetConfig returns a config for the test cases as a singleton
func GetConfig() *cfg.Config {
	if globalConfig == nil {
		pathname := makePathname()
		globalConfig = cfg.ResetTestRoot(pathname)

		// and we use random ports to run in parallel
		tm, rpc, grpc := makeAddrs()
		globalConfig.P2P.ListenAddress = tm
		globalConfig.RPC.ListenAddress = rpc
		globalConfig.RPC.CORSAllowedOrigins = []string{"https://teragrid.network/"}
		globalConfig.RPC.GRPCListenAddress = grpc
		globalConfig.TxIndex.IndexTags = "app.creator,tx.height" // see kvstore application
	}
	return globalConfig
}

func GetGRPCClient() core_grpc.BroadcastAPIClient {
	grpcAddr := globalConfig.RPC.GRPCListenAddress
	return core_grpc.StartGRPCClient(grpcAddr)
}

// StartTeragrid starts a test teragrid server in a go routine and returns when it is initialized
func StartTeragrid(app asura.Application) *nm.Node {
	node := NewTeragrid(app)
	err := node.Start()
	if err != nil {
		panic(err)
	}

	// wait for rpc
	waitForRPC()
	waitForGRPC()

	fmt.Println("Dgrid running!")

	return node
}

// StopTeragrid stops a test teragrid server, waits until it's stopped and
// cleans up test/config files.
func StopTeragrid(node *nm.Node) {
	node.Stop()
	node.Wait()
	os.RemoveAll(node.Config().RootDir)
}

// NewTeragrid creates a new teragrid server and sleeps forever
func NewTeragrid(app asura.Application) *nm.Node {
	// Create & start node
	config := GetConfig()
	logger := log.NewTGLogger(log.NewSyncWriter(os.Stdout))
	logger = log.NewFilter(logger, log.AllowError())
	pvKeyFile := config.PrivValidatorKeyFile()
	pvKeyStateFile := config.PrivValidatorStateFile()
	pv := validator.LoadOrGenFilePV(pvKeyFile, pvKeyStateFile)
	papp := proxy.NewLocalClientCreator(app)
	nodeKey, err := p2p.LoadOrGenNodeKey(config.NodeKeyFile())
	if err != nil {
		panic(err)
	}
	node, err := nm.NewNode(config, pv, nodeKey, papp,
		nm.DefaultGenesisDocProviderFunc(config),
		nm.DefaultDBProvider,
		nm.DefaultMetricsProvider(config.Instrumentation),
		logger)
	if err != nil {
		panic(err)
	}
	return node
}
