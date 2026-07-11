package captcha

import (
	"fmt"
	"strings"
	"time"

	"github.com/ligaolin/goweb/cache"
	"github.com/mojocn/base64Captcha"
)

type ImageConfig struct {
	Width      int `json:"width"`
	Height     int `json:"height"`
	Length     int `json:"length"`
	NoiseCount int `json:"noise_count"`
}

type Image struct {
	Client *cache.Client
	Config *ImageConfig
}

func NewImage(c *cache.Client, config *ImageConfig) *Image {
	if config == nil {
		config = &ImageConfig{
			Width:      240,
			Height:     80,
			Length:     4,
			NoiseCount: 3,
		}
	}
	if config.Width <= 0 {
		config.Width = 240
	}
	if config.Height <= 0 {
		config.Height = 80
	}
	if config.Length <= 0 {
		config.Length = 4
	}
	if config.NoiseCount < 0 {
		config.NoiseCount = 3
	}
	return &Image{
		Client: c,
		Config: config,
	}
}

func (i *Image) Generate(expir time.Duration) (string, string, error) {
	driver := base64Captcha.NewDriverString(
		i.Config.Height,
		i.Config.Width,
		i.Config.NoiseCount,
		base64Captcha.OptionShowHollowLine,
		i.Config.Length,
		base64Captcha.TxtNumbers+base64Captcha.TxtAlphabet,
		nil,
		nil,
		[]string{},
	)

	_, content, answer := driver.GenerateIdQuestionAnswer()
	item, err := driver.DrawCaptcha(content)
	if err != nil {
		return "", "", fmt.Errorf("生成图片验证码失败: %w", err)
	}
	b64s := item.EncodeB64string()
	uuid, err := i.Client.Set("captcha-image", value{
		Code:    answer,
		Carrier: "",
	}, expir)
	if err != nil {
		return "", "", fmt.Errorf("存储图片验证码失败: %w", err)
	}
	return uuid, b64s, nil
}

func (i *Image) Verify(uuid string, code string) error {
	var val value
	if err := i.Client.GetAndDelete(uuid, "captcha-image", &val); err != nil {
		return fmt.Errorf("验证码不存在或已过期: %w", err)
	}
	if !strings.EqualFold(val.Code, code) {
		return fmt.Errorf("验证码错误")
	}
	return nil
}

func (i *Image) Delete(uuid string) error {
	return i.Client.Delete(uuid, "captcha-image")
}
