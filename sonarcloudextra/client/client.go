package client

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	ApplicationJson   = "application/json"
	Bearer            = "Bearer"
	SonarCloudBaseUrl = "https://XXX.YYY.com/api/v1"
)

type Client struct {
	numRetries int
	retryDelay int
	httpClient *http.Client
}

func NewClient(numRetries int, retryDelay int) (*Client, error) {
	c := &Client{
		numRetries: numRetries,
		retryDelay: retryDelay,
		httpClient: &http.Client{},
	}
	return c, nil
}

func (c *Client) HttpRequest(ctx context.Context, method string, path string, query url.Values, headerMap http.Header, body *bytes.Buffer) (*bytes.Buffer, error) {
	req, err := http.NewRequest(method, c.RequestPath(path), body)
	if err != nil {
		return nil, &RequestError{StatusCode: http.StatusInternalServerError, Err: err}
	}
	//Handle query values
	if query != nil {
		requestQuery := req.URL.Query()
		for key, values := range query {
			for _, value := range values {
				requestQuery.Add(key, value)
			}
		}
		req.URL.RawQuery = requestQuery.Encode()
	}
	//Handle header values
	if headerMap != nil {
		for key, values := range headerMap {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
	}
	requestDump, err := httputil.DumpRequest(req, true)
	if err != nil {
		tflog.Info(ctx, "SonarCloud API:", map[string]any{"error": err})
	} else {
		tflog.Info(ctx, "SonarCloud API: ", map[string]any{"request": string(requestDump)})
	}
	try := 0
	var resp *http.Response
	for {
		resp, err = c.httpClient.Do(req)
		if err != nil {
			return nil, &RequestError{StatusCode: http.StatusInternalServerError, Err: err}
		}
		if (resp.StatusCode == http.StatusTooManyRequests) || (resp.StatusCode >= http.StatusInternalServerError) {
			try++
			if try >= c.numRetries {
				break
			}
			time.Sleep(time.Duration(c.retryDelay) * time.Second)
			continue
		}
		break
	}
	defer resp.Body.Close()
	respBody := new(bytes.Buffer)
	_, err = respBody.ReadFrom(resp.Body)
	if err != nil {
		return nil, &RequestError{StatusCode: resp.StatusCode, Err: err}
	}
	if (resp.StatusCode < http.StatusOK) || (resp.StatusCode >= http.StatusMultipleChoices) {
		return nil, &RequestError{StatusCode: resp.StatusCode, Err: fmt.Errorf("%s", respBody.String())}
	}
	return respBody, nil
}

func (c *Client) RequestPath(path string) string {
	return fmt.Sprintf("%s/%s", SonarCloudBaseUrl, path)
}
