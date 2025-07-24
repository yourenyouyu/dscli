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
	// 禁用默认的 completion 命令
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	
	// 设置自定义帮助模板
	rootCmd.SetHelpTemplate(`{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}

{{end}}{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}`)
	
	// 设置自定义使用模板
	rootCmd.SetUsageTemplate(`用法:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

别名:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

示例:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

可用命令:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

选项:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

全局选项:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

其他帮助主题:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

使用 "{{.CommandPath}} [command] --help" 获取命令的更多信息。{{end}}
`)
	
	// 自定义帮助标志的描述
	rootCmd.PersistentFlags().BoolP("help", "h", false, "显示 dscli 的帮助信息")
	
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
}
