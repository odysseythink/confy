# Confy 缺失功能补齐 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 按 A→B→C→D→E→F 顺序补齐 confy 缺失的环境变量、Flag 绑定、文件自动发现、多格式支持、虚拟文件系统及测试文档。

**Architecture:** 在现有 confy 核心上增量扩展。能内联实现的逻辑绝不引入外部库；仅 TOML 引入 `pelletier/go-toml/v2`；pflag 适配放入独立子包 `confy/pflag`；自定义最小 FS 接口替代 afero。

**Tech Stack:** Go 1.25.7, `pelletier/go-toml/v2`, `spf13/pflag`

---

## File Structure

| File | Action | Responsibility |
|------|--------|----------------|
| `fs.go` | Create | `FS` interface + `osFS` 默认实现 |
| `fs_test.go` | Create | FS 接口测试、mock FS |
| `file.go` | Rewrite | 多路径配置文件搜索 |
| `file_test.go` | Create | 文件发现测试 |
| `dotenv_codec.go` | Create | dotenv/env 内联编解码器 |
| `ini_codec.go` | Create | INI 内联编解码器 |
| `toml_codec.go` | Create | TOML 编解码器（wrapper go-toml/v2） |
| `encoding_test.go` | Create | 各格式编解码器测试 |
| `flags.go` | Create | `FlagValue` / `FlagValueSet` 接口；`BindFlagValue` / `BindFlagValues` |
| `flags_test.go` | Create | Flag 绑定测试 |
| `pflag/pflag.go` | Create | pflag 适配子包 |
| `pflag/pflag_test.go` | Create | pflag 适配测试 |
| `env_test.go` | Create | 环境变量功能测试 |
| `confy.go` | Modify | 新增 env/flag/fs 字段；修改 `find()` 签名与优先级链；`New()` / `Sub()` 初始化 |
| `encoding.go` | Modify | 注册 TOML / dotenv / INI codec |
| `README.md` | Rewrite | 完整使用文档 |
| `go.mod` | Modify | 新增 `pelletier/go-toml/v2` 和 `spf13/pflag` |

---

## Task Decomposition

### Task 1: E. 虚拟文件系统 (FS interface)

**Files:**
- Create: `fs.go`
- Create: `fs_test.go`
- Modify: `confy.go`（`Confy` struct 新增 `fs FS` 字段，`New()` 初始化，`SetFS()` API）
- Modify: `file.go`（`exists` 改为 `exists(fs FS, path string)`）

- [ ] **Step 1: Create `fs.go` with FS interface and osFS**

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

- [ ] **Step 2: Modify `confy.go` — add `fs` field and `SetFS` API**

In `Confy` struct, add:
```go
fs FS
```

In `New()`, add after `v.configPermissions = os.FileMode(0o644)`:
```go
v.fs = osFS{}
```

Add public API:
```go
func SetFS(fs FS) { v.SetFS(fs) }
func (v *Confy) SetFS(fs FS) {
	v.fs = fs
}
```

- [ ] **Step 3: Modify `file.go` — `exists` now takes `fs` parameter**

```go
func exists(fs FS, path string) (bool, error) {
	stat, err := fs.Stat(path)
	if err == nil {
		return !stat.IsDir(), nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
```

- [ ] **Step 4: Create `fs_test.go` with mock FS test**

```go
package confy

import (
	"os"
	"testing"
)

type mockFS struct {
	files map[string]string
}

func (m *mockFS) Open(name string) (io.ReadCloser, error) {
	if data, ok := m.files[name]; ok {
		return io.NopCloser(strings.NewReader(data)), nil
	}
	return nil, os.ErrNotExist
}

func (m *mockFS) Stat(name string) (os.FileInfo, error) {
	if _, ok := m.files[name]; ok {
		return &mockFileInfo{name: name}, nil
	}
	return nil, os.ErrNotExist
}

func (m *mockFS) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	return nil, os.ErrNotExist
}

type mockFileInfo struct{ name string }

func (m *mockFileInfo) Name() string       { return m.name }
func (m *mockFileInfo) Size() int64        { return 0 }
func (m *mockFileInfo) Mode() os.FileMode  { return 0 }
func (m *mockFileInfo) ModTime() time.Time { return time.Time{} }
func (m *mockFileInfo) IsDir() bool        { return false }
func (m *mockFileInfo) Sys() any           { return nil }

func TestSetFS(t *testing.T) {
	v := New()
	m := &mockFS{files: map[string]string{}}
	v.SetFS(m)
	if v.fs == nil {
		t.Fatal("fs should not be nil after SetFS")
	}
}
```

