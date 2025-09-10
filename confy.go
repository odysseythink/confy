package confy

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"mlib.com/confy/cast"
	"mlib.com/confy/mapstructure"

	"mlib.com/confy/internal/features"
)

// ConfigMarshalError happens when failing to marshal the configuration.
type ConfigMarshalError struct {
	err error
}

// Error returns the formatted configuration error.
func (e ConfigMarshalError) Error() string {
	return fmt.Sprintf("While marshaling config: %s", e.err.Error())
}

var v *Confy

func init() {
	v = New()
}

// UnsupportedConfigError denotes encountering an unsupported
// configuration filetype.
type UnsupportedConfigError string

// Error returns the formatted configuration error.
func (str UnsupportedConfigError) Error() string {
	return fmt.Sprintf("Unsupported Config Type %q", string(str))
}

// ConfigFileNotFoundError denotes failing to find configuration file.
type ConfigFileNotFoundError struct {
	name, locations string
}

// Error returns the formatted configuration error.
func (fnfe ConfigFileNotFoundError) Error() string {
	return fmt.Sprintf("Config File %q Not Found in %q", fnfe.name, fnfe.locations)
}

// ConfigFileAlreadyExistsError denotes failure to write new configuration file.
type ConfigFileAlreadyExistsError string

// Error returns the formatted error when configuration already exists.
func (faee ConfigFileAlreadyExistsError) Error() string {
	return fmt.Sprintf("Config File %q Already Exists", string(faee))
}

// A DecoderConfigOption can be passed to confy.Unmarshal to configure
// mapstructure.DecoderConfig options.
type DecoderConfigOption func(*mapstructure.DecoderConfig)

// DecodeHook returns a DecoderConfigOption which overrides the default
// DecoderConfig.DecodeHook value, the default is:
//
//	 mapstructure.ComposeDecodeHookFunc(
//			mapstructure.StringToTimeDurationHookFunc(),
//			mapstructure.StringToSliceHookFunc(","),
//		)
func DecodeHook(hook mapstructure.DecodeHookFunc) DecoderConfigOption {
	return func(c *mapstructure.DecoderConfig) {
		c.DecodeHook = hook
	}
}

// Confy is a prioritized configuration registry. It
// maintains a set of configuration sources, fetches
// values to populate those, and provides them according
// to the source's priority.
// The priority of the sources is the following:
// 1. overrides
// 4. config file
// 5. key/value store
// 6. defaults
//
// For example, if values from the following sources were loaded:
//
//	Defaults : {
//		"secret": "",
//		"user": "default",
//		"endpoint": "https://localhost"
//	}
//	Config : {
//		"user": "root"
//		"secret": "defaultsecret"
//	}
//
// The resulting config will have the following values:
//
//	{
//		"secret": "somesecretkey",
//		"user": "root",
//		"endpoint": "https://localhost"
//	}
//
// Note: Confys are not safe for concurrent Get() and Set() operations.
type Confy struct {
	// Delimiter that separates a list of keys
	// used to access a nested value in one go
	keyDelim string

	// A set of paths to look for the config file in
	configPaths []string

	// A set of remote providers to search for the configuration
	remoteProviders []*defaultRemoteProvider

	// Name of file to look for inside the path
	configName        string
	configFile        string
	configType        string
	configPermissions os.FileMode

	parents        []string
	config         map[string]any
	override       map[string]any
	defaults       map[string]any
	kvstore        map[string]any
	aliases        map[string]string
	typeByDefValue bool

	onConfigChange func(fsnotify.Event)

	logger *slog.Logger

	encoderRegistry EncoderRegistry
	decoderRegistry DecoderRegistry

	decodeHook mapstructure.DecodeHookFunc

	experimentalFinder     bool
	experimentalBindStruct bool
}

// New returns an initialized Confy instance.
func New() *Confy {
	v := new(Confy)
	v.keyDelim = "."
	v.configName = "config"
	v.configPermissions = os.FileMode(0o644)
	v.config = make(map[string]any)
	v.parents = []string{}
	v.override = make(map[string]any)
	v.defaults = make(map[string]any)
	v.kvstore = make(map[string]any)
	v.aliases = make(map[string]string)
	v.typeByDefValue = false
	v.logger = slog.New(&discardHandler{})

	codecRegistry := NewCodecRegistry()

	v.encoderRegistry = codecRegistry
	v.decoderRegistry = codecRegistry

	v.experimentalFinder = features.Finder
	v.experimentalBindStruct = features.BindStruct

	return v
}

// Option configures Confy using the functional options paradigm popularized by Rob Pike and Dave Cheney.
// If you're unfamiliar with this style,
// see https://commandcenter.blogspot.com/2014/01/self-referential-functions-and-design.html and
// https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis.
type Option interface {
	apply(v *Confy)
}

type optionFunc func(v *Confy)

func (fn optionFunc) apply(v *Confy) {
	fn(v)
}

// KeyDelimiter sets the delimiter used for determining key parts.
// By default it's value is ".".
func KeyDelimiter(d string) Option {
	return optionFunc(func(v *Confy) {
		v.keyDelim = d
	})
}

// StringReplacer applies a set of replacements to a string.
type StringReplacer interface {
	// Replace returns a copy of s with all replacements performed.
	Replace(s string) string
}

// WithDecodeHook sets a default decode hook for mapstructure.
func WithDecodeHook(h mapstructure.DecodeHookFunc) Option {
	return optionFunc(func(v *Confy) {
		if h == nil {
			return
		}

		v.decodeHook = h
	})
}

// NewWithOptions creates a new Confy instance.
func NewWithOptions(opts ...Option) *Confy {
	v := New()

	for _, opt := range opts {
		opt.apply(v)
	}

	return v
}

// SetOptions sets the options on the global Confy instance.
//
// Be careful when using this function: subsequent calls may override options you set.
// It's always better to use a local Confy instance.
func SetOptions(opts ...Option) {
	for _, opt := range opts {
		opt.apply(v)
	}
}

// Reset is intended for testing, will reset all to default settings.
// In the public interface for the confy package so applications
// can use it in their testing as well.
func Reset() {
	v = New()
	SupportedExts = []string{"json", "yaml", "yml"}

	resetRemote()
}

