package vanceai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (cli *Client) doAndUnmarshal(ctx context.Context, req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := cli.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if v != nil {
		if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return resp, nil
}
