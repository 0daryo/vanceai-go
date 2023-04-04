package vanceai

import (
	"context"
	"fmt"
	"net/http"
)

func (cli *Client) GetPoint(ctx context.Context) (Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cli.BaseURL.String()+"/point", nil)
	if err != nil {
		return Response{}, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	q := req.URL.Query()
	q.Add("api_token", cli.APIKey)
	req.URL.RawQuery = q.Encode()
	resStruct := Response{}
	resp, err := cli.doAndUnmarshal(ctx, req, &resStruct)
	if err != nil {
		return Response{}, fmt.Errorf("failed to do request: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return Response{}, fmt.Errorf("failed to get point: %s", resp.Status)
	}
	if err := resStruct.Error(); err != nil {
		return Response{}, fmt.Errorf("failed to get point: %w", err)
	}
	return resStruct, nil
}
