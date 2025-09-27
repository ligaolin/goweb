package captcha

import (
	"errors"
	"strings"
	"time"

	"github.com/ligaolin/goweb/cache"
	"github.com/mojocn/base64Captcha"
)

type ImageConfig struct {
	Width      int `json:"width" toml:"width" yaml:"width"`
	Height     int `json:"height" toml:"height" yaml:"height"`
	Length     int `json:"length" toml:"length" yaml:"length"`
	NoiseCount int `json:"noise_count" toml:"noise_count" yaml:"noise_count"` // 噪点数量
}
type Image struct {
	Client *cache.Client
	Config *ImageConfig
}

func NewImage(c *cache.Client, config *ImageConfig) *Image {
	return &Image{
		Client: c,
		Config: config,
	}
}

// 创建图片验证码
func (i *Image) Generate(expir time.Duration) (string, string, error) {
	driver := base64Captcha.NewDriverString(
		i.Config.Height,                    // 高度
		i.Config.Width,                     // 宽度
		i.Config.NoiseCount,                // 噪点数量
		base64Captcha.OptionShowHollowLine, // 显示线条选项
		i.Config.Length,                    // 验证码长度
		base64Captcha.TxtNumbers+base64Captcha.TxtAlphabet, // 数据源
		nil,        // &color.RGBA{R: 255, G: 255, B: 0, A: 255}, &color.RGBA{R: 195, G: 245, B: 237, A: 255}// 背景颜
		nil,        // 字体存储（可以根据需要设置）
		[]string{}, // 字体列表
	)

	_, content, answer := driver.GenerateIdQuestionAnswer()
	item, err := driver.DrawCaptcha(content)
	if err != nil {
		return "", "", err
	}
	b64s := item.EncodeB64string()
	uuid, err := i.Client.Set("captcha-image", Value{
		Code:    answer,
		Carrier: "",
	}, expir)
	return uuid, b64s, err
}

// 验证
func (i *Image) Verify(uuid string, code string) error {
	var val Value
	if err := i.Client.Get(uuid, "captcha-image", &val, false); err != nil {
		return errors.New("验证码不存在或过期")
	}
	if !strings.EqualFold(val.Code, code) {
		return errors.New("验证码错误")
	}
	return nil
}