- [ ] **Step 5: Run tests to verify no regressions**

Run: `go test ./... -run TestSetFS -v`
Expected: PASS

Run: `go test ./... -run TestWriteConfig -v`
Expected: PASS (existing test still works)

- [ ] **Step 6: Commit**

```bash
git add fs.go fs_test.go confy.go file.go
git commit -m "feat: add FS interface and osFS default implementation"
```

---

### Task 2: C. 文件发现 (multi-path config file search)

**Files:**
- Create: `file_test.go`
- Modify: `file.go` (rewrite with `findConfigFile` / `searchInPath`)
- Modify: `confy.go` (`getConfigFile()` logic, add `SetConfigName` / `SetConfigType` / `SetConfigPermissions`)

- [ ] **Step 1: Create `file_test.go` with failing tests**

```go
package confy

import (
	"testing"
)

func TestFindConfigFile(t *testing.T) {
	v := New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("/nonexistent")

	_, err := v.findConfigFile()
	if err == nil {
		t.Fatal("expected error for missing config file")
	}
}

func TestSearchInPath(t *testing.T) {
	v := New()
	v.SetConfigName("app")
	v.SetConfigType("json")

	m := &mockFS{files: map[string]string{
		"/etc/app.json": `{"key":"value"}`,
	}}
	v.SetFS(m)
	v.AddConfigPath("/etc")

	file, err := v.findConfigFile()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if file != "/etc/app.json" {
		t.Fatalf("expected /etc/app.json, got %s", file)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./... -run TestFindConfigFile -v`
Expected: FAIL — `findConfigFile` method not defined

- [ ] **Step 3: Rewrite `file.go` and modify `confy.go`**

`file.go`:
```go
package confy

import (
	"fmt"
	"path/filepath"
)

func (v *Confy) findConfigFile() (string, error) {
	v.logger.Info("searching for config in paths", "paths", v.configPaths)

	for _, cp := range v.configPaths {
		if file := v.searchInPath(cp); file != "" {
			return file, nil
		}
	}
	return "", ConfigFileNotFoundError{v.configName, fmt.Sprintf("%v", v.configPaths)}
}

func (v *Confy) searchInPath(in string) string {
	v.logger.Debug("searching for config in path", "path", in)
	for _, ext := range SupportedExts {
		fullPath := filepath.Join(in, v.configName+"."+ext)
		v.logger.Debug("checking if file exists", "file", fullPath)
		if b, _ := exists(v.fs, fullPath); b {
			v.logger.Debug("found file", "file", fullPath)
			return fullPath
		}
	}

	if v.configType != "" {
		fullPath := filepath.Join(in, v.configName)
		if b, _ := exists(v.fs, fullPath); b {
			return fullPath
		}
	}
	return ""
}

func exists(fs FS, path string) (bool, error) {
	stat, err := fs.Stat(path)
	if err == nil {
		return !stat.IsDir(), nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
```

`confy.go` — modify `getConfigFile()`:
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

Add public APIs to `confy.go`:
```go
func SetConfigName(in string) { v.SetConfigName(in) }
func (v *Confy) SetConfigName(in string) {
	if in != "" {
		v.configName = in
		v.configFile = ""
	}
}

func SetConfigType(in string) { v.SetConfigType(in) }
func (v *Confy) SetConfigType(in string) {
	if in != "" {
		v.configType = in
	}
}

func SetConfigPermissions(perm os.FileMode) { v.SetConfigPermissions(perm) }
func (v *Confy) SetConfigPermissions(perm os.FileMode) {
	v.configPermissions = perm.Perm()
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./... -run "TestFindConfigFile|TestSearchInPath" -v`
Expected: PASS

Run: `go test ./... -run TestWriteConfig -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add file.go file_test.go confy.go
git commit -m "feat: add multi-path config file discovery"
```

---

### Task 3: D. dotenv 编解码器

