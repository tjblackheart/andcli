package spinner

import (
	"io"
	"time"
)

type Option func(*Spinner)

func WithFrames(frames []string) Option {
	return func(s *Spinner) {
		s.Frames = frames
	}
}

func WithFPS(fps time.Duration) Option {
	return func(s *Spinner) {
		s.FPS = fps
	}
}

func WithWriter(w io.Writer) Option {
	return func(s *Spinner) {
		s.writer = w
	}
}

func WithColor(c Color) Option {
	return func(s *Spinner) {
		s.Color = c
	}
}

func WithSuffix(val string) Option {
	return func(s *Spinner) {
		s.Suffix = val
	}
}
