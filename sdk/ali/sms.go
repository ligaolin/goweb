package ali

import (
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dysmsapi20170525 "github.com/alibabacloud-go/dysmsapi-20170525/v5/client"
	"github.com/alibabacloud-go/tea/tea"
)

type AliSms struct {
	Client *dysmsapi20170525.Client
	Config *AliSmsConfig
}

type AliSmsConfig struct {
	AccessKeyId                  string `json:"access_key_id" toml:"access_key_id" yaml:"access_key_id"`
	AccessKeySecret              string `json:"access_key_secret" toml:"access_key_secret" yaml:"access_key_secret"`
	TemplateCodeVerificationCode string `json:"template_code_verification_code" toml:"template_code_verification_code" yaml:"template_code_verification_code"`
	SignName                     string `json:"sign_name" toml:"sign_name" yaml:"sign_name"`
}

func NewAliSms(cfg *AliSmsConfig) (*AliSms, error) {
	client, err := dysmsapi20170525.NewClient(&openapi.Config{
		AccessKeyId:     tea.String(cfg.AccessKeyId),
		AccessKeySecret: tea.String(cfg.AccessKeySecret),
		Endpoint:        tea.String("dysmsapi.aliyuncs.com"),
	})
	return &AliSms{Client: client, Config: cfg}, err
}