// SupportedExts are universally supported extensions.
var SupportedExts = []string{"json", "yaml", "yml"}

// OnConfigChange sets the event handler that is called when a config file changes.
func OnConfigChange(run func(in fsnotify.Event)) { v.OnConfigChange(run) }

// OnConfigChange sets the event handler that is called when a config file changes.
func (v *Confy) OnConfigChange(run func(in fsnotify.Event)) {
	v.onConfigChange = run
}

// WatchConfig starts watching a config file for changes.
func WatchConfig() { v.WatchConfig() }

// WatchConfig starts watching a config file for changes.
func (v *Confy) WatchConfig() {
	initWG := sync.WaitGroup{}
	initWG.Add(1)
	go func() {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			v.logger.Error(fmt.Sprintf("failed to create watcher: %s", err))
			os.Exit(1)
		}
		defer watcher.Close()
		// we have to watch the entire directory to pick up renames/atomic saves in a cross-platform way
		filename, err := v.getConfigFile()
		if err != nil {
			v.logger.Error(fmt.Sprintf("get config file: %s", err))
			initWG.Done()
			return
		}

		configFile := filepath.Clean(filename)
		configDir, _ := filepath.Split(configFile)
		realConfigFile, _ := filepath.EvalSymlinks(filename)

		eventsWG := sync.WaitGroup{}
		eventsWG.Add(1)
		go func() {
			for {
				select {
				case event, ok := <-watcher.Events:
					if !ok { // 'Events' channel is closed
						eventsWG.Done()
						return
					}
					currentConfigFile, _ := filepath.EvalSymlinks(filename)
					// we only care about the config file with the following cases:
					// 1 - if the config file was modified or created
					// 2 - if the real path to the config file changed (eg: k8s ConfigMap replacement)
					if (filepath.Clean(event.Name) == configFile &&
						(event.Has(fsnotify.Write) || event.Has(fsnotify.Create))) ||
						(currentConfigFile != "" && currentConfigFile != realConfigFile) {
						realConfigFile = currentConfigFile
						err := v.ReadInConfig()
						if err != nil {
							v.logger.Error(fmt.Sprintf("read config file: %s", err))
						}
						if v.onConfigChange != nil {
							v.onConfigChange(event)
						}
					} else if filepath.Clean(event.Name) == configFile && event.Has(fsnotify.Remove) {
						eventsWG.Done()
						return
					}

				case err, ok := <-watcher.Errors:
					if ok { // 'Errors' channel is not closed
						v.logger.Error(fmt.Sprintf("watcher error: %s", err))
					}
					eventsWG.Done()
					return
				}
			}
		}()
		err = watcher.Add(configDir)
		if err != nil {
			v.logger.Error(fmt.Sprintf("failed to add watcher: %s", err))
			initWG.Done()
			return
		}
		initWG.Done()   // done initializing the watch in this go routine, so the parent routine can move on...
		eventsWG.Wait() // now, wait for event loop to end in this go-routine...
	}()
	initWG.Wait() // make sure that the go routine above fully ended before returning
}

// SetConfigFile explicitly defines the path, name and extension of the config file.
// Confy will use this and not check any of the config paths.
func SetConfigFile(in string) { v.SetConfigFile(in) }

func (v *Confy) SetConfigFile(in string) {
	if in != "" {
		v.configFile = in
	}
}

// ConfigFileUsed returns the file used to populate the config registry.
func ConfigFileUsed() string            { return v.ConfigFileUsed() }
func (v *Confy) ConfigFileUsed() string { return v.configFile }

// AddConfigPath adds a path for Confy to search for the config file in.
// Can be called multiple times to define multiple search paths.
func AddConfigPath(in string) { v.AddConfigPath(in) }

func (v *Confy) AddConfigPath(in string) {
	if in != "" {
		absin := absPathify(v.logger, in)

		v.logger.Info("adding path to search paths", "path", absin)
		if !slices.Contains(v.configPaths, absin) {
			v.configPaths = append(v.configPaths, absin)
		}
	}
}

// searchMap recursively searches for a value for path in source map.
// Returns nil if not found.
// Note: This assumes that the path entries and map keys are lower cased.
func (v *Confy) searchMap(source map[string]any, path []string) any {
	if len(path) == 0 {
		return source
	}

	next, ok := source[path[0]]
	if ok {
		// Fast path
		if len(path) == 1 {
			return next
		}

		// Nested case
		switch next := next.(type) {
		case map[any]any:
			return v.searchMap(cast.To[map[string]any](next), path[1:])
		case map[string]any:
			// Type assertion is safe here since it is only reached
			// if the type of `next` is the same as the type being asserted
			return v.searchMap(next, path[1:])
		default:
			// got a value but nested key expected, return "nil" for not found
			return nil
		}
	}
	return nil
}

// searchIndexableWithPathPrefixes recursively searches for a value for path in source map/slice.
//
// While searchMap() considers each path element as a single map key or slice index, this
// function searches for, and prioritizes, merged path elements.
// e.g., if in the source, "foo" is defined with a sub-key "bar", and "foo.bar"
// is also defined, this latter value is returned for path ["foo", "bar"].
//
// This should be useful only at config level (other maps may not contain dots
// in their keys).
//
// Note: This assumes that the path entries and map keys are lower cased.
func (v *Confy) searchIndexableWithPathPrefixes(source any, path []string) any {
	if len(path) == 0 {
		return source
	}

	// search for path prefixes, starting from the longest one
	for i := len(path); i > 0; i-- {
		prefixKey := strings.ToLower(strings.Join(path[0:i], v.keyDelim))

		var val any
		switch sourceIndexable := source.(type) {
		case []any:
			val = v.searchSliceWithPathPrefixes(sourceIndexable, prefixKey, i, path)
		case map[string]any:
			val = v.searchMapWithPathPrefixes(sourceIndexable, prefixKey, i, path)
		}
		if val != nil {
			return val
		}
	}

	// not found
	return nil
}

