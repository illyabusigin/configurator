package configurator

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestConfigLoadFailureScenarios(t *testing.T) {
	type testConfig struct {
		Config
		Value string
	}

	// Pass a struct directly instead of pointer reference
	config := testConfig{}
	err := config.Load(config)

	assert.NotNil(t, err)
	assert.Equal(t, err, ErrNotStructPointer)

	// Pass a non-struct by reference
	badValue := []int{1, 2, 3}
	c := Config{}
	err = c.Load(&badValue)

	assert.NotNil(t, err)
	assert.Equal(t, err, ErrNotStruct)
}

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
		Counter     uint32  `env:"APP_COUNTER" default:"1"`
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
	assert.Equal(t, uint32(1), config.Counter)
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

func TestLoadConfigFromFileSuccess(t *testing.T) {
	type testFileConfig struct {
		Config

		Environment string  `file:"env" default:"development"`
		Host        string  `file:"host" default:"localhost"`
		Port        int     `file:"port" default:"3306"`
		Version     float32 `file:"version" default:"1.5"`
		TestBool    bool    `file:"enabled" default:"false"`
	}

	configData := []byte(`{
    "env": "development",
    "host": "127.0.0.1",
    "port": 3000,
    "version": 2,
    "enabled": true
	}`)

	filePath := "./config.json"
	testConfig, err := os.Create(filePath)

	defer func() {
		testConfig.Close()
		os.Remove(filePath)
	}()

	assert.Nil(t, err)

	_, err = testConfig.Write(configData)
	assert.Nil(t, err, "File write error should be nil!")
	testConfig.Sync()

	config := testFileConfig{}
	config.FileName = "config"
	config.FilePaths = []string{
		".",
	}

	err = config.Load(&config)
	assert.Nil(t, err)

	assert.Equal(t, "development", config.Environment)
	assert.Equal(t, "127.0.0.1", config.Host)
	assert.Equal(t, 3000, config.Port)
	assert.Equal(t, float32(2), config.Version)
	assert.True(t, config.TestBool)
}

func TestLoadConfigFromFileFailureBadValue(t *testing.T) {
	type testFileConfig struct {
		Config

		Environment string  `file:"env" default:"development"`
		Host        uint32  `file:"host" default:"3000"` // changed host to uint32, while it's a string in JSON file
		Port        int     `file:"port" default:"3306"`
		Version     float32 `file:"version" default:"1.5"`
		TestBool    bool    `file:"enabled" default:"false"`
	}

	configData := []byte(`{
    "env": "development",
    "host": "127.0.0.1",
    "port": 3000,
    "version": 2,
    "enabled": true
	}`)

	filePath := "./config.json"
	testConfig, err := os.Create(filePath)

	defer func() {
		testConfig.Close()
		os.Remove(filePath)
	}()

	assert.Nil(t, err)

	_, err = testConfig.Write(configData)
	assert.Nil(t, err, "File write error should be nil!")
	testConfig.Sync()

	config := testFileConfig{}
	config.FileName = "config"
	config.FilePaths = []string{
		".",
	}

	err = config.Load(&config)
	assert.NotNil(t, err)
}

func TestCompoundSourcesScenario(t *testing.T) {
	type testFileConfig struct {
		Config

		Environment string  `file:"env" env:"APP_ENV" flag:"env5" default:"development"`
		Host        string  `file:"host" env:"APP_HOST" flag:"host5" default:"3000"` // changed host to uint32, while it's a string in JSON file
		Port        uint    `file:"port" env:"APP_PORT" flag:"port5" default:"3306"`
		Version     float32 `file:"version" env:"APP_VERSION" flag:"version5" default:"1.5"`
		Restricted  bool    `file:"enabled" env:"APP_RESTRICTED" flag:"restricted" default:"false"`
	}
}
