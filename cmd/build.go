package cmd

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type BuildTarget struct {
	OS   string
	Arch string
}

// AssetConfig 资源配置，支持指定源路径和输出路径
type AssetConfig struct {
	Source string `json:"source"` // 源路径
	Output string `json:"output"` // 输出路径
}

type BuildConfig struct {
	Assets    interface{} `json:"assets"`     // 需要打包的资源文件/目录，支持字符串数组或对象数组
	Excludes  []string    `json:"excludes"`   // 排除的文件/目录
	OutputDir string      `json:"output_dir"` // 输出目录
}

var (
	buildTargets = []BuildTarget{
		{"windows", "386"},
		{"windows", "amd64"},
		{"windows", "arm64"},
		{"darwin", "amd64"},
		{"darwin", "arm64"},
		{"linux", "386"},
		{"linux", "amd64"},
		{"linux", "arm64"},
	}
	targetFlag  string
	buildConfig *BuildConfig
)

// buildCmd 代表 build 命令
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "为多个平台构建 dsserv 模块",
	Long: `为多个平台和架构构建 dsserv 模块。
此命令将为 Windows、macOS 和 Linux 的不同架构编译项目，
更新 manifest.json 文件的构建信息，并创建特定平台的 tar.gz 包。`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := buildProject(); err != nil {
			fmt.Printf("构建失败: %v\n", err)
			return
		}
		fmt.Println("\n✅ 构建完成!")
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.Flags().StringVarP(&targetFlag, "target", "t", "", "指定目标平台和架构 (格式: os/arch，如 linux/amd64) 或 'all' 编译所有平台")
}

func buildProject() error {
	// 检查是否在有效的项目目录中
	if !isValidProject() {
		return fmt.Errorf("不在有效的 dsserv 项目目录中 (未找到 manifest.json)")
	}

	// 加载构建配置
	if err := loadBuildConfig(); err != nil {
		fmt.Printf("Warning: 无法加载构建配置: %v\n", err)
		// 使用默认配置
		buildConfig = &BuildConfig{
			Assets:    []interface{}{},
			Excludes:  []string{},
			OutputDir: "dist",
		}
	}

	// 读取当前清单
	manifest, err := readManifest()
	if err != nil {
		return fmt.Errorf("读取 manifest.json 失败: %w", err)
	}

	projectName := manifest["name"].(string)
	fmt.Printf("正在构建项目: %s\n", projectName)

	// 确定要构建的目标
	targets, err := getTargetsToBuild()
	if err != nil {
		return fmt.Errorf("获取构建目标失败: %w", err)
	}

	// 确定输出目录
	distDir := buildConfig.OutputDir
	if distDir == "" {
		distDir = "dist"
	}
	if err := os.RemoveAll(distDir); err != nil {
		return fmt.Errorf("清理输出目录失败: %w", err)
	}
	if err := os.MkdirAll(distDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	// 为每个目标平台构建
	for _, target := range targets {
		fmt.Printf("正在为 %s/%s 构建...\n", target.OS, target.Arch)
		if err := buildForTarget(projectName, target, distDir); err != nil {
			fmt.Printf("警告: %s/%s 构建失败: %v\n", target.OS, target.Arch, err)
			continue
		}
	}

	fmt.Println("\n构建摘要:")
	files, _ := filepath.Glob(filepath.Join(distDir, "*.tar.gz"))
	for _, file := range files {
		info, _ := os.Stat(file)
		fmt.Printf("  %s (%.2f MB)\n", filepath.Base(file), float64(info.Size())/1024/1024)
	}

	return nil
}

func buildForTarget(projectName string, target BuildTarget, distDir string) error {
	// 设置交叉编译的环境变量
	env := os.Environ()
	env = append(env, fmt.Sprintf("GOOS=%s", target.OS))
	env = append(env, fmt.Sprintf("GOARCH=%s", target.Arch))
	env = append(env, "CGO_ENABLED=0")

	buildTime := time.Now().Format(time.RFC3339)
	var builtBinaries []string

	// 1. 构建主程序（根目录的main.go）
	if _, err := os.Stat("main.go"); err == nil {
		binaryName := projectName
		if target.OS == "windows" {
			binaryName += ".exe"
		}

		ldflags := fmt.Sprintf("-ldflags=-X main.buildDate=%s", buildTime)
		cmd := exec.Command("go", "build", ldflags, "-o", filepath.Join("bin", binaryName), ".")
		cmd.Env = env
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Printf("❌ 构建主程序失败: %v\n", err)
		} else {
			builtBinaries = append(builtBinaries, binaryName)
			fmt.Printf("✅ 构建完成: %s (主程序)\n", projectName)
		}
	}

	// 2. 自动发现并构建cmd目录下的子目录
	cmdExecutables, err := discoverCmdExecutables()
	if err != nil {
		fmt.Printf("⚠️  扫描cmd目录失败: %v\n", err)
	} else {
		for _, execName := range cmdExecutables {
			binaryName := execName
			if target.OS == "windows" {
				binaryName += ".exe"
			}

			sourcePath := "./" + filepath.Join("cmd", execName)
			ldflags := fmt.Sprintf("-ldflags=-X main.buildDate=%s", buildTime)
			cmd := exec.Command("go", "build", ldflags, "-o", filepath.Join("bin", binaryName), sourcePath)
			cmd.Env = env
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			if err := cmd.Run(); err != nil {
				fmt.Printf("❌ 构建 %s 失败: %v\n", execName, err)
				continue
			}

			builtBinaries = append(builtBinaries, binaryName)
			fmt.Printf("✅ 构建完成: %s\n", execName)
		}
	}

	if len(builtBinaries) == 0 {
		return fmt.Errorf("没有成功构建任何可执行文件")
	}

	// 为此目标更新清单
	if err := updateManifestForTarget(target, buildTime); err != nil {
		return fmt.Errorf("更新清单失败: %w", err)
	}

	// 创建包
	packageName := fmt.Sprintf("%s_%s_%s.tar.gz", projectName, target.OS, target.Arch)
	packagePath := filepath.Join(distDir, packageName)

	if err := createPackage(packagePath, builtBinaries); err != nil {
		return fmt.Errorf("failed to create package: %w", err)
	}

	// 清理二进制文件
	for _, binaryName := range builtBinaries {
		os.Remove(filepath.Join("bin", binaryName))
	}

	return nil
}

