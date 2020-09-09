package adapter

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	routing "github.com/jackwhelpton/fasthttp-routing/v2"
)

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
	pathParams, headerParams, queryParams, userValues := getTagNames(argsType)
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
				return routing.NewHTTPError(http.StatusBadRequest, fmt.Sprintf(
					"path param '%s' is empty", key,
				))
			}

			v, err := decodeType(info.Kind, param)
			if err != nil {
				return routing.NewHTTPError(http.StatusBadRequest, fmt.Sprintf(
					"could not convert path param to %s: %s", reflect.Int, err.Error(),
				))
			}

			args.Elem().Field(info.Idx).Set(v)
		}
		for key, info := range headerParams {
			param := string(ctx.Request.Header.Peek(key))
			if param == "" && info.Required {
				return routing.NewHTTPError(http.StatusBadRequest, fmt.Sprintf(
					"required header param '%s' is empty", key,
				))
			}

			v, err := decodeType(info.Kind, param)
			if err != nil {
				return routing.NewHTTPError(http.StatusBadRequest, fmt.Sprintf(
					"could not convert path param to %s: %s", reflect.Int, err.Error(),
				))
			}

			args.Elem().Field(info.Idx).Set(v)
		}
		for key, info := range queryParams {
			param := string(ctx.Request.URI().QueryArgs().Peek(key))
			if param == "" && info.Required {
				return routing.NewHTTPError(http.StatusBadRequest, fmt.Sprintf(
					"required query param '%s' is empty", key,
				))
			}

			v, err := decodeType(info.Kind, param)
			if err != nil {
				return routing.NewHTTPError(http.StatusBadRequest, fmt.Sprintf(
					"could not convert path param to %s: %s", reflect.Int, err.Error(),
				))
			}

			args.Elem().Field(info.Idx).Set(v)
		}

		for key, info := range userValues {
			param := ctx.UserValue(key)
			if info.Required && param == nil {
				return routing.NewHTTPError(http.StatusBadRequest, fmt.Sprintf(
					"required user value '%s' is empty", key,
				))
			}

			paramV := reflect.ValueOf(param)
			canConvert := paramV.Type().ConvertibleTo(info.Type)
			if !canConvert {
				return routing.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf(
					"could not convert userValue %s to type %T", param, info.Type,
				))
			}

			args.Elem().Field(info.Idx).Set(paramV.Convert(info.Type))
		}

		err, _ := v.Call([]reflect.Value{reflect.ValueOf(ctx), args.Elem()})[0].Interface().(error)
		return err
	}
}

func decodeType(kind reflect.Kind, v string) (reflect.Value, error) {
	switch kind {
	case reflect.Int:
		i, err := strconv.Atoi(v)
		return reflect.ValueOf(i), err
	case reflect.Int8:
		i, err := strconv.ParseInt(v, 10, 8)
		return reflect.ValueOf(int8(i)), err
	case reflect.Int16:
		i, err := strconv.ParseInt(v, 10, 16)
		return reflect.ValueOf(int16(i)), err
	case reflect.Int32:
		i, err := strconv.ParseInt(v, 10, 32)
		return reflect.ValueOf(int32(i)), err
	case reflect.Int64:
		i, err := strconv.ParseInt(v, 10, 64)
		return reflect.ValueOf(int64(i)), err

	case reflect.Uint:
		i, err := strconv.Atoi(v)
		return reflect.ValueOf(uint(i)), err
	case reflect.Uint8:
		i, err := strconv.ParseUint(v, 10, 8)
		return reflect.ValueOf(uint8(i)), err
	case reflect.Uint16:
		i, err := strconv.ParseUint(v, 10, 16)
		return reflect.ValueOf(uint16(i)), err
	case reflect.Uint32:
		i, err := strconv.ParseUint(v, 10, 32)
		return reflect.ValueOf(uint32(i)), err
	case reflect.Uint64:
		i, err := strconv.ParseUint(v, 10, 64)
		return reflect.ValueOf(uint64(i)), err
	}

	return reflect.ValueOf(v), nil
}

type tagInfo struct {
	Idx      int
	Required bool
	Kind     reflect.Kind
	Type     reflect.Type
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
func getTagNames(t reflect.Type) (
	pathParams map[string]tagInfo,
	headerParams map[string]tagInfo,
	queryParams map[string]tagInfo,
	userValues map[string]tagInfo,
) {
	pathParams = map[string]tagInfo{}
	headerParams = map[string]tagInfo{}
	queryParams = map[string]tagInfo{}
	userValues = map[string]tagInfo{}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		opts := strings.Split(field.Tag.Get("path"), ",")

		key := opts[0]
		if key == "" {
			continue
		}

		pathParams[key] = tagInfo{
			Idx:      i,
			Required: true,
			Kind:     field.Type.Kind(),
			Type:     field.Type,
		}
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		opts := strings.Split(field.Tag.Get("header"), ",")

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
			Kind:     field.Type.Kind(),
			Type:     field.Type,
		}
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		opts := strings.Split(field.Tag.Get("query"), ",")
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
			Kind:     field.Type.Kind(),
			Type:     field.Type,
		}
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		opts := strings.Split(field.Tag.Get("uservalue"), ",")
		key := opts[0]
		if key == "" {
			continue
		}

		required := true
		if len(opts) > 1 && opts[1] == "optional" {
			required = false
		}

		userValues[key] = tagInfo{
			Idx:      i,
			Required: required,
			Kind:     field.Type.Kind(),
			Type:     field.Type,
		}
	}

	return
}
