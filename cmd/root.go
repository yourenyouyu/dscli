package cmd

import (
	"github.com/spf13/cobra"
)

// rootCmd 代表没有子命令时调用的基础命令
var rootCmd = &cobra.Command{
	Use:   "dscli",
	Short: "dscli 是一个用于 dsserv 模块开发的脚手架工具",
	Long: `dscli 是一个用于创建和构建 dsserv 模块的 CLI 工具。
它提供类似 vue-cli 的项目脚手架和构建命令。`,
}

// Execute 将所有子命令添加到根命令并适当设置标志。
// 这由 main.main() 调用。对于 rootCmd 只需要执行一次。
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// 添加自定义的帮助命令
	helpCmd := &cobra.Command{
		Use:   "help [command]",
		Short: "查看任何命令的帮助信息",
		Long:  "查看任何命令的帮助信息。",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				rootCmd.Help()
			} else {
				subCmd, _, err := rootCmd.Find(args)
				if err != nil {
					cmd.Printf("未知命令 \"%s\"\n", args[0])
					return
				}
				subCmd.Help()
			}
		},
	}
	
	// 替换默认的帮助命令
	rootCmd.SetHelpCommand(helpCmd)
	
	// Cobra 也支持本地标志，只有在直接调用此操作时才会运行。
	rootCmd.Flags().BoolP("toggle", "t", false, "切换选项的帮助信息")
}