**Files:**
- Create: `dotenv_codec.go`
- Modify: `encoding.go` (register dotenv codec)
- Modify: `encoding_test.go` (add dotenv tests)

- [ ] **Step 0: Update `SupportedExts` in `confy.go`**

Change:
```go
var SupportedExts = []string{"json", "toml", "yaml", "yml", "properties", "props", "prop", "dotenv", "env", "ini"}
```

And update `Reset()` to match.

- [ ] **Step 1: Create `dotenv_codec.go`**

```go
package confy

import (
	"errors"
	"fmt"
	"strings"
)

type dotenvCodec struct{}

func (dotenvCodec) Encode(v map[string]any) ([]byte, error) {
	var b strings.Builder
	for key, val := range v {
		b.WriteString(fmt.Sprintf("%s=%v\n", key, val))
	}
	return []byte(b.String()), nil
}

func (dotenvCodec) Decode(b []byte, v map[string]any) error {
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		v[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}
	return nil
}
```

- [ ] **Step 2: Modify `encoding.go` to register dotenv codec**

In `DefaultCodecRegistry.codec()`, add:
```go
case "dotenv", "env":
	return dotenvCodec{}, true
```

- [ ] **Step 3: Modify `encoding_test.go` with dotenv tests**

```go
package confy

import (
	"testing"
)

func TestDotenvCodec(t *testing.T) {
	codec := dotenvCodec{}

	input := map[string]any{
		"KEY1": "value1",
		"KEY2": "value2",
	}
	b, err := codec.Encode(input)
	if err != nil {
		t.Fatal(err)
	}

	output := make(map[string]any)
	err = codec.Decode(b, output)
	if err != nil {
		t.Fatal(err)
	}

	if output["KEY1"] != "value1" {
		t.Fatalf("expected value1, got %v", output["KEY1"])
	}
}

func TestDotenvCodecDecode(t *testing.T) {
	codec := dotenvCodec{}
	data := []byte(`
# comment
KEY1=value1
KEY2 = value2
`)
	output := make(map[string]any)
	err := codec.Decode(data, output)
	if err != nil {
		t.Fatal(err)
	}
	if output["KEY1"] != "value1" || output["KEY2"] != "value2" {
		t.Fatalf("unexpected output: %v", output)
	}
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./... -run "TestDotenv" -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add dotenv_codec.go encoding.go encoding_test.go
git commit -m "feat: add dotenv codec support"
```

---

### Task 4: D. INI 编解码器

**Files:**
- Create: `ini_codec.go`
- Modify: `encoding.go` (register ini codec)
- Modify: `encoding_test.go` (add INI tests)

- [ ] **Step 1: Create `ini_codec.go`**

```go
package confy

import (
	"errors"
	"fmt"
	"strings"
)

type iniCodec struct{}

func (iniCodec) Encode(v map[string]any) ([]byte, error) {
	var b strings.Builder
	for key, val := range v {
		b.WriteString(fmt.Sprintf("%s=%v\n", key, val))
	}
	return []byte(b.String()), nil
}

func (iniCodec) Decode(b []byte, v map[string]any) error {
	var section string
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.ToLower(line[1 : len(line)-1])
			continue
		}

		sep := "="
		if strings.Contains(line, ":") && !strings.Contains(line, "=") {
			sep = ":"
		}
		key, value, found := strings.Cut(line, sep)
		if !found {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if section != "" {
			key = section + "." + key
		}
		v[strings.ToLower(key)] = value
	}
	return nil
}
```

- [ ] **Step 2: Modify `encoding.go` to register INI codec**

In `DefaultCodecRegistry.codec()`, add:
```go
case "ini":
	return iniCodec{}, true
```

- [ ] **Step 3: Modify `encoding_test.go` with INI tests**

```go
func TestIniCodecDecode(t *testing.T) {
	codec := iniCodec{}
	data := []byte(`
[database]
host = localhost
port = 3306

[server]
name = myapp
`)
	output := make(map[string]any)
	err := codec.Decode(data, output)
	if err != nil {
		t.Fatal(err)
	}
	if output["database.host"] != "localhost" {
		t.Fatalf("expected localhost, got %v", output["database.host"])
	}
	if output["database.port"] != "3306" {
		t.Fatalf("expected 3306, got %v", output["database.port"])
	}
	if output["server.name"] != "myapp" {
		t.Fatalf("expected myapp, got %v", output["server.name"])
	}
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./... -run "TestIni" -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add ini_codec.go encoding.go encoding_test.go
git commit -m "feat: add INI codec support"
```

