package confy

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"strings"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/odysseythink/confy/mapstructure"
)

func TestWriteConfig(t *testing.T) {
	Reset()
	tmpFile := filepath.Join(t.TempDir(), "config.yml")
	SetConfigFile(tmpFile)
	SetConfigType("yaml")
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

func TestKeyDelimiter(t *testing.T) {
	Reset()
	v := NewWithOptions(KeyDelimiter("/"))
	v.SetDefault("a/b", "val")
	if v.GetString("a/b") != "val" {
		t.Fatal("expected val")
	}
}

func TestWithDecodeHook(t *testing.T) {
	Reset()
	hook := mapstructure.ComposeDecodeHookFunc(mapstructure.StringToTimeDurationHookFunc())
	v := NewWithOptions(WithDecodeHook(hook))
	if v.decodeHook == nil {
		t.Fatal("expected decode hook")
	}
}

func TestWithDecodeHookNil(t *testing.T) {
	v := NewWithOptions(WithDecodeHook(nil))
	if v.decodeHook != nil {
		t.Fatal("expected nil decode hook")
	}
}

func TestNewWithOptions(t *testing.T) {
	v := NewWithOptions(KeyDelimiter("/"))
	if v.keyDelim != "/" {
		t.Fatalf("expected /, got %s", v.keyDelim)
	}
}

func TestSetOptions(t *testing.T) {
	Reset()
	SetOptions(KeyDelimiter("/"))
	if v.keyDelim != "/" {
		t.Fatalf("expected /, got %s", v.keyDelim)
	}
}

func TestOnConfigChange(t *testing.T) {
	Reset()
	OnConfigChange(func(e fsnotify.Event) {})
	if v.onConfigChange == nil {
		t.Fatal("expected onConfigChange to be set")
	}
}

func TestWatchConfig(t *testing.T) {
	Reset()
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configFile, []byte("key: value\n"), 0644); err != nil {
		t.Fatal(err)
	}
	SetConfigFile(configFile)
	SetConfigType("yaml")
	called := make(chan bool, 1)
	OnConfigChange(func(e fsnotify.Event) {
		select {
		case called <- true:
		default:
		}
	})
	WatchConfig()
	time.Sleep(50 * time.Millisecond)
	if err := os.WriteFile(configFile, []byte("key: value2\n"), 0644); err != nil {
		t.Fatal(err)
	}
	select {
	case <-called:
	case <-time.After(3 * time.Second):
		t.Fatal("expected OnConfigChange to be called")
	}
	os.Remove(configFile)
	time.Sleep(50 * time.Millisecond)
}

func TestWatchConfigNoConfigFile(t *testing.T) {
	Reset()
	// Should not panic even when config file is missing
	WatchConfig()
}

func TestConfigFileUsed(t *testing.T) {
	Reset()
	SetConfigFile("/etc/app.yaml")
	if ConfigFileUsed() != "/etc/app.yaml" {
		t.Fatalf("expected /etc/app.yaml, got %s", ConfigFileUsed())
	}
}

func TestAddConfigPath(t *testing.T) {
	Reset()
	AddConfigPath("/tmp")
	if len(v.configPaths) != 1 {
		t.Fatalf("expected 1 path, got %d", len(v.configPaths))
	}
}

func TestSetTypeByDefaultValue(t *testing.T) {
	Reset()
	SetTypeByDefaultValue(true)
	SetDefault("num", 0)
	Set("num", "123")
	if GetInt("num") != 123 {
		t.Fatalf("expected 123, got %d", GetInt("num"))
	}
}

func TestGetGeneric(t *testing.T) {
	Reset()
	Set("str", "hello")
	if Get[string]("str") != "hello" {
		t.Fatal("expected hello")
	}
	Set("num", 42)
	if Get[int]("num") != 42 {
		t.Fatalf("expected 42, got %d", Get[int]("num"))
	}
}

func TestGetWithDefault(t *testing.T) {
	Reset()
	if GetWithDefault("missing", "fallback") != "fallback" {
		t.Fatal("expected fallback")
	}
	SetConfigType("json")
	ReadConfig(strings.NewReader(`{"present": "value"}`))
	if GetWithDefault("present", "fallback") != "value" {
		t.Fatal("expected value")
	}
}

func TestGetSlice(t *testing.T) {
	Reset()
	Set("items", []map[string]any{{"name": "a"}, {"name": "b"}})
	res := GetSlice[map[string]any]("items")
	if len(res) != 2 || res[0]["name"] != "a" {
		t.Fatalf("unexpected: %v", res)
	}
}

func TestTypedGetters(t *testing.T) {
	Reset()
	Set("bool", true)
	Set("float", 3.14)
	Set("time", "2023-01-01T00:00:00Z")
	Set("duration", "1h")
	Set("intslice", []int{1, 2, 3})
	Set("stringslice", []string{"a", "b"})
	Set("stringmap", map[string]any{"k": "v"})
	Set("stringmapstring", map[string]string{"k": "v"})
	Set("stringmapstringslice", map[string][]string{"k": {"v1", "v2"}})
	Set("size", "1GB")

	if !GetBool("bool") {
		t.Fatal("expected true")
	}
	if GetFloat64("float") != 3.14 {
		t.Fatal("expected 3.14")
	}
	if GetTime("time").Year() != 2023 {
		t.Fatal("expected 2023")
	}
	if GetDuration("duration") != time.Hour {
		t.Fatal("expected 1h")
	}
	if len(GetIntSlice("intslice")) != 3 {
		t.Fatal("expected 3 ints")
	}
	if len(GetStringSlice("stringslice")) != 2 {
		t.Fatal("expected 2 strings")
	}
	if GetStringMap("stringmap")["k"] != "v" {
		t.Fatal("expected v")
	}
	if GetStringMapString("stringmapstring")["k"] != "v" {
		t.Fatal("expected v")
	}
	if len(GetStringMapStringSlice("stringmapstringslice")["k"]) != 2 {
		t.Fatal("expected 2 items")
	}
	if GetSizeInBytes("size") != 1<<30 {
		t.Fatalf("expected %d, got %d", 1<<30, GetSizeInBytes("size"))
	}
}

func TestSub(t *testing.T) {
	Reset()
	Set("parent.child", "value")
	sub := Sub("parent")
	if sub == nil {
		t.Fatal("expected sub")
	}
	if sub.GetString("child") != "value" {
		t.Fatal("expected value")
	}
}

func TestSubNil(t *testing.T) {
	Reset()
	if Sub("missing") != nil {
		t.Fatal("expected nil")
	}
}

func TestSetEnvKeyReplacer(t *testing.T) {
	Reset()
	replacer := strings.NewReplacer("-", "_")
	SetEnvKeyReplacer(replacer)
	SetEnvPrefix("confy")
	AutomaticEnv()
	os.Setenv("CONFY_MY_KEY", "replaced")
	defer os.Unsetenv("CONFY_MY_KEY")
	if GetString("my-key") != "replaced" {
		t.Fatalf("expected replaced, got %s", GetString("my-key"))
	}
}

func TestMustBindEnv(t *testing.T) {
	Reset()
	MustBindEnv("key1", "ENV_KEY1")
	os.Setenv("ENV_KEY1", "val1")
	defer os.Unsetenv("ENV_KEY1")
	if GetString("key1") != "val1" {
		t.Fatal("expected val1")
	}
}

func TestMustBindEnvPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	Reset()
	MustBindEnv()
}

func TestRegisterAlias(t *testing.T) {
	Reset()
	Set("realkey", "value")
	RegisterAlias("aliaskey", "realkey")
	if GetString("aliaskey") != "value" {
		t.Fatal("expected value via alias")
	}
}

func TestRegisterAliasCircular(t *testing.T) {
	Reset()
	// alias == key should log warning but not panic
	RegisterAlias("same", "same")
}

func TestRegisterAliasMoveValue(t *testing.T) {
	Reset()
	Set("oldname", "moved")
	RegisterAlias("oldname", "newname")
	if GetString("newname") != "moved" {
		t.Fatal("expected moved")
	}
	if GetString("oldname") != "moved" {
		t.Fatal("expected alias to resolve")
	}
}

func TestInConfig(t *testing.T) {
	Reset()
	SetConfigType("json")
	ReadConfig(strings.NewReader(`{"existing": "value"}`))
	if !InConfig("existing") {
		t.Fatal("expected true")
	}
	if InConfig("missing") {
		t.Fatal("expected false")
	}
}

func TestIsSet(t *testing.T) {
	Reset()
	Set("setkey", "value")
	if !IsSet("setkey") {
		t.Fatal("expected true")
	}
	if IsSet("unsetkey") {
		t.Fatal("expected false")
	}
}

func TestWriteConfigTo(t *testing.T) {
	Reset()
	SetConfigType("json")
	Set("key", "value")
	var buf bytes.Buffer
	err := WriteConfigTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "key") {
		t.Fatalf("expected key in output, got %s", buf.String())
	}
}

func TestWriteConfigToUnsupported(t *testing.T) {
	Reset()
	SetConfigType("unsupported")
	var buf bytes.Buffer
	err := WriteConfigTo(&buf)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSafeWriteConfig(t *testing.T) {
	Reset()
	tmpDir := t.TempDir()
	AddConfigPath(tmpDir)
	SetConfigName("app")
	SetConfigType("json")
	Set("key", "value")
	err := SafeWriteConfig()
	if err != nil {
		t.Fatal(err)
	}
	// second call should fail
	err = SafeWriteConfig()
	if err == nil {
		t.Fatal("expected already exists error")
	}
}

func TestWriteConfigAs(t *testing.T) {
	Reset()
	tmpFile := filepath.Join(t.TempDir(), "out.yaml")
	SetConfigType("yaml")
	Set("key", "value")
	err := WriteConfigAs(tmpFile)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "key") {
		t.Fatal("expected key in file")
	}
}

func TestSafeWriteConfigAs(t *testing.T) {
	Reset()
	tmpFile := filepath.Join(t.TempDir(), "safe.json")
	SetConfigType("json")
	Set("key", "value")
	err := SafeWriteConfigAs(tmpFile)
	if err != nil {
		t.Fatal(err)
	}
	err = SafeWriteConfigAs(tmpFile)
	if err == nil {
		t.Fatal("expected already exists error")
	}
}

func TestReadConfig(t *testing.T) {
	Reset()
	SetConfigType("json")
	err := ReadConfig(strings.NewReader(`{"key": "value"}`))
	if err != nil {
		t.Fatal(err)
	}
	if GetString("key") != "value" {
		t.Fatal("expected value")
	}
}

func TestMergeConfig(t *testing.T) {
	Reset()
	Set("existing", "old")
	SetConfigType("json")
	err := MergeConfig(strings.NewReader(`{"new": "value"}`))
	if err != nil {
		t.Fatal(err)
	}
	if GetString("existing") != "old" {
		t.Fatal("expected old")
	}
	if GetString("new") != "value" {
		t.Fatal("expected value")
	}
}

func TestMergeInConfig(t *testing.T) {
	Reset()
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.json")
	os.WriteFile(configFile, []byte(`{"key": "value"}`), 0644)
	SetConfigFile(configFile)
	SetConfigType("json")
	err := MergeInConfig()
	if err != nil {
		t.Fatal(err)
	}
	if GetString("key") != "value" {
		t.Fatal("expected value")
	}
}

func TestAllKeys(t *testing.T) {
	Reset()
	SetDefault("default.key", "val")
	Set("override.key", "val")
	Set("config.key", "val")
	keys := AllKeys()
	m := make(map[string]bool)
	for _, k := range keys {
		m[k] = true
	}
	if !m["default.key"] || !m["override.key"] || !m["config.key"] {
		t.Fatalf("unexpected keys: %v", keys)
	}
}

func TestAllSettingsMerge(t *testing.T) {
	Reset()
	SetDefault("a", "default")
	Set("a", "override")
	all := AllSettings()
	if all["a"] != "override" {
		t.Fatal("expected override")
	}
}

func TestSetConfigPermissions(t *testing.T) {
	Reset()
	SetConfigPermissions(0755)
	if v.configPermissions != 0755 {
		t.Fatalf("expected 0755, got %o", v.configPermissions)
	}
}

func TestDebugTo(t *testing.T) {
	Reset()
	var buf bytes.Buffer
	DebugTo(&buf)
	if !strings.Contains(buf.String(), "Aliases:") {
		t.Fatal("expected Aliases in output")
	}
}

func TestUnmarshalExact(t *testing.T) {
	Reset()
	Set("name", "test")
	type S struct {
		Name string `mapstructure:"name"`
	}
	var s S
	err := UnmarshalExact(&s)
	if err != nil {
		t.Fatal(err)
	}
	if s.Name != "test" {
		t.Fatal("expected test")
	}
}

func TestUnmarshalExactUnusedField(t *testing.T) {
	Reset()
	Set("unknown", "value")
	type S struct {
		Name string `mapstructure:"name"`
	}
	var s S
	err := UnmarshalExact(&s)
	if err == nil {
		t.Fatal("expected error for unused field")
	}
}

func TestStringToWeakSliceHookFunc(t *testing.T) {
	Reset()
	Set("list", "a,b,c")
	type S struct {
		List []string `mapstructure:"list"`
	}
	var s S
	err := Unmarshal(&s)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.List) != 3 || s.List[0] != "a" {
		t.Fatalf("unexpected list: %v", s.List)
	}
}

func TestReadAsCSV(t *testing.T) {
	res, err := readAsCSV("a,b,c")
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != 3 {
		t.Fatal("expected 3 items")
	}
	res2, err := readAsCSV("")
	if err != nil {
		t.Fatal(err)
	}
	if len(res2) != 0 {
		t.Fatal("expected empty")
	}
}

func TestStringToStringConv(t *testing.T) {
	val := stringToStringConv("k1=v1,k2=v2")
	m, ok := val.(map[string]any)
	if !ok {
		t.Fatalf("expected map, got %T", val)
	}
	if m["k1"] != "v1" || m["k2"] != "v2" {
		t.Fatalf("unexpected: %v", m)
	}
	empty := stringToStringConv("")
	if len(empty.(map[string]any)) != 0 {
		t.Fatal("expected empty map")
	}
	invalid := stringToStringConv("novalue")
	if invalid != nil {
		t.Fatal("expected nil")
	}
}

