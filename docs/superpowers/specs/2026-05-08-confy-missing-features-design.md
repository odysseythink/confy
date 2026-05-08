# Confy 缺失功能补齐设计文档

## 1. 背景与目标

当前 `confy` 是基于 `viper` 裁剪而来的轻量级配置库，仅支持 JSON/YAML 文件读写。与 viper v1.21.0 相比，缺失环境变量、Flag 绑定、文件自动发现、多格式支持、虚拟文件系统等核心能力。

本设计目标是在**最小外部依赖**原则下，按 `A→B→C→D→E→F` 顺序补齐这些功能，使 confy 成为功能完整但保持轻量的配置解决方案。

## 2. 设计原则

1. **最小外部依赖**：能内联实现的逻辑绝不引入库；仅 TOML 解析器（`pelletier/go-toml/v2`）因复杂度允许引入。
2. **向后兼容**：所有现有 API 行为不变，新增功能通过新增方法暴露。
3. **与 viper API 兼容**：方法签名尽量与 viper 保持一致，降低用户迁移成本。
4. **分层解耦**：pflag 适配放入独立子包 `confy/pflag`，核心包保持零依赖。

## 3. 功能模块设计

### 3.1 A. 环境变量支持

#### 3.1.1 新增字段（`Confy` struct）

```go
envPrefix           string
automaticEnvApplied bool
envKeyReplacer      StringReplacer  // 复用已有 interface
allowEmptyEnv       bool
env                 map[string][]string  // key -> []envVarName
```

#### 3.1.2 新增公共 API

```go
func SetEnvPrefix(in string)
func GetEnvPrefix() string
func AutomaticEnv()
func AllowEmptyEnv(bool)
func BindEnv(input ...string) error
func MustBindEnv(input ...string)
func SetEnvKeyReplacer(r StringReplacer)
```

#### 3.1.3 优先级链更新

`find()` 方法中新增 env 查找阶段，位于 pflag 之后、config file 之前：

```
override → pflag → env → config file → kvstore → default
```

#### 3.1.4 内部方法

- `getEnv(key string) (string, bool)`：封装 `os.LookupEnv`，应用 `envKeyReplacer` 替换，根据 `allowEmptyEnv` 判断空值是否有效。
- `mergeWithEnvPrefix(in string) string`：若设置了前缀，自动拼接为 `PREFIX_KEY` 格式并转大写。
- `isPathShadowedInAutoEnv(path []string) string`：自动环境下嵌套键的 shadow 检测。

#### 3.1.5 默认行为

- `allowEmptyEnv = false`：空字符串环境变量视为未设置，回退到下一优先级。
- 环境变量名大小写敏感（与 viper 一致）。

---

### 3.2 B. Flag 绑定

#### 3.2.1 分层架构

| 层级 | 包 | 依赖 | 说明 |
|------|-----|------|------|
| 核心接口 | `confy` | 0 | `FlagValue` / `FlagValueSet` 接口 + `BindFlagValue` |
| pflag 适配 | `confy/pflag` | `spf13/pflag` | 为 pflag/Cobra 用户提供便捷绑定 |

#### 3.2.2 核心接口（`confy` 包，零依赖）

```go
type FlagValue interface {
    HasChanged() bool
    Name() string
    ValueString() string
    ValueType() string
}

type FlagValueSet interface {
    VisitAll(fn func(FlagValue))
}

func BindFlagValue(key string, flag FlagValue) error
func BindFlagValues(flags FlagValueSet) error
```

#### 3.2.3 `Confy` 新增字段

```go
pflags map[string]FlagValue
```

#### 3.2.4 `find()` 中 flag 查找逻辑

位于 override 之后、env 之前：

1. 若 `flag.HasChanged()` 为 true，按 `flag.ValueType()` 转换值：
   - `int/int8/int16/int32/int64` → `cast.ToInt`
   - `bool` → `cast.ToBool`
   - `stringSlice/stringArray` → CSV 解析为 `[]string`
   - `stringToString` → `stringToStringConv`
   - `durationSlice` → `cast.ToDurationSlice`
   - 默认 → 原字符串
2. 若 `flagDefault = true` 且 flag 未改变，返回 flag 默认值。

#### 3.2.5 pflag 适配子包（`confy/pflag`）