---

### Task 5: D. TOML 编解码器

**Files:**
- Create: `toml_codec.go`
- Modify: `encoding.go` (register toml codec)
- Modify: `encoding_test.go` (add TOML tests)
- Modify: `go.mod` (add `pelletier/go-toml/v2`)

- [ ] **Step 1: Add dependency**

Run: `go get github.com/pelletier/go-toml/v2`
Expected: `go.mod` and `go.sum` updated

- [ ] **Step 2: Create `toml_codec.go`**

```go
package confy

import (
	"github.com/pelletier/go-toml/v2"
)

type tomlCodec struct{}

func (tomlCodec) Encode(v map[string]any) ([]byte, error) {
	return toml.Marshal(v)
}

func (tomlCodec) Decode(b []byte, v map[string]any) error {
	return toml.Unmarshal(b, &v)
}
```

- [ ] **Step 3: Modify `encoding.go` to register TOML codec**

In `DefaultCodecRegistry.codec()`, add:
```go
case "toml":
	return tomlCodec{}, true
```

- [ ] **Step 4: Modify `encoding_test.go` with TOML tests**

```go
func TestTomlCodec(t *testing.T) {
	codec := tomlCodec{}
	input := map[string]any{
		"title": "Test",
		"owner": map[string]any{
			"name": "Tom",
		},
	}
	b, err := codec.Encode(input)
	if err != nil {
		t.Fatal(err)
	}

	output := make(map[string]any)
	err = codec.Decode(b, output)
	if err != nil {
		t.Fatal(err)
	}

	if output["title"] != "Test" {
		t.Fatalf("expected Test, got %v", output["title"])
	}
}
```

- [ ] **Step 5: Run tests**

Run: `go test ./... -run "TestToml" -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add toml_codec.go encoding.go encoding_test.go go.mod go.sum
git commit -m "feat: add TOML codec support using go-toml/v2"
```


---

### Task 6: A. 环境变量支持

**Files:**
- Create: `env_test.go`
- Modify: `confy.go` (add env fields to `Confy`, add env APIs, modify `find()` signature and logic, modify `New()` / `Sub()`)

- [ ] **Step 1: Create `env_test.go` with failing tests**

```go
package confy

import (
	"os"
	"testing"
)

func TestSetEnvPrefix(t *testing.T) {
	Reset()
	SetEnvPrefix("myapp")
	if GetEnvPrefix() != "myapp" {
		t.Fatalf("expected myapp, got %s", GetEnvPrefix())
	}
}

func TestAutomaticEnv(t *testing.T) {
	Reset()
	SetEnvPrefix("confy")
	AutomaticEnv()
	os.Setenv("CONFY_TESTKEY", "from_env")
	defer os.Unsetenv("CONFY_TESTKEY")

	val := GetString("testkey")
	if val != "from_env" {
		t.Fatalf("expected from_env, got %s", val)
	}
}

func TestBindEnv(t *testing.T) {
	Reset()
	BindEnv("mykey", "CUSTOM_ENV_VAR")
	os.Setenv("CUSTOM_ENV_VAR", "bound_value")
	defer os.Unsetenv("CUSTOM_ENV_VAR")

	val := GetString("mykey")
	if val != "bound_value" {
		t.Fatalf("expected bound_value, got %s", val)
	}
}

func TestAllowEmptyEnv(t *testing.T) {
	Reset()
	AllowEmptyEnv(true)
	SetEnvPrefix("confy")
	AutomaticEnv()
	os.Setenv("CONFY_EMPTY", "")
	defer os.Unsetenv("CONFY_EMPTY")

	val := GetString("empty")
	if val != "" {
		t.Fatalf("expected empty string, got %s", val)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./... -run "TestSetEnvPrefix|TestAutomaticEnv|TestBindEnv|TestAllowEmptyEnv" -v`
Expected: FAIL — env methods not defined

- [ ] **Step 3: Modify `confy.go` — add env fields and APIs**

In `Confy` struct, add:
```go
envPrefix           string
automaticEnvApplied bool
envKeyReplacer      StringReplacer
allowEmptyEnv       bool
env                 map[string][]string
```

