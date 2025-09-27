package captcha

import (
	"time"
)

type InputCaptcha interface {
	Generate(carrier string, expir time.Duration) (string, error)
	Verify(carrier string, uuid, code string) error
}

type OutputCaptcha interface {
	Generate(expir time.Duration) (string, string, error)
	Verify(uuid, code string) error
}

type CaptchaConfig struct {
	Expir int64 `json:"expir" toml:"expir" yaml:"expir"` // 过期时间
}

type Value struct {
	Code    string
	Carrier string
}