// searchSliceWithPathPrefixes searches for a value for path in sourceSlice
//
// This function is part of the searchIndexableWithPathPrefixes recurring search and
// should not be called directly from functions other than searchIndexableWithPathPrefixes.
func (v *Confy) searchSliceWithPathPrefixes(
	sourceSlice []any,
	prefixKey string,
	pathIndex int,
	path []string,
) any {
	// if the prefixKey is not a number or it is out of bounds of the slice
	index, err := strconv.Atoi(prefixKey)
	if err != nil || len(sourceSlice) <= index {
		return nil
	}

	next := sourceSlice[index]

	// Fast path
	if pathIndex == len(path) {
		return next
	}

	switch n := next.(type) {
	case map[any]any:
		return v.searchIndexableWithPathPrefixes(cast.ToStringMap(n), path[pathIndex:])
	case map[string]any, []any:
		return v.searchIndexableWithPathPrefixes(n, path[pathIndex:])
	default:
		// got a value but nested key expected, do nothing and look for next prefix
	}

	// not found
	return nil
}

// searchMapWithPathPrefixes searches for a value for path in sourceMap
//
// This function is part of the searchIndexableWithPathPrefixes recurring search and
// should not be called directly from functions other than searchIndexableWithPathPrefixes.
func (v *Confy) searchMapWithPathPrefixes(
	sourceMap map[string]any,
	prefixKey string,
	pathIndex int,
	path []string,
) any {
	next, ok := sourceMap[prefixKey]
	if !ok {
		return nil
	}

	// Fast path
	if pathIndex == len(path) {
		return next
	}

	// Nested case
	switch n := next.(type) {
	case map[any]any:
		return v.searchIndexableWithPathPrefixes(cast.ToStringMap(n), path[pathIndex:])
	case map[string]any, []any:
		return v.searchIndexableWithPathPrefixes(n, path[pathIndex:])
	default:
		// got a value but nested key expected, do nothing and look for next prefix
	}

	// not found
	return nil
}

// isPathShadowedInDeepMap makes sure the given path is not shadowed somewhere
// on its path in the map.
// e.g., if "foo.bar" has a value in the given map, it “shadows”
//
//	"foo.bar.baz" in a lower-priority map
func (v *Confy) isPathShadowedInDeepMap(path []string, m map[string]any) string {
	var parentVal any
	for i := 1; i < len(path); i++ {
		parentVal = v.searchMap(m, path[0:i])
		if parentVal == nil {
			// not found, no need to add more path elements
			return ""
		}
		switch parentVal.(type) {
		case map[any]any:
			continue
		case map[string]any:
			continue
		default:
			// parentVal is a regular value which shadows "path"
			return strings.Join(path[0:i], v.keyDelim)
		}
	}
	return ""
}

// isPathShadowedInFlatMap makes sure the given path is not shadowed somewhere
// in a sub-path of the map.
// e.g., if "foo.bar" has a value in the given map, it “shadows”
//
//	"foo.bar.baz" in a lower-priority map
func (v *Confy) isPathShadowedInFlatMap(path []string, mi any) string {
	// unify input map
	var m map[string]interface{}
	switch miv := mi.(type) {
	case map[string]string:
		m = castMapStringToMapInterface(miv)
	default:
		return ""
	}

	// scan paths
	var parentKey string
	for i := 1; i < len(path); i++ {
		parentKey = strings.Join(path[0:i], v.keyDelim)
		if _, ok := m[parentKey]; ok {
			return parentKey
		}
	}
	return ""
}

// SetTypeByDefaultValue enables or disables the inference of a key value's
// type when the Get function is used based upon a key's default value as
// opposed to the value returned based on the normal fetch logic.
//
// For example, if a key has a default value of []string{} and the same key
// is set via an environment variable to "a b c", a call to the Get function
// would return a string slice for the key if the key's type is inferred by
// the default value and the Get function would return:
//
//	[]string {"a", "b", "c"}
//
// Otherwise the Get function would return:
//
//	"a b c"
func SetTypeByDefaultValue(enable bool) { v.SetTypeByDefaultValue(enable) }

func (v *Confy) SetTypeByDefaultValue(enable bool) {
	v.typeByDefValue = enable
}

// GetConfy gets the global Confy instance.
func GetConfy() *Confy {
	return v
}

// Get can retrieve any value given the key to use.
// Get is case-insensitive for a key.
// Get has the behavior of returning the value associated with the first
// place from where it is set. Confy will check in the following order:
// override, config file, key/value store, default
//
// Get returns an interface. For a specific value use one of the Get____ methods.
func Get(key string) any { return v.Get(key) }

func (v *Confy) Get(key string) any {
	lcaseKey := strings.ToLower(key)
	val := v.find(lcaseKey)
	if val == nil {
		return nil
	}

	if v.typeByDefValue {
		// TODO(bep) this branch isn't covered by a single test.
		valType := val
		path := strings.Split(lcaseKey, v.keyDelim)
		defVal := v.searchMap(v.defaults, path)
		if defVal != nil {
			valType = defVal
		}

		switch valType.(type) {
		case bool:
			return cast.ToBool(val)
		case string:
			return cast.ToString(val)
		case int32, int16, int8, int:
			return cast.ToInt(val)
		case uint:
			return cast.ToUint(val)
		case uint32:
			return cast.ToUint32(val)
		case uint64:
			return cast.ToUint64(val)
		case int64:
			return cast.ToInt64(val)
		case float64, float32:
			return cast.ToFloat64(val)
		case time.Time:
			return cast.ToTime(val)
		case time.Duration:
			return cast.ToDuration(val)
		case []string:
			return cast.ToStringSlice(val)
		case []int:
			return cast.ToIntSlice(val)
		case []time.Duration:
			return cast.ToDurationSlice(val)
		}
	}

	return val
}

// Sub returns new Confy instance representing a sub tree of this instance.
// Sub is case-insensitive for a key.
func Sub(key string) *Confy { return v.Sub(key) }