```go
package pflag

import (
    "github.com/spf13/pflag"
    "github.com/odysseythink/confy"
)

func BindPFlag(v *confy.Confy, key string, flag *pflag.Flag) error
func BindPFlags(v *confy.Confy, flags *pflag.FlagSet) error
```

---

### 3.3 C. 文件发现

#### 3.3.1 目标

当用户未显式调用 `SetConfigFile()` 时，通过 `SetConfigName()` + `AddConfigPath()` 自动在多路径中搜索配置文件。

#### 3.3.2 修改 `getConfigFile()`

```go
func (v *Confy) getConfigFile() (string, error) {
    if v.configFile != "" {
        return v.configFile, nil
    }
    if v.configName == "" {
        return "", errors.New("no config file or config name provided")
    }
    return v.findConfigFile()
}
```

#### 3.3.3 新增内部方法

```go
func (v *Confy) findConfigFile() (string, error) {
    for _, cp := range v.configPaths {
        if file := v.searchInPath(cp); file != "" {
            return file, nil
        }
    }
    return "", ConfigFileNotFoundError{v.configName, fmt.Sprintf("%v", v.configPaths)}
}

func (v *Confy) searchInPath(in string) string {
    for _, ext := range SupportedExts {
        if b, _ := exists(v.fs, filepath.Join(in, v.configName+"."+ext)); b {
            return filepath.Join(in, v.configName+"."+ext)
        }
    }
    if v.configType != "" {
        if b, _ := exists(v.fs, filepath.Join(in, v.configName)); b {
            return filepath.Join(in, v.configName)
        }
    }
    return ""
}
```

#### 3.3.4 新增公共 API

```go
func SetConfigName(in string)
func SetConfigType(in string)
func SetConfigPermissions(perm os.FileMode)
```

#### 3.3.5 搜索顺序

按 `configPaths` 定义的顺序遍历，每个路径内按 `SupportedExts` 定义的顺序匹配扩展名。

---

### 3.4 D. 更多格式支持

#### 3.4.1 扩展 `SupportedExts`

```go
var SupportedExts = []string{
    "json", "toml", "yaml", "yml",
    "properties", "props", "prop",
    "dotenv", "env", "ini",
}
```

> HCL / tfvars **不在本设计范围内**。HCL 解析器过重，使用场景逐年减少，若后续有需求可单独评估。

#### 3.4.2 编解码器注册表扩展（`DefaultCodecRegistry.codec()`）

| 格式 | 实现 | 外部依赖 |
|------|------|---------|
| JSON | 标准库 `encoding/json` | 0 |
| YAML | `gopkg.in/yaml.v3`（已有） | 0 |
| TOML | `pelletier/go-toml/v2` | 1 |
| dotenv / env | 内联解析 | 0 |
| INI | 内联解析 | 0 |

#### 3.4.3 dotenv 内联解析

按行分割，规则：
- 忽略空行和以 `#` 开头的注释行
- 每行按第一个 `=` 分割为 key / value
- key 和 value 做 `strings.TrimSpace`
- 结果存入 `map[string]any`

#### 3.4.4 INI 内联解析

- 支持 `[section]` 分段
- 键值对用 `=` 或 `:` 分割
- 键名大小写不敏感（与 confy 整体行为一致）
- 扁平化为 `section.key` 格式存储到 `map[string]any`

---

### 3.5 E. 虚拟文件系统

#### 3.5.1 目标

不引入 `afero`，自定义最小 FS 接口，使测试可注入 mock 文件系统。

#### 3.5.2 新增文件 `fs.go`

```go
package confy

import (
    "io"
    "os"
)

type FS interface {
    Open(name string) (io.ReadCloser, error)
    Stat(name string) (os.FileInfo, error)
    OpenFile(name string, flag int, perm os.FileMode) (*os.File, error)
}

type osFS struct{}

func (osFS) Open(name string) (io.ReadCloser, error) { return os.Open(name) }
func (osFS) Stat(name string) (os.FileInfo, error)   { return os.Stat(name) }
func (osFS) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
    return os.OpenFile(name, flag, perm)
}
```

#### 3.5.3 `Confy` struct 变更

新增字段：
```go
fs FS
```

`New()` 初始化：`v.fs = osFS{}`

#### 3.5.4 新增公共 API

```go
func SetFS(fs FS)
```

#### 3.5.5 影响范围

