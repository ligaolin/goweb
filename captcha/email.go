package captcha

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ligaolin/gin_lin/v2"
	"github.com/ligaolin/gin_lin/v2/cache"
	"github.com/ligaolin/gin_lin/v2/email"
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

func (c *Email) Generate(email any, expir time.Duration) (string, error) {
	value := Value{
		Code:    fmt.Sprintf("%d", gin_lin.Random(6)),
		Carrier: email.(string),
	}
	uuid, err := c.Client.Set("captcha-email", value, expir)
	if err != nil {
		return "", err
	}

	if err = c.SendEmailCode(email.(string), value.Code, "邮件验证码", expir); err != nil {
		return "", err
	}
	return uuid, nil
}

func (e *Email) Verify(email any, uuid string, code string) error {
	var val Value
	if err := e.Client.Get(uuid, "captcha-email", &val, false); err != nil {
		return errors.New("验证码不存在或过期")
	}
	if !strings.EqualFold(val.Code, code) {
		return errors.New("验证码错误")
	}
	if val.Carrier != email.(string) {
		return errors.New("不是接收验证码的邮箱")
	}
	return nil
}

func (e *Email) SendEmailCode(to string, code string, subject string, expir time.Duration) error {
	return e.Email.Send([]string{to}, subject, fmt.Sprintf(`尊敬的用户：

	您好！
	您正在进行邮箱验证操作，验证码为：%s。
	此验证码有效期为 %.f分钟，请尽快完成验证。
	
	如非本人操作，请忽略此邮件。
	
	感谢您的支持！
	
	【系统邮件，请勿直接回复】`, code, expir.Minutes()))
}