func (v *Confy) Sub(key string) *Confy {
	subv := New()
	data := v.Get(key)
	if data == nil {
		return nil
	}

	if reflect.TypeOf(data).Kind() == reflect.Map {
		subv.parents = append([]string(nil), v.parents...)
		subv.parents = append(subv.parents, strings.ToLower(key))
		subv.keyDelim = v.keyDelim
		subv.config = cast.ToStringMap(data)
		return subv
	}
	return nil
}

// GetString returns the value associated with the key as a string.
func GetString(key string) string { return v.GetString(key) }

func (v *Confy) GetString(key string) string {
	return cast.ToString(v.Get(key))
}

// GetBool returns the value associated with the key as a boolean.
func GetBool(key string) bool { return v.GetBool(key) }

func (v *Confy) GetBool(key string) bool {
	return cast.ToBool(v.Get(key))
}

// GetInt returns the value associated with the key as an integer.
func GetInt(key string) int { return v.GetInt(key) }

func (v *Confy) GetInt(key string) int {
	return cast.ToInt(v.Get(key))
}

// GetInt32 returns the value associated with the key as an integer.
func GetInt32(key string) int32 { return v.GetInt32(key) }

func (v *Confy) GetInt32(key string) int32 {
	return cast.ToInt32(v.Get(key))
}

// GetInt64 returns the value associated with the key as an integer.
func GetInt64(key string) int64 { return v.GetInt64(key) }

func (v *Confy) GetInt64(key string) int64 {
	return cast.ToInt64(v.Get(key))
}

// GetUint8 returns the value associated with the key as an unsigned integer.
func GetUint8(key string) uint8 { return v.GetUint8(key) }

func (v *Confy) GetUint8(key string) uint8 {
	return cast.ToUint8(v.Get(key))
}

// GetUint returns the value associated with the key as an unsigned integer.
func GetUint(key string) uint { return v.GetUint(key) }

func (v *Confy) GetUint(key string) uint {
	return cast.ToUint(v.Get(key))
}

// GetUint16 returns the value associated with the key as an unsigned integer.
func GetUint16(key string) uint16 { return v.GetUint16(key) }

func (v *Confy) GetUint16(key string) uint16 {
	return cast.ToUint16(v.Get(key))
}

// GetUint32 returns the value associated with the key as an unsigned integer.
func GetUint32(key string) uint32 { return v.GetUint32(key) }

func (v *Confy) GetUint32(key string) uint32 {
	return cast.ToUint32(v.Get(key))
}

// GetUint64 returns the value associated with the key as an unsigned integer.
func GetUint64(key string) uint64 { return v.GetUint64(key) }

func (v *Confy) GetUint64(key string) uint64 {
	return cast.ToUint64(v.Get(key))
}

// GetFloat64 returns the value associated with the key as a float64.
func GetFloat64(key string) float64 { return v.GetFloat64(key) }

func (v *Confy) GetFloat64(key string) float64 {
	return cast.ToFloat64(v.Get(key))
}

// GetTime returns the value associated with the key as time.
func GetTime(key string) time.Time { return v.GetTime(key) }

func (v *Confy) GetTime(key string) time.Time {
	return cast.ToTime(v.Get(key))
}

// GetDuration returns the value associated with the key as a duration.
func GetDuration(key string) time.Duration { return v.GetDuration(key) }

func (v *Confy) GetDuration(key string) time.Duration {
	return cast.ToDuration(v.Get(key))
}

// GetIntSlice returns the value associated with the key as a slice of int values.
func GetIntSlice(key string) []int { return v.GetIntSlice(key) }

func (v *Confy) GetIntSlice(key string) []int {
	return cast.ToIntSlice(v.Get(key))
}

// GetStringSlice returns the value associated with the key as a slice of strings.
func GetStringSlice(key string) []string { return v.GetStringSlice(key) }

func (v *Confy) GetStringSlice(key string) []string {
	return cast.ToStringSlice(v.Get(key))
}

// GetStringMap returns the value associated with the key as a map of interfaces.
func GetStringMap(key string) map[string]any { return v.GetStringMap(key) }

func (v *Confy) GetStringMap(key string) map[string]any {
	return cast.ToStringMap(v.Get(key))
}

// GetStringMapString returns the value associated with the key as a map of strings.
func GetStringMapString(key string) map[string]string { return v.GetStringMapString(key) }

func (v *Confy) GetStringMapString(key string) map[string]string {
	return cast.ToStringMapString(v.Get(key))
}

// GetStringMapStringSlice returns the value associated with the key as a map to a slice of strings.
func GetStringMapStringSlice(key string) map[string][]string { return v.GetStringMapStringSlice(key) }

func (v *Confy) GetStringMapStringSlice(key string) map[string][]string {
	return cast.ToStringMapStringSlice(v.Get(key))
}

// GetSizeInBytes returns the size of the value associated with the given key
// in bytes.
func GetSizeInBytes(key string) uint { return v.GetSizeInBytes(key) }

func (v *Confy) GetSizeInBytes(key string) uint {
	sizeStr := cast.ToString(v.Get(key))
	return parseSizeInBytes(sizeStr)
}

// UnmarshalKey takes a single key and unmarshals it into a Struct.
func UnmarshalKey(key string, rawVal any, opts ...DecoderConfigOption) error {
	return v.UnmarshalKey(key, rawVal, opts...)
}

func (v *Confy) UnmarshalKey(key string, rawVal any, opts ...DecoderConfigOption) error {
	return decode(v.Get(key), v.defaultDecoderConfig(rawVal, opts...))
}

// Unmarshal unmarshals the config into a Struct. Make sure that the tags
// on the fields of the structure are properly set.
func Unmarshal(rawVal any, opts ...DecoderConfigOption) error {
	return v.Unmarshal(rawVal, opts...)
}

func (v *Confy) Unmarshal(rawVal any, opts ...DecoderConfigOption) error {
	keys := v.AllKeys()

	if v.experimentalBindStruct {
		// TODO: make this optional?
		structKeys, err := v.decodeStructKeys(rawVal, opts...)
		if err != nil {
			return err
		}

		keys = append(keys, structKeys...)
	}

	// TODO: struct keys should be enough?
	return decode(v.getSettings(keys), v.defaultDecoderConfig(rawVal, opts...))
}

