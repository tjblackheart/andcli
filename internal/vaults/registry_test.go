package vaults_test

import (
	"context"
	"reflect"
	"strings"
	"testing"

	_ "github.com/tjblackheart/andcli/v2/internal/vaults/aegis"
	_ "github.com/tjblackheart/andcli/v2/internal/vaults/andotp"
	_ "github.com/tjblackheart/andcli/v2/internal/vaults/keepass"
	_ "github.com/tjblackheart/andcli/v2/internal/vaults/protonpass"
	_ "github.com/tjblackheart/andcli/v2/internal/vaults/stratum"
	_ "github.com/tjblackheart/andcli/v2/internal/vaults/twofas"

	"github.com/tjblackheart/andcli/v2/internal/vaults"
)

func TestRegister(t *testing.T) {
	for _, vt := range vaults.Types() {
		_, err := vaults.Open(context.Background(), ".", nil, vt)
		if err == nil {
			t.Fatalf("%s: expected an error, got none", vt)
		}

		if strings.Contains(err.Error(), "not implemented") {
			t.Fatalf("%s: missing registered openFn: %s", vt, err)
		}
	}
}

func TestTypes(t *testing.T) {
	tests := []struct {
		name string
		want []vaults.VaultType
	}{
		{
			"returns defined types",
			[]vaults.VaultType{
				vaults.AEGIS,
				vaults.ANDOTP,
				vaults.KEEPASS,
				vaults.PROTON,
				vaults.STRATUM,
				vaults.TWOFAS,
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
