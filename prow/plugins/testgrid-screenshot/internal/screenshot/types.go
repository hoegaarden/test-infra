package screenshot

import (
	"io"
	"net/http"
)

const (
	defaultHeight = 2500
	defaultWidth  = 3000
)

//go:generate counterfeiter . Capture

type Capture func(site string, w io.Writer, opts Options) error

type Options struct {
	Width  int
	Height int

	Client *http.Client
}

func (o Options) Default() Options {
	defaulted := o

	if defaulted.Width == 0 {
		defaulted.Width = defaultWidth
	}

	if defaulted.Height == 0 {
		defaulted.Height = defaultHeight
	}

	if defaulted.Client == nil {
		defaulted.Client = &http.Client{}
	}

	return defaulted
}
