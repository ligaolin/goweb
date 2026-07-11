package response

import (
	"encoding/json"
	"net/http"
)

// Response HTTP响应封装
type Response struct {
	Message string              `json:"message"`
	Code    int                 `json:"code"`
	Data    any                 `json:"data"`
	Writer  http.ResponseWriter `json:"-"`
	written bool
}

func NewResponse(w http.ResponseWriter) *Response {
	return &Response{
		Message: "success",
		Code:    200,
		Writer:  w,
	}
}

func (r *Response) SetData(data any) *Response {
	r.Data = data
	return r
}

func (r *Response) SetMessage(message string) *Response {
	r.Message = message
	return r
}

func (r *Response) SetCode(code int) *Response {
	r.Code = code
	return r
}

func (r *Response) Write() {
	if r.written {
		return
	}
	r.written = true
	r.Writer.Header().Set("Content-Type", "application/json")
	r.Writer.WriteHeader(http.StatusOK)
	json.NewEncoder(r.Writer).Encode(r)
}
