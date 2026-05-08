package confy

import (
	"bytes"
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
