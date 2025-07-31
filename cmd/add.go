package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
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

		if name == "" {
			fmt.Println("错误: 需要可执行文件名称")
			return
		}

		if err := createExecutableInCmd(name); err != nil {
			fmt.Printf("添加可执行文件时出错: %v\n", err)
			return
		}

		// 更新manifest.json的executable字段
		if err := updateManifestExecutable(name); err != nil {
			fmt.Printf("更新manifest.json时出错: %v\n", err)
			return
		}

		fmt.Printf("\n✅ 可执行文件 '%s' 添加成功!\n", name)
		fmt.Printf("\n下一步操作:\n")
		fmt.Printf("  1. 编辑 cmd/%s/main.go 实现您的逻辑\n", name)
		fmt.Printf("  2. 运行 'dscli build' 自动构建所有可执行文件\n")
	},
}

func init() {
	rootCmd.AddCommand(addCmd)

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

// updateManifestExecutable 更新manifest.json的executable字段
func updateManifestExecutable(name string) error {
	// 读取现有的manifest.json
	data, err := os.ReadFile("manifest.json")
	if err != nil {
		return fmt.Errorf("读取manifest.json失败: %w", err)
	}

	var manifest map[string]interface{}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return fmt.Errorf("解析manifest.json失败: %w", err)
	}

	// 获取现有的executable数组
	executables, ok := manifest["executable"].([]interface{})
	if !ok {
		executables = []interface{}{}
	}

	// 根据manifest.json中的OS字段确定可执行文件名
	executableName := name
	if osValue, ok := manifest["os"].(string); ok && osValue == "windows" {
		executableName += ".exe"
	}

	// 检查新的可执行文件是否已存在
	newExecutable := fmt.Sprintf("./bin/%s", executableName)
	for _, exec := range executables {
		if execStr, ok := exec.(string); ok && execStr == newExecutable {
			fmt.Printf("可执行文件 %s 已存在于manifest.json中\n", newExecutable)
			return nil
		}
	}

	// 添加新的可执行文件
	executables = append(executables, newExecutable)
	manifest["executable"] = executables

	// 写回manifest.json
	updatedData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化manifest.json失败: %w", err)
	}

	if err := os.WriteFile("manifest.json", updatedData, 0644); err != nil {
		return fmt.Errorf("写入manifest.json失败: %w", err)
	}

	fmt.Printf("已更新manifest.json，添加可执行文件: %s\n", newExecutable)
	return nil
}
