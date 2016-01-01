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

/*
How to specify config name?
How to specify config file paths?
How to wath config file for changes?
How to specify remote configurations?
How to not override already specified config values?
*/

const (
	tagEnv     = "env"
	tagFlag    = "flag"
	tagFile    = "file"
	tagDefault = "default"
)

var (
	ErrNoConfigValuesDetected = errors.New("No configuration values detected!")
	ErrNotStruct              = errors.New("Value does not appear to be a struct!")
	ErrNotStructPointer       = errors.New("Value passed was not a struct pointer!")
	ErrUnsupportedFieldType   = errors.New("Unsupported struct field type!")
)

type Config struct {
	ConfigFileName  string // name of config file (without extension)
	ConfigFilePaths []string
	WatchConfigFile bool

	externalConfig *interface{}
	viper          *viper.Viper
}

// Load blah blah
func (c *Config) Load(val interface{}) error {
	c.viper = viper.New()

	canLoadErr := c.canLoad(val)
	if canLoadErr != nil {
		return canLoadErr
	}

	ptrRef := reflect.ValueOf(val)
	ref := ptrRef.Elem()

	return c.parseStructConfigValues(ref, val)
}

func (c *Config) canLoad(val interface{}) error {
	ptrRef := reflect.ValueOf(val)
	if ptrRef.Kind() != reflect.Ptr {
		return ErrNotStructPointer
	}
	elemRef := ptrRef.Elem()
	if elemRef.Kind() != reflect.Struct {
		return ErrNotStruct
	}

	return nil
}

// Check to see if we can load, value must me a struct pointer
// Parse configuration values
// Populate defaults
// Populate flags, env, config file values
// FUTURE: Populate remote config
// Fetch values from viper, populate struct

func (c *Config) parseStructConfigValues(structRef reflect.Value, val interface{}) error {
	// Parse configurator values on our struct
	defaultValues := parseDefaultValues(structRef)
	envValues := parseEnvValues(structRef)
	flagValues := parseFlagValues(structRef)
	// configValues := parseConfigFileValues(structRef)

	c.populateDefaults(defaultValues)
	c.bindEnvValues(envValues)
	c.bindFlagValues(flagValues)
	// c.bindConfigFileValues(configValues)

	// Populate config values
	err := c.populateConfigStruct(structRef)

	return err
}

//////////
// Parsing
//////////

type parsedValue struct {
	tagValue  string
	fieldType reflect.Type
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

///////////
// Populate
///////////

func (c *Config) populateDefaults(defaultValues map[string]parsedValue) {
	for k, v := range defaultValues {
		fmt.Printf("Setting default <%v> for field: <%s>\n", v.tagValue, k)
		c.viper.SetDefault(k, v.tagValue)
	}
}

func (c *Config) populateConfigStruct(structRef reflect.Value) error {
	structType := structRef.Type()
	for i := 0; i < structType.NumField(); i++ {
		structField := structType.Field(i)
		configValue := c.viper.Get(structField.Name)
		fmt.Printf("configValue: <%v> Field: <%s>\n", configValue, structField.Name)
		if configValue != nil {
			err := populateStructField(structField, structRef.Field(i), fmt.Sprintf("%v", configValue))

			if err != nil {
				return err
			}
		}
	}

	return nil
}

//////////
// Binding
//////////

func (c *Config) bindEnvValues(envValues map[string]parsedValue) {
	for k, v := range envValues {
		c.viper.BindEnv(k, v.tagValue)
	}
}

func (c *Config) bindFlagValues(flagValues map[string]parsedValue) *pflag.FlagSet {
	flagSet := pflag.NewFlagSet("configurator", pflag.PanicOnError)

	for k, v := range flagValues {

		// Introspect the variable type, instantiate a pflag
		switch v.fieldType.Kind() {
		case reflect.String:
			pflag.String(v.tagValue, "", "")
		case reflect.Bool:
			pflag.Bool(v.tagValue, false, "")
		case reflect.Float32, reflect.Float64:
			pflag.Float64(v.tagValue, 0, "")
		case reflect.Int:
			pflag.Int64(v.tagValue, 0, "")
		}

		flag := pflag.Lookup(v.tagValue)
		if flag != nil {
			c.viper.BindPFlag(k, flag)
			flagSet.AddFlag(flag)
		}
	}

	return flagSet
}

func (c *Config) bindConfigFileValues(configValues map[string]parsedValue) {
	// for k, v := range configValues {
	// 	//c.viper.BindEnv(k, v)
	// }
}

func populateStructField(field reflect.StructField, fieldValue reflect.Value, value string) error {
	switch fieldValue.Kind() {
	case reflect.String:
		fieldValue.SetString(value)
	case reflect.Bool:
		bvalue, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("Unable to convert value (%s) for to bool for field: %s! Error: %s", value, field.Name, err.Error())
		}
		fieldValue.SetBool(bvalue)
	case reflect.Float32, reflect.Float64:
		floatValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("Unable to convert value (%s) for to float for field: %s! Error: %s", value, field.Name, err.Error())
		}
		fieldValue.SetFloat(floatValue)
	case reflect.Int:
		intValue, err := strconv.ParseInt(value, 10, 64)
		fmt.Printf("populateStructField INT <%s> Value: <%v>, Field: <%s>\n", value, intValue, field.Name)
		if err != nil {
			return fmt.Errorf("Unable to convert value (%s) for to int for field: %s! Error: %s", value, field.Name, err.Error())
		}
		fieldValue.SetInt(intValue)
	default:
		return ErrUnsupportedFieldType
	}
	return nil
}

func isZeroOfUnderlyingType(x interface{}) bool {
	return x == reflect.Zero(reflect.TypeOf(x)).Interface()
}
