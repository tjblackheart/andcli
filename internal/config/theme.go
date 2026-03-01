package config

import (
	"log"
	"regexp"
	"strconv"
	"strings"
)

type Theme struct {
	Base   string `yaml:"base"`
	Green  string `yaml:"green"`
	Yellow string `yaml:"yellow"`
	Red    string `yaml:"red"`
	Grey   string `yaml:"grey"`
	Black  string `yaml:"black"`
	White  string `yaml:"white"`
}

var DefaultTheme = Theme{
	Base:   "#39A02E",
	Green:  "#39A02E",
	Yellow: "#DB9F1F",
	Red:    "#f10000",
	Grey:   "#424242",
	Black:  "#000000",
	White:  "#FFFFFF",
}

var hexRx = regexp.MustCompile(`^#?([a-f0-9]{6}|[a-f0-9]{3})$`)

func (t *Theme) validate() {
	invalid := []string{}

	if _, err := strconv.ParseUint(t.Base, 16, 32); err != nil {
		t.Base = DefaultTheme.Base
		invalid = append(invalid, "base")
	}

	if !hexRx.MatchString(t.Green) {
		t.Green = DefaultTheme.Green
		invalid = append(invalid, "green")
	}

	if !hexRx.MatchString(t.Yellow) {
		t.Yellow = DefaultTheme.Yellow
		invalid = append(invalid, "yellow")
	}

	if !hexRx.MatchString(t.Red) {
		t.Red = DefaultTheme.Red
		invalid = append(invalid, "red")
	}

	if !hexRx.MatchString(t.Grey) {
		t.Grey = DefaultTheme.Grey
		invalid = append(invalid, "grey")
	}

	if !hexRx.MatchString(t.Black) {
		t.Black = DefaultTheme.Black
		invalid = append(invalid, "black")
	}

	if !hexRx.MatchString(t.White) {
		t.White = DefaultTheme.White
		invalid = append(invalid, "white")
	}

	if len(invalid) > 0 {
		msg := "Theme errors where found and substituted by fallbacks"
		log.Printf("andcli: config: %s. Invalid fields: %s\n", msg, strings.Join(invalid, ", "))
	}
}