func (v *Confy) decodeStructKeys(input any, opts ...DecoderConfigOption) ([]string, error) {
	var structKeyMap map[string]any

	err := decode(input, v.defaultDecoderConfig(&structKeyMap, opts...))
	if err != nil {
		return nil, err
	}

	flattenedStructKeyMap := v.flattenAndMergeMap(map[string]bool{}, structKeyMap, "")

	r := make([]string, 0, len(flattenedStructKeyMap))
	for v := range flattenedStructKeyMap {
		r = append(r, v)
	}

	return r, nil
}

// defaultDecoderConfig returns default mapstructure.DecoderConfig with support
// of time.Duration values & string slices.
func (v *Confy) defaultDecoderConfig(output any, opts ...DecoderConfigOption) *mapstructure.DecoderConfig {
	decodeHook := v.decodeHook
	if decodeHook == nil {
		decodeHook = mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			// mapstructure.StringToSliceHookFunc(","),
			stringToWeakSliceHookFunc(","),
		)
	}

	c := &mapstructure.DecoderConfig{
		Metadata:         nil,
		WeaklyTypedInput: true,
		DecodeHook:       decodeHook,
	}

	for _, opt := range opts {
		opt(c)
	}

	// Do not allow overwriting the output
	c.Result = output

	return c
}

// As of mapstructure v2.0.0 StringToSliceHookFunc checks if the return type is a string slice.
// This function removes that check.
// TODO: implement a function that checks if the value can be converted to the return type and use it instead.
func stringToWeakSliceHookFunc(sep string) mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{},
	) (interface{}, error) {
		if f.Kind() != reflect.String || t.Kind() != reflect.Slice {
			return data, nil
		}

		raw := data.(string)
		if raw == "" {
			return []string{}, nil
		}

		return strings.Split(raw, sep), nil
	}
}

// decode is a wrapper around mapstructure.Decode that mimics the WeakDecode functionality.
func decode(input any, config *mapstructure.DecoderConfig) error {
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}
	return decoder.Decode(input)
}

// UnmarshalExact unmarshals the config into a Struct, erroring if a field is nonexistent
// in the destination struct.
func UnmarshalExact(rawVal any, opts ...DecoderConfigOption) error {
	return v.UnmarshalExact(rawVal, opts...)
}

func (v *Confy) UnmarshalExact(rawVal any, opts ...DecoderConfigOption) error {
	config := v.defaultDecoderConfig(rawVal, opts...)
	config.ErrorUnused = true

	keys := v.AllKeys()

	if v.experimentalBindStruct {
		// TODO: make this optional?
		structKeys, err := v.decodeStructKeys(rawVal, opts...)
		if err != nil {
			return err
		}

		keys = append(keys, structKeys...)
	}

	// TODO: struct keys should be enough?
	return decode(v.getSettings(keys), config)
}

// Given a key, find the value.
//
// Confy will check to see if an alias exists first.
// Confy will then check in the following order:
// config file, key/value store.
//
// Note: this assumes a lower-cased key given.
func (v *Confy) find(lcaseKey string) any {
	var (
		val    any
		path   = strings.Split(lcaseKey, v.keyDelim)
		nested = len(path) > 1
	)

	// compute the path through the nested maps to the nested value
	if nested && v.isPathShadowedInDeepMap(path, castMapStringToMapInterface(v.aliases)) != "" {
		return nil
	}

	// if the requested key is an alias, then return the proper key
	lcaseKey = v.realKey(lcaseKey)
	path = strings.Split(lcaseKey, v.keyDelim)
	nested = len(path) > 1

	// Set() override first
	val = v.searchMap(v.override, path)
	if val != nil {
		return val
	}
	if nested && v.isPathShadowedInDeepMap(path, v.override) != "" {
		return nil
	}

	// Config file next
	val = v.searchIndexableWithPathPrefixes(v.config, path)
	if val != nil {
		return val
	}
	if nested && v.isPathShadowedInDeepMap(path, v.config) != "" {
		return nil
	}

	// K/V store next
	val = v.searchMap(v.kvstore, path)
	if val != nil {
		return val
	}
	if nested && v.isPathShadowedInDeepMap(path, v.kvstore) != "" {
		return nil
	}

	// Default next
	val = v.searchMap(v.defaults, path)
	if val != nil {
		return val
	}
	if nested && v.isPathShadowedInDeepMap(path, v.defaults) != "" {
		return nil
	}

	return nil
}

func readAsCSV(val string) ([]string, error) {
	if val == "" {
		return []string{}, nil
	}
	stringReader := strings.NewReader(val)
	csvReader := csv.NewReader(stringReader)
	return csvReader.Read()
}

// alterations are: errors are swallowed, map[string]any is returned in order to enable cast.ToStringMap.
func stringToStringConv(val string) any {
	val = strings.Trim(val, "[]")
	// An empty string would cause an empty map
	if val == "" {
		return map[string]any{}
	}
	r := csv.NewReader(strings.NewReader(val))
	ss, err := r.Read()
	if err != nil {
		return nil
	}
	out := make(map[string]any, len(ss))
	for _, pair := range ss {
		k, vv, found := strings.Cut(pair, "=")
		if !found {
			return nil
		}
		out[k] = vv
	}
	return out
}

// alterations are: errors are swallowed, map[string]any is returned in order to enable cast.ToStringMap.
func stringToIntConv(val string) any {
	val = strings.Trim(val, "[]")
	// An empty string would cause an empty map
	if val == "" {
		return map[string]any{}
	}
	ss := strings.Split(val, ",")
	out := make(map[string]any, len(ss))
	for _, pair := range ss {
		k, vv, found := strings.Cut(pair, "=")
		if !found {
			return nil
		}
		var err error
		out[k], err = strconv.Atoi(vv)
		if err != nil {
			return nil
		}
	}
	return out
}

// IsSet checks to see if the key has been set in any of the data locations.
// IsSet is case-insensitive for a key.
func IsSet(key string) bool { return v.IsSet(key) }

func (v *Confy) IsSet(key string) bool {
	lcaseKey := strings.ToLower(key)
	val := v.find(lcaseKey)
	return val != nil
}

