package vanceai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
)

type Client struct {
	APIKey     string
	httpClient *http.Client
	BaseURL    *url.URL
}

func NewClient(apiKey string) (*Client, error) {
	baseURL, err := url.Parse("https://api-service.vanceai.com/web_api/v1")
	if err != nil {
		return nil, err
	}
	return &Client{
		APIKey:     apiKey,
		httpClient: http.DefaultClient,
		BaseURL:    baseURL,
	}, nil
}

type Response struct {
	Code   int64  `json:"code"`
	CSCode int64  `json:"cscode"`
	IP     string `json:"ip"`
	Data   Data   `json:"data"`
}

type Data struct {
	UID       string `json:"uid"`
	Name      string `json:"name"`
	Thumbnail string `json:"thumbnail"`
	W         int64  `json:"w"`
	H         int64  `json:"h"`
	FileSize  int64  `json:"filesize"`
}

func (cli *Client) UploadImage(
	ctx context.Context,
	reader io.Reader,
	name string, //file name
) (Response, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", name)
	if err != nil {
		return Response{}, fmt.Errorf("failed to create form file: %w", err)
	}
	_, err = io.Copy(part, reader)
	if err != nil {
		return Response{}, fmt.Errorf("failed to copy file: %w", err)
	}
	_ = writer.WriteField("api_token", cli.APIKey)
	err = writer.Close()
	if err != nil {
		return Response{}, fmt.Errorf("failed to close writer: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cli.BaseURL.String()+"/upload", body)
	if err != nil {
		return Response{}, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resStruct := Response{}
	resp, err := cli.doAndUnmarshal(ctx, req, &resStruct)
	if err != nil {
		return Response{}, fmt.Errorf("failed to do request: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return Response{}, fmt.Errorf("failed to upload image: %s", resp.Status)
	}
	return resStruct, nil
}

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
