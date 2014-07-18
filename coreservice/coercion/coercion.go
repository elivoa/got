package coercion

import (
	"github.com/elivoa/got/config"
	"github.com/elivoa/got/core/exception"
	"github.com/elivoa/got/logs"
	"github.com/gorilla/schema"
	"reflect"
	"time"
)

var coercionLogger = logs.Get(logs.LOGGER_INJECTION_VALUE_COERCION)

// global tools
var SchemaDecoder = schema.NewDecoder()

// helper variables.
var invalidValue = reflect.Value{} // stealed from gorilla/context

func init() {
	// Empty velues are set to empty, not ignore as default.
	SchemaDecoder.ZeroEmpty(true)
	SchemaDecoder.IgnoreUnknownKeys(true)

	SchemaDecoder.RegisterConverter(time.Now(), timeCoercion)
}

// string into time.Time, used programmingly.
func Time(value string) (*time.Time, error) {
	for _, format := range config.ValidTimeFormats {
		if t, err := time.Parse(format, value); err == nil {
			if coercionLogger.Debug() {
				coercionLogger.Printf("Translate '%s' into [%v]", value, t)
			}
			return &t, nil
		}
	}
	e := exception.NewCoercionErrorf("Can't parse '%s' into time using any format in config.ValidTimeFormat.",
		value)
	return nil, e
}

func timeCoercion(value string) reflect.Value {
	t, err := Time(value)
	if err != nil {
		if coercionLogger.Info() {
			coercionLogger.Print(err.Error())
		}
		if config.IgnoreInjectionParseError {
			return config.DefaultTimeReflectValue
		} else {
			return invalidValue // will return error when return invalidValue.
		}
	}
	return reflect.ValueOf(*t) // t is an address.
}
