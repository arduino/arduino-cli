package configs_test

import (
	"reflect"
	"testing"
)

func TestNavigate(t *testing.T) {
	type args struct {
		root string
		pwd  string
	}
	tests := []struct {
		name string
		args args
		want Configuration
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Navigate(tt.args.root, tt.args.pwd); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Navigate() = %v, want %v", got, tt.want)
			}
		})
	}
}