// RegisterAlias creates an alias that provides another accessor for the same key.
// This enables one to change a name without breaking the application.
func RegisterAlias(alias, key string) { v.RegisterAlias(alias, key) }

func (v *Confy) RegisterAlias(alias, key string) {
	v.registerAlias(alias, strings.ToLower(key))
}

func (v *Confy) registerAlias(alias, key string) {
	alias = strings.ToLower(alias)
	if alias != key && alias != v.realKey(key) {
		_, exists := v.aliases[alias]

		if !exists {
			// if we alias something that exists in one of the maps to another
			// name, we'll never be able to get that value using the original
			// name, so move the config value to the new realkey.
			if val, ok := v.config[alias]; ok {
				delete(v.config, alias)
				v.config[key] = val
			}
			if val, ok := v.kvstore[alias]; ok {
				delete(v.kvstore, alias)
				v.kvstore[key] = val
			}
			if val, ok := v.defaults[alias]; ok {
				delete(v.defaults, alias)
				v.defaults[key] = val
			}
			if val, ok := v.override[alias]; ok {
				delete(v.override, alias)
				v.override[key] = val
			}
			v.aliases[alias] = key
		}
	} else {
		v.logger.Warn("creating circular reference alias", "alias", alias, "key", key, "real_key", v.realKey(key))
	}
}

func (v *Confy) realKey(key string) string {
	newkey, exists := v.aliases[key]
	if exists {
		v.logger.Debug("key is an alias", "alias", key, "to", newkey)

		return v.realKey(newkey)
	}
	return key
}

// InConfig checks to see if the given key (or an alias) is in the config file.
func InConfig(key string) bool { return v.InConfig(key) }

func (v *Confy) InConfig(key string) bool {
	lcaseKey := strings.ToLower(key)

	// if the requested key is an alias, then return the proper key
	lcaseKey = v.realKey(lcaseKey)
	path := strings.Split(lcaseKey, v.keyDelim)

	return v.searchIndexableWithPathPrefixes(v.config, path) != nil
}

// SetDefault sets the default value for this key.
// SetDefault is case-insensitive for a key.
// Default only used when no value is provided by the user via flag, config or ENV.
func SetDefault(key string, value any) { v.SetDefault(key, value) }

func (v *Confy) SetDefault(key string, value any) {
	// If alias passed in, then set the proper default
	key = v.realKey(strings.ToLower(key))
	value = toCaseInsensitiveValue(value)

	path := strings.Split(key, v.keyDelim)
	lastKey := strings.ToLower(path[len(path)-1])
	deepestMap := deepSearch(v.defaults, path[0:len(path)-1])

	// set innermost value
	deepestMap[lastKey] = value
}

// Set sets the value for the key in the override register.
// Set is case-insensitive for a key.
// Will be used instead of values obtained via
// flags, config file, ENV, default, or key/value store.
func Set(key string, value any) { v.Set(key, value) }

func (v *Confy) Set(key string, value any) {
	// If alias passed in, then set the proper override
	key = v.realKey(strings.ToLower(key))
	value = toCaseInsensitiveValue(value)

	path := strings.Split(key, v.keyDelim)
	lastKey := strings.ToLower(path[len(path)-1])
	deepestMap := deepSearch(v.override, path[0:len(path)-1])

	// set innermost value
	deepestMap[lastKey] = value
}

// ReadInConfig will discover and load the configuration file from disk
// and key/value stores, searching in one of the defined paths.
func ReadInConfig() error { return v.ReadInConfig() }

func (v *Confy) ReadInConfig() error {
	v.logger.Info("attempting to read in config file")
	filename, err := v.getConfigFile()
	if err != nil {
		return err
	}

	if !slices.Contains(SupportedExts, v.getConfigType()) {
		return UnsupportedConfigError(v.getConfigType())
	}

	v.logger.Debug("reading file", "file", filename)
	file, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	config := make(map[string]any)

	err = v.unmarshalReader(bytes.NewReader(file), config)
	if err != nil {
		return err
	}

	v.config = config
	return nil
}

// MergeInConfig merges a new configuration with an existing config.
func MergeInConfig() error { return v.MergeInConfig() }

func (v *Confy) MergeInConfig() error {
	v.logger.Info("attempting to merge in config file")
	filename, err := v.getConfigFile()
	if err != nil {
		return err
	}

	if !slices.Contains(SupportedExts, v.getConfigType()) {
		return UnsupportedConfigError(v.getConfigType())
	}

	file, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	return v.MergeConfig(bytes.NewReader(file))
}

// ReadConfig will read a configuration file, setting existing keys to nil if the
// key does not exist in the file.
func ReadConfig(in io.Reader) error { return v.ReadConfig(in) }

func (v *Confy) ReadConfig(in io.Reader) error {
	config := make(map[string]any)

	err := v.unmarshalReader(in, config)
	if err != nil {
		return err
	}

	v.config = config

	return nil
}

// MergeConfig merges a new configuration with an existing config.
func MergeConfig(in io.Reader) error { return v.MergeConfig(in) }

func (v *Confy) MergeConfig(in io.Reader) error {
	config := make(map[string]any)

	if err := v.unmarshalReader(in, config); err != nil {
		return err
	}

	return v.MergeConfigMap(config)
}

// MergeConfigMap merges the configuration from the map given with an existing config.
// Note that the map given may be modified.
func MergeConfigMap(cfg map[string]any) error { return v.MergeConfigMap(cfg) }

func (v *Confy) MergeConfigMap(cfg map[string]any) error {
	if v.config == nil {
		v.config = make(map[string]any)
	}
	insensitiviseMap(cfg)
	mergeMaps(cfg, v.config, nil)
	return nil
}

// WriteConfig writes the current configuration to a file.
func WriteConfig() error { return v.WriteConfig() }

func (v *Confy) WriteConfig() error {
	filename, err := v.getConfigFile()
	if err != nil {
		return err
	}
	return v.writeConfig(filename, true)
}

// SafeWriteConfig writes current configuration to file only if the file does not exist.
func SafeWriteConfig() error { return v.SafeWriteConfig() }