func TestStringToIntConv(t *testing.T) {
	val := stringToIntConv("k1=1,k2=2")
	m, ok := val.(map[string]any)
	if !ok {
		t.Fatalf("expected map, got %T", val)
	}
	if m["k1"] != 1 || m["k2"] != 2 {
		t.Fatalf("unexpected: %v", m)
	}
	empty := stringToIntConv("")
	if len(empty.(map[string]any)) != 0 {
		t.Fatal("expected empty map")
	}
	invalid := stringToIntConv("bad")
	if invalid != nil {
		t.Fatal("expected nil")
	}
}

func TestCastToMapStringInterface(t *testing.T) {
	src := map[any]any{"key": "value"}
	dst := castToMapStringInterface(src)
	if dst["key"] != "value" {
		t.Fatal("expected value")
	}
}

func TestCastMapStringSliceToMapInterface(t *testing.T) {
	src := map[string][]string{"key": {"a", "b"}}
	dst := castMapStringSliceToMapInterface(src)
	if len(dst["key"].([]string)) != 2 {
		t.Fatal("expected 2 items")
	}
}

func TestCastMapFlagToMapInterface(t *testing.T) {
	f := map[string]FlagValue{"flag": mockFlag{name: "f", val: "v"}}
	m := castMapFlagToMapInterface(f)
	if m["flag"] == nil {
		t.Fatal("expected flag")
	}
}

func TestCastMapStringToMapInterface(t *testing.T) {
	src := map[string]string{"k": "v"}
	dst := castMapStringToMapInterface(src)
	if dst["k"] != "v" {
		t.Fatal("expected v")
	}
}

func TestMergeMaps(t *testing.T) {
	Reset()
	src := map[string]any{
		"a": map[string]any{"b": "new"},
		"c": "val",
	}
	tgt := map[string]any{
		"a": map[string]any{"d": "old"},
	}
	mergeMaps(src, tgt, nil)
	if tgt["a"].(map[string]any)["b"] != "new" {
		t.Fatal("expected new")
	}
	if tgt["a"].(map[string]any)["d"] != "old" {
		t.Fatal("expected old")
	}
	if tgt["c"] != "val" {
		t.Fatal("expected val")
	}
}

func TestMergeMapsMapAnyAny(t *testing.T) {
	Reset()
	src := map[string]any{
		"a": map[any]any{"b": "new"},
	}
	tgt := map[string]any{
		"a": map[any]any{"d": "old"},
	}
	mergeMaps(src, tgt, nil)
	// Just ensure no panic; coverage is the goal.
}

func TestMergeFlatMap(t *testing.T) {
	Reset()
	shadow := map[string]bool{"a": true}
	m := map[string]any{"a.b": "value", "c": "value2"}
	result := v.mergeFlatMap(shadow, m)
	if result["a.b"] {
		t.Fatal("expected a.b to be shadowed")
	}
	if !result["c"] {
		t.Fatal("expected c to be present")
	}
}

func TestSearchSliceWithPathPrefixes(t *testing.T) {
	Reset()
	SetConfigType("json")
	ReadConfig(strings.NewReader(`{"list": [{"name": "first"}, {"name": "second"}]}`))
	if GetString("list.0.name") != "first" {
		t.Fatalf("expected first, got %s", GetString("list.0.name"))
	}
	if GetString("list.1.name") != "second" {
		t.Fatalf("expected second, got %s", GetString("list.1.name"))
	}
}

func TestSearchIndexableWithPathPrefixes(t *testing.T) {
	Reset()
	Set("foo.bar", "val")
	if GetString("foo.bar") != "val" {
		t.Fatal("expected val")
	}
}

func TestIsPathShadowedInDeepMap(t *testing.T) {
	Reset()
	m := map[string]any{"a": map[string]any{"b": "val"}}
	shadow := v.isPathShadowedInDeepMap([]string{"a", "b", "c"}, m)
	if shadow != "a.b" {
		t.Fatalf("expected a.b, got %s", shadow)
	}
	noShadow := v.isPathShadowedInDeepMap([]string{"x", "y"}, m)
	if noShadow != "" {
		t.Fatal("expected no shadow")
	}
}

func TestIsPathShadowedInFlatMap(t *testing.T) {
	Reset()
	m := map[string]string{"a.b": "val"}
	shadow := v.isPathShadowedInFlatMap([]string{"a", "b", "c"}, m)
	if shadow != "a.b" {
		t.Fatalf("expected a.b, got %s", shadow)
	}
	noShadow := v.isPathShadowedInFlatMap([]string{"x", "y"}, m)
	if noShadow != "" {
		t.Fatal("expected no shadow")
	}
}

func TestFlattenAndMergeMap(t *testing.T) {
	Reset()
	m := map[string]any{"a": map[string]any{"b": "v"}}
	shadow := v.flattenAndMergeMap(nil, m, "")
	if !shadow["a.b"] {
		t.Fatal("expected a.b")
	}
}

func TestKeyExists(t *testing.T) {
	m := map[string]any{"Key": "value"}
	if keyExists("key", m) != "Key" {
		t.Fatal("expected Key")
	}
	if keyExists("missing", m) != "" {
		t.Fatal("expected empty")
	}
}

