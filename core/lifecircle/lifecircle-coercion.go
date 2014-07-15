package lifecircle

import (
	"github.com/elivoa/got/config"
	"github.com/elivoa/got/logs"
	"github.com/gorilla/schema"
	"reflect"
	"time"
)

var coercionLogger = logs.Get(logs.LOGGER_INJECTION_VALUE_COERCION)

// global tools
var SchemaDecoder = schema.NewDecoder()

// helper variables.
var invalidValue = reflect.Value{}

func init() {
	// Empty velues are set to empty, not ignore as default.
	SchemaDecoder.ZeroEmpty(true)
	SchemaDecoder.IgnoreUnknownKeys(true)

	SchemaDecoder.RegisterConverter(time.Now(), convertTime)
}

func convertTime(value string) reflect.Value {
	for _, format := range config.ValidTimeFormats {
		if t, err := time.Parse(format, value); err == nil {
			return reflect.ValueOf(t)
		}
	}
	if coercionLogger.Info() {
		coercionLogger.Printf("Can't parse '%s' into time using any format in config.ValidTimeFormat.", value)
	}
	if config.IgnoreInjectionParseError {
		return config.DefaultTimeReflectValue
	} else {
		return invalidValue // will return error when return invalidValue.
	}
}