In `New()`, add after `v.aliases = make(map[string]string)`:
```go
v.env = make(map[string][]string)
```

Add public APIs:
```go
func SetEnvPrefix(in string) { v.SetEnvPrefix(in) }
func (v *Confy) SetEnvPrefix(in string) {
	if in != "" {
		v.envPrefix = in
	}
}

func GetEnvPrefix() string { return v.GetEnvPrefix() }
func (v *Confy) GetEnvPrefix() string {
	return v.envPrefix
}

func AutomaticEnv() { v.AutomaticEnv() }
func (v *Confy) AutomaticEnv() {
	v.automaticEnvApplied = true
}

func AllowEmptyEnv(allowEmptyEnv bool) { v.AllowEmptyEnv(allowEmptyEnv) }
func (v *Confy) AllowEmptyEnv(allowEmptyEnv bool) {
	v.allowEmptyEnv = allowEmptyEnv
}

func SetEnvKeyReplacer(r StringReplacer) { v.SetEnvKeyReplacer(r) }
func (v *Confy) SetEnvKeyReplacer(r StringReplacer) {
	v.envKeyReplacer = r
}

func BindEnv(input ...string) error { return v.BindEnv(input...) }
func (v *Confy) BindEnv(input ...string) error {
	if len(input) == 0 {
		return fmt.Errorf("missing key to bind to")
	}
	key := strings.ToLower(input[0])
	if len(input) == 1 {
		v.env[key] = append(v.env[key], v.mergeWithEnvPrefix(key))
	} else {
		v.env[key] = append(v.env[key], input[1:]...)
	}
	return nil
}

func MustBindEnv(input ...string) { v.MustBindEnv(input...) }
func (v *Confy) MustBindEnv(input ...string) {
	if err := v.BindEnv(input...); err != nil {
		panic(fmt.Sprintf("error while binding environment variable: %v", err))
	}
}
```

Add internal methods:
```go
func (v *Confy) mergeWithEnvPrefix(in string) string {
	if v.envPrefix != "" {
		return strings.ToUpper(v.envPrefix + "_" + in)
	}
	return strings.ToUpper(in)
}

func (v *Confy) getEnv(key string) (string, bool) {
	if v.envKeyReplacer != nil {
		key = v.envKeyReplacer.Replace(key)
	}
	val, ok := os.LookupEnv(key)
	return val, ok && (v.allowEmptyEnv || val != "")
}

func (v *Confy) isPathShadowedInAutoEnv(path []string) string {
	var parentKey string
	for i := 1; i < len(path); i++ {
		parentKey = strings.Join(path[0:i], v.keyDelim)
		if _, ok := v.getEnv(v.mergeWithEnvPrefix(parentKey)); ok {
			return parentKey
		}
	}
	return ""
}
```

- [ ] **Step 4: Modify `confy.go` — update `find()` signature and env logic**

Change `func (v *Confy) find(lcaseKey string) any` to:
```go
func (v *Confy) find(lcaseKey string, flagDefault bool) any
```

Update all callers:
- `Get()`: `val := v.find(lcaseKey, true)`
- `IsSet()`: `val := v.find(lcaseKey, false)`

Insert env lookup into `find()`, after pflag block (before config file block):
```go
// Env override next
if v.automaticEnvApplied {
	envKey := strings.Join(append(v.parents, lcaseKey), ".")
	if val, ok := v.getEnv(v.mergeWithEnvPrefix(envKey)); ok {
		return val
	}
	if nested && v.isPathShadowedInAutoEnv(path) != "" {
		return nil
	}
}
envkeys, exists := v.env[lcaseKey]
if exists {
	for _, envkey := range envkeys {
		if val, ok := v.getEnv(envkey); ok {
			return val
		}
	}
}
if nested && v.isPathShadowedInFlatMap(path, v.env) != "" {
	return nil
}
```

Also update `Sub()` to copy env-related fields:
```go
subv.automaticEnvApplied = v.automaticEnvApplied
subv.envPrefix = v.envPrefix
subv.envKeyReplacer = v.envKeyReplacer
subv.allowEmptyEnv = v.allowEmptyEnv
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test ./... -run "TestSetEnvPrefix|TestAutomaticEnv|TestBindEnv|TestAllowEmptyEnv" -v`
Expected: PASS

