package main

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"reflect"
	"testing"

	"github.com/xlzd/gotp"
)

func TestEntryGenerateHasher(t *testing.T) {
	tests := []struct {
		name  string
		entry *entry
		want  *gotp.Hasher
	}{
		{
			"generates default hasher",
			&entry{},
			&gotp.Hasher{HashName: "sha1", Digest: sha1.New},
		},
		{
			"generates sha256 hasher",
			&entry{Algorithm: "SHA256"},
			&gotp.Hasher{HashName: "sha256", Digest: sha256.New},
		},
		{
			"generates sha512 hasher",
			&entry{Algorithm: "SHA512"},
			&gotp.Hasher{HashName: "sha512", Digest: sha512.New},
		},
		{
			"ignores cases",
			&entry{Algorithm: "sha1"},
			&gotp.Hasher{HashName: "sha1", Digest: sha1.New},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.entry.generateHasher()
			if !reflect.DeepEqual(tt.want.HashName, h.HashName) {
				t.Errorf("entry.generateHasher() = %v, want %v", h, tt.want)
			}

			fn1 := reflect.Indirect(reflect.ValueOf(tt.want.Digest))
			fn2 := reflect.Indirect(reflect.ValueOf(h.Digest))
			if !reflect.DeepEqual(fn1, fn2) {
				t.Errorf("entry.generateHasher() = %v, want %v", h, tt.want)
			}
		})
	}
}

func TestEntriesFilter(t *testing.T) {
	list := []entry{
		{Choice: "aaa"},
		{Choice: "aab"},
		{Choice: "ccc"},
		{Choice: "ddd"},
	}

	tests := []struct {
		name string
		e    entries
		args string
		want entries
	}{
		{
			"finds one",
			list,
			"aaa",
			[]entry{{Choice: "aaa"}},
		},
		{
			"finds multiple",
			list,
			"aa",
			[]entry{{Choice: "aaa"}, {Choice: "aab"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.filter(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("entries.filter() = %v, want %v", got, tt.want)
			}
		})
	}
}
