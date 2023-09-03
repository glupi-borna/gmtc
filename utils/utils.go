package utils

import (
	"fmt"
	"errors"
)

func IsBetween(char byte, start byte, end byte) bool {
	return char >= start && char <= end
}

func IsWhitespace(char byte) bool {
	return char == ' ' || char == '\n' || char == '\t' || char == '\r'
}

func IsWhitespaceNoNL(char byte) bool {
	return char == ' ' || char == '\t'
}

func IsHexNumber(char byte) bool {
	return IsNumber(char) || IsBetween(char, 'a', 'f') || IsBetween(char, 'A', 'f')
}

func IsNumber(char byte) bool {
	return IsBetween(char, '0', '9')
}

func IsLetter(char byte) bool {
	return IsBetween(char, 'A', 'Z') || IsBetween(char, 'a', 'z')
}

func IsIdentChar(char byte, i int) bool {
	if IsLetter(char) || char == '_' {
		return true
	}
	return i != 0 && IsNumber(char)
}

type Errors []error

func (e Errors) Add(err error) Errors {
	return append(e, err)
}

func (e Errors) AddPrefix(prefix string, err error) Errors {
	return append(e, fmt.Errorf("%v: %w", prefix, err))
}

func (e Errors) AddPostfix(err error, postfix string) Errors {
	return append(e, fmt.Errorf("%w: %v", err, postfix))
}

func (e Errors) Extend(es Errors) Errors {
	return append(e, es...)
}

func MapMerge[K comparable, V any](maps ...map[K]V) map[K]V {
	size := 0
	for _, m := range maps {
		size += len(m)
	}

	out := make(map[K]V, size)

	for _, m := range maps {
		for k, v := range m {
			out[k] = v
		}
	}

	return out
}

func (e Errors) Has(err_type interface{}) bool {
	for _, err := range e {
		if errors.As(err, err_type) { return true }
	}
	return false
}

func ErrorCount[K any](es Errors) int {
	count := 0
	for _, e := range es {
		_, ok := e.(K)
		if ok {
			count++
		}
	}
	return count
}

func AnyNil(vals ...any) bool {
	for _, v := range vals {
		if v == nil { return true }
	}
	return false
}
