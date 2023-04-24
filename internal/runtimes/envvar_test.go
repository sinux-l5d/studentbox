package runtimes_test

import (
	"fmt"
	"testing"

	"github.com/sinux-l5d/studentbox/internal/runtimes"
)

func TestEnvVarApplyModifiers(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		envVar      runtimes.EnvVar
		expected    string
		expectError bool
	}{
		{
			name:  "No modifiers",
			input: "",
			envVar: runtimes.EnvVar{
				Name:         "MARIADB_USER",
				DefaultValue: "student",
				Modifiers:    []runtimes.EnvModifierParams{},
			},
			expected:    "student",
			expectError: false,
		},
		{
			name:  "Fail empty w default wo input",
			input: "",
			envVar: runtimes.EnvVar{
				Name:         "MARIADB_DATABASE",
				DefaultValue: "app",
				Modifiers: []runtimes.EnvModifierParams{
					{
						Name:   "failempty",
						Params: []string{},
					},
				},
			},
			expected:    "app",
			expectError: false,
		},
		{
			name:  "Fail empty wo default wo input",
			input: "",
			envVar: runtimes.EnvVar{
				Name:         "MARIADB_DATABASE",
				DefaultValue: "",
				Modifiers: []runtimes.EnvModifierParams{
					{
						Name:   "failempty",
						Params: []string{},
					},
				},
			},
			expected:    "",
			expectError: true,
		},
		{
			name:  "Fail empty wo default w input",
			input: "test",
			envVar: runtimes.EnvVar{
				Name:         "MARIADB_DATABASE",
				DefaultValue: "",
				Modifiers: []runtimes.EnvModifierParams{
					{
						Name:   "failempty",
						Params: []string{},
					},
				},
			},
			expected:    "test",
			expectError: false,
		},
		{
			name:  "Password modifier no input",
			input: "",
			envVar: runtimes.EnvVar{
				Name:         "MARIADB_PASSWORD",
				DefaultValue: "",
				Modifiers: []runtimes.EnvModifierParams{
					{
						Name:   "password",
						Params: []string{"30"},
					},
				},
			},
			// We cannot predict the generated password, so we leave it empty.
			expected:    "",
			expectError: false,
		},
		{
			name:  "Password modifier with input",
			input: "definedpassword",
			envVar: runtimes.EnvVar{
				Name:         "MARIADB_PASSWORD",
				DefaultValue: "",
				Modifiers: []runtimes.EnvModifierParams{
					{
						Name:   "password",
						Params: []string{"30"},
					},
				},
			},
			expected:    "definedpassword",
			expectError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := &test.input
			// empty string is considered as nil for tests
			if test.input == "" {
				input = nil
			}
			result, err := test.envVar.ApplyModifiersWithInput(input)
			if test.expectError && err == nil {
				t.Errorf("Expected an error, but didn't get one")
			}
			if !test.expectError && err != nil {
				t.Errorf("Didn't expect an error, but got one: %v", err)
			}
			if test.expected != "" && result != test.expected {
				t.Errorf("Expected %s, got %s", test.expected, result)
			}
		})
	}
}

func TestModifierPassword(t *testing.T) {
	lengths := []int{0, 30}
	for _, l := range lengths {
		lS := fmt.Sprint(l)
		t.Run("No input "+lS+" chars", func(t *testing.T) {
			input := ""
			result, err := runtimes.EnvModifiers["password"].Modify(input, lS)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if len(result) != l {
				t.Errorf("Expected %d characters, got %d", l, len(result))
			}
		})
		t.Run("With input "+lS+" chars", func(t *testing.T) {
			input := "definedpassword"
			result, err := runtimes.EnvModifiers["password"].Modify(input, lS)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if result != input {
				t.Errorf("Expected %s, got %s", input, result)
			}
		})
	}
}
