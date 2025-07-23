package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	// 这些将由构建标志设置
	Version = "0.0.2"
)

// versionCmd 代表 version 命令
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "打印 dscli 的版本号",
	Long:  `打印 dscli 的版本号和其他构建信息。`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("dscli version %s\n", Version)
		fmt.Printf("Go version: %s\n", runtime.Version())
		fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