func (v *Confy) SafeWriteConfig() error {
	if len(v.configPaths) < 1 {
		return errors.New("missing configuration for 'configPath'")
	}
	return v.SafeWriteConfigAs(filepath.Join(v.configPaths[0], v.configName+"."+v.configType))
}

// WriteConfigAs writes current configuration to a given filename.
func WriteConfigAs(filename string) error { return v.WriteConfigAs(filename) }

func (v *Confy) WriteConfigAs(filename string) error {
	return v.writeConfig(filename, true)
}

// WriteConfigTo writes current configuration to an [io.Writer].
func WriteConfigTo(w io.Writer) error { return v.WriteConfigTo(w) }

func (v *Confy) WriteConfigTo(w io.Writer) error {
	format := strings.ToLower(v.getConfigType())

	if !slices.Contains(SupportedExts, format) {
		return UnsupportedConfigError(format)
	}

	return v.marshalWriter(w, format)
}

// SafeWriteConfigAs writes current configuration to a given filename if it does not exist.
func SafeWriteConfigAs(filename string) error { return v.SafeWriteConfigAs(filename) }

func (v *Confy) SafeWriteConfigAs(filename string) error {
	alreadyExists, err := exists(filename)
	if alreadyExists && err == nil {
		return ConfigFileAlreadyExistsError(filename)
	}
	return v.writeConfig(filename, false)
}

func (v *Confy) writeConfig(filename string, force bool) error {
	v.logger.Info("attempting to write configuration to file")

	var configType string

	ext := filepath.Ext(filename)
	if ext != "" && ext != filepath.Base(filename) {
		configType = ext[1:]
	} else {
		configType = v.configType
	}
	if configType == "" {
		return fmt.Errorf("config type could not be determined for %s", filename)
	}

	if !slices.Contains(SupportedExts, configType) {
		return UnsupportedConfigError(configType)
	}
	if v.config == nil {
		v.config = make(map[string]any)
	}
	flags := os.O_CREATE | os.O_TRUNC | os.O_WRONLY
	if !force {
		flags |= os.O_EXCL
	}
	f, err := os.OpenFile(filename, flags, v.configPermissions)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := v.marshalWriter(f, configType); err != nil {
		return err
	}

	return f.Sync()
}

func (v *Confy) unmarshalReader(in io.Reader, c map[string]any) error {
	format := strings.ToLower(v.getConfigType())
	if format == "" {
		return errors.New("cannot decode configuration: unable to determine config type")
	}

	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(in)
	if err != nil {
		return fmt.Errorf("failed to read configuration from input: %w", err)
	}

	// TODO: remove this once SupportedExts is deprecated/removed
	if !slices.Contains(SupportedExts, format) {
		return UnsupportedConfigError(format)
	}

	// TODO: return [UnsupportedConfigError] if the registry does not contain the format
	// TODO: consider deprecating this error type
	decoder, err := v.decoderRegistry.Decoder(format)
	if err != nil {
		return ConfigParseError{err}
	}

	err = decoder.Decode(buf.Bytes(), c)
	if err != nil {
		return ConfigParseError{err}
	}

	insensitiviseMap(c)
	return nil
}

// Marshal a map into Writer.
func (v *Confy) marshalWriter(w io.Writer, configType string) error {
	c := v.AllSettings()

	encoder, err := v.encoderRegistry.Encoder(configType)
	if err != nil {
		return ConfigMarshalError{err}
	}

	b, err := encoder.Encode(c)
	if err != nil {
		return ConfigMarshalError{err}
	}

	_, err = w.Write(b)
	if err != nil {
		return ConfigMarshalError{err}
	}

	return nil
}

func keyExists(k string, m map[string]any) string {
	lk := strings.ToLower(k)
	for mk := range m {
		lmk := strings.ToLower(mk)
		if lmk == lk {
			return mk
		}
	}
	return ""
}

func castToMapStringInterface(
	src map[any]any,
) map[string]any {
	tgt := map[string]any{}
	for k, v := range src {
		tgt[fmt.Sprintf("%v", k)] = v
	}
	return tgt
}

func castMapStringSliceToMapInterface(src map[string][]string) map[string]any {
	tgt := map[string]any{}
	for k, v := range src {
		tgt[k] = v
	}
	return tgt
}

func castMapStringToMapInterface(src map[string]string) map[string]any {
	tgt := map[string]any{}
	for k, v := range src {
		tgt[k] = v
	}
	return tgt
}

// mergeMaps merges two maps. The `itgt` parameter is for handling go-yaml's
// insistence on parsing nested structures as `map[any]any`
// instead of using a `string` as the key for nest structures beyond one level
// deep. Both map types are supported as there is a go-yaml fork that uses
// `map[string]any` instead.
func mergeMaps(src, tgt map[string]any, itgt map[any]any) {
	for sk, sv := range src {
		tk := keyExists(sk, tgt)
		if tk == "" {
			v.logger.Debug("", "tk", "\"\"", fmt.Sprintf("tgt[%s]", sk), sv)
			tgt[sk] = sv
			if itgt != nil {
				itgt[sk] = sv
			}
			continue
		}

		tv, ok := tgt[tk]
		if !ok {
			v.logger.Debug("", fmt.Sprintf("ok[%s]", tk), false, fmt.Sprintf("tgt[%s]", sk), sv)
			tgt[sk] = sv
			if itgt != nil {
				itgt[sk] = sv
			}
			continue
		}

		svType := reflect.TypeOf(sv)
		tvType := reflect.TypeOf(tv)

		v.logger.Debug(
			"processing",
			"key", sk,
			"st", svType,
			"tt", tvType,
			"sv", sv,
			"tv", tv,
		)

		switch ttv := tv.(type) {
		case map[any]any:
			v.logger.Debug("merging maps (must convert)")
			tsv, ok := sv.(map[any]any)
			if !ok {
				v.logger.Error(
					"Could not cast sv to map[any]any",
					"key", sk,
					"st", svType,
					"tt", tvType,
					"sv", sv,
					"tv", tv,
				)
				continue
			}

			ssv := castToMapStringInterface(tsv)
			stv := castToMapStringInterface(ttv)
			mergeMaps(ssv, stv, ttv)
		case map[string]any:
			v.logger.Debug("merging maps")
			tsv, ok := sv.(map[string]any)
			if !ok {
				v.logger.Error(
					"Could not cast sv to map[string]any",
					"key", sk,
					"st", svType,
					"tt", tvType,
					"sv", sv,
					"tv", tv,
				)
				continue
			}
			mergeMaps(tsv, ttv, nil)
		default:
			v.logger.Debug("setting value")
			tgt[tk] = sv
			if itgt != nil {
				itgt[tk] = sv
			}
		}
	}
}

