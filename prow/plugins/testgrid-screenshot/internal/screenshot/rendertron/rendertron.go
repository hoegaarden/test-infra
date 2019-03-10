package rendertron

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"

	"k8s.io/test-infra/prow/plugins/testgrid-screenshot/internal/screenshot"
)

const (
	rendertronURLTmpl = "https://rendertron.appspot.com/screenshot/%s?%s"
)

func Capture(site string, w io.Writer, opts screenshot.Options) error {
	opts = opts.Default()
	renderURL := getURL(site, opts)

	client := opts.Client

	res, err := client.Get(renderURL)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "URl: %s\n", renderURL)
	_ = res

	return nil
}

func getURL(site string, opts screenshot.Options) string {
	query := make(url.Values)
	query.Set("width", strconv.Itoa(opts.Width))
	query.Set("height", strconv.Itoa(opts.Height))

	return fmt.Sprintf(
		rendertronURLTmpl,
		url.QueryEscape(site),
		query.Encode(),
	)
}