以下方法全部改为通过 `v.fs` 操作文件：
- `ReadInConfig`（读）
- `MergeInConfig`（读）
- `writeConfig`（写）
- `exists`（检查存在性）
- `findConfigFile` / `searchInPath`（搜索）

---

### 3.6 F. 测试与文档

#### 3.6.1 测试策略

| 测试文件 | 覆盖内容 |
|---------|---------|
| `confy_test.go` | 核心功能回归（读写、merge、marshal） |
| `env_test.go` | 环境变量前缀、自动环境、BindEnv、空值处理 |
| `flags_test.go` | FlagValue 接口绑定、shadow 逻辑 |
| `file_test.go` | 多路径搜索、无扩展名文件、ConfigFileNotFoundError |
| `encoding_test.go` | TOML / dotenv / INI 编解码器 |
| `fs_test.go` | mock FS 注入、ReadInConfig 通过 FS |

测试利用 `SetFS` 注入内存 mock 文件系统，避免污染真实磁盘。

#### 3.6.2 文档更新

重写 `README.md`，包含：
- 项目简介与安装
- 快速开始示例
- 支持的配置格式列表
- 配置源优先级说明
- 环境变量使用示例
- Flag 绑定示例（含 pflag 子包）
- 配置文件自动发现说明
- 虚拟文件系统（测试）说明

---

## 4. 最终配置源优先级链

```
1. explicit Set()          override
2. command-line flags      pflags
3. environment variables   env (automatic + BindEnv)
4. config file             config file (支持自动发现)
5. remote key/value store  kvstore
6. defaults                defaults
```

每条链路都支持大小写不敏感键名、嵌套键（`.` 分隔）、alias 转发。

---

## 5. 文件变更清单

| 文件 | 操作 | 说明 |
|------|------|------|
| `confy.go` | 修改 | 新增 env/flag/fs 字段；修改 `find(lcaseKey string)` → `find(lcaseKey string, flagDefault bool)` / `New()` / `Sub()` |
| `flags.go` | 新增 | `FlagValue` / `FlagValueSet` 接口；`BindFlagValue` / `BindFlagValues` |
| `file.go` | 重写 | 多路径搜索；`findConfigFile` / `searchInPath` |
| `fs.go` | 新增 | `FS` interface + `osFS` 默认实现 |
| `encoding.go` | 修改 | 注册 TOML / dotenv / INI codec |
| `json_codec.go` | 无变更 | — |
| `yaml_codec.go` | 无变更 | — |
| `toml_codec.go` | 新增 | TOML 编解码器（wrapper go-toml/v2） |
| `dotenv_codec.go` | 新增 | dotenv 内联编解码器 |
| `ini_codec.go` | 新增 | INI 内联编解码器 |
| `remote.go` | 无变更 | — |
| `util.go` | 无变更 | — |
| `logger.go` | 无变更 | — |
| `experimental.go` | 无变更 | — |
| `pflag/` | 新增子包 | pflag 适配，作为根模块子目录，不独立 go.mod |
| `*_test.go` | 新增/扩展 | 各功能单元测试 |
| `README.md` | 重写 | 完整使用文档 |
| `go.mod` | 修改 | 新增 `pelletier/go-toml/v2`；新增 `spf13/pflag`（若 pflag 子包合并到根模块则引入） |

---

## 6. 依赖变化

### 当前依赖
```
fsnotify, etcd/api, etcd/client, crypto, yaml.v3
```

### 新增依赖
```
github.com/pelletier/go-toml/v2    # TOML 解析（唯一非标准库新增）
github.com/spf13/pflag             # pflag 适配子包所需（零间接依赖，极轻量）
```

### 移除/不变
- `cast` 保持内嵌
- `mapstructure` 保持内嵌
- 不引入 `afero`、`locafero`、`gotenv`

---

## 7. 风险与注意事项

1. **HCL 未实现**：若用户需要 HCL，需后续评估引入 `hcl` 解析器的成本。
2. **INI 扁平化**：INI 的 `section.key` 扁平化可能与用户预期不同，文档中需明确说明。
3. **pflag 子包**：作为根模块子目录，不独立 go.mod，根模块直接引入 `spf13/pflag`（该库零间接依赖，代价极小）。
4. **并发安全**：本次补齐不改动并发模型，仍保持“非并发安全”的现有约定。
