package wechat

import (
	"math/rand"
	"context"
	"os"

	"github.com/go-pay/gopay"
	"github.com/go-pay/gopay/wechat/v3"
	"github.com/go-pay/xlog"
)

type WechatPayConfig struct {
	AppID      string `json:"app_id"`
	MchID      string `json:"mch_id"`
	SerialNo   string `json:"serial_no"`
	ApiV3Key   string `json:"api_v3_key"`
	PrivateKey string `json:"private_key"`
	NotifyUrl  string `json:"notify_url"`
}

type WechatMerchant struct {
	Client *wechat.ClientV3
	Config *WechatPayConfig
}

func NewWechatMerchant(cfg *WechatPayConfig) (*WechatMerchant, error) {
	privateKey, err := os.ReadFile(cfg.PrivateKey)
	if err != nil {
		xlog.Error(err)
		return nil, err
	}
	client, err := wechat.NewClientV3(cfg.MchID, cfg.SerialNo, cfg.ApiV3Key, string(privateKey))
	if err != nil {
		xlog.Error(err)
		return nil, err
	}

	err = client.AutoVerifySign()
	if err != nil {
		xlog.Error(err)
		return nil, err
	}

	client.DebugSwitch = gopay.DebugOn
	return &WechatMerchant{Client: client, Config: cfg}, nil
}

func (wm *WechatMerchant) NativePay(c context.Context, tradeNo string, description string, price float32, ip string) (*wechat.NativeRsp, error) {
	totalFee := int64(price * 100)
	bm := make(gopay.BodyMap)
	bm.Set("appid", wm.Config.AppID).
		Set("description", description).
		Set("out_trade_no", tradeNo).
		Set("notify_url", wm.Config.NotifyUrl).
		SetBodyMap("amount", func(bm gopay.BodyMap) {
			bm.Set("total", totalFee).
				Set("currency", "CNY")
		})
	return wm.Client.V3TransactionNative(c, bm)
}

func GenerateOutTradeNo() string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, 22)
	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