func updateManifestForTarget(target BuildTarget, buildTime string) error {
	manifest, err := readManifest()
	if err != nil {
		return err
	}

	// 更新构建信息
	manifest["build_date"] = buildTime
	manifest["os"] = target.OS
	manifest["arch"] = target.Arch

	// 写入更新的清单
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile("manifest.json", data, 0644)
}

// parseAssets 解析assets配置，支持字符串数组和对象数组两种格式
func parseAssets(assets interface{}) ([]AssetConfig, error) {
	if assets == nil {
		return []AssetConfig{}, nil
	}

	// 尝试解析为字符串数组
	if assetsArray, ok := assets.([]interface{}); ok {
		var result []AssetConfig
		for _, asset := range assetsArray {
			if assetStr, ok := asset.(string); ok {
				// 字符串格式：直接使用相同的源路径和输出路径
				result = append(result, AssetConfig{
					Source: assetStr,
					Output: assetStr,
				})
			} else if assetMap, ok := asset.(map[string]interface{}); ok {
				// 对象格式：解析source和output字段
				source, sourceOk := assetMap["source"].(string)
				output, outputOk := assetMap["output"].(string)
				if sourceOk && outputOk {
					result = append(result, AssetConfig{
						Source: source,
						Output: output,
					})
				} else {
					return nil, fmt.Errorf("asset对象必须包含source和output字段")
				}
			} else {
				return nil, fmt.Errorf("不支持的asset格式")
			}
		}
		return result, nil
	}

	return nil, fmt.Errorf("assets必须是数组格式")
}

func createPackage(packagePath string, builtBinaries []string) error {
	file, err := os.Create(packagePath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// 先添加bin目录
	if len(builtBinaries) > 0 {
		binDirHeader := &tar.Header{
			Name:     "bin/",
			Mode:     0755,
			Typeflag: tar.TypeDir,
		}
		if err := tarWriter.WriteHeader(binDirHeader); err != nil {
			fmt.Printf("⚠️  无法创建bin目录: %v\n", err)
		}
	}

	// 添加所有二进制文件到bin目录
	for _, binaryName := range builtBinaries {
		binPath := filepath.Join("bin", binaryName)
		if err := addFileToTar(tarWriter, binPath, binPath); err != nil {
			fmt.Printf("⚠️  无法添加二进制文件 %s: %v\n", binaryName, err)
			continue
		}
	}

	// 添加清单
	if err := addFileToTar(tarWriter, "manifest.json", "manifest.json"); err != nil {
		return err
	}

	// 添加配置文件中指定的资源
	assets, err := parseAssets(buildConfig.Assets)
	if err != nil {
		fmt.Printf("⚠️  解析assets配置失败: %v\n", err)
	} else {
		for _, asset := range assets {
			// 检查是否被排除
			if isExcluded(asset.Source) {
				fmt.Printf("ℹ️  跳过被排除的资源: %s\n", asset.Source)
				continue
			}

			info, err := os.Stat(asset.Source)
			if err != nil {
				fmt.Printf("⚠️  资源文件不存在: %s\n", asset.Source)
				continue
			}

			if info.IsDir() {
				if err := addDirToTar(tarWriter, asset.Source, asset.Output); err != nil {
					fmt.Printf("⚠️  无法添加目录 %s: %v\n", asset.Source, err)
				}
			} else {
				if err := addFileToTar(tarWriter, asset.Source, asset.Output); err != nil {
					fmt.Printf("⚠️  无法添加文件 %s: %v\n", asset.Source, err)
				}
			}
		}
	}

	return nil
}

func addFileToTar(tarWriter *tar.Writer, filePath, tarPath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}
	header.Name = tarPath

	// 为二进制文件设置可执行权限
	if strings.HasSuffix(tarPath, ".exe") || (!strings.Contains(tarPath, ".") && tarPath != "manifest.json") {
		header.Mode = 0755
	}

	if err := tarWriter.WriteHeader(header); err != nil {
		return err
	}

	_, err = io.Copy(tarWriter, file)
	return err
}

