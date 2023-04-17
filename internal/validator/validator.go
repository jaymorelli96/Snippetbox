package validator

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

type Validator struct {
	NonFieldErrors []string
	FieldErrors    map[string]string
}

var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// Valid returns true if FieldErrors is empty
func (v *Validator) Valid() bool {
	return len(v.FieldErrors) == 0 && len(v.NonFieldErrors) == 0
}

// AddFieldError add a new error message entry into the FieldErrors map
func (v *Validator) AddFieldError(key, message string) {
	if v.FieldErrors == nil {
		v.FieldErrors = make(map[string]string)
	}

	if _, exists := v.FieldErrors[key]; !exists {
		v.FieldErrors[key] = message
	}
}

// AddNonFieldError add non field specifc errors into the NonFieldErrors slice
func (v *Validator) AddNonFieldError(message string) {
	v.NonFieldErrors = append(v.NonFieldErrors, message)
}

// CheckField adds an error message to the FieldErrors map only if a
// validation check is not 'ok'.
func (v *Validator) CheckField(ok bool, key, message string) {
	if !ok {
		v.AddFieldError(key, message)
	}
}

// NotBlank returns true if a value is not an empty string.
func NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

// MaxChars returns true if a value contains no more than n characters.
func MaxChars(value string, n int) bool {
	return utf8.RuneCountInString(value) <= n
}

// MinChars returns true if a value contains at least n characters.
func MinChars(value string, n int) bool {
	return utf8.RuneCountInString(value) >= n
}

// PermittedInt returns true if a value is in a list of permitted integers.
func PermittedValue[T comparable](value T, permittedValues ...T) bool {
	for i := range permittedValues {
		if value == permittedValues[i] {
			return true
		}
	}
	return false
}

// Matches returns true if a value matches a provided compiled regular
// expression pattern.
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

func Equals[T comparable](got, expected T) bool {
	return got == expected
}
