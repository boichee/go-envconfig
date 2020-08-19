# Go Config-by-Environment

Current practice is to use the runtime environment, whenever possible, to configure an application. In Golang, this often means writing out a configuration struct and providing a specific loader within your application or using global variables in a package. Overall, it tends to be ugly, hard to maintain, and wasteful as you do the same thing for each application.

This simple package seeks to provide a solution.

### Usage:

The `envconfig` package uses Golang's reflection tools to read a provided concrete struct; it uses the type and tag information associated with each field to load configuration from the runtime environment into that struct and return it.


```go
package example

import (
	"github.com/boichee/go-envconfig"
	"fmt"
)

// First, create a configuration struct with 1 field for each configuration value
type Configuration struct {
	Foo int     `env:"FOO_VALUE"`
	Bar string  `env:"BAR_VALUE" required:"true"`
	Baz bool    `env:"BAZ_VALUE" default:"1"`
}

// Next, you use the loader to get the configuration into a concrete struct
func main() {
	var cfg Configuration
	if err := envconfig.Process(&cfg, true); err != nil {
		panic("Something went terribly wrong loading configuration!")
	}
	
	// Now do stuff with configuration values!
	fmt.Println("Foo is:", cfg.Foo)
}
```

Notice the struct tags. The `env` tag defines the environment variable that will contain the value for the configuration field. When set, `required` indicates to the loader that this field must have a value. The `default` tag allows you to provide a default value if one is not found in the environment.

The signature for `Process` is:

```golang
envconfig.Process(cfg interface{}, showErrors bool)
```

If you set `showErrors` to true, configuration loading errors will be logged to `stderr` automatically (errors will be retuned by the function regardless of this setting).
  
### Types:

So far, the environment configuration loader can handle the following types:

  - `int`
  - `int8`
  - `int16`
  - `int32`
  - `int64`
  
  - `float32`
  - `float64`
  
  - `string`
  - `bool`

> Note: For the `bool` type, `1` evaluates to true, and `0` to false.
  
The flag configuration loader handles the following types:
  - `int64`
  - `uint64`
  - `float64`
  - `string`
  - `bool`
  
  
### Flag Processing:

You can also process flags by providing a struct as a spec. A few key differences:

1. The name of the flag will default to the **lowercased** name of the struct field unless overriden using the `flag` tag
2. Flags are never required. They will default to the zero value for their type if not set
3. You can provide a `usage` tag that will be used to add usage instructions for the flag

To process a struct as flags, use the `ProcessFlags` function instead.

```go
envconfig.ProcessFlags(cfg interface{}) error
```
  

### Struct Tags:

The following struct tags are meaningful when using this package:

| Name | Purpose | Allowed Values |
| ---- | ------- | -------------- |
| `env` | Defines the environment variable that contains the value for a field in the struct | `[A-z_0-9]` |
| `flag` | Sets the flag name | `[A-z0-9_-]` |
| `required` | Marks a field as required. If the env variable cannot be found or is empty, an error will be returned. | N/A, if the tag is set, its value is irrelevant |
| `default` | Allows a default value to be provided for a field | Any valid value for the field's type |
