package validators

import (
	"errors"
	"strings"
)

var (
	MinValidationError      = errors.New("minimum validation failed")
	MaxValidationError      = errors.New("maximum validation failed")
	LenValidationError      = errors.New("len validation failed")
	InValidationError       = errors.New("in validation failed")
	NotEmptyValidationError = errors.New("not empty validation failed")
)

type Validator interface {
	Validate(any) error
}

type MinValidator struct {
	Arg int
}

func (mv MinValidator) Validate(val any) error {
	switch v := val.(type) {
	case string:
		if len(v) < mv.Arg {
			return MinValidationError
		}
	case int:
		if v < mv.Arg {
			return MinValidationError
		}
	case []int:
		return validateSlice(v, mv)
	case []string:
		return validateSlice(v, mv)
	default:
	}
	return nil
}

type MaxValidator struct {
	Arg int
}

func (mv MaxValidator) Validate(val any) error {
	switch v := val.(type) {
	case string:
		if len(v) > mv.Arg {
			return MaxValidationError
		}
	case int:
		if v > mv.Arg {
			return MaxValidationError
		}
	case []int:
		return validateSlice(v, mv)
	case []string:
		return validateSlice(v, mv)
	default:
	}
	return nil
}

type LenValidator struct {
	Arg int
}

func (lv LenValidator) Validate(val any) error {
	switch v := val.(type) {
	case string:
		if len(v) != lv.Arg {
			return LenValidationError
		}
	case []int:
		return validateSlice(v, lv)
	case []string:
		return validateSlice(v, lv)
	default:
	}
	return nil
}

type InValidator[T comparable] struct {
	Args []T
}

func (iv InValidator[T]) Validate(val any) error {
	switch v := val.(type) {
	case []int:
		return validateSlice(v, iv)
	case []string:
		return validateSlice(v, iv)
	case int, string:
		if vv, ok := val.(T); ok {
			for _, vs := range iv.Args {
				if vs == vv {
					return nil
				}
			}
		}
	}
	return InValidationError
}

func validateSlice[T string | int](s []T, validator Validator) error {
	for _, v := range s {
		if err := validator.Validate(v); err != nil {
			return err
		}
	}
	return nil
}

type NotEmptyValidator struct{}

func (nev NotEmptyValidator) Validate(val any) error {
	switch v := val.(type) {
	case string:
		if len(strings.TrimSpace(v)) == 0 {
			return NotEmptyValidationError
		}
	case []string:
		return validateSlice(v, nev)
	default:
	}
	return nil
}
