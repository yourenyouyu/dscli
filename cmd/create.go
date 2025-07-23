package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

type ProjectConfig struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Author      string `json:"author"`
}

type ManifestData struct {
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	Version         string   `json:"version"`
	ManifestVersion int      `json:"manifest_version"`
	Author          string   `json:"author"`
	BuildDate       string   `json:"build_date"`
	OS              string   `json:"os"`
	Arch            string   `json:"arch"`
	LogDir          string   `json:"log_dir"`
	Executable      []string `json:"executable"`
}

var (
	description    string
	version        string
	author         string
	nonInteractive bool
)

// createCmd 代表 create 命令
var createCmd = &cobra.Command{
	Use:   "create [project-name]",
	Short: "创建一个新的 dsserv 模块项目",
	Long: `创建一个具有指定名称的新 dsserv 模块项目。
此命令将创建一个包含项目结构和文件的新目录。`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var projectName string
		if len(args) > 0 {
			projectName = args[0]
		}

		var config *ProjectConfig
		var err error

		if nonInteractive {
			config = &ProjectConfig{
				Name:        projectName,
				Description: description,
				Version:     version,
				Author:      author,
			}
			// 如果未提供则设置默认值
			if config.Name == "" {
				fmt.Println("错误: 非交互模式下需要项目名称")
				return
			}
			if config.Description == "" {
				config.Description = "A dsserv module"
			}
			if config.Version == "" {
				config.Version = "1.0.0"
			}
			if config.Author == "" {
				config.Author = "DataShell Team"
			}
		} else {
			config, err = promptForProjectInfo(projectName)
			if err != nil {
				fmt.Printf("Error getting project information: %v\n", err)
				return
			}
		}

		if err := createProject(config); err != nil {
			fmt.Printf("Error creating project: %v\n", err)
			return
		}

		fmt.Printf("\n✅ Project '%s' created successfully!\n", config.Name)
		fmt.Printf("\nNext steps:\n")
		fmt.Printf("  cd %s\n", config.Name)
		fmt.Printf("  go mod tidy\n")
		fmt.Printf("  dscli build\n")
	},
}

func init() {
	rootCmd.AddCommand(createCmd)

	// 添加标志
	createCmd.Flags().StringVarP(&description, "description", "d", "", "项目描述")
	createCmd.Flags().StringVarP(&version, "version", "v", "", "项目版本")
	createCmd.Flags().StringVarP(&author, "author", "a", "", "项目作者")
	createCmd.Flags().BoolVar(&nonInteractive, "non-interactive", false, "以非交互模式运行")
}

func promptForProjectInfo(initialName string) (*ProjectConfig, error) {
	config := &ProjectConfig{}

	questions := []*survey.Question{
		{
			Name: "name",
			Prompt: &survey.Input{
				Message: "项目名称:",
				Default: initialName,
			},
			Validate: survey.Required,
		},
		{
			Name: "description",
			Prompt: &survey.Input{
				Message: "项目描述:",
				Default: "一个 dsserv 模块",
			},
			Validate: survey.Required,
		},
		{
			Name: "version",
			Prompt: &survey.Input{
				Message: "版本:",
				Default: "1.0.0",
			},
			Validate: survey.Required,
		},
		{
			Name: "author",
			Prompt: &survey.Input{
				Message: "作者:",
				Default: "DataShell Team",
			},
		},
	}

	err := survey.Ask(questions, config)
	return config, err
}

func createProject(config *ProjectConfig) error {
	// 创建项目目录
	projectDir := config.Name
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return fmt.Errorf("创建项目目录失败: %w", err)
	}

	// 创建子目录
	dirs := []string{
		filepath.Join(projectDir, "cmd"),
		filepath.Join(projectDir, "internal"),
		filepath.Join(projectDir, "pkg"),
		filepath.Join(projectDir, "logs"),
		filepath.Join(projectDir, "bin"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建目录 %s 失败: %w", dir, err)
		}
	}

	// 创建文件
	if err := createGoMod(projectDir, config); err != nil {
		return err
	}

	if err := createMainGo(projectDir, config); err != nil {
		return err
	}

	if err := createManifest(projectDir, config); err != nil {
		return err
	}

	// 注意：不再自动创建configs目录和配置文件
	// 用户可以根据需要手动创建配置文件

	if err := createReadme(projectDir, config); err != nil {
		return err
	}

	if err := createGitignore(projectDir); err != nil {
		return err
	}

	if err := createDscliConfig(projectDir); err != nil {
		return err
	}

	return nil
}

