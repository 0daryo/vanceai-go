package vanceai

import "errors"

var (
	ErrorIllegalParameter = errors.New("illegal parameter")
	VanceAIInternalError  = errors.New("vanceai internal error")
	FileNotFound          = errors.New("file not found")
	SizeExceedsLimit      = errors.New("size exceeds limit")
	JParamParseError      = errors.New("jparam parse error")
	JobFailed             = errors.New("job failed")
	InvalidAPIKey         = errors.New("invalid api key")
	InsufficientBalance   = errors.New("insufficient balance")
)

func (r *Response) Error() error {
	switch r.Code {
	case 10001:
		return ErrorIllegalParameter
	case 10010:
		return VanceAIInternalError
	case 10011:
		return FileNotFound
	case 10012:
		return SizeExceedsLimit
	case 10013:
		return JParamParseError
	case 10014:
		return JobFailed
	case 30001:
		return InvalidAPIKey
	case 30004:
		return InsufficientBalance
	default:
		return nil
	}
}