Run: `go test ./... -run TestWriteConfig -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add confy.go env_test.go
git commit -m "feat: add environment variable support (AutomaticEnv, BindEnv, SetEnvPrefix)"
```

---

### Task 7: B. Flag 核心接口

**Files:**
- Create: `flags.go`
- Create: `flags_test.go`
- Modify: `confy.go` (add `pflags` field, modify `find()` for flag lookup, update `New()` / `AllKeys()`)

- [ ] **Step 1: Create `flags.go` with FlagValue interfaces**

```go
package confy

import "strings"

// FlagValueSet is an interface that users can implement
type FlagValueSet interface {
	VisitAll(fn func(FlagValue))
}

// FlagValue is an interface that users can implement
type FlagValue interface {
	HasChanged() bool
	Name() string
	ValueString() string
	ValueType() string
}

func BindFlagValues(flags FlagValueSet) error { return v.BindFlagValues(flags) }
func (v *Confy) BindFlagValues(flags FlagValueSet) (err error) {
	flags.VisitAll(func(flag FlagValue) {
		if err = v.BindFlagValue(flag.Name(), flag); err != nil {
			return
		}
	})
	return nil
}

func BindFlagValue(key string, flag FlagValue) error { return v.BindFlagValue(key, flag) }
func (v *Confy) BindFlagValue(key string, flag FlagValue) error {
	if flag == nil {
		return fmt.Errorf("flag for %q is nil", key)
	}
	v.pflags[strings.ToLower(key)] = flag
	return nil
}
```

- [ ] **Step 2: Create `flags_test.go` with mock flag tests**

```go
package confy

import (
	"testing"
)

type mockFlag struct {
	name      string
	val       string
	valType   string
	changed   bool
}

func (m mockFlag) HasChanged() bool { return m.changed }
func (m mockFlag) Name() string     { return m.name }
func (m mockFlag) ValueString() string { return m.val }
func (m mockFlag) ValueType() string   { return m.valType }

func TestBindFlagValue(t *testing.T) {
	Reset()
	flag := mockFlag{name: "port", val: "8080", valType: "int", changed: true}
	err := BindFlagValue("port", flag)
	if err != nil {
		t.Fatal(err)
	}

	val := GetInt("port")
	if val != 8080 {
		t.Fatalf("expected 8080, got %d", val)
	}
}

func TestBindFlagValueNotChanged(t *testing.T) {
	Reset()
	flag := mockFlag{name: "port", val: "8080", valType: "int", changed: false}
	err := BindFlagValue("port", flag)
	if err != nil {
		t.Fatal(err)
	}

	// flag not changed and no default -> should return nil
	val := Get("port")
	if val != nil {
		t.Fatalf("expected nil, got %v", val)
	}
}
```

- [ ] **Step 3: Run tests to verify they fail**

Run: `go test ./... -run "TestBindFlag" -v`
Expected: FAIL — `pflags` field not initialized, `find()` doesn't check flags

- [ ] **Step 4: Modify `confy.go`**

In `Confy` struct, add:
```go
pflags map[string]FlagValue
```

In `New()`, add after `v.env = make(map[string][]string)`:
```go
v.pflags = make(map[string]FlagValue)
```

In `find()`, after override block and before env block, insert:
```go
// PFlag override next
flag, exists := v.pflags[lcaseKey]
if exists && flag.HasChanged() {
	switch flag.ValueType() {
	case "int", "int8", "int16", "int32", "int64":
		return cast.ToInt(flag.ValueString())
	case "bool":
		return cast.ToBool(flag.ValueString())
	case "stringSlice", "stringArray":
		s := strings.TrimPrefix(flag.ValueString(), "[")
		s = strings.TrimSuffix(s, "]")
		res, _ := readAsCSV(s)
		return res
	case "stringToString":
		return stringToStringConv(flag.ValueString())
	case "durationSlice":
		s := strings.TrimPrefix(flag.ValueString(), "[")
		s = strings.TrimSuffix(s, "]")
		slice := strings.Split(s, ",")
		return cast.ToDurationSlice(slice)
	default:
		return flag.ValueString()
	}
}
if nested && v.isPathShadowedInFlatMap(path, v.pflags) != "" {
	return nil
}
```