func createGoMod(projectDir string, config *ProjectConfig) error {
	content := fmt.Sprintf(`module %s

go 1.21

require (
	github.com/spf13/cobra v1.8.0
	github.com/spf13/viper v1.18.2
)
`, config.Name)

	return writeFile(filepath.Join(projectDir, "go.mod"), content)
}

func createMainGo(projectDir string, config *ProjectConfig) error {
	tmpl := `package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	configFile string
	version    = "{{.Version}}"
	buildDate  = "unknown"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "{{.Name}}",
		Short: "{{.Description}}",
		Version: version,
		Run: func(cmd *cobra.Command, args []string) {
			run()
		},
	}

	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "./config.json", "config file path")

	if err := rootCmd.Execute(); err != nil {
			log.Fatal(err)
		}
}

func run() {
	// 加载配置
	if configFile != "" {
		viper.SetConfigFile(configFile)
		if err := viper.ReadInConfig(); err != nil {
			log.Printf("警告: 无法读取配置文件: %v", err)
		}
	}

	fmt.Printf("启动 {{.Name}} v%s (构建时间: %s)\n", version, buildDate)
	fmt.Println("{{.Description}}")

	// 设置优雅关闭
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// 主服务循环
	go func() {
		for {
			// 在这里编写您的主要服务逻辑
			fmt.Println("服务正在运行...")
			time.Sleep(10 * time.Second)
		}
	}()

	// 等待关闭信号
	<-c
	fmt.Println("\n正在优雅关闭...")
	// 清理逻辑在这里
	fmt.Println("服务已停止。")
}
`

	t, err := template.New("main").Parse(tmpl)
	if err != nil {
		return err
	}

	file, err := os.Create(filepath.Join(projectDir, "main.go"))
	if err != nil {
		return err
	}
	defer file.Close()

	return t.Execute(file, config)
}

func createManifest(projectDir string, config *ProjectConfig) error {
	manifest := ManifestData{
		Name:            config.Name,
		Description:     config.Description,
		Version:         config.Version,
		ManifestVersion: 1,
		Author:          config.Author,
		BuildDate:       time.Now().Format(time.RFC3339),
		OS:              "linux",
		Arch:            "amd64",
		LogDir:          "./logs",
		Executable:      []string{fmt.Sprintf("./bin/%s", config.Name)},
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}

	return writeFile(filepath.Join(projectDir, "manifest.json"), string(data))
}

func createReadme(projectDir string, config *ProjectConfig) error {
	tmpl := `# {{.Name}}

{{.Description}}

## Installation

` + "```bash\n" +
		`go mod tidy
` + "```\n\n" +
		`## Usage

` + "```bash\n" +
		`# Run the application
go run main.go

# Or build and run
go build -o bin/{{.Name}}
./bin/{{.Name}}
` + "```\n\n" +
		`## Build

Use dscli to build the module:

` + "```bash\n" +
		`dscli build
` + "```\n\n" +
		`This will create platform-specific binaries and packages.
		
## Author

{{.Author}}

## Version

{{.Version}}
`

	t, err := template.New("readme").Parse(tmpl)
	if err != nil {
		return err
	}

	file, err := os.Create(filepath.Join(projectDir, "README.md"))
	if err != nil {
		return err
	}
	defer file.Close()

	return t.Execute(file, config)
}

func createGitignore(projectDir string) error {
	content := `# Binaries for programs and plugins
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test binary, built with ` + "`go test -c`" + `
*.test

# Output of the go coverage tool, specifically when used with LiteIDE
*.out

# Dependency directories (remove the comment below to include it)
# vendor/

# Go workspace file
go.work

# Build output
bin/
dist/
*.zip
*.tar.gz

# Logs
logs/*.log

# IDE
.vscode/
.idea/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db
`

	return writeFile(filepath.Join(projectDir, ".gitignore"), content)
}

func createDscliConfig(projectDir string) error {
	dscliConfig := map[string]interface{}{
		"assets": []string{
			"config/",
			"templates/",
		},
		"excludes": []string{
			"*.log",
			"*.tmp",
			".git/",
		},
		"output_dir": "dist",
	}

	data, err := json.MarshalIndent(dscliConfig, "", "  ")
	if err != nil {
		return err
	}

	return writeFile(filepath.Join(projectDir, ".dscli.json"), string(data))
}

func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}
