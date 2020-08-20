package envconfig

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"unsafe"
)

func handleError(s string, showError bool) error {
	if showError {
		fmt.Fprintln(os.Stderr, s)
	}

	return errors.New(s)
}

// ProcessFlags works mostly the same as Process, but expects values in the spec to be provided
// as command line flags instead of as environment variables
// Differences:
// 1. Only (u)int64 is supported (compared to all int/uint sizes)
// 2. Flag name is automatically determined by lowercasing the field name. This can be overriden by providing a "flag" tag
// 3. A "usage" tag can be provided to add usage instructions
// 4. showErrors support not available. Errors will never be printed to stdErr
func ProcessFlags(spec interface{}) error {
	// Check that spec is a pointer to struct (otherwise it won't be mutable)
	if reflect.ValueOf(spec).Kind() != reflect.Ptr {
		return handleError("spec param must be a pointer to struct", false)
	}

	// Get the concrete, specific instance pointed to by "spec"
	concrete := reflect.ValueOf(spec).Elem()

	// We iterate over the struct, and extract the type information from each field along with the tags
	// supplied. We then grab an unsafe pointer to each field in the struct and cast it to the correct
	// pointer type for that field. Then that "safe" pointer is used as the target for the call to flag
	for i := 0; i < concrete.NumField(); i++ {
		typ := concrete.Type().Field(i)
		defaultVal, usageVal := typ.Tag.Get("default"), typ.Tag.Get("usage")
		name := strings.ToLower(typ.Name)
		if fName, ok := typ.Tag.Lookup("flag"); ok {
			// explicit flag name was provided, so it overrides the default of the lowercased field name
			name = fName
		}

		fld := concrete.Field(i)
		uptr := unsafe.Pointer(fld.UnsafeAddr())
		switch fld.Type().Kind() {
		case reflect.Int64:
			ptr := (*int64)(uptr)
			def, _ := strconv.ParseInt(defaultVal, 10, 64) // No need to deal with error as the value on error is 0 which is what we want
			flag.Int64Var(ptr, name, def, usageVal)
		case reflect.Uint64:
			ptr := (*uint64)(uptr)
			def, _ := strconv.ParseUint(defaultVal, 10, 64)
			flag.Uint64Var(ptr, name, def, usageVal)
		case reflect.Float32, reflect.Float64:
			ptr := (*float64)(uptr)
			def, _ := strconv.ParseFloat(defaultVal, 64)
			flag.Float64Var(ptr, name, def, usageVal)
		case reflect.String:
			ptr := (*string)(uptr)
			flag.StringVar(ptr, name, defaultVal, usageVal)
		case reflect.Bool:
			ptr := (*bool)(uptr)
			flag.BoolVar(ptr, name, false, usageVal) // We set the default to false so that this is only true when set
		default:
			return handleError(fmt.Sprintf("The type '%s' of the field '%s' is not supported", fld.Type().Kind(), typ.Name), false)
		}
	}

	flag.Parse()
	return nil
}

// Process reads a struct with fields and some specific tags and reaches into the runtime environment to fill in values
// for the fields of that struct. The Env variable to field associations are defined using the `env` tag.
// Additionally, you can set 2 tags to control the behavior of the configuration loader:
// 1. `default`: Allows you to set a default value for the field in the event the environment variable is not set
// 2. `required`: Causes a panic if no value is defined in the environment variable specified by `env` tag
func Process(spec interface{}, showErrors bool) error {
	// Check that spec is a pointer to struct
	if reflect.ValueOf(spec).Kind() != reflect.Ptr {
		return handleError("spec param must be a pointer to struct", showErrors)
	}

	// Get value from struct and dereference it
	el := reflect.ValueOf(spec).Elem()

	// For each field in spec struct, load relevant env var, and attempt to cast to the correct type
	for i := 0; i < el.NumField(); i++ {
		// Get the raw environment value based on env tag
		typField := el.Type().Field(i)

		// Get env tag and ensure it was set
		envTag, ok := typField.Tag.Lookup("env")
		if !ok {
			s := fmt.Sprintf("'env' tag not found for field %s", typField.Name)
			return handleError(s, showErrors)
		}

		// Extract the value from the environment
		raw := os.Getenv(envTag)
		if raw == "" { // Check if Raw Env value is empty, if so we have a few fallback positions
			if def := typField.Tag.Get("default"); def != "" {
				// raw is missing, first check for a default setting
				raw = def
			} else if _, ok := typField.Tag.Lookup("required"); ok {
				// no default, so check if required. If yes, we panic out since we cannot set this value
				s := fmt.Sprintf("Env variable %s is required by field %s\n", envTag, typField.Name)
				return handleError(s, showErrors)
			}
		}

		// Extract the concrete field for this iteration
		fld := el.Field(i)

		switch fld.Type().Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			conv, err := strconv.Atoi(raw)
			if err != nil {
				s := fmt.Sprintf("Unable to convert value found in environment variable %s ('%s') to int. Aborting.", envTag, raw)
				return handleError(s, showErrors)
			}

			fld.SetInt(int64(conv))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			conv, err := strconv.ParseUint(raw, 10, 64)
			if err != nil {
				s := fmt.Sprintf("Unable to convert value found in environment variable %s ('%s') to uint. Aborting.", envTag, raw)
				return handleError(s, showErrors)
			}

			fld.SetUint(uint64(conv))
		case reflect.String:
			fld.SetString(raw)
		case reflect.Float32, reflect.Float64:
			conv, err := strconv.ParseFloat(raw, 64)
			if err != nil {
				s := fmt.Sprintf("Unable to convert value found in environment variable %s ('%s') to float. Aborting.", envTag, raw)
				return handleError(s, showErrors)
			}

			fld.SetFloat(conv)
		case reflect.Bool:
			switch raw {
			case "0":
				fld.SetBool(false)
			case "1":
				fld.SetBool(true)
			default:
				s := fmt.Sprintf("Unable to convert value found in environment variable %s ('%s') to bool (should be: 1 or 0). Aborting.", envTag, raw)
				return handleError(s, showErrors)
			}
		}
	}

	return nil
}

// LoadConfig present for backwards compatibility
func LoadConfig(cfg interface{}, showErrors bool) (interface{}, error) {
	err := Process(cfg, showErrors)
	return cfg, err
}



