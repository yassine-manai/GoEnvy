package goenvy

import (
	"fmt"
	"os"
	"reflect"
	"strings"
)

// parseStruct walks all fields of the struct value v and populates them from
// environment variables based on their `env` tags.
//
// For each field:
//  1. If it is a nested struct (but not time.Time), recurse into it.
//  2. If it has no env tag, skip it.
//  3. Otherwise, resolve the env value and set the field via setField.
func parseStruct(v reflect.Value) error {
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		if field.Kind() == reflect.Struct && field.Type() != timeType {
			if err := parseStruct(field); err != nil {
				return err
			}
			continue
		}

		tag := fieldType.Tag.Get("env")
		if tag == "" {
			continue
		}

		key, defaultVal, hasDefault := parseTag(tag)

		rawVal, err := resolveValue(key, defaultVal, hasDefault, fieldType.Name)
		if err != nil {
			return err
		}

		if err := setField(field, fieldType, rawVal); err != nil {
			return err
		}
	}

	return nil
}

// parseTag splits an env tag value into its components.
// Tag format: "KEY" or "KEY,default_value".
// Returns the key, the default value (if any), and whether a default was provided.
func parseTag(tag string) (string, string, bool) {
	parts := strings.SplitN(tag, ",", 2)
	key := strings.TrimSpace(parts[0])

	if len(parts) == 2 {
		return key, strings.TrimSpace(parts[1]), true
	}

	return key, "", false
}

// resolveValue retrieves the value for an environment variable.
// Priority order:
//  1. Actual environment variable (os.Getenv) — whitespace trimmed
//  2. Default value from the env tag (if hasDefault is true) — whitespace trimmed
//  3. Error: missing required variable
//
// Trimming prevents parse failures from values like " 8080 " or " true"
// that can appear in .env files with inline spacing.
// The fieldName parameter is used only in error messages for clarity.
func resolveValue(key, defaultVal string, hasDefault bool, fieldName string) (string, error) {
	val := strings.TrimSpace(os.Getenv(key))
	if val != "" {
		return val, nil
	}

	if hasDefault {
		return strings.TrimSpace(defaultVal), nil
	}

	return "", fmt.Errorf("[goenvy] missing required env variable %q (for field %q)", key, fieldName)
}
