// Package configurator implements simple configuration built on top of Viper.
//
// It aims to make application configuration as easy as possible.
// This is accomplish by allowing you to annotate your struct with env, file,
// flag, default annotations that will tell Viper where to where to look for
// configuration values. Since configurator is built on top of Viper the
// same source precedence is applicable.
//
// The priority of config sources is the following:
// 1. Overrides, or setting the config struct field directly.
// 2. Flags - note that github.com/spf13/pflag is used  in lieu of the standard flag package.
// 3. Environment variables.
// 4. Configuration file values.
// 5. Default values.
//
// NOTE: Viper key/value store and/or watching config sources is not yet supported.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package configurator

import (
	"errors"
	"fmt"
	"go/ast"
	"reflect"
	"strconv"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	tagEnv     = "env"
	tagFlag    = "flag"
	tagFile    = "file"
	tagDefault = "default"
)

var (
	// ErrValueNotStruct is returned when value passed to config.Load() is not a struct.
	ErrValueNotStruct = errors.New("Value does not appear to be a struct!")

	// ErrValueNotStructPointer is returned when value passed to config.Load() is not a pointer to a struct.
	ErrValueNotStructPointer = errors.New("Value passed was not a struct pointer!")
)

var c *Config

func init() {
	c = &Config{
		FileName:  "config",
		FilePaths: []string{"."},
	}
}

// Config is a convenience configuration struct built on top of Viper. You
// use Config by annotating your configuration struct with env, flag, file,
// and default tags which will be parsed by Config. You can either
// embed Configurator.Config in your struct or reference configurator.Load()
// directly. The priority of the sources is the same Viper:
// 1. overrides
// 2. flags
// 3. env. variables
// 4. config file
// 5. defaults
//
// For example, if you embedded configurator.Config in your struct and
// configured it like so:
//
//  type AppConfig struct {
//	configurator.Config
//	Secret      string `file:"secret" env:"APP_SECRET" flag:"secret" default:"abc123xyz"`
//	User        string `file:"user" env:"APP_USER" flag:"user" default:"root"`
//	Environment string `file:"env" env:"APP_ENV" flag:"env" default:"dev"`
//   }
//
// Assuming your source values were the following:
//  File : {
// 	"user": "test_user"
//	"secret": "defaultsecret"
//  }
//  Env : {
//  	"APP_SECRET": "somesecretkey"
//  }
//
// This is how you would load the configuration:
//
//  func loadConfig() {
// 	config := AppConfig{}
// 	err := config.Load(&config)
//
// 	if err != nil {
// 		// Always handle your errors
// 		log.Fatalf("Unable to load application configuration! Error: %s", err.Error())
// 	}
//
// 	fmt.Println("config.Secret =", config.Secret)           // somesecretkey, from env
// 	fmt.Println("config.User =", config.User)               // test_user, from file
// 	fmt.Println("config.Environment =", config.Environment) // dev, from defaults
//  }
//
type Config struct {
	// FileName is the name of the configuration file without any extensions.
	FileName string

	// FilePaths is an array of configuration file paths to search for the configuration file.
	FilePaths []string

	externalConfig *interface{}
	viper          *viper.Viper
}

// Load attempts to populate the struct with configuration values.
// The value passed to load must be a struct reference or an error
// will be returned.
func Load(structRef interface{}) error {
	return c.Load(structRef)
}

// SetFileName specifies the name of the configuration file without any extensions.
func SetFileName(fileName string) {
	c.FileName = fileName
}

// SetFilePaths specifies the array of configuration file paths to search for the configuration file.
func SetFilePaths(filePaths []string) {
	c.FilePaths = filePaths
}

// Load attempts to populate the struct with configuration values.
// The value passed to load must be a struct reference or an error
// will be returned.
func (c *Config) Load(structRef interface{}) error {
	c.viper = viper.New()

	canLoadErr := c.canLoad(structRef)
	if canLoadErr != nil {
		return canLoadErr
	}

	ptrRef := reflect.ValueOf(structRef)
	ref := ptrRef.Elem()

	return c.parseStructConfigValues(ref, structRef)
}

func (c *Config) canLoad(structRef interface{}) error {
	ptrRef := reflect.ValueOf(structRef)
	if ptrRef.Kind() != reflect.Ptr {
		return ErrValueNotStructPointer
	}
	elemRef := ptrRef.Elem()
	if elemRef.Kind() != reflect.Struct {
		return ErrValueNotStruct
	}

	return nil
}

/////////////
// Parsing //
/////////////

type parsedValue struct {
	tagValue  string
	fieldType reflect.Type
}

func (c *Config) parseStructConfigValues(structRef reflect.Value, val interface{}) error {
	// Parse configurator values on our struct
	defaultValues := parseDefaultValues(structRef)
	envValues := parseEnvValues(structRef)
	flagValues := parseFlagValues(structRef)
	configValues := parseConfigFileValues(structRef)

	c.populateDefaults(defaultValues)
	c.bindEnvValues(envValues)
	c.bindFlagValues(flagValues)
	c.bindConfigFileValues(configValues)

	err := c.populateConfigStruct(structRef)

	return err
}

