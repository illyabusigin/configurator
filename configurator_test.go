package configurator

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestLoadFromDefaultsSuccess(t *testing.T) {
	type testEnvConfigWithDefaults struct {
		Config

		// Annotation your properties with where config data can come from
		Environment string  `env:"APP_ENV" default:"development"`
		Host        string  `env:"APP_HOST" default:"localhost"`
		Port        int     `env:"APP_PORT" default:"3306"`
		TestFloat   float32 `env:"APP_FLOAT" default:"1.5"`
		TestBool    bool    `env:"APP_BOOL" default:"true"`
	}

	config := testEnvConfigWithDefaults{}
	os.Clearenv()

	err := config.Load(&config)
	assert.Nil(t, err)

	assert.Equal(t, "development", config.Environment)
	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, 3306, config.Port)
	assert.Equal(t, float32(1.5), config.TestFloat)
	assert.True(t, config.TestBool)
}

func TestLoadFromDefaultsWithoutOverridingAlreadySetValues(t *testing.T) {
	type testEnvConfigWithDefaults struct {
		Config

		// Annotation your properties with where config data can come from
		Environment string  `env:"APP_ENV" default:"development"`
		Host        string  `env:"APP_HOST" default:"localhost"`
		Port        int     `env:"APP_PORT" default:"3306"`
		TestFloat   float32 `env:"APP_FLOAT" default:"1.5"`
		TestBool    bool    `env:"APP_BOOL" default:"true"`
	}

	config := testEnvConfigWithDefaults{}
	config.Host = "testhost"
	config.Port = 1000
	os.Clearenv()

	err := config.Load(&config)
	assert.Nil(t, err)

	assert.Equal(t, "development", config.Environment)
	assert.Equal(t, "testhost", config.Host)
	assert.Equal(t, 1000, config.Port)
	assert.Equal(t, float32(1.5), config.TestFloat)
	assert.True(t, config.TestBool)
}

func TestLoadFromDefaultsFailureBadFloatDefault(t *testing.T) {
	type badDefaults struct {
		Config
		TestFloat float32 `env:"APP_FLOAT" default:"badvalue"`
	}

	config := badDefaults{}
	os.Clearenv()

	err := config.Load(&config)
	assert.NotNil(t, err)
}

func TestLoadFromDefaultsFailureBadBoolDefault(t *testing.T) {
	type badDefaults struct {
		Config
		TestBool bool `env:"APP_BOOL" default:"badvalue"`
	}

	config := badDefaults{}
	os.Clearenv()

	err := config.Load(&config)
	assert.NotNil(t, err)
}

func TestLoadFromDefaultsFailureBadIntDefault(t *testing.T) {
	type badDefaults struct {
		Config
		TestPort int `env:"APP_PORT" default:"badvalue"`
	}

	config := badDefaults{}
	os.Clearenv()

	err := config.Load(&config)
	assert.NotNil(t, err)
}

func TestLoadFromEnvVariablesSuccess(t *testing.T) {
	type testEnvConfigWithDefaults struct {
		Config

		// Annotation your properties with where config data can come from
		Environment string  `env:"APP_ENV" default:"development"`
		Host        string  `env:"APP_HOST" default:"localhost"`
		Port        int     `env:"APP_PORT" default:"3306"`
		TestFloat   float32 `env:"APP_FLOAT" default:"1.5"`
		TestBool    bool    `env:"APP_BOOL" default:"true"`
	}

	config := testEnvConfigWithDefaults{}
	os.Clearenv()
	os.Setenv("APP_ENV", "dev")
	os.Setenv("APP_HOST", "127.0.0.1")
	os.Setenv("APP_PORT", "4000")
	os.Setenv("APP_FLOAT", "3.14159")
	os.Setenv("APP_BOOL", "false")

	err := config.Load(&config)
	assert.Nil(t, err)

	assert.Equal(t, "dev", config.Environment)
	assert.Equal(t, "127.0.0.1", config.Host)
	assert.Equal(t, 4000, config.Port)
	assert.Equal(t, float32(3.14159), config.TestFloat)
	assert.False(t, config.TestBool)
}

