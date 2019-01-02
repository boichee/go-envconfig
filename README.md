# Go Config-by-Environment

Current practice is to use the runtime environment, whenever possible, to configure an application. In Golang, this often means writing out a configuration struct and providing a specific loader within your application or using global variables in a package. Overall, it tends to be ugly, hard to maintain, and wasteful as you do the same thing for each application.

This simple package seeks to provide a solution.

### Usage:

The `envconfig` package uses Golang's reflection tools to read a provided concrete struct; it uses the type and tag information associated with each field to load configuration from the runtime environment into that struct and return it.

First, create a configuration struct with 1 field for each configuration value you need:

```go
package example

type Configuration struct {
	Foo int     `env:"FOO_VALUE"`
	Bar string  `env:"BAR_VALUE" required:"true"`
	Baz bool    `env:"BAZ_VALUE" default:"true"`
}
```

Notice the tags. The `env` tag defines the environment variable that will contain the value for the configuration field. When set, `required` indicates to the loader that this field must have a value. The `default` tag allows you to provide a default value if one is not found in the environment.

Next, you simply use the `LoadConfiguration` function found in the `envconfig` package:

```go
package example

import (
	"fmt"
	"os"
)

type Configuration struct {
	Foo int     `env:"FOO_VALUE"`
	Bar string  `env:"BAR_VALUE" required:"true"`
	Baz bool    `env:"BAZ_VALUE" default:"true"`
}

func load() *Configuration {
	result, err := envconfig.LoadConfiguration(&Configuration{}, false) // The 2nd param determines whether errors are printed to stderr
	if err != nil {
		fmt.Println(err) // We print the error ourselves since we set `showErrors` to false above
		os.Exit(127)
	}
	
	return result.(*Configuration) // You must cast the interface{} back to your configuration type using a type assertion
}

var Cfg = load() // One possible way to use this. You could also use the init() function to get the configuration loaded
```

Note that the `LoadConfiguration` function:

  1. takes two arguments: The first must be a pointer to a `struct` with the format shown above. The second is a `bool` that tells the configuration loader whether or not to print any errors it discovers to `stderr`. Either way, the error will be returned to the caller so you can choose to print it in your calling code as I have done here.
  2. returns an `interface{}` type. This is because you have to define the struct that it processes. As such, the final step is always to make a type assertion, converting the `interface{}` to your `*Configuration` struct type—just as I have done above.
  
  
 