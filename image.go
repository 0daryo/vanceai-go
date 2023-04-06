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
	APIKey         string
	httpClient     *http.Client
	BaseURL        *url.URL
	ProcessWebhook string
}

func NewClient(apiKey, processWebhook string) (*Client, error) {
	baseURL, err := url.Parse("https://api-service.vanceai.com/web_api/v1")
	if err != nil {
		return nil, err
	}
	return &Client{
		APIKey:         apiKey,
		httpClient:     http.DefaultClient,
		BaseURL:        baseURL,
		ProcessWebhook: processWebhook,
	}, nil
}

type job interface {
	jsonString() (string, error)
}

func (jc *JobConfig) jsonString() (string, error) {
	b, err := json.Marshal(jc)
	if err != nil {
		return "", fmt.Errorf("failed to marshal job config: %w", err)
	}
	return string(b), nil
}

func (jc *MultipleJob) jsonString() (string, error) {
	b, err := json.Marshal(jc)
	if err != nil {
		return "", fmt.Errorf("failed to marshal job config: %w", err)
	}
	return string(b), nil
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
	if err := writer.WriteField("api_token", cli.APIKey); err != nil {
		return Response{}, fmt.Errorf("failed to write field api_token: %w", err)
	}
	if err := writer.Close(); err != nil {
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
	if err := resStruct.Error(); err != nil {
		return Response{}, fmt.Errorf("failed to upload image %s: %w", resStruct.MsgString(), err)
	}
	return resStruct, nil
}

func (cli *Client) ProcessImage(
	ctx context.Context,
	uid string,
	jobConfig job,
) (Response, error) {
	jc, err := jobConfig.jsonString()
	if err != nil {
		return Response{}, fmt.Errorf("failed to marshal job config: %w", err)
	}
	jsonBytes, err := json.Marshal(ProcessRequest{
		APIToken:  cli.APIKey,
		UID:       uid,
		Webhook:   cli.ProcessWebhook,
		JobConfig: jc,
	})
	if err != nil {
		return Response{}, fmt.Errorf("failed to encode json: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cli.BaseURL.String()+"/transform", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return Response{}, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resStruct := Response{}
	resp, err := cli.doAndUnmarshal(ctx, req, &resStruct)
	if err != nil {
		return Response{}, fmt.Errorf("failed to do request: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return Response{}, fmt.Errorf("failed to process image: %s", resp.Status)
	}
	if err := resStruct.Error(); err != nil {
		return Response{}, fmt.Errorf("failed to upload image %s: %w", resStruct.MsgString(), err)
	}
	return resStruct, nil
}

func (cli *Client) GetProgress(
	ctx context.Context,
	transID string,
) (Response, error) {

	jsonBytes, err := json.Marshal(ProgressRequest{
		APIToken: cli.APIKey,
		TransID:  transID,
	})
	if err != nil {
		return Response{}, fmt.Errorf("failed to encode json: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cli.BaseURL.String()+"/progress", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return Response{}, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resStruct := Response{}
	resp, err := cli.doAndUnmarshal(ctx, req, &resStruct)
	if err != nil {
		return Response{}, fmt.Errorf("failed to do request: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return Response{}, fmt.Errorf("failed to get progress status is not ok: %s", resp.Status)
	}
	if err := resStruct.Error(); err != nil {
		return Response{}, fmt.Errorf("failed to get progress error code returned%s: %w", resStruct.MsgString(), err)
	}
	return resStruct, nil
}

func (cli *Client) Download(ctx context.Context, transID string) (io.ReadCloser, error) {
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	if err := writer.WriteField("api_token", cli.APIKey); err != nil {
		return nil, fmt.Errorf("failed to write field api_token: %w", err)
	}
	if err := writer.WriteField("trans_id", transID); err != nil {
		return nil, fmt.Errorf("failed to write field trans_id: %w", err)
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cli.BaseURL.String()+"/download", payload)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := cli.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to do request: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download image: %s", resp.Status)
	}
	return resp.Body, nil
}
