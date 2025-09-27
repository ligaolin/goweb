package wechat

import (
	"context"
	"os"

	"github.com/go-pay/gopay"
	"github.com/go-pay/gopay/wechat/v3"
	"github.com/go-pay/xlog"
	"github.com/ligaolin/gin_lin/v2"
)

type WechatPayConfig struct {
	AppID      string `json:"app_id" toml:"app_id" yaml:"app_id"`
	MchID      string `json:"mch_id" toml:"mch_id" yaml:"mch_id"`
	SerialNo   string `json:"serial_no" toml:"serial_no" yaml:"serial_no"`
	ApiV3Key   string `json:"api_v3_key" toml:"api_v3_key" yaml:"api_v3_key"`
	PrivateKey string `json:"private_key" toml:"private_key" yaml:"private_key"` // 文件路径
	NotifyUrl  string `json:"notify_url" toml:"notify_url" yaml:"notify_url"`
}

type WechatMerchant struct {
	Client *wechat.ClientV3
	Config *WechatPayConfig
}

func NewWechatMerchant(cfg *WechatPayConfig) (*WechatMerchant, error) {
	// NewClientV3 初始化微信客户端 v3
	// mchid：商户ID 或者服务商模式的 sp_mchid
	// serialNo：商户证书的证书序列号
	// apiV3Key：apiV3Key，商户平台获取
	// privateKey：私钥 apiclient_key.pem 读取后的内容
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

	// 注意：以下两种自动验签方式二选一
	// 微信支付公钥自动同步验签（新微信支付用户推荐）
	// err = wm.Client.AutoVerifySignByPublicKey([]byte("微信支付公钥内容"), "微信支付公钥ID")
	// if err != nil {
	// 	xlog.Error(err)
	// 	return
	// }
	//// 微信平台证书自动获取证书+同步验签（并自动定时更新微信平台API证书）
	err = client.AutoVerifySign()
	if err != nil {
		xlog.Error(err)
		return nil, err
	}

	// 自定义配置http请求接收返回结果body大小，默认 10MB
	// wm.Client.SetBodySize() // 没有特殊需求，可忽略此配置

	// 设置自定义RequestId生成方法，非必须
	// wm.Client.SetRequestIdFunc()

	// 打开Debug开关，输出日志，默认是关闭的
	client.DebugSwitch = gopay.DebugOn
	return &WechatMerchant{Client: client, Config: cfg}, nil
}

func (wm *WechatMerchant) NativePay(c context.Context, tradeNo string, description string, price float32, ip string) (*wechat.NativeRsp, error) {
	// 设置支付参数
	totalFee := int64(price * 100)
	bm := make(gopay.BodyMap)
	bm.Set("appid", wm.Config.AppID).
		Set("description", description).        // 商品描述
		Set("out_trade_no", tradeNo).           // 商户订单号
		Set("notify_url", wm.Config.NotifyUrl). // 支付结果通知地址
		SetBodyMap("amount", func(bm gopay.BodyMap) {
			bm.Set("total", totalFee).
				Set("currency", "CNY")
		})
	return wm.Client.V3TransactionNative(c, bm)
}

// 生成商户订单号
func GenerateOutTradeNo() string {
	return gin_lin.GenerateRandomAlphanumeric(22)
}