func TestWriteConfigUnsupportedExt(t *testing.T) {
	Reset()
	SetConfigFile("config.xml")
	SetConfigType("xml")
	err := WriteConfig()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestReadInConfigUnsupportedExt(t *testing.T) {
	Reset()
	SetConfigFile("config.xml")
	err := ReadInConfig()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestWriteConfigMissingType(t *testing.T) {
	Reset()
	SetConfigFile("config")
	v.configType = ""
	err := WriteConfig()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMergeInConfigMissingFile(t *testing.T) {
	Reset()
	SetConfigName("missing")
	SetConfigType("json")
	err := MergeInConfig()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSafeWriteConfigMissingPath(t *testing.T) {
	Reset()
	v.configPaths = []string{}
	err := SafeWriteConfig()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetConfigType(t *testing.T) {
	Reset()
	SetConfigType("yaml")
	if v.getConfigType() != "yaml" {
		t.Fatalf("expected yaml, got %s", v.getConfigType())
	}
	v.configType = ""
	SetConfigFile("app.json")
	if v.getConfigType() != "json" {
		t.Fatalf("expected json, got %s", v.getConfigType())
	}
}

func TestSetConfigName(t *testing.T) {
	Reset()
	SetConfigName("app")
	if v.configName != "app" {
		t.Fatalf("expected app, got %s", v.configName)
	}
	if v.configFile != "" {
		t.Fatal("expected configFile to be cleared")
	}
}

func TestSetConfigType(t *testing.T) {
	Reset()
	SetConfigType("toml")
	if v.configType != "toml" {
		t.Fatalf("expected toml, got %s", v.configType)
	}
}

func TestUnmarshalReaderEmptyType(t *testing.T) {
	Reset()
	v.configType = ""
	err := v.unmarshalReader(strings.NewReader("{}"), make(map[string]any))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUnmarshalReaderUnsupportedType(t *testing.T) {
	Reset()
	v.configType = "xml"
	err := v.unmarshalReader(strings.NewReader("<a/>"), make(map[string]any))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestExperimentalBindStruct(t *testing.T) {
	v := NewWithOptions(ExperimentalBindStruct())
	if !v.experimentalBindStruct {
		t.Fatal("expected experimentalBindStruct to be true")
	}
}

func TestWithLogger(t *testing.T) {
	logger := slog.New(&discardHandler{})
	v := NewWithOptions(WithLogger(logger))
	if v.logger != logger {
		t.Fatal("expected custom logger")
	}
}

// --- coverage boosters ---

func TestDebug(t *testing.T) {
	Reset()
	Debug()
}

func TestSetFSGlobal(t *testing.T) {
	Reset()
	m := &mockFS{files: map[string]string{}}
	SetFS(m)
	if v.fs == nil {
		t.Fatal("expected fs")
	}
}

func TestGetInt32(t *testing.T) {
	Reset()
	Set("i32", int32(32))
	if GetInt32("i32") != 32 {
		t.Fatal()
	}
}

func TestGetInt64(t *testing.T) {
	Reset()
	Set("i64", int64(64))
	if GetInt64("i64") != 64 {
		t.Fatal()
	}
}

func TestGetUint8(t *testing.T) {
	Reset()
	Set("u8", uint8(8))
	if GetUint8("u8") != 8 {
		t.Fatal()
	}
}

func TestGetUint(t *testing.T) {
	Reset()
	Set("u", uint(10))
	if GetUint("u") != 10 {
		t.Fatal()
	}
}

func TestGetUint16(t *testing.T) {
	Reset()
	Set("u16", uint16(16))
	if GetUint16("u16") != 16 {
		t.Fatal()
	}
}

func TestGetUint32(t *testing.T) {
	Reset()
	Set("u32", uint32(32))
	if GetUint32("u32") != 32 {
		t.Fatal()
	}
}

func TestGetUint64(t *testing.T) {
	Reset()
	Set("u64", uint64(64))
	if GetUint64("u64") != 64 {
		t.Fatal()
	}
}

func TestUnmarshalKey(t *testing.T) {
	Reset()
	Set("person", map[string]any{"name": "Alice", "age": 30})
	type Person struct {
		Name string `mapstructure:"name"`
		Age  int    `mapstructure:"age"`
	}
	var p Person
	err := UnmarshalKey("person", &p)
	if err != nil {
		t.Fatal(err)
	}
	if p.Name != "Alice" || p.Age != 30 {
		t.Fatalf("unexpected %+v", p)
	}
}

func TestUnmarshalWithExperimentalBindStruct(t *testing.T) {
	v := NewWithOptions(ExperimentalBindStruct())
	v.Set("name", "test")
	type S struct {
		Name string `mapstructure:"name"`
	}
	var s S
	err := v.Unmarshal(&s)
	if err != nil {
		t.Fatal(err)
	}
	if s.Name != "test" {
		t.Fatal("expected test")
	}
}

func TestBindFlagValues(t *testing.T) {
	Reset()
	flags := &mockFlagSet{
		flags: []FlagValue{
			mockFlag{name: "port", val: "8080", valType: "int", changed: true},
			mockFlag{name: "host", val: "localhost", valType: "string", changed: true},
		},
	}
	err := BindFlagValues(flags)
	if err != nil {
		t.Fatal(err)
	}
	if GetInt("port") != 8080 {
		t.Fatal()
	}
	if GetString("host") != "localhost" {
		t.Fatal()
	}
}

type mockFlagSet struct {
	flags []FlagValue
}

func (m *mockFlagSet) VisitAll(fn func(FlagValue)) {
	for _, f := range m.flags {
		fn(f)
	}
}

func TestBindFlagValueNil(t *testing.T) {
	Reset()
	err := BindFlagValue("key", nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestIniCodecEncode(t *testing.T) {
	codec := iniCodec{}
	input := map[string]any{"section.key": "value"}
	b, err := codec.Encode(input)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "section.key=value") {
		t.Fatalf("unexpected: %s", string(b))
	}
}

func TestDiscardHandlerMethods(t *testing.T) {
	h := &discardHandler{}
	if h.Enabled(nil, slog.LevelDebug) {
		t.Fatal("expected false")
	}
	if err := h.Handle(nil, slog.Record{}); err != nil {
		t.Fatal(err)
	}
	if h.WithAttrs(nil) == nil {
		t.Fatal("expected non-nil")
	}
	if h.WithGroup("g") == nil {
		t.Fatal("expected non-nil")
	}
}

func TestFindShadowedInOverride(t *testing.T) {
	Reset()
	Set("foo", "shadow")
	SetConfigType("json")
	ReadConfig(strings.NewReader(`{"foo": {"bar": "val"}}`))
	if v.find("foo.bar", true) != nil {
		t.Fatal("expected nil because foo shadows foo.bar")
	}
}

func TestFindFlagDefault(t *testing.T) {
	Reset()
	flag := mockFlag{name: "port", val: "8080", valType: "int", changed: false}
	BindFlagValue("port", flag)
	if v.find("port", true) == nil {
		t.Fatal("expected flag default value")
	}
	if v.find("port", false) != nil {
		t.Fatal("expected nil when flagDefault is false")
	}
}

func TestFindAutoEnvShadowed(t *testing.T) {
	Reset()
	AutomaticEnv()
	SetEnvPrefix("confy")
	os.Setenv("CONFY_FOO", "shadow")
	defer os.Unsetenv("CONFY_FOO")
	SetConfigType("json")
	ReadConfig(strings.NewReader(`{"foo": {"bar": "val"}}`))
	if v.find("foo.bar", true) != nil {
		t.Fatal("expected nil because env foo shadows foo.bar")
	}
}

func TestFindFlagShadowed(t *testing.T) {
	Reset()
	flag := mockFlag{name: "foo", val: "v", changed: true}
	BindFlagValue("foo", flag)
	SetConfigType("json")
	ReadConfig(strings.NewReader(`{"foo": {"bar": "val"}}`))
	if v.find("foo.bar", true) != nil {
		t.Fatal("expected nil because flag foo shadows foo.bar")
	}
}

func TestFindEnvShadowed(t *testing.T) {
	Reset()
	BindEnv("foo", "ENV_FOO")
	os.Setenv("ENV_FOO", "v")
	defer os.Unsetenv("ENV_FOO")
	SetConfigType("json")
	ReadConfig(strings.NewReader(`{"foo": {"bar": "val"}}`))
	if v.find("foo.bar", true) != nil {
		t.Fatal("expected nil because env foo shadows foo.bar")
	}
}

func TestGetTypeByDefaultValueAllTypes(t *testing.T) {
	Reset()
	SetTypeByDefaultValue(true)
	SetDefault("bool", false)
	SetDefault("str", "")
	SetDefault("int", 0)
	SetDefault("uint", uint(0))
	SetDefault("uint32", uint32(0))
	SetDefault("uint64", uint64(0))
	SetDefault("int64", int64(0))
	SetDefault("float", float64(0))
	SetDefault("t", time.Time{})
	SetDefault("dur", time.Duration(0))
	SetDefault("ss", []string{})
	SetDefault("si", []int{})

	Set("bool", "true")
	Set("str", "hello")
	Set("int", "42")
	Set("uint", "42")
	Set("uint32", "42")
	Set("uint64", "42")
	Set("int64", "42")
	Set("float", "3.14")
	Set("t", "2023-01-01T00:00:00Z")
	Set("dur", "1h")
	Set("ss", []string{"a", "b"})
	Set("si", []int{1, 2})

	if GetBool("bool") != true {
		t.Fatal()
	}
	if GetString("str") != "hello" {
		t.Fatal()
	}
	if GetInt("int") != 42 {
		t.Fatal()
	}
	if GetUint("uint") != 42 {
		t.Fatal()
	}
	if GetUint32("uint32") != 42 {
		t.Fatal()
	}
	if GetUint64("uint64") != 42 {
		t.Fatal()
	}
	if GetInt64("int64") != 42 {
		t.Fatal()
	}
	if GetFloat64("float") != 3.14 {
		t.Fatal()
	}
	if GetTime("t").Year() != 2023 {
		t.Fatal()
	}
	if GetDuration("dur") != time.Hour {
		t.Fatal()
	}
	if len(GetStringSlice("ss")) != 2 {
		t.Fatal()
	}
	if len(GetIntSlice("si")) != 2 {
		t.Fatal()
	}
}

func TestGetSliceError(t *testing.T) {
	Reset()
	Set("bad", make(chan int))
	res := GetSlice[string]("bad")
	if res != nil {
		t.Fatal("expected nil")
	}
}

func TestMergeMapsItgt(t *testing.T) {
	Reset()
	src := map[string]any{"a": "val"}
	tgt := map[string]any{}
	itgt := map[any]any{}
	mergeMaps(src, tgt, itgt)
	if tgt["a"] != "val" {
		t.Fatal()
	}
	if itgt["a"] != "val" {
		t.Fatal()
	}
}

func TestMergeMapsItgtDefault(t *testing.T) {
	Reset()
	src := map[string]any{"a": "new"}
	tgt := map[string]any{"a": "old"}
	itgt := map[any]any{"a": "old"}
	mergeMaps(src, tgt, itgt)
	if tgt["a"] != "new" {
		t.Fatal()
	}
	if itgt["a"] != "new" {
		t.Fatal()
	}
}

func TestMergeMapsTypeMismatch(t *testing.T) {
	Reset()
	src := map[string]any{"a": "string"}
	tgt := map[string]any{"a": map[string]any{"b": "old"}}
	mergeMaps(src, tgt, nil)
	if tgt["a"].(map[string]any)["b"] != "old" {
		t.Fatal("expected old")
	}
}

func TestRegisterAliasMoveFromDefaults(t *testing.T) {
	Reset()
	SetDefault("old", "defval")
	RegisterAlias("old", "new")
	if GetString("new") != "defval" {
		t.Fatal()
	}
}

func TestRegisterAliasMoveFromOverride(t *testing.T) {
	Reset()
	Set("old", "ovval")
	RegisterAlias("old", "new")
	if GetString("new") != "ovval" {
		t.Fatal()
	}
}

func TestRegisterAliasMoveFromKVStore(t *testing.T) {
	Reset()
	v.kvstore["old"] = "kvval"
	RegisterAlias("old", "new")
	if GetString("new") != "kvval" {
		t.Fatal()
	}
}

func TestSearchSliceWithPathPrefixesBounds(t *testing.T) {
	Reset()
	v.config = map[string]any{
		"list": []any{"a"},
	}
	if v.Get("list.5") != nil {
		t.Fatal("expected nil")
	}
}

func TestSearchSliceWithPathPrefixesNonMap(t *testing.T) {
	Reset()
	v.config = map[string]any{
		"list": []any{"a"},
	}
	if v.Get("list.0.name") != nil {
		t.Fatal("expected nil")
	}
}

func TestMarshalWriterEncodeError(t *testing.T) {
	Reset()
	reg := NewCodecRegistry()
	_ = reg.RegisterCodec("json", errorCodec{})
	v := NewWithOptions(WithEncoderRegistry(reg))
	v.SetConfigType("json")
	v.Set("key", "value")
	var buf bytes.Buffer
	err := v.WriteConfigTo(&buf)
	if err == nil {
		t.Fatal("expected error")
	}
}

type errorCodec struct{}

func (e errorCodec) Encode(v map[string]any) ([]byte, error) {
	return nil, errors.New("encode error")
}

func (e errorCodec) Decode(b []byte, v map[string]any) error {
	return nil
}

func TestMarshalWriterWriteError(t *testing.T) {
	Reset()
	SetConfigType("json")
	Set("key", "value")
	err := WriteConfigTo(&errorWriter{})
	if err == nil {
		t.Fatal("expected error")
	}
}

type errorWriter struct{}

func (e *errorWriter) Write(p []byte) (n int, err error) {
	return 0, errors.New("write error")
}

func TestSearchMapMapAnyAny(t *testing.T) {
	Reset()
	v.config = map[string]any{
		"a": map[any]any{"b": "val"},
	}
	if GetString("a.b") != "val" {
		t.Fatal()
	}
}

func TestWatchConfigRemove(t *testing.T) {
	Reset()
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configFile, []byte("key: value\n"), 0644); err != nil {
		t.Fatal(err)
	}
	SetConfigFile(configFile)
	SetConfigType("yaml")
	WatchConfig()
	time.Sleep(50 * time.Millisecond)
	os.Remove(configFile)
	time.Sleep(100 * time.Millisecond)
}

func TestMergeInConfigUnsupportedExt(t *testing.T) {
	Reset()
	SetConfigFile("config.xml")
	err := MergeInConfig()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMergeInConfigOpenError(t *testing.T) {
	Reset()
	SetConfigFile("/nonexistent/config.json")
	SetConfigType("json")
	err := MergeInConfig()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestReadInConfigOpenError(t *testing.T) {
	Reset()
	SetConfigFile("/nonexistent/config.json")
	SetConfigType("json")
	err := ReadInConfig()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUnmarshalReaderDecodeError(t *testing.T) {
	Reset()
	SetConfigType("json")
	err := v.unmarshalReader(strings.NewReader("not json"), make(map[string]any))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetConfigFileMissingName(t *testing.T) {
	Reset()
	v.configFile = ""
	v.configName = ""
	_, err := v.getConfigFile()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetRemoteConfigError(t *testing.T) {
	Reset()
	RemoteConfig = &mockRemoteConfigErr{}
	defer func() { RemoteConfig = nil }()
	_ = AddRemoteProvider("etcd", "http://localhost:2379", "/config")
	err := v.getKeyValueConfig()
	if err == nil {
		t.Fatal("expected error")
	}
}

type mockRemoteConfigErr struct{}

func (m *mockRemoteConfigErr) Get(rp RemoteProvider) (io.Reader, error) {
	return nil, errors.New("remote error")
}

func (m *mockRemoteConfigErr) Watch(rp RemoteProvider) (io.Reader, error) {
	return nil, errors.New("remote error")
}

func (m *mockRemoteConfigErr) WatchChannel(rp RemoteProvider) (<-chan *RemoteResponse, chan bool) {
	return nil, nil
}

func TestWatchRemoteConfigError(t *testing.T) {
	Reset()
	RemoteConfig = &mockRemoteConfigErr{}
	defer func() { RemoteConfig = nil }()
	_ = AddRemoteProvider("etcd", "http://localhost:2379", "/config")
	err := v.watchKeyValueConfig()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestExistsError(t *testing.T) {
	v := New()
	v.SetFS(&errorFS{})
	exists, err := exists(v.fs, "/any")
	if err == nil {
		t.Fatal("expected error")
	}
	if exists {
		t.Fatal("expected false")
	}
}

type errorFS struct{}

func (e *errorFS) Open(name string) (io.ReadCloser, error) {
	return nil, errors.New("fs error")
}

func (e *errorFS) Stat(name string) (os.FileInfo, error) {
	return nil, errors.New("fs error")
}

func (e *errorFS) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	return nil, errors.New("fs error")
}

func TestExistsDir(t *testing.T) {
	exists, err := exists(&dirFS{}, "/dir")
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal("expected false for dir")
	}
}

type dirFS struct{}

func (d *dirFS) Open(name string) (io.ReadCloser, error) {
	return nil, os.ErrNotExist
}

func (d *dirFS) Stat(name string) (os.FileInfo, error) {
	return &dirFileInfo{}, nil
}

func (d *dirFS) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	return nil, os.ErrNotExist
}

type dirFileInfo struct{}

func (d *dirFileInfo) Name() string       { return "dir" }
func (d *dirFileInfo) Size() int64        { return 0 }
func (d *dirFileInfo) Mode() os.FileMode  { return os.ModeDir }
func (d *dirFileInfo) ModTime() time.Time { return time.Time{} }
func (d *dirFileInfo) IsDir() bool        { return true }
func (d *dirFileInfo) Sys() any           { return nil }

func TestSearchInPathNoExt(t *testing.T) {
	v := New()
	v.SetConfigName("app")
	v.SetConfigType("json")
	m := &mockFS{files: map[string]string{
		"/etc/app": `{"key":"value"}`,
	}}
	v.SetFS(m)
	v.AddConfigPath("/etc")
	file, err := v.findConfigFile()
	if err != nil {
		t.Fatal(err)
	}
	if file != "/etc/app" {
		t.Fatalf("expected /etc/app, got %s", file)
	}
}

func TestFlattenAndMergeMapShadowed(t *testing.T) {
	Reset()
	shadow := map[string]bool{"a": true}
	m := map[string]any{"a": map[string]any{"b": "v"}, "c": "v2"}
	result := v.flattenAndMergeMap(shadow, m, "")
	if result["a.b"] {
		t.Fatal("expected a.b to be shadowed")
	}
	if !result["c"] {
		t.Fatal("expected c")
	}
}

func TestIsPathShadowedInFlatMapSlice(t *testing.T) {
	Reset()
	m := map[string][]string{"a.b": {"v"}}
	shadow := v.isPathShadowedInFlatMap([]string{"a", "b", "c"}, m)
	if shadow != "a.b" {
		t.Fatalf("expected a.b, got %s", shadow)
	}
}

func TestIsPathShadowedInFlatMapFlag(t *testing.T) {
	Reset()
	m := map[string]FlagValue{"a.b": mockFlag{name: "f"}}
	shadow := v.isPathShadowedInFlatMap([]string{"a", "b", "c"}, m)
	if shadow != "a.b" {
		t.Fatalf("expected a.b, got %s", shadow)
	}
}

func TestBindEnvEmpty(t *testing.T) {
	Reset()
	err := BindEnv()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestWriteConfigOpenFileError(t *testing.T) {
	Reset()
	SetFS(&errorFS{})
	SetConfigFile("/tmp/config.json")
	SetConfigType("json")
	err := WriteConfig()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetConfigTypeEmpty(t *testing.T) {
	Reset()
	v.configType = ""
	v.configFile = ""
	v.configName = ""
	if v.getConfigType() != "" {
		t.Fatalf("expected empty, got %s", v.getConfigType())
	}
}

func TestDotenvCodecDecodeNoEqual(t *testing.T) {
	codec := dotenvCodec{}
	output := make(map[string]any)
	err := codec.Decode([]byte("noequal\n"), output)
	if err != nil {
		t.Fatal(err)
	}
	if len(output) != 0 {
		t.Fatal("expected empty")
	}
}

func TestIniCodecDecodeColon(t *testing.T) {
	codec := iniCodec{}
	output := make(map[string]any)
	err := codec.Decode([]byte("key:value\n"), output)
	if err != nil {
		t.Fatal(err)
	}
	if output["key"] != "value" {
		t.Fatalf("expected value, got %v", output["key"])
	}
}

func TestSubNonMap(t *testing.T) {
	Reset()
	Set("scalar", "value")
	if Sub("scalar") != nil {
		t.Fatal("expected nil")
	}
}

func TestSearchMapWithPathPrefixesNested(t *testing.T) {
	Reset()
	v.config = map[string]any{
		"foo": map[string]any{"bar": "val"},
	}
	if GetString("foo.bar") != "val" {
		t.Fatal()
	}
}

func TestStringToStringConvCSVError(t *testing.T) {
	val := stringToStringConv(`"bad`)
	if val != nil {
		t.Fatal("expected nil")
	}
}

func TestStringToIntConvAtoiError(t *testing.T) {
	val := stringToIntConv("k=bad")
	if val != nil {
		t.Fatal("expected nil")
	}
}

func TestMergeWithEnvPrefixEmpty(t *testing.T) {
	Reset()
	v.envPrefix = ""
	if v.mergeWithEnvPrefix("key") != "KEY" {
		t.Fatal()
	}
}

func TestWithEncoderRegistryNil(t *testing.T) {
	v := NewWithOptions(WithEncoderRegistry(nil))
	if v.encoderRegistry == nil {
		t.Fatal("expected default registry")
	}
}

func TestWithDecoderRegistryNil(t *testing.T) {
	v := NewWithOptions(WithDecoderRegistry(nil))
	if v.decoderRegistry == nil {
		t.Fatal("expected default registry")
	}
}

// --- final coverage boosters ---

func TestDecodeHook(t *testing.T) {
	Reset()
	Set("dur", "1h")
	type S struct {
		Dur time.Duration `mapstructure:"dur"`
	}
	var s S
	err := Unmarshal(&s, DecodeHook(mapstructure.StringToTimeDurationHookFunc()))
	if err != nil {
		t.Fatal(err)
	}
	if s.Dur != time.Hour {
		t.Fatal()
	}
}

func TestUnmarshalExactExperimental(t *testing.T) {
	v := NewWithOptions(ExperimentalBindStruct())
	v.Set("name", "test")
	type S struct {
		Name string `mapstructure:"name"`
	}
	var s S
	err := v.UnmarshalExact(&s)
	if err != nil {
		t.Fatal(err)
	}
	if s.Name != "test" {
		t.Fatal()
	}
}

func TestFindFlagTypes(t *testing.T) {
	Reset()
	flag1 := mockFlag{name: "list", val: "[a,b,c]", valType: "stringSlice", changed: true}
	BindFlagValue("list", flag1)
	if len(v.find("list", true).([]string)) != 3 {
		t.Fatal("expected 3 items")
	}
	flag2 := mockFlag{name: "arr", val: "[a,b]", valType: "stringArray", changed: true}
	BindFlagValue("arr", flag2)
	if len(v.find("arr", true).([]string)) != 2 {
		t.Fatal("expected 2 items")
	}
	flag3 := mockFlag{name: "map", val: "[k1=v1,k2=v2]", valType: "stringToString", changed: true}
	BindFlagValue("map", flag3)
	m := v.find("map", true).(map[string]any)
	if m["k1"] != "v1" {
		t.Fatal()
	}
	flag4 := mockFlag{name: "durs", val: "[1h,30m]", valType: "durationSlice", changed: true}
	BindFlagValue("durs", flag4)
	ds := v.find("durs", true).([]time.Duration)
	if len(ds) != 2 {
		t.Fatal()
	}
	flag5 := mockFlag{name: "other", val: "hello", valType: "unknown", changed: true}
	BindFlagValue("other", flag5)
	if v.find("other", true) != "hello" {
		t.Fatal()
	}
}

func TestFindConfigDotKey(t *testing.T) {
	Reset()
	SetConfigType("json")
	ReadConfig(strings.NewReader(`{"foo": "shadow", "foo.bar": "val"}`))
	// searchIndexableWithPathPrefixes prioritizes "foo.bar" key over "foo" shadow
	if v.find("foo.bar", true) != "val" {
		t.Fatal("expected val")
	}
}

func TestFindKVStoreShadowed(t *testing.T) {
	Reset()
	v.kvstore["foo"] = "shadow"
	v.kvstore["foo.bar"] = "val"
	if v.find("foo.bar", true) != nil {
		t.Fatal("expected nil")
	}
}

func TestFindDefaultsShadowed(t *testing.T) {
	Reset()
	v.defaults["foo"] = "shadow"
	v.defaults["foo.bar"] = "val"
	if v.find("foo.bar", true) != nil {
		t.Fatal("expected nil")
	}
}

func TestFindBindEnvMultiple(t *testing.T) {
	Reset()
	BindEnv("key", "ENV1", "ENV2")
	os.Setenv("ENV2", "val2")
	defer os.Unsetenv("ENV2")
	if v.find("key", true) != "val2" {
		t.Fatal("expected val2")
	}
}

func TestFindEnvNestedShadow(t *testing.T) {
	Reset()
	BindEnv("foo.bar", "ENV_FOO_BAR")
	SetDefault("foo", "shadow")
	if v.find("foo.bar", true) != nil {
		t.Fatal("expected nil")
	}
}

func TestWatchConfigAddWatcherError(t *testing.T) {
	Reset()
	SetConfigFile("config.yaml")
	WatchConfig()
}

func TestMergeConfigUnmarshalError(t *testing.T) {
	Reset()
	SetConfigType("json")
	err := MergeConfig(strings.NewReader("not json"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestWriteConfigMissingFile(t *testing.T) {
	Reset()
	v.configFile = ""
	v.configName = ""
	err := WriteConfig()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetSliceUnmarshalError(t *testing.T) {
	Reset()
	Set("bad", map[string]any{"k": "v"})
	res := GetSlice[int]("bad")
	if res != nil {
		t.Fatal("expected nil")
	}
}

func TestYamlCodec(t *testing.T) {
	codec := yamlCodec{}
	input := map[string]any{"key": "value"}
	b, err := codec.Encode(input)
	if err != nil {
		t.Fatal(err)
	}
	output := make(map[string]any)
	err = codec.Decode(b, output)
	if err != nil {
		t.Fatal(err)
	}
	if output["key"] != "value" {
		t.Fatal()
	}
}

func TestMergeConfigMapNilConfig(t *testing.T) {
	Reset()
	v.config = nil
	err := MergeConfigMap(map[string]any{"key": "value"})
	if err != nil {
		t.Fatal(err)
	}
	if GetString("key") != "value" {
		t.Fatal()
	}
}

func TestSearchSliceWithPathPrefixesMapAnyAny(t *testing.T) {
	Reset()
	v.config = map[string]any{
		"list": []any{map[any]any{"name": "first"}},
	}
	if GetString("list.0.name") != "first" {
		t.Fatal()
	}
}

func TestMergeMapsMapAnyAnyItgt(t *testing.T) {
	Reset()
	src := map[string]any{"a": map[any]any{"b": "new"}}
	tgt := map[string]any{"a": map[any]any{"b": "old"}}
	itgt := tgt["a"].(map[any]any)
	mergeMaps(src, tgt, itgt)
	if itgt["b"] != "new" {
		t.Fatalf("expected new, got %v", itgt["b"])
	}
}

func TestSearchMapEmptyPath(t *testing.T) {
	Reset()
	m := map[string]any{"a": "b"}
	res := v.searchMap(m, []string{})
	rm, ok := res.(map[string]any)
	if !ok || rm["a"] != "b" {
		t.Fatal()
	}
}

func TestIsPathShadowedInAutoEnvNoShadow(t *testing.T) {
	Reset()
	AutomaticEnv()
	SetEnvPrefix("confy")
	if v.isPathShadowedInAutoEnv([]string{"foo", "bar"}) != "" {
		t.Fatal()
	}
}

func TestGetWithDefaultCastError(t *testing.T) {
	Reset()
	SetConfigType("json")
	ReadConfig(strings.NewReader(`{"num": "notanumber"}`))
	if GetWithDefault("num", 42) != 42 {
		t.Fatal("expected default on cast error")
	}
}

func TestRegisterAliasExists(t *testing.T) {
	Reset()
	RegisterAlias("alias", "key")
	RegisterAlias("alias", "key")
	if GetString("alias") != "" {
		t.Fatal()
	}
}

func TestStringToWeakSliceHookFuncEmpty(t *testing.T) {
	Reset()
	Set("empty", "")
	type S struct {
		List []string `mapstructure:"empty"`
	}
	var s S
	err := Unmarshal(&s)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.List) != 0 {
		t.Fatalf("expected empty, got %v", s.List)
	}
}

func TestBindEnvMergeWithEnvPrefix(t *testing.T) {
	Reset()
	SetEnvPrefix("myapp")
	BindEnv("key")
	os.Setenv("MYAPP_KEY", "val")
	defer os.Unsetenv("MYAPP_KEY")
	if GetString("key") != "val" {
		t.Fatal()
	}
}

func TestDefaultDecoderConfigOpts(t *testing.T) {
	Reset()
	Set("name", "test")
	type S struct {
		Name string `mapstructure:"name"`
	}
	var s S
	opt := func(c *mapstructure.DecoderConfig) {
		c.TagName = "mapstructure"
	}
	err := Unmarshal(&s, opt)
	if err != nil {
		t.Fatal(err)
	}
	if s.Name != "test" {
		t.Fatal()
	}
}

func TestGetConfigTypeNoExt(t *testing.T) {
	Reset()
	v.configType = ""
	SetConfigFile("config")
	if v.getConfigType() != "" {
		t.Fatalf("expected empty, got %s", v.getConfigType())
	}
}

func TestIsPathShadowedInDeepMapMapAnyAny(t *testing.T) {
	Reset()
	m := map[string]any{"a": map[any]any{"b": "val"}}
	shadow := v.isPathShadowedInDeepMap([]string{"a", "b", "c"}, m)
	if shadow != "a.b" {
		t.Fatalf("expected a.b, got %s", shadow)
	}
}

func TestSearchIndexableWithPathPrefixesMissing(t *testing.T) {
	Reset()
	v.config = map[string]any{
		"foo": map[string]any{"bar": "val"},
	}
	if v.Get("foo.missing") != nil {
		t.Fatal()
	}
}

func TestFindKVStore(t *testing.T) {
	Reset()
	v.kvstore["key"] = "val"
	if v.find("key", true) != "val" {
		t.Fatal()
	}
}

func TestFindDefaults(t *testing.T) {
	Reset()
	SetDefault("key", "val")
	if v.find("key", true) != "val" {
		t.Fatal()
	}
}

func TestFindNestedOverride(t *testing.T) {
	Reset()
	Set("a.b", "val")
	if v.find("a.b", true) != "val" {
		t.Fatal()
	}
}

func TestFindNestedConfig(t *testing.T) {
	Reset()
	SetConfigType("json")
	ReadConfig(strings.NewReader(`{"a": {"b": "val"}}`))
	if v.find("a.b", true) != "val" {
		t.Fatal()
	}
}

func TestFindNestedDefaults(t *testing.T) {
	Reset()
	SetDefault("a.b", "val")
	if v.find("a.b", true) != "val" {
		t.Fatal()
	}
}

func TestFindNestedKVStore(t *testing.T) {
	Reset()
	v.kvstore["a"] = map[string]any{"b": "val"}
	if v.find("a.b", true) != "val" {
		t.Fatal()
	}
}

func TestReadInConfigReadError(t *testing.T) {
	Reset()
	v.SetFS(&readErrorFS{mockFS{files: map[string]string{"/etc/app.json": `{"key":"value"}`}}})
	v.SetConfigFile("/etc/app.json")
	v.SetConfigType("json")
	err := v.ReadInConfig()
	if err == nil {
		t.Fatal("expected error")
	}
}

type errorReadCloser struct{}

func (e *errorReadCloser) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}

func (e *errorReadCloser) Close() error {
	return nil
}

type readErrorFS struct {
	mockFS
}

func (r *readErrorFS) Open(name string) (io.ReadCloser, error) {
	return &errorReadCloser{}, nil
}

func TestWriteConfigAsUnsupportedExt(t *testing.T) {
	Reset()
	SetConfigType("xml")
	err := WriteConfigAs("/tmp/config.xml")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestReadConfigUnsupportedExt(t *testing.T) {
	Reset()
	SetConfigType("xml")
	err := ReadConfig(strings.NewReader("<a/>"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSafeWriteConfigAsUnsupportedExt(t *testing.T) {
	Reset()
	SetConfigType("xml")
	err := SafeWriteConfigAs("/tmp/config.xml")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSafeWriteConfigUnsupportedExt(t *testing.T) {
	Reset()
	AddConfigPath(t.TempDir())
	SetConfigName("app")
	SetConfigType("xml")
	err := SafeWriteConfig()
	if err == nil {
		t.Fatal("expected error")
	}
}

// --- push to 95%+ ---

func TestCodecRegistryYAML(t *testing.T) {
	Reset()
	SetConfigType("yaml")
	Set("key", "value")
	var buf bytes.Buffer
	err := WriteConfigTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "key") {
		t.Fatal()
	}
}

func TestCodecRegistryTOML(t *testing.T) {
	Reset()
	SetConfigType("toml")
	Set("key", "value")
	var buf bytes.Buffer
	err := WriteConfigTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "key") {
		t.Fatal()
	}
}

func TestSearchSliceWithPathPrefixesFastPath(t *testing.T) {
	Reset()
	v.config = map[string]any{
		"list": []any{"a", "b"},
	}
	if GetString("list.0") != "a" {
		t.Fatal()
	}
}

func TestWriteConfigNilConfig(t *testing.T) {
	Reset()
	tmpFile := filepath.Join(t.TempDir(), "out.json")
	SetConfigType("json")
	v.config = nil
	err := WriteConfigAs(tmpFile)
	if err != nil {
		t.Fatal(err)
	}
}

func TestIsPathShadowedInFlatMapDefault(t *testing.T) {
	Reset()
	shadow := v.isPathShadowedInFlatMap([]string{"a", "b"}, 42)
	if shadow != "" {
		t.Fatal()
	}
}

func TestFlattenAndMergeMapMapAnyAny(t *testing.T) {
	Reset()
	m := map[string]any{"a": map[any]any{"b": "v"}}
	shadow := v.flattenAndMergeMap(nil, m, "")
	if !shadow["a.b"] {
		t.Fatal()
	}
}

func TestGetSliceNil(t *testing.T) {
	Reset()
	res := GetSlice[string]("missing")
	if res != nil {
		t.Fatal()
	}
}

func TestMergeInConfigReadError(t *testing.T) {
	Reset()
	v.SetFS(&readErrorFS{mockFS{files: map[string]string{"/etc/app.json": `{"key":"value"}`}}})
	v.SetConfigFile("/etc/app.json")
	v.SetConfigType("json")
	err := v.MergeInConfig()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUnmarshalReaderReadError(t *testing.T) {
	Reset()
	SetConfigType("json")
	err := v.unmarshalReader(&errorReader{}, make(map[string]any))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFindConfigDeepShadow(t *testing.T) {
	Reset()
	v.config = map[string]any{"foo": "shadow"}
	if v.find("foo.bar", true) != nil {
		t.Fatal("expected nil")
	}
}

func TestFindFlagDefaultTypes(t *testing.T) {
	Reset()
	flag1 := mockFlag{name: "list", val: "[a,b]", valType: "stringSlice", changed: false}
	BindFlagValue("list", flag1)
	if len(v.find("list", true).([]string)) != 2 {
		t.Fatal()
	}
	flag2 := mockFlag{name: "map", val: "[k=v]", valType: "stringToString", changed: false}
	BindFlagValue("map", flag2)
	m := v.find("map", true).(map[string]any)
	if m["k"] != "v" {
		t.Fatal()
	}
	flag3 := mockFlag{name: "durs", val: "[1h]", valType: "durationSlice", changed: false}
	BindFlagValue("durs", flag3)
	ds := v.find("durs", true).([]time.Duration)
	if len(ds) != 1 {
		t.Fatal()
	}
}

func TestRegisterAliasRealKey(t *testing.T) {
	Reset()
	RegisterAlias("b", "c")
	RegisterAlias("c", "b")
}

func TestSearchMapWithPathPrefixesMapAnyAny(t *testing.T) {
	Reset()
	v.config = map[string]any{
		"foo": map[any]any{"bar": "val"},
	}
	if GetString("foo.bar") != "val" {
		t.Fatal()
	}
}

func TestMergeConfigMapMapAnyAny(t *testing.T) {
	Reset()
	v.config = map[string]any{"a": map[any]any{"b": "old"}}
	err := MergeConfigMap(map[string]any{"a": map[string]any{"b": "new"}})
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnmarshalExperimentalError(t *testing.T) {
	v := NewWithOptions(ExperimentalBindStruct())
	v.Set("key", "val")
	var ch chan int
	err := v.Unmarshal(&ch)
	if err == nil {
		t.Fatal("expected error")
	}
}

type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}

func TestFindFlagDefaultBool(t *testing.T) {
	Reset()
	flag := mockFlag{name: "flag", val: "true", valType: "bool", changed: false}
	BindFlagValue("flag", flag)
	if v.find("flag", true) != true {
		t.Fatal("expected true")
	}
}

func TestMergeMapsSliceAnyAny(t *testing.T) {
	Reset()
	dst := map[string]any{"a": []any{map[any]any{"b": 1}}}
	src := map[string]any{"a": []any{map[string]any{"b": 2}}}
	mergeMaps(src, dst, nil)
	if dst["a"].([]any)[0].(map[string]any)["b"] != 2 {
		t.Fatal()
	}
}

func TestAbsPathifyRelative(t *testing.T) {
	Reset()
	got := absPathify(v.logger, "./config.yml")
	if got == "./config.yml" {
		t.Fatal("expected absolute path")
	}
}
