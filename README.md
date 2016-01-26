Ridiculously simple application configuration built on top of [Viper](https://github.com/spf13/viper).

[![Build Status](https://travis-ci.org/illyabusigin/configurator.svg)](https://travis-ci.org/illyabusigin/configurator)
[![Build Status](https://img.shields.io/coveralls/illyabusigin/configurator.svg)](https://img.shields.io/coveralls/illyabusigin/configurator.svg) [![](https://godoc.org/github.com/illyabusigin/configurator?status.svg)](http://godoc.org/github.com/illyabusigin/configurator)

## What is Configurator?

Configurator is an application configuration solution built on top of [Viper](https://github.com/spf13/viper). It was designed to be as easy to use as possible. Since Configurator is built on top of Viper it supports:

* setting defaults
* reading from JSON, TOML, YAML and HCL config files
* reading from environment variables
* reading from command line flags
* setting explicit values

Configurator does not currently support the following Viper features:

* live watching and re-reading of config files (optional)
* reading from remote config systems (etcd or Consul), and watching changes
* reading from buffer

**NOTE**: Full Viper support is planned in future releases.


## Why Configurator?

Viper is awesome and has seen widespread use in many popular Go packages. The goal of Configurator was to make Viper easier to use.

Every application needs some form of configuration and Viper and Configurator make ingesting configuration data easy no matter what the source. If you haven't  familiarized yourself what Viper is, [please do so](https://github.com/spf13/viper#what-is-viper).

Since configurator is built on top of Viper the same source precedence are applicable:

* Overrides, or setting the config struct field directly.
* Flags - note that [github.com/spf13/pflag](https://github.com/spf13/pflag) is used in lieu of the standard flag package
* Environment variables
* Configuration file values
* Default values


## Usage

There are two ways of using Configurator. You can embed configurator.Config into your configuration struct or call configurator.Load(&yourStruct) directly. Whichever method you pick Configurator requires you to annotate your struct with the configuration sources like so:

```go
type AppConfig struct {
	configurator.Config

	Secret      string `env:"APP_SECRET" file:"secret" flag:"secret" default:"asecretvalue"`
	Port        string `env:"APP_PORT" file:"port" flag:"port" default:"3000"`
	Environment string `env:"APP_ENV" file:"env" flag:"env" default:"dev"`
}
```

### Embedded configurator.Config
```go
import (
	"fmt"
	"log"

	"github.com/illyabusigin/configurator"
)

type AppConfig struct {
	configurator.Config

	Secret      string `env:"APP_SECRET" file:"secret" flag:"secret" default:"asecretvalue"`
	Port        string `env:"APP_PORT" file:"port" flag:"port" default:"3000"`
	Environment string `env:"APP_ENV" file:"env" flag:"env" default:"dev"`
}

var Config AppConfig

func main() {
	Config = AppConfig{}
	Config.FileName = "config"
	Config.FilePaths = []string{
		".",
	}
	
	err := Config.Load(& Config)
	
	if err != nil {
		// Always handle your errors
		log.Fatalf("Unable to load application configuration! Error: %s", err.Error())
	}
}
```

### Load Configuration Directly
```go
import (
	"fmt"
	"log"

	"github.com/illyabusigin/configurator"
)

type AppConfig struct {
	Secret      string `env:"APP_SECRET" file:"secret" flag:"secret" default:"asecretvalue"`
	Port        string `env:"APP_PORT" file:"port" flag:"port" default:"3000"`
	Environment string `env:"APP_ENV" file:"env" flag:"env" default:"dev"`
}

var Config AppConfig

func main() {
	Config = AppConfig{}
	configurator.SetFileName("config")
	configurator.SetFilePaths([]string{
		".",
	})
	
	Config.Secret = "my test secret" // Because we set Secret directly prior to calling configurator.Load() it won't be overridden by configurator.Load()
	err := configurator.Load(&Config)
	
	if err != nil {
		// Always handle your errors
		log.Fatalf("Unable to load application configuration! Error: %s", err.Error())
	}
}
```


## Bugs & Feature Requests

There is **no support** offered with this component. If you would like a feature or find a bug, please submit a feature request through the [GitHub issue tracker](https://github.com/illyabusigin/configurator/issues).

Pull-requests for bug-fixes and features are welcome!


## Attribution

| Component     | Description   | License  |
| :------------ |:-------------| :----:|
| [viper](https://github.com/spf13/viper)      | Go configuration with fangs | [MIT](https://github.com/spf13/viper/blob/master/LICENSE) |
| [pflag](https://github.com/spf13/pflag)      | Drop-in replacement for Go's flag package, implementing POSIX/GNU-style --flags.      |   [MIT](https://github.com/spf13/pflag/blob/master/LICENSE) |
