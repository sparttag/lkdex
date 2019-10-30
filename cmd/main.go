package main

import (
	"os"
	"path/filepath"

	"github.com/lianxiangcloud/linkchain/libs/cli"
	cfg "github.com/lianxiangcloud/lkdex/config"
	nm "github.com/lianxiangcloud/lkdex/node"
)

func main() {
	rootCmd := RootCmd
	nodeFunc := nm.DefaultNewNode

	rootCmd.AddCommand(VersionCmd)
	// Create & start node
	rootCmd.AddCommand(NewRunNodeCmd(nodeFunc))

	cmd := cli.PrepareBaseCmd(rootCmd, "TM", os.ExpandEnv(filepath.Join("$HOME", cfg.DefaultDexDir)))
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
