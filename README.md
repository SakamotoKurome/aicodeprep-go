# aicodeprep-go

一个用于快速生成包含多个代码文件内容的 LLM Prompt 的 Go 命令行工具。

## 功能特性

- **文件选择**: 支持通过命令行参数或交互式输入选择文件
- **模式匹配**: 支持通配符模式（如 `src/*.go`, `**/*.js`）和排除模式
- **格式化输出**: 生成结构化的 Prompt 文本
- **剪贴板集成**: 自动复制到系统剪贴板（支持 Windows、macOS、Linux）
- **配置文件**: 支持 YAML 配置文件
- **交互式模式**: 提供友好的交互式界面
- **进度显示**: 处理大量文件时显示进度条

## 安装

```bash
# 从源码编译
git clone <repository-url>
cd aicodeprep-go
go build -o aicodeprep-go cmd/aicodeprep-go/main.go

# 或者直接运行
go run cmd/aicodeprep-go/main.go [flags]
```

## 使用方法

### 基本用法

```bash
# 处理当前目录下的所有 .go 文件
./aicodeprep-go -f "*.go" -p "请帮我重构这些代码"

# 处理多种文件类型
./aicodeprep-go -f "src/**/*.go" -f "internal/**/*.go" -e "*_test.go"

# 交互式模式
./aicodeprep-go -i

# 预览将要处理的文件
./aicodeprep-go --dry-run -f "*.go"

# 输出到文件而不是剪贴板
./aicodeprep-go -f "*.go" -o prompt.txt
```

### 命令行参数

- `-f, --files`: 文件模式（可多次使用）
- `-e, --exclude`: 排除模式（可多次使用）
- `-p, --prompt`: Prompt 文本
- `-i, --interactive`: 交互式模式
- `-o, --output`: 输出文件路径（默认输出到剪贴板）
- `-c, --config`: 配置文件路径
- `--dry-run`: 只显示将要处理的文件列表
- `--max-size`: 最大文件大小限制（默认 1MB）
- `-v, --verbose`: 详细输出模式

### 配置文件

可以使用 YAML 配置文件来设置默认选项：

```yaml
# .aicodeprep.yaml
files:
  - "src/**/*.go"
  - "internal/**/*.go"
exclude:
  - "vendor/**"
  - "*_test.go"
  - "*.generated.go"
prompt: |
  请帮我重构这些代码，提高可读性和性能。
  重点关注：
  1. 代码结构优化
  2. 错误处理改进
  3. 性能优化
max_file_size: 1048576  # 1MB
output: "prompt.txt"
```

配置文件查找顺序：
1. 通过 `-c` 参数指定的路径
2. 当前目录的 `.aicodeprep.yaml` 或 `.aicodeprep.yml`
3. `$HOME/.config/aicodeprep/config.yaml`
4. `$HOME/.aicodeprep.yaml`

### 输出格式

生成的 Prompt 格式如下：

```
=== 用户需求 ===
<用户输入的 Prompt>

=== 文件内容开始 ===
--- 文件: path/to/file1.go ---
<file1 内容>

--- 文件: path/to/file2.go ---
<file2 内容>
=== 文件内容结束 ===

=== 用户需求 ===
<用户输入的 Prompt>
```

## 支持的文件模式

### 基本通配符

- `*.go` - 当前目录下的所有 .go 文件
- `src/*.js` - src 目录下的所有 .js 文件
- `*.{go,js,ts}` - 所有 .go、.js、.ts 文件

### 递归模式

- `**/*.go` - 所有目录下的 .go 文件
- `src/**/*.js` - src 目录及其子目录下的所有 .js 文件
- `internal/**/*` - internal 目录下的所有文件

### 排除模式

- `vendor/**` - 排除 vendor 目录下的所有文件
- `*_test.go` - 排除所有测试文件
- `*.generated.*` - 排除所有生成的文件

## 剪贴板支持

工具会自动检测系统并使用相应的剪贴板命令：

- **Windows**: `clip.exe`
- **macOS**: `pbcopy`
- **Linux**: `xclip` 或 `xsel`

如果剪贴板不可用，会自动回退到文件输出模式。

## 交互式模式

使用 `-i` 参数启动交互式模式，将引导您：

1. 输入功能描述（Prompt）
2. 选择文件模式
3. 设置排除模式
4. 确认文件列表
5. 选择输出方式

## 项目结构

```
aicodeprep-go/
├── cmd/aicodeprep-go/main.go     # 主程序入口
├── internal/
│   ├── clipboard/clipboard.go    # 剪贴板操作
│   ├── selector/selector.go      # 文件选择逻辑
│   ├── formatter/formatter.go    # Prompt 格式化
│   ├── interactive/interactive.go # 交互式输入
│   └── config/config.go          # 配置文件处理
├── go.mod
├── go.sum
├── README.md
└── .gitignore
```

## 依赖

- [cobra](https://github.com/spf13/cobra) - 命令行界面
- [yaml.v3](https://gopkg.in/yaml.v3) - YAML 配置文件解析
- [progressbar](https://github.com/schollz/progressbar) - 进度条显示

## 许可证

MIT License