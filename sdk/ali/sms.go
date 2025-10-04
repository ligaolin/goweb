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
	AccessKeyId                  string
	AccessKeySecret              string
	TemplateCodeVerificationCode string
	SignName                     string
}

func NewAliSms(cfg *AliSmsConfig) (*AliSms, error) {
	client, err := dysmsapi20170525.NewClient(&openapi.Config{
		AccessKeyId:     tea.String(cfg.AccessKeyId),
		AccessKeySecret: tea.String(cfg.AccessKeySecret),
		Endpoint:        tea.String("dysmsapi.aliyuncs.com"),
	})
	return &AliSms{Client: client, Config: cfg}, err
}
