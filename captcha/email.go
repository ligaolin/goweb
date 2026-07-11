package captcha

import (
	"fmt"
	"strings"
	"time"

	"github.com/ligaolin/goweb/cache"
	"github.com/ligaolin/goweb/email"
)

type Email struct {
	Client *cache.Client
	Email  *email.Email
}

func NewEmail(c *cache.Client, e *email.Email) *Email {
	return &Email{
		Client: c,
		Email:  e,
	}
}

func (e *Email) Generate(carrier string, expir time.Duration) (string, error) {
	code := generateCode()
	v := value{
		Code:    code,
		Carrier: carrier,
	}
	uuid, err := e.Client.Set("captcha-email", v, expir)
	if err != nil {
		return "", fmt.Errorf("存储邮箱验证码失败: %w", err)
	}
	if err := e.sendEmailCode(carrier, code, expir); err != nil {
		return "", fmt.Errorf("发送邮箱验证码失败: %w", err)
	}
	return uuid, nil
}

func (e *Email) Verify(carrier string, uuid string, code string) error {
	var val value
	if err := e.Client.GetAndDelete(uuid, "captcha-email", &val); err != nil {
		return fmt.Errorf("验证码不存在或已过期: %w", err)
	}
	if !strings.EqualFold(val.Code, code) {
		return fmt.Errorf("验证码错误")
	}
	if val.Carrier != carrier {
		return fmt.Errorf("不是接收验证码的邮箱")
	}
	return nil
}

func (e *Email) Delete(uuid string) error {
	return e.Client.Delete(uuid, "captcha-email")
}

func (e *Email) sendEmailCode(to string, code string, expir time.Duration) error {
	subject := "邮箱验证码"
	body := fmt.Sprintf("您好！\n\n您的验证码为：%s。\n此验证码有效期为 %.f 分钟，请尽快完成验证。\n\n如非本人操作，请忽略此邮件。", code, expir.Minutes())
	return e.Email.Send([]string{to}, subject, body)
}
