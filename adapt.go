package adapter

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"

	routing "github.com/jackwhelpton/fasthttp-routing/v2"
)

type ValidationError struct {
	error
	Code ErrCode
}

func NewValidationError(msg string, args ...interface{}) ValidationError {
	return ValidationError{
		error: fmt.Errorf(msg, args...),
	}
}

func NewMissingRequiredParamError(msg string, args ...interface{}) ValidationError {
	return ValidationError{
		error: fmt.Errorf(msg, args...),
		Code:  MissingRequiredParamError,
	}
}

type ErrCode uint

const (
	NoError ErrCode = iota
	UnexpectedError
	MissingRequiredParamError
)

func ErrorCode(err error) ErrCode {
	if err == nil {
		return NoError
	}

	validationError, ok := err.(ValidationError)
	if !ok {
		return UnexpectedError
	}

	return validationError.Code
}

var ctxType = reflect.TypeOf(&routing.Context{})
var errType = reflect.TypeOf(new(error)).Elem()

func Adapt(fn interface{}) func(ctx *routing.Context) error {
	t := reflect.TypeOf(fn)
	v := reflect.ValueOf(fn)

	if t.Kind() != reflect.Func {
		log.Fatal("adapt's argument must be a function!")
	}

	if t.NumIn() != 2 {
		log.Fatal("received function must have 2 arguments!")
	}

	if t.In(0) != ctxType {
		log.Fatal("first argument must be of type *routing.Context!")
	}

	if t.NumOut() != 1 {
		log.Fatal("received function must have a single return value!")
	}

	if t.Out(0) != errType {
		log.Fatal("first return value must be of type error")
	}

	argsType := t.In(1)
	if argsType.Kind() != reflect.Struct {
		log.Fatal("second argument must a struct!")
	}

	bodyInfo := getBodyInfo(argsType)
	pathParams, headerParams, queryParams := getTagNames(argsType)
	return func(ctx *routing.Context) error {
		args := reflect.New(argsType)

		if bodyInfo != nil {
			param := reflect.New(argsType.Field(bodyInfo.Idx).Type)
			err := json.Unmarshal(ctx.PostBody(), param.Interface())
			if err != nil {
				return routing.NewHTTPError(http.StatusBadRequest, fmt.Sprintf(
					"could not parse body as JSON: %s", err.Error(),
				))
			}

			args.Elem().Field(bodyInfo.Idx).Set(param.Elem())
		}

		for key, info := range pathParams {
			param := ctx.Param(key)
			if param == "" && info.Required {
				return NewMissingRequiredParamError("path param '%s' is empty", key)
			}

			args.Elem().Field(info.Idx).Set(reflect.ValueOf(param))
		}
		for key, info := range headerParams {
			param := string(ctx.Request.Header.Peek(key))
			if param == "" && info.Required {
				return NewMissingRequiredParamError("required header param '%s' is empty", key)
			}

			args.Elem().Field(info.Idx).Set(reflect.ValueOf(param))
		}
		for key, info := range queryParams {
			param := string(ctx.Request.URI().QueryArgs().Peek(key))
			if param == "" && info.Required {
				return NewMissingRequiredParamError("required query param '%s' is empty", key)
			}

			args.Elem().Field(info.Idx).Set(reflect.ValueOf(param))
		}

		err, _ := v.Call([]reflect.Value{reflect.ValueOf(ctx), args.Elem()})[0].Interface().(error)
		return err
	}
}

type tagInfo struct {
	Idx      int
	Required bool
}

func getBodyInfo(t reflect.Type) *tagInfo {
	for i := 0; i < t.NumField(); i++ {
		name := t.Field(i).Name
		if name == "JSONBody" {
			return &tagInfo{
				Idx:      i,
				Required: true,
			}
		}
	}

	return nil
}

// This function collects only the names
// that will be used from the type
// this should save several calls to `Field(i).Tag.Get("foo")`
// which might improve the performance by a lot.
func getTagNames(t reflect.Type) (map[string]tagInfo, map[string]tagInfo, map[string]tagInfo) {
	pathParams := map[string]tagInfo{}
	headerParams := map[string]tagInfo{}
	queryParams := map[string]tagInfo{}

	for i := 0; i < t.NumField(); i++ {
		opts := strings.Split(t.Field(i).Tag.Get("path"), ",")

		key := opts[0]
		if key == "" {
			continue
		}

		pathParams[key] = tagInfo{
			Idx:      i,
			Required: true,
		}
	}

	for i := 0; i < t.NumField(); i++ {
		opts := strings.Split(t.Field(i).Tag.Get("header"), ",")

		key := opts[0]
		if key == "" {
			continue
		}

		required := true
		if len(opts) > 1 && opts[1] == "optional" {
			required = false
		}

		headerParams[key] = tagInfo{
			Idx:      i,
			Required: required,
		}
	}

	for i := 0; i < t.NumField(); i++ {
		opts := strings.Split(t.Field(i).Tag.Get("query"), ",")
		key := opts[0]
		if key == "" {
			continue
		}

		required := false
		if len(opts) > 1 && opts[1] == "required" {
			required = true
		}

		queryParams[key] = tagInfo{
			Idx:      i,
			Required: required,
		}
	}

	return pathParams, headerParams, queryParams
}
