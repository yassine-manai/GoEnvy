package goenvy

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Sentinel type values used for reflection-based type detection.
// Precomputing these avoids repeated reflect.TypeOf calls during field processing.
var (
	durationType = reflect.TypeOf(time.Duration(0))
	timeType     = reflect.TypeOf(time.Time{})
)

// setField dispatches to the appropriate type-specific setter based on
// the field's type or kind.
func setField(field reflect.Value, fieldType reflect.StructField, rawVal string) error {
	switch {
	case field.Type() == durationType:
		return setDurationField(field, fieldType.Name, rawVal)
	case field.Type() == timeType:
		return setTimeField(field, fieldType, rawVal)
	}

	switch field.Kind() {
	case reflect.String:
		return setStringField(field, fieldType.Name, rawVal)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return setIntField(field, fieldType.Name, rawVal)
	case reflect.Bool:
		return setBoolField(field, fieldType.Name, rawVal)
	case reflect.Float32, reflect.Float64:
		return setFloatField(field, fieldType.Name, rawVal)
	case reflect.Slice:
		return setSliceField(field, fieldType.Name, rawVal)
	case reflect.Map:
		return setMapField(field, fieldType.Name, rawVal)
	default:
		return fmt.Errorf("[goenvy] field %q: unsupported type %s", fieldType.Name, field.Kind())
	}
}

// setStringField assigns the raw string value directly to a string struct field.
func setStringField(field reflect.Value, fieldName, rawVal string) error {
	field.SetString(rawVal)
	return nil
}

// setIntField parses the raw string as a base-10 integer and assigns it to an int struct field.
// Supports int, int8, int16, int32, and int64 kinds.
func setIntField(field reflect.Value, fieldName, rawVal string) error {
	n, err := strconv.ParseInt(rawVal, 10, 64)
	if err != nil {
		return fmt.Errorf("[goenvy] field %q: cannot parse %q as int: %w", fieldName, rawVal, err)
	}
	field.SetInt(n)
	return nil
}

// setBoolField parses the raw string as a boolean and assigns it to a bool struct field.
// Accepts true/false, 1/0, and t/f via strconv.ParseBool.
func setBoolField(field reflect.Value, fieldName, rawVal string) error {
	b, err := strconv.ParseBool(rawVal)
	if err != nil {
		return fmt.Errorf("[goenvy] field %q: cannot parse %q as bool: %w", fieldName, rawVal, err)
	}
	field.SetBool(b)
	return nil
}

// setFloatField parses the raw string as a 64-bit float and assigns it to a float struct field.
// Supports float32 and float64 kinds.
func setFloatField(field reflect.Value, fieldName, rawVal string) error {
	f, err := strconv.ParseFloat(rawVal, 64)
	if err != nil {
		return fmt.Errorf("[goenvy] field %q: cannot parse %q as float: %w", fieldName, rawVal, err)
	}
	field.SetFloat(f)
	return nil
}

// setDurationField parses the raw string as a time.Duration and assigns it to a duration struct field.
// Accepts Go duration strings like "5s", "1m30s", "2h", "500ms".
func setDurationField(field reflect.Value, fieldName, rawVal string) error {
	d, err := time.ParseDuration(rawVal)
	if err != nil {
		return fmt.Errorf("[goenvy] field %q: cannot parse %q as duration (use format like 5s, 1m30s, 2h): %w", fieldName, rawVal, err)
	}
	field.SetInt(int64(d))
	return nil
}

// setSliceField parses a comma-separated raw string into a slice and assigns it to a slice struct field.
// Supported element types: string, int, float64.
// If rawVal is empty, an empty slice is assigned.
func setSliceField(field reflect.Value, fieldName, rawVal string) error {
	elemKind := field.Type().Elem().Kind()

	if rawVal == "" {
		field.Set(reflect.MakeSlice(field.Type(), 0, 0))
		return nil
	}

	parts := strings.Split(rawVal, ",")

	switch elemKind {
	case reflect.String:
		result := make([]string, 0, len(parts))
		for _, p := range parts {
			result = append(result, strings.TrimSpace(p))
		}
		field.Set(reflect.ValueOf(result))

	case reflect.Int:
		result := make([]int, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			n, err := strconv.Atoi(p)
			if err != nil {
				return fmt.Errorf("[goenvy] field %q: cannot parse %q as int: %w", fieldName, p, err)
			}
			result = append(result, n)
		}
		field.Set(reflect.ValueOf(result))

	case reflect.Float64:
		result := make([]float64, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			f, err := strconv.ParseFloat(p, 64)
			if err != nil {
				return fmt.Errorf("[goenvy] field %q: cannot parse %q as float64: %w", fieldName, p, err)
			}
			result = append(result, f)
		}
		field.Set(reflect.ValueOf(result))

	default:
		return fmt.Errorf("[goenvy] field %q: unsupported slice element type %s", fieldName, elemKind)
	}

	return nil
}

// setMapField parses a comma-separated "key=value" string into map[string]string and assigns it.
// Entries without an "=" separator are silently skipped.
// If rawVal is empty, an empty map is assigned.
func setMapField(field reflect.Value, fieldName, rawVal string) error {
	if field.Type().Key().Kind() != reflect.String || field.Type().Elem().Kind() != reflect.String {
		return fmt.Errorf("[goenvy] field %q: unsupported map type, only map[string]string is supported", fieldName)
	}

	result := make(map[string]string)

	if rawVal == "" {
		field.Set(reflect.ValueOf(result))
		return nil
	}

	entries := strings.Split(rawVal, ",")

	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		pair := strings.SplitN(entry, "=", 2)
		if len(pair) != 2 {
			continue
		}
		key := strings.TrimSpace(pair[0])
		val := strings.TrimSpace(pair[1])
		result[key] = val
	}

	field.Set(reflect.ValueOf(result))
	return nil
}

// setTimeField parses the raw string as a time.Time and assigns it to a time struct field.
// Uses the field's env_format tag for parsing; defaults to time.RFC3339 if no format is specified.
// Example: `env:"DATE" env_format:"2006-01-02"` parses "2024-07-11" correctly.
func setTimeField(field reflect.Value, fieldType reflect.StructField, rawVal string) error {
	format := fieldType.Tag.Get("env_format")
	if format == "" {
		format = time.RFC3339
	}

	t, err := time.Parse(format, rawVal)
	if err != nil {
		return fmt.Errorf("[goenvy] field %q: cannot parse %q as time with format %q: %w",
			fieldType.Name, rawVal, format, err)
	}

	field.Set(reflect.ValueOf(t))
	return nil
}
