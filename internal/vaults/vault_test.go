package vaults

import (
	"reflect"
	"testing"
)

func TestTypes(t *testing.T) {
	tests := []struct {
		name string
		want []string
	}{
		{
			"returns defined types",
			[]string{TYPE_ANDOTP, TYPE_AEGIS, TYPE_TWOFAS, TYPE_STRATUM},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Types(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Types() = %v, want %v", got, tt.want)
			}
		})
	}
}
