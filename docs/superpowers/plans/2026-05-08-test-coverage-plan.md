# Test Coverage补齐 Plan (Target: 95%+)

## Current State
- Overall: 42.8%
- confy main: 37.2% (209 uncovered functions)
- cast: 0.0%
- pflag: 50.0%
- mapstructure: 92.0%
- remote: 0.0%

## Parallel Tasks

### Task A: 核心 confy 包测试 (confy.go + util.go + remote.go + file.go + flags.go + encoding.go + codecs)
Target: 95%+
Key uncovered areas:
- Get/GetWithDefault/GetSlice generics and all typed getters
- WatchConfig
- SafeWriteConfig/SafeWriteConfigAs/WriteConfigTo
- MergeMaps/mergeFlatMap/copyAndInsensitiviseMap
- All error type Error() methods
- RegisterAlias complex scenarios
- Size parsing (parseSizeInBytes, safeMul)
- absPathify, userHomeDir, deepSearch edge cases
- Config permissions
- Remote provider methods

### Task B: cast 包测试
Target: 95%+
Files: cast/alias.go, cast/basic.go, cast/cast.go, cast/indirect.go, cast/map.go, cast/number.go, cast/slice.go, cast/time.go, cast/zz_generated.go, cast/internal/*.go

### Task C: pflag 包 + 其他子包补齐
Target: 95%+
- pflag/BindPFlags test
- mapstructure: edge cases to reach 95%
- remote.go error types and provider methods
