package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeader(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"underlines word", "test", "test\n====\n0 entries.\n\nType to search: \n\n"},
		{"underlines words", " test test ", " test test \n===========\n0 entries.\n\nType to search: \n\n"},
		{"ignores empty", "", "\n"},
	}

	for _, tt := range tests {
		have := model{}.header(tt.input)
		assert.Equal(t, tt.want, have)
	}
}

func TestFooter(t *testing.T) {
	tests := []struct {
		name    string
		copyCmd string
		view    string
		want    string
	}{
		{
			"generates list footer",
			"",
			VIEW_LIST,
			"\n[esc] quit\n",
		},
		{
			"generates detail footer",
			"",
			VIEW_DETAIL,
			"\n[esc] back | [q] quit | [enter] toggle visibility\n",
		},
		{
			"generates detail footer with copy",
			"xcopy",
			VIEW_DETAIL,
			"\n[esc] back | [q] quit | [enter] toggle visibility | [c] copy\n",
		},
	}

	for _, tt := range tests {
		copyCmd = tt.copyCmd
		have := model{view: tt.view}.footer()
		assert.Equal(t, tt.want, have)
	}
}
