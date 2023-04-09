package govalidator

import (
	"errors"
	"fmt"
	"github.com/AndrewDanilin/govalidator/validators"
	"reflect"
	"strconv"
	"strings"
)

var ErrNotStruct = errors.New("wrong argument given, should be a struct")
var ErrInvalidValidatorSyntax = errors.New("invalid validator syntax")
var ErrValidateForUnexportedFields = errors.New("validation for unexported field is not allowed")

var ErrUnsupportedType = errors.New("unsupported type")

type ValidationError struct {
	Err error
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	if len(v) == 1 {
		return strings.Split(v[0].Err.Error(), "err: ")[1]
	}

	arr := make([]string, len(v))
	for i, valErr := range v {
		arr[i] = valErr.Err.Error()
	}
	return strings.Join(arr, "\n")
}

func Validate(v any) error {
	errs := ValidationErrors{}
	vType := reflect.TypeOf(v)
	vVal := reflect.ValueOf(v)

	if vType.Kind() != reflect.Struct {
		return ErrNotStruct
	}

	for i := 0; i < vType.NumField(); i++ {
		f := vType.Field(i)
		if vStr, ok := f.Tag.Lookup("validate"); ok {
			if !f.IsExported() {
				errs = append(errs, buildErr(f.Name, ErrValidateForUnexportedFields))
				continue
			}

			var vrs []validators.Validator
			var err error

			switch f.Type.Kind() {
			case reflect.String:
				vrs, err = parseValidators(reflect.String, vStr)
			case reflect.Int:
				vrs, err = parseValidators(reflect.Int, vStr)
			case reflect.Slice:
				vrs, err = parseValidators(f.Type.Elem().Kind(), vStr)
			default:
				errs = append(errs, buildErr(f.Name, ErrUnsupportedType))
				continue
			}

			if err != nil {
				errs = append(errs, buildErr(f.Name, err))
				continue
			}

			for _, vr := range vrs {
				err = vr.Validate(vVal.Field(i).Interface())

				if err != nil {
					errs = append(errs, buildErr(f.Name, err))
				}
			}
		}
	}

	if len(errs) == 0 {
		return nil
	}

	return errs
}

func buildErr(name string, err error) ValidationError {
	return ValidationError{Err: fmt.Errorf("name: %s, err: %s", name, err)}
}

func parseValidators(t reflect.Kind, vStr string) ([]validators.Validator, error) {
	parts := strings.Split(vStr, ";")
	vrs := make([]validators.Validator, len(parts))
	for i, v := range parts {
		vr, err := parseValidator(t, v)
		if err != nil {
			return nil, err
		}
		vrs[i] = vr
	}
	return vrs, nil
}

func parseValidator(
	t reflect.Kind,
	vStr string,
) (validators.Validator, error) {
	parts := strings.Split(vStr, ":")
	switch parts[0] {
	case "max":
		arg, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, ErrInvalidValidatorSyntax
		}
		return validators.MaxValidator{Arg: arg}, nil
	case "min":
		arg, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, ErrInvalidValidatorSyntax
		}
		return validators.MinValidator{Arg: arg}, nil
	case "len":
		arg, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, ErrInvalidValidatorSyntax
		}
		switch t {
		case reflect.String:
			return validators.LenValidator{Arg: arg}, nil
		default:
			return nil, ErrUnsupportedType
		}
	case "in":
		args := strings.Split(parts[1], ",")
		if len(args) == 1 && args[0] == parts[1] {
			args = make([]string, 0)
		}

		switch t {
		case reflect.String:
			return validators.InValidator[string]{Args: args}, nil
		case reflect.Int:
			argsTyped := make([]int, len(args))
			for i := range args {
				n, err := strconv.Atoi(args[i])
				if err != nil {
					return nil, ErrInvalidValidatorSyntax
				}
				argsTyped[i] = n
			}
			return validators.InValidator[int]{Args: argsTyped}, nil
		default:
		}
	case "not_empty":
		switch t {
		case reflect.String:
			return validators.NotEmptyValidator{}, nil
		default:
			return nil, ErrUnsupportedType
		}
	default:
		return nil, ErrInvalidValidatorSyntax
	}
	return nil, ErrInvalidValidatorSyntax
}