func TestLoadConfigFromFlagsSuccess(t *testing.T) {
	type testFlagConfig struct {
		Config

		// Annotation your properties with where config data can come from
		Environment string  `flag:"env" default:"development"`
		Host        string  `flag:"host" default:"localhost"`
		Port        int     `flag:"port" default:"3306"`
		Version     float32 `flag:"version" default:"1.5"`
		TestBool    bool    `flag:"testbool" default:"false"`
	}

	config := testFlagConfig{}
	config.viper = viper.New()
	os.Clearenv()

	ptrRef := reflect.ValueOf(&config)
	structRef := ptrRef.Elem()
	flagValues := parseFlagValues(structRef)
	flagSet := config.bindFlagValues(flagValues)

	fmt.Println("flag vals", flagValues)

	var expectedFlagValues = map[string]string{
		"env":      "dev",
		"host":     "127.0.0.1",
		"port":     "4000",
		"version":  "3.14159",
		"testbool": "true",
	}

	// Update the flag values with actual
	flagSet.VisitAll(func(flag *pflag.Flag) {
		flag.Value.Set(expectedFlagValues[flag.Name])
		flag.Changed = true
	})

	defaultValues := parseDefaultValues(structRef)
	config.populateDefaults(defaultValues)
	config.populateConfigStruct(structRef)

	assert.Equal(t, "dev", config.Environment)
	assert.Equal(t, "127.0.0.1", config.Host)
	assert.Equal(t, 4000, config.Port)
	assert.Equal(t, float32(3.14159), config.Version)
	assert.True(t, config.TestBool)
}

func TestLoadConfigFromFlagsFailureUseDefaults(t *testing.T) {
	type testFlagConfig struct {
		Config

		// Annotation your properties with where config data can come from
		Environment string  `flag:"env1" default:"development"`
		Host        string  `flag:"host1" default:"localhost"`
		Port        int     `flag:"port1" default:"3306"`
		Version     float32 `flag:"version1" default:"1.5"`
		TestBool    bool    `flag:"testbool1" default:"false"`
	}

	config := testFlagConfig{}
	config.Load(&config)

	// Config should be loaded with defaults
	assert.Equal(t, "development", config.Environment)
	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, 3306, config.Port)
	assert.Equal(t, float32(1.5), config.Version)
	assert.False(t, config.TestBool)
}

func TestLoadConfigFromFlagsFailureBadValues(t *testing.T) {
	jww.SetStdoutThreshold(jww.LevelTrace)
	type testFlagConfig struct {
		Config

		// Annotation your properties with where config data can come from
		Environment string  `flag:"env2" default:"development"`
		Host        string  `flag:"host2" default:"localhost"`
		Port        int     `flag:"port2" default:"3306"`
		Version     float32 `flag:"version2" default:"1.5"`
		TestBool    bool    `flag:"testbool2" default:"false"`
	}

	config := testFlagConfig{}
	config.viper = viper.New()
	os.Clearenv()

	ptrRef := reflect.ValueOf(&config)
	structRef := ptrRef.Elem()
	flagValues := parseFlagValues(structRef)
	flagSet := config.bindFlagValues(flagValues)

	var expectedFlagValues = map[string]string{
		"env2":      "dev",
		"host2":     "127.0.0.1",
		"port2":     "asb",
		"version2":  "3.14159",
		"testbool2": "true",
	}

	// Update the flag values with actual
	flagSet.VisitAll(func(flag *pflag.Flag) {
		flag.Value.Set(expectedFlagValues[flag.Name])
		flag.Changed = true
	})

	defaultValues := parseDefaultValues(structRef)
	config.populateDefaults(defaultValues)
	err := config.populateConfigStruct(structRef)

	assert.NotNil(t, err)
}