func parseDefaultValues(structRef reflect.Value) map[string]parsedValue {
	values := parseValuesForTag(structRef, tagDefault)
	return values
}

func parseEnvValues(structRef reflect.Value) map[string]parsedValue {
	values := parseValuesForTag(structRef, tagEnv)
	return values
}

func parseFlagValues(structRef reflect.Value) map[string]parsedValue {
	values := parseValuesForTag(structRef, tagFlag)
	return values
}

func parseConfigFileValues(structRef reflect.Value) map[string]parsedValue {
	values := parseValuesForTag(structRef, tagFile)
	return values
}

func parseValuesForTag(structRef reflect.Value, tagName string) map[string]parsedValue {
	values := map[string]parsedValue{}

	structType := structRef.Type()
	for i := 0; i < structType.NumField(); i++ {
		structField := structType.Field(i)
		tag := structField.Tag
		tagValue := tag.Get(tagName)

		if tagValue != "" && ast.IsExported(structField.Name) {
			values[structField.Name] = parsedValue{tagValue, structField.Type}
		}
	}

	return values
}

/////////////
// Binding //
/////////////

func (c *Config) bindEnvValues(envValues map[string]parsedValue) {
	for k, v := range envValues {
		c.viper.BindEnv(k, v.tagValue)
	}
}

func (c *Config) bindFlagValues(flagValues map[string]parsedValue) *pflag.FlagSet {
	flagSet := pflag.NewFlagSet("configurator", pflag.PanicOnError)

	for k, v := range flagValues {
		pflag.String(v.tagValue, "", "")
		flag := pflag.Lookup(v.tagValue)

		c.viper.BindPFlag(k, flag)
		flagSet.AddFlag(flag)
	}

	return flagSet
}

func (c *Config) bindConfigFileValues(configValues map[string]parsedValue) {
	c.viper.SetConfigName(c.FileName)

	for _, filePath := range c.FilePaths {
		c.viper.AddConfigPath(filePath)
	}

	// Map the config file keys to our variable
	for k, v := range configValues {
		c.viper.RegisterAlias(k, v.tagValue)
	}
}

//////////////
// Populate //
//////////////

func (c *Config) populateDefaults(defaultValues map[string]parsedValue) {
	for k, v := range defaultValues {
		c.viper.SetDefault(k, v.tagValue)
	}
}

func (c *Config) populateConfigStruct(structRef reflect.Value) error {
	c.viper.ReadInConfig()

	structType := structRef.Type()
	for i := 0; i < structType.NumField(); i++ {
		structField := structType.Field(i)
		configValue := c.viper.Get(structField.Name)
		if configValue != nil {
			err := populateStructField(structField, structRef.Field(i), fmt.Sprintf("%v", configValue))

			if err != nil {
				return err
			}
		}
	}

	return nil
}

func populateStructField(field reflect.StructField, fieldValue reflect.Value, value string) error {
	switch fieldValue.Kind() {
	case reflect.String:
		if isZeroOfUnderlyingType(fieldValue.Interface()) {
			fieldValue.SetString(value)
		}

	case reflect.Bool:
		bvalue, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("Unable to convert value (%s) for to bool for field: %s! Error: %s", value, field.Name, err.Error())
		}

		if isZeroOfUnderlyingType(fieldValue.Interface()) {
			fieldValue.SetBool(bvalue)
		}

	case reflect.Float32, reflect.Float64:
		floatValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("Unable to convert value (%s) for to float for field: %s! Error: %s", value, field.Name, err.Error())
		}

		if isZeroOfUnderlyingType(fieldValue.Interface()) {
			fieldValue.SetFloat(floatValue)
		}

	case reflect.Int, reflect.Int8, reflect.Int32, reflect.Int64:
		intValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("Unable to convert value (%s) for to int for field: %s! Error: %s", value, field.Name, err.Error())
		}

		if isZeroOfUnderlyingType(fieldValue.Interface()) {
			fieldValue.SetInt(intValue)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint32, reflect.Uint64:
		intValue, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return fmt.Errorf("Unable to convert value (%s) for to unsigned int for field: %s! Error: %s", value, field.Name, err.Error())
		}

		if isZeroOfUnderlyingType(fieldValue.Interface()) {
			fieldValue.SetUint(intValue)
		}
	}
	return nil
}

/////////////
// Utility //
/////////////

func isZeroOfUnderlyingType(x interface{}) bool {
	// Source: http://stackoverflow.com/questions/13901819/quick-way-to-detect-empty-values-via-reflection-in-go
	return x == reflect.Zero(reflect.TypeOf(x)).Interface()
}