At end of `find()`, before `return nil`, add flag default fallback:
```go
if flagDefault {
	if flag, exists := v.pflags[lcaseKey]; exists {
		switch flag.ValueType() {
		case "int", "int8", "int16", "int32", "int64":
			return cast.ToInt(flag.ValueString())
		case "bool":
			return cast.ToBool(flag.ValueString())
		case "stringSlice", "stringArray":
			s := strings.TrimPrefix(flag.ValueString(), "[")
			s = strings.TrimSuffix(s, "]")
			res, _ := readAsCSV(s)
			return res
		case "stringToString":
			return stringToStringConv(flag.ValueString())
		case "durationSlice":
			s := strings.TrimPrefix(flag.ValueString(), "[")
			s = strings.TrimSuffix(s, "]")
			slice := strings.Split(s, ",")
			return cast.ToDurationSlice(slice)
		default:
			return flag.ValueString()
		}
	}
}
```

Modify `isPathShadowedInFlatMap` to handle `map[string]FlagValue`:
```go
func (v *Confy) isPathShadowedInFlatMap(path []string, mi any) string {
	var m map[string]interface{}
	switch miv := mi.(type) {
	case map[string]string:
		m = castMapStringToMapInterface(miv)
	case map[string][]string:
		m = castMapStringSliceToMapInterface(miv)
	case map[string]FlagValue:
		m = castMapFlagToMapInterface(miv)
	default:
		return ""
	}
	// ... rest unchanged
}
```

Add helper:
```go
func castMapFlagToMapInterface(src map[string]FlagValue) map[string]any {
	tgt := map[string]any{}
	for k, v := range src {
		tgt[k] = v
	}
	return tgt
}
```

Update `AllKeys()` to include pflags:
```go
m = v.mergeFlatMap(m, castMapFlagToMapInterface(v.pflags))
m = v.mergeFlatMap(m, castMapStringSliceToMapInterface(v.env))
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test ./... -run "TestBindFlag" -v`
Expected: PASS

Run: `go test ./... -run TestWriteConfig -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add flags.go flags_test.go confy.go
git commit -m "feat: add FlagValue interface and flag binding support"
```

---

### Task 8: B. pflag 适配子包

**Files:**
- Create: `pflag/pflag.go`
- Create: `pflag/pflag_test.go`
- Modify: `go.mod` (add `spf13/pflag`)

- [ ] **Step 1: Add dependency**

Run: `go get github.com/spf13/pflag`
Expected: `go.mod` and `go.sum` updated

- [ ] **Step 2: Create `pflag/pflag.go`**

```go
package pflag

import (
	"github.com/odysseythink/confy"
	"github.com/spf13/pflag"
)

func BindPFlag(v *confy.Confy, key string, flag *pflag.Flag) error {
	if flag == nil {
		return fmt.Errorf("flag for %q is nil", key)
	}
	return v.BindFlagValue(key, pflagValue{flag})
}

func BindPFlags(v *confy.Confy, flags *pflag.FlagSet) error {
	return v.BindFlagValues(pflagValueSet{flags})
}

type pflagValueSet struct {
	flags *pflag.FlagSet
}

func (p pflagValueSet) VisitAll(fn func(confy.FlagValue)) {
	p.flags.VisitAll(func(flag *pflag.Flag) {
		fn(pflagValue{flag})
	})
}

type pflagValue struct {
	flag *pflag.Flag
}

func (p pflagValue) HasChanged() bool { return p.flag.Changed }
func (p pflagValue) Name() string     { return p.flag.Name }
func (p pflagValue) ValueString() string { return p.flag.Value.String() }
func (p pflagValue) ValueType() string   { return p.flag.Value.Type() }
```

Wait — `fmt` is used but not imported. Fix:

```go
package pflag

import (
	"fmt"

	"github.com/odysseythink/confy"
	"github.com/spf13/pflag"
)
```

- [ ] **Step 3: Create `pflag/pflag_test.go`**

