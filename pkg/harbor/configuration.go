package harbor

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

// ApplyConfiguration applies configuration to harbor instance.
func (c *client) ApplyConfiguration(ctx context.Context, config []byte) error {
	u := fmt.Sprintf("%s/api/v2.0/configurations", strings.TrimSuffix(c.url, "/"))

	req, err := http.NewRequestWithContext(ctx, "PUT", u, bytes.NewReader(config))
	if err != nil {
		return fmt.Errorf("new request error: %w", err)
	}
	// with header
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	// with auth
	if c.opts.credential != nil {
		req.SetBasicAuth(c.opts.credential.username, c.opts.credential.password)
	}
	// request harbor api
	resp, err := c.opts.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("client request error: %w", err)
	}

	defer resp.Body.Close()

	code := resp.StatusCode
	if code != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("read http response body error: %w", err)
		}

		return errors.Errorf("response status code is %d, body is %s", code, string(body))
	}

	return nil
}
