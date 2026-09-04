package spinner

import (
	"fmt"
	"io"
	"os"
	"time"
)

type Spinner struct {
	Frames   []string
	FPS      time.Duration
	Color    Color
	Suffix   string
	stopChan chan struct{}
	writer   io.Writer
}

type Color string

func (c Color) String() string { return string(c) }

var (
	// taken from https://github.com/charmbracelet/bubbles/blob/v1.0.0/spinner/spinner.go
	Line    = Spinner{Frames: []string{"|", "/", "-", "\\"}, FPS: time.Second / 10}
	Dot     = Spinner{Frames: []string{"⣾ ", "⣽ ", "⣻ ", "⢿ ", "⡿ ", "⣟ ", "⣯ ", "⣷ "}, FPS: time.Second / 10}
	MiniDot = Spinner{Frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}, FPS: time.Second / 12}
	Jump    = Spinner{Frames: []string{"⢄", "⢂", "⢁", "⡁", "⡈", "⡐", "⡠"}, FPS: time.Second / 10}

	// taken from https://github.com/Homebrew/brew/blob/fb33e11256bb5dc850803cf9e53a7a011f7fe64f/Library/Homebrew/download_queue.rb#L425
	Brew = Spinner{Frames: []string{"⠙", "⠚", "⠞", "⠖", "⠦", "⠴", "⠲", "⠳", "⠓"}, FPS: time.Second / 12}

	Default = Line

	// colors
	Reset   Color = "\033[0m"
	Red     Color = "\033[31m"
	Green   Color = "\033[32m"
	Yellow  Color = "\033[33m"
	Blue    Color = "\033[34m"
	Magenta Color = "\033[35m"
	Cyan    Color = "\033[36m"
	Gray    Color = "\033[37m"
	White   Color = "\033[97m"
)

func New(opts ...Option) *Spinner {
	s := Default
	s.Color = White
	s.stopChan = make(chan struct{}, 1)
	s.writer = os.Stderr

	for _, opt := range opts {
		opt(&s)
	}

	return &s
}

func (s *Spinner) Start() {
	if s.stopChan == nil {
		s.stopChan = make(chan struct{}, 1)
	}

	if s.writer == nil {
		s.writer = os.Stderr
	}

	if s.Color == "" {
		s.Color = White
	}

	go func() {
		for {
			for _, frame := range s.Frames {
				fmt.Fprintf(s.writer, "%s%s%s %s", s.Color, frame, Reset, s.Suffix)
				time.Sleep(s.FPS)
				fmt.Fprintf(s.writer, "\r")

				select {
				case <-s.stopChan:
					return
				default:
				}
			}
		}
	}()
}

func (s *Spinner) ClearLine(appends ...rune) {
	fmt.Fprintf(s.writer, "\033[2K\r")
	for _, r := range appends {
		fmt.Fprintf(s.writer, "%c", r)
	}
}

func (s *Spinner) Stop() {
	s.stopChan <- struct{}{}
	s.ClearLine()
}

func (s *Spinner) SetColor(c Color) *Spinner {
	s.Color = c
	return s
}

func (s *Spinner) SetSuffix(val string) *Spinner {
	s.Suffix = val
	return s
}
