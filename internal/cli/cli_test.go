package cli

import (
	"reflect"
	"testing"
)

func TestGetOptions(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    Options
		wantErr bool
	}{
		{
			"no args",
			[]string{},
			Options{false, false, nil},
			false,
		},
		{
			"help with --help",
			[]string{"--help"},
			Options{true, false, nil},
			false,
		},
		{
			"help with -h",
			[]string{"-h"},
			Options{true, false, nil},
			false,
		},
		{
			"help with invalid flags",
			[]string{"--help", "-jkldwasgh", "--dhwalk"},
			Options{true, false, nil},
			true,
		},
		{
			"version with --version",
			[]string{"--version"},
			Options{false, true, nil},
			false,
		},
		{
			"version with -v",
			[]string{"-v"},
			Options{false, true, nil},
			false,
		},
		{
			"version with invalid flags",
			[]string{"--version", "--wacsfdfg", "-0fwafas"},
			Options{false, true, nil},
			true,
		},
		{
			"invalid file path",
			[]string{"some_invalid_name"},
			Options{false, false, new("some_invalid_name")},
			true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, gotErr := GetOptions(test.args)

			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("GetOptions() = %s, _; want %s.", got, test.want)
			}

			if gotErr != nil && !test.wantErr {
				t.Error("Expected no error, but something failed.")
			} else if gotErr == nil && test.wantErr {
				t.Error("Expected an error to occur, but it didn't.")
			}
		})
	}
}
