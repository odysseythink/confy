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
