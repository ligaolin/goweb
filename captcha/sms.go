package captcha

import (
	"errors"
	"fmt"
	"strings"
	"time"

	dysmsapi20170525 "github.com/alibabacloud-go/dysmsapi-20170525/v5/client"
	"github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/ligaolin/goweb"
	"github.com/ligaolin/goweb/cache"
	"github.com/ligaolin/goweb/sdk/ali"
)

type Sms struct {
	Client *cache.Client
	AliSms *ali.AliSms
}

func NewSms(c *cache.Client, a *ali.AliSms) *Sms {
	return &Sms{
		Client: c,
		AliSms: a,
	}
}

func (s *Sms) Generate(mobile any, expir time.Duration) (string, error) {
	value := Value{
		Code:    fmt.Sprintf("%d", goweb.Random(6)),
		Carrier: mobile.(string),
	}
	uuid, err := s.Client.Set("captcha-sms", value, expir)
	if err != nil {
		return "", err
	}

	err = s.SendSmsCode(mobile.(string), value.Code)
	if err != nil {
		return "", err
	}
	return uuid, nil
}

func (s *Sms) Verify(mobile any, uuid string, code string) error {
	var val Value
	if err := s.Client.Get(uuid, "captcha-sms", &val, false); err != nil {
		return errors.New("验证码不存在或过期")
	}
	if !strings.EqualFold(val.Code, code) {
		return errors.New("验证码错误")
	}
	if val.Carrier != mobile.(string) {
		return errors.New("不是接收验证码的手机号")
	}
	return nil
}

func (s Sms) SendSmsCode(mobile string, code string) error {
	if _, err := s.AliSms.Client.SendSmsWithOptions(&dysmsapi20170525.SendSmsRequest{
		PhoneNumbers:  tea.String(mobile),
		SignName:      tea.String(s.AliSms.Config.SignName),
		TemplateCode:  tea.String(s.AliSms.Config.TemplateCodeVerificationCode),
		TemplateParam: tea.String(fmt.Sprintf(`{"code":"%s"}`, code)),
	}, &service.RuntimeOptions{}); err != nil {
		return err
	}
	return nil
}
