# GoEnvy

[![Go Reference](https://pkg.go.dev/badge/github.com/yassine-manai/GoEnvy.svg)](https://pkg.go.dev/github.com/yassine-manai/GoEnvy)
[![Go Report Card](https://goreportcard.com/badge/github.com/yassine-manai/GoEnvy)](https://goreportcard.com/report/github.com/yassine-manai/GoEnvy)
[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

A lightweight, type-safe environment variable loader for Go applications with automatic struct field mapping using tags.

**GitHub:** [github.com/yassine-manai/GoEnvy](https://github.com/yassine-manai/GoEnvy)

---

## What is GoEnvy?

GoEnvy reads environment variables (or a `.env` file) and fills your Go struct automatically. You just write `env:"KEY_NAME"` tags on your struct fields, and GoEnvy does the rest — looking up the variable, parsing it to the right type, and handling errors.

**No manual `os.Getenv`, no messy `strconv` calls, no boilerplate.**

---

## Features

- Simple `env` tag-based configuration mapping
- Support for 10 Go types: `string`, `int`, `bool`, `float64`, `time.Duration`, `[]string`, `[]int`, `[]float64`, `map[string]string`, `time.Time`
- Nested struct support (group config by category: Server, Database, etc.)
- Default value handling (optional variables)
- Required variables (error if missing)
- `.env` file support (via godotenv)
- Reflection-based automatic type casting (you don't need to parse anything)

---

## Installation

Open your terminal in your Go project and run:

```bash
go get github.com/yassine-manai/GoEnvy
```

This downloads the library and adds it to your `go.mod` file.

### Requirements

- Go 1.16 or higher

---

## Quick Start (Complete Example)

### 1. Create a `.env` file in your project root

```
# .env
APP_HOST=0.0.0.0
APP_PORT=9000
APP_DEBUG=true
APP_TIMEOUT=45s
APP_TAGS=api,web,admin
```

### 2. Create `main.go`

```go
package main

import (
	"fmt"
	"log"
	"time"

	goenvy "github.com/yassine-manai/GoEnvy"
)

// Config holds your application configuration.
// Each field has an env tag that tells GoEnvy which env variable to read.
// The value after the comma is the default (used if the env var is not set).
type Config struct {
	Host    string        `env:"APP_HOST,localhost"`
	Port    int           `env:"APP_PORT,8080"`
	Debug   bool          `env:"APP_DEBUG,false"`
	Timeout time.Duration `env:"APP_TIMEOUT,30s"`
	Tags    []string      `env:"APP_TAGS,go,env,loader"`
}

func main() {
	var cfg Config

	// LoadEnvVariable reads .env (if it exists) + actual env vars + fills the struct
	err := goenvy.LoadEnvVariable(&cfg)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// All fields are now populated with the correct types
	fmt.Printf("Host:    %s\n", cfg.Host)
	fmt.Printf("Port:    %d\n", cfg.Port)
	fmt.Printf("Debug:   %v\n", cfg.Debug)
	fmt.Printf("Timeout: %v\n", cfg.Timeout)
	fmt.Printf("Tags:    %v\n", cfg.Tags)
}
```

### 3. Run it

```bash
go run main.go
```

**Output:**
```
Host:    0.0.0.0
Port:    9000
Debug:   true
Timeout: 45s
Tags:    [api web admin]
```

Now try deleting the `.env` file and run again — you'll see the defaults kick in:
```
Host:    localhost
Port:    8080
Debug:   false
Timeout: 30s
Tags:    [go env loader]
```

> **Tip:** The `.env` file is optional. GoEnvy logs a warning if it's missing but continues normally using env vars or defaults.

---

## Understanding Tags

Tags are Go struct annotations (the backtick-quoted text after the field type). They tell GoEnvy how to map your struct fields to environment variables.

```go
type Example struct {
	FieldName type `env:"ENV_VAR_NAME,default_value"`
}
//                  └──────────┬──────────┘
//                  key        └── default (optional)
```

| Tag | Meaning | Example |
|-----|---------|---------|
| `env:"PORT"` | **Required.** Error if `PORT` is not set | — |
| `env:"PORT,8080"` | **Optional.** Uses `8080` if `PORT` is not set | `.env` file: `PORT=3000` → result: `3000` |

If you set the env var (or put it in `.env`), that value is used. If you don't, the default is used. If there's no default and the var is missing, you get an error.

---

## Supported Types — Complete Reference

| Go Type in Struct | What to write in `.env` | Example `.env` value |
|-------------------|------------------------|---------------------|
| `string` | Plain text | `HELLO=hello world` |
| `int` | Whole number | `PORT=8080` |
| `bool` | `true` / `false` / `1` / `0` / `t` / `f` | `DEBUG=true` |
| `float64` | Decimal number | `RATE=3.14` |
| `time.Duration` | Go duration string: `Ns`, `Nms`, `Ns`, `Nm`, `Nh` | `TIMEOUT=5s` |
| `[]string` | Comma-separated values | `TAGS=a,b,c` |
| `[]int` | Comma-separated numbers | `IDS=1,2,3` |
| `[]float64` | Comma-separated decimals | `PRICES=1.99,2.99` |
| `map[string]string` | Comma-separated `key=value` pairs | `MAP=host=a,port=b` |
| `time.Time` | Date/time string | `DATE=2024-01-15T10:30:00Z` |

### Type Examples in Code

```go
type AllTypes struct {
	Name     string            `env:"NAME,default_name"`       // string
	Count    int               `env:"COUNT,100"`               // integer
	Enabled  bool              `env:"ENABLED,true"`            // boolean
	Price    float64           `env:"PRICE,19.99"`             // float
	Duration time.Duration     `env:"DURATION,30s"`            // duration
	Tags     []string          `env:"TAGS,tag1,tag2"`          // string slice
	Scores   []int             `env:"SCORES,10,20,30"`         // int slice
	Ratios   []float64         `env:"RATIOS,1.5,2.5"`          // float64 slice
	Meta     map[string]string `env:"META,key1=val1,key2=val2"` // map
	Date     time.Time         `env:"DATE,2024-07-11T00:00:00Z"` // time
}
```

---

## time.Time — Custom Date/Time Formats

By default, GoEnvy expects `time.Time` values in **RFC3339** format, which looks like this:

```
2024-07-11T15:04:05Z
2024-07-11T10:30:00+02:00
```

But if your env variable uses a different format, you can tell GoEnvy exactly how to parse it using the **`env_format`** tag.

### How to set a custom format

Add `env_format:"layout"` next to your `env` tag, where `layout` uses Go's reference time:

```go
type Config struct {
	EventDate time.Time `env:"EVENT_DATE" env_format:"2006-01-02"`
	Birthday  time.Time `env:"BIRTHDAY" env_format:"01/02/2006"`
	Timestamp time.Time `env:"TIMESTAMP" env_format:"2006-01-02 15:04:05"`
}
```

### Understanding Go's Reference Time

Go doesn't use `YYYY-MM-DD` or `MM/DD/YYYY` like other languages. Instead, Go uses a **specific date and time** as the reference:

```
Mon Jan 2 15:04:05 MST 2006
│   │   │  │    │   │    │
│   │   │  │    │   │    └── year (2006)
│   │   │  │    │   └────── timezone (MST)
│   │   │  │    └───────── second (05)
│   │   │  └───────────── minute (04)
│   │   └──────────────── hour (15 = 3 PM, 03 = 3 AM/PM)
│   └──────────────────── day (2)
└──────────────────────── month (January = 01)
```

So to build any format, just arrange these numbers:

| You Want This Format | Write This Layout |
|----------------------|-------------------|
| `2024-07-11` | `2006-01-02` |
| `07/11/2024` | `01/02/2006` |
| `11-07-2024` | `02-01-2006` |
| `2024-07-11 15:04:05` | `2006-01-02 15:04:05` |
| `2024-07-11T15:04:05Z` | `2006-01-02T15:04:05Z07:00` |
| `11-Jul-2024` | `02-Jan-2006` |

### Real .env Example

```
# .env file
EVENT_DATE=2024-12-25
BIRTHDAY=01/15/1990
TIMESTAMP=2024-07-11 14:30:00
```

```go
type Config struct {
	EventDate time.Time `env:"EVENT_DATE" env_format:"2006-01-02"`
	Birthday  time.Time `env:"BIRTHDAY" env_format:"01/02/2006"`
	Timestamp time.Time `env:"TIMESTAMP" env_format:"2006-01-02 15:04:05"`
}

var cfg Config
goenvy.MustLoadEnvVariable(&cfg)

fmt.Println(cfg.EventDate) // 2024-12-25 00:00:00 +0000 UTC
fmt.Println(cfg.Birthday)  // 1990-01-15 00:00:00 +0000 UTC
fmt.Println(cfg.Timestamp) // 2024-07-11 14:30:00 +0000 UTC
```

> **Note:** If you don't specify `env_format`, GoEnvy uses `time.RFC3339` (`"2006-01-02T15:04:05Z07:00"`) automatically.

---

## Nested Structs — Organize Your Config

Group related settings into nested structs for cleaner organization:

```go
type Config struct {
	Server struct {
		Host string `env:"SERVER_HOST,localhost"`
		Port int    `env:"SERVER_PORT,8080"`
	}
	Database struct {
		URL      string `env:"DB_URL"`
		PoolSize int    `env:"DB_POOL,10"`
		Password string `env:"DB_PASSWORD"` // required (no default!)
	}
	Cache struct {
		Enabled bool   `env:"CACHE_ENABLED,true"`
		TTL     string `env:"CACHE_TTL,5m"`
	}
}
```

**.env file:**
```
SERVER_HOST=0.0.0.0
SERVER_PORT=3000
DB_URL=postgres://localhost:5432/mydb
DB_PASSWORD=secret123
CACHE_ENABLED=false
```

**Usage:**
```go
var cfg Config
goenvy.MustLoadEnvVariable(&cfg)

fmt.Println(cfg.Server.Host)     // "0.0.0.0"
fmt.Println(cfg.Database.URL)    // "postgres://localhost:5432/mydb"
fmt.Println(cfg.Cache.Enabled)   // false
```

---

## API Reference

### `LoadEnvVariable(cfg interface{}) error`

Reads environment variables (and `.env` if it exists) and populates the struct.

**Returns an error if:**
- A required variable (no default) is missing from the environment
- A value can't be converted to the target type (e.g., `"abc"` for an `int` field)
- `cfg` is not a pointer to a struct

```go
type Config struct {
	APIKey string `env:"API_KEY"`        // required — will error if missing
	Port   int    `env:"PORT,8080"`       // optional — has default
}

var cfg Config
if err := goenvy.LoadEnvVariable(&cfg); err != nil {
	log.Fatalf("Error loading config: %v", err)
}
```

**Example error messages:**

```
// Missing required variable:
[goenvy] missing required env variable "API_KEY" (for field "APIKey")

// Type conversion failure:
[goenvy] field "Port": cannot parse "not_a_number" as int: ...

// Wrong argument type:
[goenvy] expected a pointer to a struct, got string
```

### `MustLoadEnvVariable(cfg interface{})`

Same as `LoadEnvVariable` but **panics** (crashes the program) on error. Use this when you want the program to fail fast at startup rather than handle errors manually.

```go
// In main() or init() — crashes immediately if config is invalid
goenvy.MustLoadEnvVariable(&cfg)
```

---

## Real-World Example

A complete web server configuration:

```go
package main

import (
	"fmt"
	"log"
	"time"

	goenvy "github.com/yassine-manai/GoEnvy"
)

type ServerConfig struct {
	Server struct {
		Host         string        `env:"SRV_HOST,0.0.0.0"`
		Port         int           `env:"SRV_PORT,8080"`
		ReadTimeout  time.Duration `env:"SRV_READ_TIMEOUT,30s"`
		WriteTimeout time.Duration `env:"SRV_WRITE_TIMEOUT,30s"`
	}
	Database struct {
		Host     string `env:"DB_HOST,localhost"`
		Port     int    `env:"DB_PORT,5432"`
		Name     string `env:"DB_NAME,appdb"`
		User     string `env:"DB_USER"`         // required
		Password string `env:"DB_PASSWORD"`     // required
	}
	Redis struct {
		Host     string `env:"REDIS_HOST,localhost"`
		Port     int    `env:"REDIS_PORT,6379"`
		Password string `env:"REDIS_PASSWORD"`
	}
	App struct {
		Name      string            `env:"APP_NAME,MyApp"`
		Version   string            `env:"APP_VERSION,1.0.0"`
		Debug     bool              `env:"APP_DEBUG,false"`
		AllowedIP []string          `env:"ALLOWED_IPS,127.0.0.1"`
		Features  map[string]string `env:"FEATURES,auth=enabled,logs=disabled"`
		LaunchAt  time.Time         `env:"LAUNCH_AT" env_format:"2006-01-02"`
	}
}

func main() {
	var cfg ServerConfig
	goenvy.MustLoadEnvVariable(&cfg)

	fmt.Printf("Server: %s:%d\n", cfg.Server.Host, cfg.Server.Port)
	fmt.Printf("DB: %s@%s:%d/%s\n", cfg.Database.User, cfg.Database.Host,
		cfg.Database.Port, cfg.Database.Name)
	fmt.Printf("Debug: %v\n", cfg.App.Debug)
	fmt.Printf("Launch: %s\n", cfg.App.LaunchAt.Format("January 2, 2006"))
}
```

---

## `.env` File Format

GoEnvy uses the standard `.env` format (via the [godotenv](https://github.com/joho/godotenv) library):

```ini
# This is a comment
KEY=VALUE
APP_PORT=3000
DB_URL=postgres://localhost/mydb
MULTI_WORD="value with spaces"
EMPTY=
```

**Rules:**
- Lines starting with `#` are comments
- Format: `KEY=VALUE`
- Quotes around values are stripped
- Empty values are treated as empty strings
- GoEnvy puts the `.env` file in your project root (where you run the app)

---

## Best Practices

1. **Use nested structs** to group related settings (Server, Database, Cache, etc.)
2. **Always set defaults** for non-critical values (port numbers, timeouts, etc.)
3. **Omit the default** for required values (passwords, API keys, secrets)
4. **Use UPPERCASE** for environment variable names (industry convention)
5. **Validate your config** after loading (check port ranges, URL formats, etc.)
6. **Keep a `.env.example`** file in your repo (without real secrets) so other devs know what variables are needed
7. **Separate config into its own package** — move your Config struct into a dedicated `config` package to keep `main.go` clean and make config reusable across files (see below)

### 🔧 Recommended Project Layout

For real applications, separate your configuration into its own package:

```
myapp/
├── main.go
├── config/
│   └── app_config.go
├── .env
└── go.mod
```

**`config/app_config.go`** — define the struct and auto-load with `init()`:

```go
package config

import (
	goenvy "github.com/yassine-manai/GoEnvy"
)

// AppConfig holds all application configuration loaded from environment variables.
type AppConfig struct {
	Port  int    `env:"PORT,8080"`
	Host  string `env:"HOST,localhost"`
	Debug bool   `env:"DEBUG,false"`
}

// Cfg is the global configuration, populated automatically when this package is imported.
var Cfg AppConfig

func LoadEnvVars() {
	// load env vars 
	err := goenvy.LoadEnvVariable(&Cfg)
	if err != nil {
		log.Error(err).Msg("Failed to Load Env Variables")
	}
}

```

**`main.go`** — just import the package, config loads automatically before `main()` starts:

```go
package main

import (
	"fmt"
	"myapp/config" // importing triggers config.init()
)

func init(){
	config.LoadEnvVars() // Env var Loading
}

// Main Function
func main() {
	// config.Cfg is already populated — ready to use
	fmt.Println("Server:", config.Cfg.Host+":"+config.Cfg.Port)
	fmt.Println("Debug:", config.Cfg.Debug)
}
```

> **Why this pattern?** The `init()` function runs automatically when the `config` package is first imported, so your configuration is populated before `main()` begins executing. This keeps `main.go` focused on application logic instead of setup code, and the `config` package can be imported from any file in your project.

---

## Troubleshooting

| Problem | Likely Cause | Solution |
|---------|-------------|----------|
| "missing required env variable" | You forgot to set the env var or add it to `.env` | Set the variable or add a default value in the tag |
| "cannot parse ... as int/bool/float" | The env value doesn't match the struct field type | Check your `.env` file or exported env var value |
| "expected a pointer to a struct" | You passed the struct value, not a pointer | Use `&cfg` instead of `cfg` |
| Warning about `.env` file | No `.env` file in the current directory | Create one, or ignore the warning if you're using real env vars |
| `time.Time` parsing error | Wrong format for the `env_format` tag | Double-check your layout using Go's reference time: `2006-01-02` |

---

## Run the Included Example

```bash
go test -run Example -v
```

---

## License

MIT License — see [LICENSE](LICENSE).

**Author:** [Yassine Manai](https://github.com/yassine-manai)
