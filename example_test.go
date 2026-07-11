package goenvy_test

import (
	"fmt"
	"os"
	"time"

	goenvy "github.com/yassine-manai/GoEnvy"
)

// Example demonstrates loading environment variables into a Go struct.
// It shows:
//   - env tags with default values (Port, Debug, Timeout, Tags)
//   - env tags overridden by actual env vars (Host, Tags)
//   - Multiple supported types: string, int, bool, time.Duration, []string
func Example() {
	type AppConfig struct {
		Host    string        `env:"HOST,localhost"`
		Port    int           `env:"PORT,3000"`
		Debug   bool          `env:"DEBUG,false"`
		Timeout time.Duration `env:"TIMEOUT,30s"`
		Tags    []string      `env:"TAGS,go,env,loader"`
	}

	os.Setenv("HOST", "0.0.0.0")
	os.Setenv("TAGS", "alpha,beta,gamma")

	var cfg AppConfig
	goenvy.LoadEnvVariable(&cfg)

	fmt.Println(cfg.Host)
	fmt.Println(cfg.Port)
	fmt.Println(cfg.Debug)
	fmt.Println(cfg.Timeout)
	fmt.Println(cfg.Tags)

	// Output:
	// 0.0.0.0
	// 3000
	// false
	// 30s
	// [alpha beta gamma]
}
