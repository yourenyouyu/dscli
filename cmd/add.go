package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

var (
	addNonInteractive bool
)

// addCmd 代表 add 命令
var addCmd = &cobra.Command{
	Use:   "add [executable-name]",
	Short: "向项目添加新的可执行文件",
	Long: `向现有的 dsserv 项目添加新的可执行文件。
此命令将在 cmd 目录下创建新的子目录和 main.go 模板文件。
构建时会自动发现并构建 cmd 目录下的所有子目录。`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if !isValidProject() {
			fmt.Println("错误: 不在有效的 dsserv 项目目录中 (未找到 manifest.json)")
			return
		}

		var name string
		if len(args) > 0 {
			name = args[0]
		}

		if addNonInteractive {
			if name == "" {
				fmt.Println("错误: 非交互模式下需要可执行文件名称")
				return
			}
		} else {
			var err error
			name, err = promptForExecutableName(name)
			if err != nil {
				fmt.Printf("Error getting executable name: %v\n", err)
				return
			}
		}

		if err := createExecutableInCmd(name); err != nil {
			fmt.Printf("Error adding executable: %v\n", err)
			return
		}

		fmt.Printf("\n✅ Executable '%s' added successfully!\n", name)
		fmt.Printf("\nNext steps:\n")
		fmt.Printf("  1. 编辑 cmd/%s/main.go 实现您的逻辑\n", name)
		fmt.Printf("  2. 运行 'dscli build' 自动构建所有可执行文件\n")
	},
}

func init() {
	rootCmd.AddCommand(addCmd)

	// 添加标志
	addCmd.Flags().BoolVar(&addNonInteractive, "non-interactive", false, "非交互模式")
}

func promptForExecutableName(initialName string) (string, error) {
	var name string

	prompt := &survey.Input{
		Message: "可执行文件名称:",
		Default: initialName,
	}

	err := survey.AskOne(prompt, &name, survey.WithValidator(survey.Required))
	return name, err
}

func createExecutableInCmd(name string) error {
	// 创建cmd目录（如果不存在）
	if err := os.MkdirAll("cmd", 0755); err != nil {
		return fmt.Errorf("创建cmd目录失败: %w", err)
	}

	// 创建可执行文件的源码目录
	execDir := filepath.Join("cmd", name)
	if err := os.MkdirAll(execDir, 0755); err != nil {
		return fmt.Errorf("创建目录 %s 失败: %w", execDir, err)
	}

	// 检查是否已存在 main.go
	mainGoPath := filepath.Join(execDir, "main.go")
	if _, err := os.Stat(mainGoPath); err == nil {
		fmt.Printf("源码文件已存在: %s\n", mainGoPath)
		return nil
	}

	// 创建基础的 main.go 模板
	mainGoTemplate := fmt.Sprintf(`package main

import (
	"fmt"
	"log"
)

func main() {
	fmt.Printf("Starting %s...\n")
	// TODO: 在这里实现您的 %s 逻辑
	log.Println("%s is running")
}
`, name, name, name)

	if err := os.WriteFile(mainGoPath, []byte(mainGoTemplate), 0644); err != nil {
		return fmt.Errorf("创建文件 %s 失败: %w", mainGoPath, err)
	}

	fmt.Printf("创建了模板文件: %s\n", mainGoPath)
	return nil
}