```go
package pflag

import (
	"testing"

	"github.com/odysseythink/confy"
	"github.com/spf13/pflag"
)

func TestBindPFlag(t *testing.T) {
	v := confy.New()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.Int("port", 8080, "port number")
	_ = fs.Set("port", "9090")

	err := BindPFlag(v, "port", fs.Lookup("port"))
	if err != nil {
		t.Fatal(err)
	}

	val := v.GetInt("port")
	if val != 9090 {
		t.Fatalf("expected 9090, got %d", val)
	}
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./pflag/... -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add pflag/ go.mod go.sum
git commit -m "feat: add pflag adapter subpackage"
```

---

### Task 9: F. 测试补充与 README 文档

**Files:**
- Modify: `confy_test.go` (expand)
- Modify: `README.md` (rewrite)

- [ ] **Step 1: Expand `confy_test.go` with integration tests**

```go
package confy

import (
	"testing"
)

func TestWriteConfig(t *testing.T) {
	SetConfigFile("config.yml")
	ReadInConfig()
	Set("mysql.path", "127.0.0.1")
	Set("mysql.port", "3306")
	Set("mysql.config", "charset=utf8mb4&parseTime=True&loc=Local")
	Set("mysql.db-name", "")
	Set("mysql.username", "test")
	Set("mysql.password", "test")
	Set("mysql.prefix", "")
	Set("mysql.singular", false)
	Set("mysql.engine", "")
	Set("mysql.max-idle-conns", 10)
	Set("mysql.max-open-conns", 100)
	Set("mysql.log-mode", "error")
	Set("jwt.signing-key", "131a5a9e-ccf4-434f-b17c-ed46bda2c4da")
	err := WriteConfig()
	if err != nil {
		t.Errorf("write config failed:%v", err)
	}
}

func TestAllSettings(t *testing.T) {
	Reset()
	SetDefault("name", "default")
	Set("name", "override")
	all := AllSettings()
	if all["name"] != "override" {
		t.Fatalf("expected override, got %v", all["name"])
	}
}

func TestMergeConfigMap(t *testing.T) {
	Reset()
	Set("a.b", "1")
	MergeConfigMap(map[string]any{
		"a": map[string]any{
			"c": "2",
		},
	})
	if GetString("a.b") != "1" {
		t.Fatalf("expected 1, got %s", GetString("a.b"))
	}
	if GetString("a.c") != "2" {
		t.Fatalf("expected 2, got %s", GetString("a.c"))
	}
}
```

- [ ] **Step 2: Rewrite `README.md`**

```markdown
# confy

A lightweight, extensible configuration management library for Go. Inspired by [viper](https://github.com/spf13/viper), confy supports multiple configuration sources with a clear priority order.

## Features

- Read and write JSON, YAML, TOML, INI, and dotenv configuration files
- Automatic config file discovery across multiple search paths
- Environment variable binding with prefix support
- Command-line flag binding (generic interface + optional pflag adapter)
- Remote configuration support (etcd, consul, etc.)
- Live config file watching
- Virtual filesystem support for easy testing

## Install

```bash
go get github.com/odysseythink/confy
```

For pflag support:
```bash
go get github.com/odysseythink/confy/pflag
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/odysseythink/confy"
)

func main() {
    confy.SetConfigName("config")
    confy.SetConfigType("yaml")
    confy.AddConfigPath(".")
    confy.ReadInConfig()

    fmt.Println(confy.GetString("database.host"))
}
```

## Configuration Priority

1. Explicit `Set()` call
2. Command-line flags
3. Environment variables
4. Config file
5. Remote key/value store
6. Defaults

## Environment Variables

```go
confy.SetEnvPrefix("MYAPP")
confy.AutomaticEnv()

// Will read MYAPP_PORT from environment
port := confy.GetString("port")
```

## Flag Binding

```go
import confypflag "github.com/odysseythink/confy/pflag"
import "github.com/spf13/pflag"

fs := pflag.NewFlagSet("myapp", pflag.ContinueOnError)
fs.String("config", "", "config file path")
confypflag.BindPFlags(confy.GetConfy(), fs)
```

## Testing with Mock FS

```go
import "testing"

func TestConfig(t *testing.T) {
    v := confy.New()
    v.SetFS(&myMockFS{})
    v.ReadInConfig()
}
```
```

- [ ] **Step 3: Run all tests**

Run: `go test ./... -v`
Expected: ALL PASS

- [ ] **Step 4: Commit**

```bash
git add confy_test.go README.md
git commit -m "docs: rewrite README and expand test coverage"
```

