package envconfig

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
)

// LoadConfig reads a struct with fields and some specific tags and reaches into the runtime environment to fill in values
// for the fields of that struct. The Env variable to field associations are defined using the `env` tag.
// Additionally, you can set 2 tags to control the behavior of the configuration loader:
// 1. `default`: Allows you to set a default value for the field in the event the environment variable is not set
// 2. `required`: Causes a panic if no value is defined in the environment variable specified by `env` tag
func LoadConfig(cfg interface{}) interface{} {
	// Check that the cfg is a pointer to a struct
	if reflect.ValueOf(cfg).Kind() != reflect.Ptr {
		fmt.Fprintln(os.Stderr, "'cfg' must be a pointer to a concrete struct")
		os.Exit(1)
	}

	// Get struct Value and dereference it to get the underlying memory space
	el := reflect.ValueOf(cfg).Elem()

	// Go over each field in the config struct, load the relevant environment variable, and attempt to cast to the correct type
	for i := 0; i < el.NumField(); i++ {
		// Get the raw environment value based on env tag
		typField := el.Type().Field(i)

		// Get env tag and ensure it was set
		envTag, ok := typField.Tag.Lookup("env")
		if !ok {
			fmt.Fprintf(os.Stderr,"'env' tag not found for field %s", typField.Name)
			os.Exit(127)
		}

		// Extract the value from the environment
		raw := os.Getenv(envTag)
		if raw == "" { // Check if Raw Env value is empty, if so we have a few fallback positions
			if def := typField.Tag.Get("default"); def != "" {
				// raw is missing, first check for a default setting
				raw = def
			} else if _, ok := typField.Tag.Lookup("required"); ok {
				// no default, so check if required. If yes, we panic out since we cannot set this value
				fmt.Fprintf(os.Stderr,"Env variable %s is required by field %s\n", envTag, typField.Name)
				os.Exit(127)
			}
		}

		// Extract the concrete field for this iteration
		fld := el.Field(i)

		switch fld.Type().Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			conv, err := strconv.Atoi(raw)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Unable to convert value found in environment variable %s ('%s') to int. Aborting.\n", envTag, raw)
				os.Exit(127)
			}

			fld.SetInt(int64(conv))
		case reflect.String:
			fld.SetString(raw)
		case reflect.Float32, reflect.Float64:
			conv, err := strconv.ParseFloat(raw, 64)
			if err != nil {
				fmt.Fprintf(os.Stderr,"Unable to convert value found in environment variable %s ('%s') to float. Aborting.\n", envTag, raw)
			}

			fld.SetFloat(conv)
		case reflect.Bool:
			fld.SetBool(raw != "")
		}
	}

	return cfg
}
