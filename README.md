# dscli - dsserv模块脚手架工具

dscli是一个用于开发dsserv服务模块的脚手架工具，提供类似vue-cli的命令行体验，帮助开发者快速创建、构建和管理dsserv模块项目。

## 功能特性

- 🚀 **项目创建**: 交互式创建新的dsserv模块项目
- 🔨 **跨平台构建**: 支持Windows、macOS、Linux多平台编译
- 📦 **自动打包**: 生成包含二进制文件和配置的压缩包
- 📋 **清单管理**: 自动生成和更新manifest.json文件
- 🎯 **多架构支持**: 支持386、amd64、arm64架构

## 安装

### 从源码构建

```bash
git clone https://github.com/yourenyouyu/dscli.git
cd dscli
make build
sudo make install
```

### 使用Go安装

```bash
go install github.com/yourenyouyu/dscli@main
```

## 快速开始

### 1. 创建新项目

```bash
dscli create my-project
```

系统会提示您输入项目信息：
- 项目名称
- 项目描述  
- 版本号
- 作者信息

### 2. 进入项目目录

```bash
cd my-project
go mod tidy
```

### 3. 构建项目

```bash
dscli build
```

这将为所有支持的平台和架构生成二进制文件和压缩包。

## 命令参考

### `dscli create [project-name]`

创建新的dsserv模块项目。如果不提供项目名称，系统会提示输入。

**示例:**
```bash
dscli create my-awesome-module
```

### `dscli build`

构建当前项目，支持灵活的目标平台选择。

**构建模式:**
- **默认构建**: 构建当前平台和架构，二进制文件输出到 `bin/` 目录
- **指定平台构建**: 使用 `-t os/arch` 指定目标平台，二进制文件输出到 `bin/` 目录
- **全平台构建**: 使用 `-t all` 构建所有支持的平台，生成zip压缩包到 `dist/` 目录

**支持的平台:**
- Windows: 386, amd64, arm64
- macOS (Darwin): amd64, arm64
- Linux: 386, amd64, arm64

**示例:**
```bash
# 构建当前平台
dscli build

# 构建指定平台
dscli build -t linux/amd64

# 构建所有平台
dscli build -t all
```

**选项:**
- `-t, --target`: 指定构建目标，支持 `os/arch` 格式或 `all`

### `dscli version`

显示dscli工具的版本信息。

## 构建配置

### .dscli.json 配置文件

可以在项目根目录创建 `.dscli.json` 文件来自定义构建行为。如果没有此文件，将使用默认配置：所有构建都会生成ZIP压缩包并输出到 `dist` 目录。

```json
{
  "assets": [
    "README.md",
    "configs",
    "docs",
    "static"
  ],
  "excludes": [
    "*.log",
    "*.tmp",
    ".git",
    "node_modules"
  ],
  "output_dir": "dist"
}
```

#### 配置字段说明

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `assets` | array | `["config/", "templates/"]` | 打包时包含的额外文件或目录 |
| `excludes` | array | `["*.log", "*.tmp", ".git/"]` | 打包时排除的文件模式 |
| `output_dir` | string | `"dist"` | 输出目录，所有构建都会生成ZIP压缩包 |

#### 字段详细说明

**assets** - 资源文件列表
- 指定构建时需要包含到压缩包中的文件或目录
- 支持相对路径，相对于项目根目录
- 支持两种格式：
  - **字符串数组格式**：`["README.md", "configs", "docs"]`
  - **对象数组格式**：`[{"src": "configs", "dest": "config"}, {"src": "docs", "dest": "documentation"}]`
- 对象格式支持重命名：`src` 为源路径，`dest` 为目标路径
- 常见用途：包含配置文件、文档、静态资源等

**excludes** - 排除文件模式
- 指定打包时需要排除的文件或目录模式
- 支持通配符匹配（`*`、`?`、`**`等）
- 优先级高于 `assets`，即使在 `assets` 中指定的文件，如果匹配 `excludes` 模式也会被排除
- 示例：`["*.log", "*.tmp", ".git", "node_modules", "**/.DS_Store"]`

**output_dir** - 输出目录
- 所有构建都会生成ZIP压缩包到此目录
- 默认值为 `"dist"`
- 目录会自动创建（如果不存在）

#### 使用示例

**基础配置（项目创建时自动生成）:**
```json
{
  "assets": ["config/", "templates/"],
  "excludes": ["*.log", "*.tmp", ".git/"],
  "output_dir": "dist"
}
```

**生产环境配置:**
```json
{
  "assets": [
    "README.md",
    "configs",
    "docs",
    "static",
    "templates"
  ],
  "excludes": [
    "*.log",
    "*.tmp",
    "*.test",
    ".git",
    "node_modules",
    "**/.DS_Store",
    "**/.vscode"
  ],
  "output_dir": "release"
}
```

**使用对象数组格式的配置示例:**
```json
{
  "assets": [
    "README.md",
    {"src": "configs", "dest": "config"},
    {"src": "docs", "dest": "documentation"},
    {"src": "static/images", "dest": "assets/img"},
    "templates"
  ],
  "excludes": [
    "*.log",
    "*.tmp",
    ".git"
  ],
  "output_dir": "dist"
}
```

**注意事项:**
- 如果项目根目录没有 `.dscli.json` 文件，将使用默认配置
- 使用 `dscli create` 创建项目时会自动生成默认的 `.dscli.json` 文件
- 配置文件格式必须是有效的JSON，语法错误会导致使用默认配置
- 文件路径区分大小写，请确保路径正确

## 项目结构

使用`dscli create`创建的项目具有以下结构：

```
my-project/
├── main.go              # 主程序入口
├── go.mod               # Go模块文件
├── manifest.json        # 模块清单文件
├── README.md            # 项目说明
├── .gitignore          # Git忽略文件
├── .dscli.json         # 构建配置文件（可选，默认生成ZIP到dist目录）
├── cmd/                 # 命令行相关代码
├── internal/            # 内部包
├── pkg/                 # 公共包
├── logs/               # 日志目录
└── bin/                # 编译输出目录
```

#### manifest.json 结构

```json
{
  "name": "data-processor",
  "description": "数据处理模块",
  "version": "1.2.0",
  "manifest_version": 1,
  "author": "DataShell Team",
  "build_date": "2024-01-15T10:30:00Z",
  "os": "darwin",
  "arch": "amd64",
  "log_dir": "./logs",
  "executable": ["./bin/processor --config config.json"]
}
```

#### manifest.json 字段说明

| 字段 | 类型 | 必需 | 说明 |
|------|------|------|------|
| `name` | string | ✅ | 模块名称，必须与配置中的 name 一致 |
| `description` | string | ✅ | 模块功能描述 |
| `version` | string | ✅ | 模块版本号 |
| `manifest_version` | int | ✅ | manifest 文件格式版本，当前为 1 |
| `executable` | array | ✅ | 模块启动命令和参数列表 |
| `os` | string | ✅ | 模块支持的操作系统 |
| `arch` | string | ✅ | 模块支持的架构 |
| `log_dir` | string | ❌ | 日志文件存放路径（配置后可能被服务端采集用于问题排查）|
| `author` | string | ❌ | 模块作者或开发团队 |
| `build_date` | string | ❌ | 模块构建时间（ISO 8601 格式）|

> 💡 **提示**：manifest.json 文件确保了模块的标准化管理，Agent 会根据此文件中的 `executable` 字段启动模块进程。