func addDirToTar(tarWriter *tar.Writer, dirPath, tarDirPath string) error {
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// 检查文件是否被排除（检查完整路径和文件名）
		if isExcluded(path) {
			fmt.Printf("ℹ️  跳过被排除的文件: %s\n", path)
			return nil // 跳过被排除的文件
		}

		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			return err
		}

		tarPath := filepath.Join(tarDirPath, relPath)
		// 为 tar 转换为正斜杠
		tarPath = strings.ReplaceAll(tarPath, "\\", "/")

		return addFileToTar(tarWriter, path, tarPath)
	})
}

func readManifest() (map[string]interface{}, error) {
	data, err := os.ReadFile("manifest.json")
	if err != nil {
		return nil, err
	}

	var manifest map[string]interface{}
	err = json.Unmarshal(data, &manifest)
	return manifest, err
}

func getTargetsToBuild() ([]BuildTarget, error) {
	if targetFlag == "" {
		// 默认构建当前平台
		return []BuildTarget{{runtime.GOOS, runtime.GOARCH}}, nil
	}

	if targetFlag == "all" {
		// 构建所有平台
		return buildTargets, nil
	}

	// 解析指定的目标
	parts := strings.Split(targetFlag, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("无效的目标格式: %s，应为 os/arch 格式", targetFlag)
	}

	os, arch := parts[0], parts[1]
	// 验证目标是否支持
	for _, target := range buildTargets {
		if target.OS == os && target.Arch == arch {
			return []BuildTarget{{os, arch}}, nil
		}
	}

	return nil, fmt.Errorf("不支持的目标平台: %s/%s", os, arch)
}

func isValidProject() bool {
	_, err := os.Stat("manifest.json")
	return err == nil
}

func loadBuildConfig() error {
	// 尝试加载 .dscli.json 配置文件
	data, err := os.ReadFile(".dscli.json")
	if err != nil {
		// 如果文件不存在，使用默认配置
		buildConfig = &BuildConfig{
			Assets:    []interface{}{},
			Excludes:  []string{},
			OutputDir: "dist",
		}
		return nil
	}

	err = json.Unmarshal(data, &buildConfig)
	if err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 设置默认值
	if buildConfig.OutputDir == "" {
		buildConfig.OutputDir = "dist"
	}

	return nil
}

func isExcluded(path string) bool {
	for _, exclude := range buildConfig.Excludes {
		if matched, _ := filepath.Match(exclude, path); matched {
			return true
		}
		// 也检查完整路径匹配
		if exclude == path {
			return true
		}
		// 检查基础文件名匹配
		if matched, _ := filepath.Match(exclude, filepath.Base(path)); matched {
			return true
		}
	}
	return false
}

// discoverCmdExecutables 自动发现cmd目录下的子目录，每个子目录代表一个可执行文件
func discoverCmdExecutables() ([]string, error) {
	cmdDir := "cmd"
	if _, err := os.Stat(cmdDir); os.IsNotExist(err) {
		return []string{}, nil // cmd目录不存在，返回空列表
	}

	entries, err := os.ReadDir(cmdDir)
	if err != nil {
		return nil, fmt.Errorf("读取cmd目录失败: %w", err)
	}

	var executables []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue // 跳过文件，只处理目录
		}

		execName := entry.Name()
		// 检查子目录是否包含main.go文件
		mainGoPath := filepath.Join(cmdDir, execName, "main.go")
		if _, err := os.Stat(mainGoPath); err == nil {
			executables = append(executables, execName)
		}
	}

	return executables, nil
}
