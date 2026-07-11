package captcha

import (
	"fmt"
	"strings"
	"time"

	dysmsapi20170525 "github.com/alibabacloud-go/dysmsapi-20170525/v5/client"
	"github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
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

func (s *Sms) Generate(carrier string, expir time.Duration) (string, error) {
	code := generateCode()
	v := value{
		Code:    code,
		Carrier: carrier,
	}
	uuid, err := s.Client.Set("captcha-sms", v, expir)
	if err != nil {
		return "", fmt.Errorf("存储短信验证码失败: %w", err)
	}
	if err := s.sendSmsCode(carrier, code); err != nil {
		return "", fmt.Errorf("发送短信验证码失败: %w", err)
	}
	return uuid, nil
}

func (s *Sms) Verify(carrier string, uuid string, code string) error {
	var val value
	if err := s.Client.GetAndDelete(uuid, "captcha-sms", &val); err != nil {
		return fmt.Errorf("验证码不存在或已过期: %w", err)
	}
	if !strings.EqualFold(val.Code, code) {
		return fmt.Errorf("验证码错误")
	}
	if val.Carrier != carrier {
		return fmt.Errorf("不是接收验证码的手机号")
	}
	return nil
}

func (s *Sms) Delete(uuid string) error {
	return s.Client.Delete(uuid, "captcha-sms")
}

func (s *Sms) sendSmsCode(mobile string, code string) error {
	_, err := s.AliSms.Client.SendSmsWithOptions(&dysmsapi20170525.SendSmsRequest{
		PhoneNumbers:  tea.String(mobile),
		SignName:      tea.String(s.AliSms.Config.SignName),
		TemplateCode:  tea.String(s.AliSms.Config.TemplateCodeVerificationCode),
		TemplateParam: tea.String(fmt.Sprintf(`{"code":"%s"}`, code)),
	}, &service.RuntimeOptions{})
	return err
}
