package vanceai

import "fmt"

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
	Status    Status `json:"status"`
}

type JobConfig struct {
	// https://docs.vanceai.com/?shell#description-of-config-file
	Job    string `json:"job"`
	Config Config `json:"config"`
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
	Rescale       int64
}
type ProcessRequest struct {
	APIToken  string `json:"api_token"`
	UID       string `json:"uid"`
	Webhook   string `json:"webhook"`
	JobConfig string `json:"jconfig"`
}

type ProgressRequest struct {
	APIToken string `json:"api_token"`
	TransID  string `json:"trans_id"`
}

type Status string

const (
	// https://docs.vanceai.com/?shell#status-of-job-processing
	Finish  Status = "finish"
	Wait    Status = "wait"
	Fatal   Status = "fatal"
	Process Status = "process"
	Webhook Status = "webhook"
	Busy    Status = "busy"
)
