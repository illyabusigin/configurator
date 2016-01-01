package configurator

import (
	_ "fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testEnvConfigWithDefaults struct {
	Config

	// Annotation your properties with where config data can come from
	Environment string  `env:"APP_ENV" default:"development"`
	Host        string  `env:"APP_HOST" default:"localhost"`
	Port        int     `env:"APP_PORT" default:"3306"`
	TestFloat   float32 `env:"APP_FLOAT" default:"1.5"`
	TestBool    bool    `env:"APP_BOOL" default:"true"`
}

func TestLoadFromDefaultsSuccess(t *testing.T) {
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
