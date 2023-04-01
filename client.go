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

type Response struct {
	Code   int64       `json:"code"`
	CSCode int64       `json:"cscode"`
	IP     string      `json:"ip"`
	Data   Data        `json:"data"`
	Msg    interface{} `json:"msg"`
}

func (resp *Response) MsgString() string {
	if resp.Msg == nil {
		return ""
	}
	msg, ok := resp.Msg.(string)
	if ok {
		return msg
	}
	return fmt.Sprintf("%v", resp.Msg)
}

type Data struct {
	UID       string `json:"uid"`
	Name      string `json:"name"`
	Thumbnail string `json:"thumbnail"`
	W         int64  `json:"w"`
	H         int64  `json:"h"`
	FileSize  int64  `json:"filesize"`
	TransID   string `json:"trans_id"`
	Status    string `json:"status"`
}

type JobConfig struct {
	Job    string `json:"job"`
	Config Config `json:"config"`
}

func (jc *JobConfig) jsonString() (string, error) {
	b, err := json.Marshal(jc)
	if err != nil {
		return "", fmt.Errorf("failed to marshal job config: %w", err)
	}
	return string(b), nil
}

type Config struct {
	Module       string       `json:"module"`
	ModuleParams ModuleParams `json:"module_params"`
	OutParams    OutParams    `json:"out_params"`
}
type OutParams struct{}

type ModuleParams struct {
	ModelName     string `json:"model_name"`
	SuppressNoise int64  `json:"suppress_noise"`
	RemoveBlur    int64  `json:"remove_blur"`
	Scale         string `json:"scale"`
}
type ProcessRequest struct {
	APIToken  string `json:"api_token"`
	UID       string `json:"uid"`
	Webhook   string `json:"webhook"`
	JobConfig string `json:"jconfig"`
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
	if err := resStruct.Error(); err != nil {
		return Response{}, fmt.Errorf("failed to upload image %s: %w", resStruct.MsgString(), err)
	}
	return resStruct, nil
}

func (cli *Client) ProcessImage(
	ctx context.Context,
	uid string,
	jobConfig JobConfig,
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

type ProgressRequest struct {
	APIToken string `json:"api_token"`
	TransID  string `json:"trans_id"`
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
