// Package goenvy provides a type-safe environment variable loader for Go applications.
//
// It maps environment variables to struct fields using `env` tags, supporting
// multiple types including string, int, bool, float64, time.Duration, slices,
// maps, and time.Time. Nested structs are fully supported.
//
// Tag format:
//
//	`env:"KEY"`             - required variable (error if missing)
//	`env:"KEY,default"`     - optional variable (uses default if missing)
//
// For time.Time fields, use an additional env_format tag:
//
//	`env:"DATE" env_format:"2006-01-02"`
package goenvy

import (
	"fmt"
	"log"
	"reflect"

	"github.com/joho/godotenv"
)

// LoadEnvVariable reads environment variables into cfg using env tags.
//
// It first attempts to load a .env file via godotenv (logs a warning if not found).
// cfg must be a non-nil pointer to a struct. Fields are populated based on their
// env tags and converted to the appropriate type.
//
// Supported types: string, int, bool, float64, time.Duration,
// []string, []int, []float64, map[string]string, time.Time.
func LoadEnvVariable(cfg interface{}) error {
	if err := godotenv.Load(); err != nil {
		log.Printf("[goenvy] Warning: could not load .env file: %v", err)
	}

	v := reflect.ValueOf(cfg)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("[goenvy] expected a pointer to a struct, got %T", cfg)
	}

	return parseStruct(v.Elem())
}

// MustLoadEnvVariable is like LoadEnvVariable but panics on error.
// Useful for initialization where a missing variable is a fatal error.
func MustLoadEnvVariable(cfg interface{}) {
	if err := LoadEnvVariable(cfg); err != nil {
		panic(fmt.Sprintf("[goenvy] failed to load environment: %v", err))
	}
}
