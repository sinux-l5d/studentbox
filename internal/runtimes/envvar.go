package runtimes

import (
	"errors"
	"strconv"
	"strings"

	"github.com/sinux-l5d/studentbox/internal/tools"
)

type EnvModifierParams struct {
	Name   string
	Params []string
}

type EnvVar struct {
	Name         string
	DefaultValue string
	Modifiers    []EnvModifierParams
}

type EnvModifier interface {
	Modify(previousValue string, args ...string) (string, error)
}

type ErrModifierParams struct {
	Name          string
	PreviousValue string
	Args          []string
}

func (e *ErrModifierParams) Error() string {
	return "modifier \"" + e.Name + "\" failed with previous value \"" + e.PreviousValue + "\" and args: \"" + strings.Join(e.Args, ", ") + "\""
}

// EnvModifiers is a map of all available modifiers
var EnvModifiers = map[string]EnvModifier{
	"password":  &PasswordModifier{},
	"failempty": &FailEmptyModifier{},
}

// Apply modifiers to env var with default value
func (e EnvVar) ApplyModifiers() (string, error) {
	return e.ApplyModifiersWithInput(nil)
}

// Apply modifiers to env var with input value
func (e EnvVar) ApplyModifiersWithInput(inputValue *string) (string, error) {
	var err error
	value := e.DefaultValue
	if inputValue != nil {
		value = *inputValue
	}
	for _, modifier := range e.Modifiers {
		value, err = EnvModifiers[modifier.Name].Modify(value, modifier.Params...)
		if err != nil {
			return "", err
		}
	}
	return value, nil
}

// Set value to random password if previous value is empty
type PasswordModifier struct{}

func (p *PasswordModifier) Modify(previousValue string, args ...string) (string, error) {
	paramErr := &ErrModifierParams{
		Name:          "password",
		PreviousValue: previousValue,
		Args:          args,
	}

	if len(args) == 0 {
		return "", paramErr
	}

	// If previous value is not empty, return it
	if previousValue != "" {
		return previousValue, nil
	}

	length, err := strconv.Atoi(args[0])
	if err != nil {
		return "", paramErr
	}

	return tools.GenerateRandomString(length)
}

type FailEmptyModifier struct{}

func (f *FailEmptyModifier) Modify(previousValue string, args ...string) (string, error) {
	if previousValue == "" {
		return "", errors.New("value is empty")
	}
	return previousValue, nil
}