// AllKeys returns all keys holding a value, regardless of where they are set.
// Nested keys are returned with a v.keyDelim separator.
func AllKeys() []string { return v.AllKeys() }

func (v *Confy) AllKeys() []string {
	m := map[string]bool{}
	// add all paths, by order of descending priority to ensure correct shadowing
	m = v.flattenAndMergeMap(m, castMapStringToMapInterface(v.aliases), "")
	m = v.flattenAndMergeMap(m, v.override, "")
	m = v.flattenAndMergeMap(m, v.config, "")
	m = v.flattenAndMergeMap(m, v.kvstore, "")
	m = v.flattenAndMergeMap(m, v.defaults, "")

	// convert set of paths to list
	a := make([]string, 0, len(m))
	for x := range m {
		a = append(a, x)
	}
	return a
}

// flattenAndMergeMap recursively flattens the given map into a map[string]bool
// of key paths (used as a set, easier to manipulate than a []string):
//   - each path is merged into a single key string, delimited with v.keyDelim
//   - if a path is shadowed by an earlier value in the initial shadow map,
//     it is skipped.
//
// The resulting set of paths is merged to the given shadow set at the same time.
func (v *Confy) flattenAndMergeMap(shadow map[string]bool, m map[string]any, prefix string) map[string]bool {
	if shadow != nil && prefix != "" && shadow[prefix] {
		// prefix is shadowed => nothing more to flatten
		return shadow
	}
	if shadow == nil {
		shadow = make(map[string]bool)
	}

	var m2 map[string]any
	if prefix != "" {
		prefix += v.keyDelim
	}
	for k, val := range m {
		fullKey := prefix + k
		switch val := val.(type) {
		case map[string]any:
			m2 = val
		case map[any]any:
			m2 = cast.ToStringMap(val)
		default:
			// immediate value
			shadow[strings.ToLower(fullKey)] = true
			continue
		}
		// recursively merge to shadow map
		shadow = v.flattenAndMergeMap(shadow, m2, fullKey)
	}
	return shadow
}

// mergeFlatMap merges the given maps, excluding values of the second map
// shadowed by values from the first map.
func (v *Confy) mergeFlatMap(shadow map[string]bool, m map[string]any) map[string]bool {
	// scan keys
outer:
	for k := range m {
		path := strings.Split(k, v.keyDelim)
		// scan intermediate paths
		var parentKey string
		for i := 1; i < len(path); i++ {
			parentKey = strings.Join(path[0:i], v.keyDelim)
			if shadow[parentKey] {
				// path is shadowed, continue
				continue outer
			}
		}
		// add key
		shadow[strings.ToLower(k)] = true
	}
	return shadow
}

// AllSettings merges all settings and returns them as a map[string]any.
func AllSettings() map[string]any { return v.AllSettings() }

func (v *Confy) AllSettings() map[string]any {
	return v.getSettings(v.AllKeys())
}

func (v *Confy) getSettings(keys []string) map[string]any {
	m := map[string]any{}
	// start from the list of keys, and construct the map one value at a time
	for _, k := range keys {
		value := v.Get(k)
		if value == nil {
			// should not happen, since AllKeys() returns only keys holding a value,
			// check just in case anything changes
			continue
		}
		path := strings.Split(k, v.keyDelim)
		lastKey := strings.ToLower(path[len(path)-1])
		deepestMap := deepSearch(m, path[0:len(path)-1])
		// set innermost value
		deepestMap[lastKey] = value
	}
	return m
}

// SetConfigName sets name for the config file.
// Does not include extension.
func SetConfigName(in string) { v.SetConfigName(in) }

func (v *Confy) SetConfigName(in string) {
	if in != "" {
		v.configName = in
		v.configFile = ""
	}
}

// SetConfigType sets the type of the configuration returned by the
// remote source, e.g. "json".
func SetConfigType(in string) { v.SetConfigType(in) }

func (v *Confy) SetConfigType(in string) {
	if in != "" {
		v.configType = in
	}
}

// SetConfigPermissions sets the permissions for the config file.
func SetConfigPermissions(perm os.FileMode) { v.SetConfigPermissions(perm) }

func (v *Confy) SetConfigPermissions(perm os.FileMode) {
	v.configPermissions = perm.Perm()
}

func (v *Confy) getConfigType() string {
	if v.configType != "" {
		return v.configType
	}

	cf, err := v.getConfigFile()
	if err != nil {
		return ""
	}

	ext := filepath.Ext(cf)

	if len(ext) > 1 {
		return ext[1:]
	}

	return ""
}

func (v *Confy) getConfigFile() (string, error) {
	if v.configFile == "" {
		return "", errors.New("no config file provided")
	}
	return v.configFile, nil
}

// Debug prints all configuration registries for debugging
// purposes.
func Debug()              { v.Debug() }
func DebugTo(w io.Writer) { v.DebugTo(w) }

func (v *Confy) Debug() { v.DebugTo(os.Stdout) }

func (v *Confy) DebugTo(w io.Writer) {
	fmt.Fprintf(w, "Aliases:\n%#v\n", v.aliases)
	fmt.Fprintf(w, "Override:\n%#v\n", v.override)
	fmt.Fprintf(w, "Key/Value Store:\n%#v\n", v.kvstore)
	fmt.Fprintf(w, "Config:\n%#v\n", v.config)
	fmt.Fprintf(w, "Defaults:\n%#v\n", v.defaults)
}
