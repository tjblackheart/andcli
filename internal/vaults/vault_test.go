package vaults_test

import (
	"reflect"
	"testing"

	"github.com/tjblackheart/andcli/v2/internal/vaults"
)

func TestTypes(t *testing.T) {
	tests := []struct {
		name string
		want []vaults.Type
	}{
		{
			"returns defined types",
			[]vaults.Type{
				vaults.ANDOTP,
				vaults.AEGIS,
				vaults.TWOFAS,
				vaults.STRATUM,
				vaults.KEEPASS,
				vaults.PROTON,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := vaults.Types(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Types() = %v, want %v", got, tt.want)
			}
		})
	}
}
