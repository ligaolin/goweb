package gin_lin

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Message string              `json:"message"`
	Code    int                 `json:"code"`
	Data    any                 `json:"data"`
	Writer  http.ResponseWriter `json:"-"`
}

func NewResponse(w http.ResponseWriter) *Response {
	return &Response{
		Message: "success",
		Code:    200,
		Data:    nil,
		Writer:  w,
	}
}

func (r *Response) Response() {
	r.Writer.Header().Set("Content-Type", "application/json")
	r.Writer.WriteHeader(http.StatusOK)
	jsonStr, _ := json.Marshal(r)
	r.Writer.Write(jsonStr)
}

func (r *Response) SetData(data any) {
	r.Data = data
	r.Response()
}

func (r *Response) SetMessage(message string) {
	r.Message = message
	r.Response()
}

func (r *Response) SetCode(code int) {
	r.Code = code
	r.Response()
}

func (r *Response) Success(message string, data any) {
	r.Message = message
	r.Data = data
	r.Response()
}

func (r *Response) Error(err error) {
	r.Message = err.Error()
	r.Code = 400
	r.Response()
}

func (r *Response) ErrorMessage(message string) {
	r.Message = message
	r.Code = 400
	r.Response()
}

func (r *Response) LoginError(err error) {
	r.Message = err.Error()
	r.Code = 401
	r.Response()
}

func (r *Response) AuthError(err error) {
	r.Message = err.Error()
	r.Code = 402
	r.Response()
}
