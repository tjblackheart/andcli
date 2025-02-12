package input

import (
	"reflect"
	"testing"
)

func TestAskHidden(t *testing.T) {
	tests := []struct {
		name    string
		want    []byte
		wantErr bool
	}{
		{
			"returns correct input",
			[]byte("test"),
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AskHidden("?", tt.want)
			if (err != nil) != tt.wantErr {
				t.Errorf("AskHidden() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AskHidden() = %v, want %v", got, tt.want)
			}
		})
	}
}
