package ali

import (
	"context"
	"fmt"
	"os"

	"github.com/go-pay/gopay"
	"github.com/go-pay/gopay/alipay/v3"
	"github.com/go-pay/util"
	"github.com/go-pay/xlog"
)

type AliPayConfig struct {
	AppID               string `json:"app_id" toml:"app_id" yaml:"app_id"`
	PrivateKey          string `json:"private_key" toml:"private_key" yaml:"private_key"`                                  // 文件路径
	AppPublicCert       string `json:"app_public_cert" toml:"app_public_cert" yaml:"app_public_cert"`                      // 文件路径
	AlipayRootCert      string `json:"alipay_root_cert" toml:"alipay_root_cert" yaml:"alipay_root_cert"`                   // 文件路径
	AlipayCertPublicKey string `json:"alipay_cert_public_key" toml:"alipay_cert_public_key" yaml:"alipay_cert_public_key"` // 文件路径
}

type AliPay struct {
	Client *alipay.ClientV3
	Config *AliPayConfig
}

func NewAliPay(cfg *AliPayConfig) (*AliPay, error) {
	privateKey, err := os.ReadFile(cfg.PrivateKey)
	if err != nil {
		xlog.Error(err)
		return nil, err
	}

	// 初始化支付宝客V3户端
	// appid：应用ID
	// privateKey：应用私钥，支持PKCS1和PKCS8
	// isProd：是否是正式环境，沙箱环境请选择新版沙箱应用。
	client, err := alipay.NewClientV3(cfg.AppID, string(privateKey), true)
	if err != nil {
		xlog.Error(err)
		return nil, err
	}

	// 自定义配置http请求接收返回结果body大小，默认 10MB
	// client.SetBodySize() // 没有特殊需求，可忽略此配置

	// 设置自定义RequestId生成方法，非必须
	// client.SetRequestIdFunc()

	// 打开Debug开关，输出日志，默认关闭
	client.DebugSwitch = gopay.DebugOn

	// 设置biz_content加密KEY，设置此参数默认开启加密（目前不可用）
	//client.SetAESKey("1234567890123456")

	// 传入证书内容
	app_public_cert, err := os.ReadFile(cfg.AppPublicCert)
	if err != nil {
		xlog.Error(err)
		return nil, err
	}

	// 传入证书内容
	alipay_root_cert, err := os.ReadFile(cfg.AlipayRootCert)
	if err != nil {
		xlog.Error(err)
		return nil, err
	}

	// 传入证书内容
	alipay_cert_public_key, err := os.ReadFile(cfg.AlipayCertPublicKey)
	if err != nil {
		xlog.Error(err)
		return nil, err
	}
	err = client.SetCert(app_public_cert, alipay_root_cert, alipay_cert_public_key)
	return &AliPay{Client: client, Config: cfg}, err
}

func (ap *AliPay) NativePay(ctx context.Context, subject string, price float32) (*alipay.TradePrecreateRsp, error) {
	// 请求参数
	bm := make(gopay.BodyMap)
	bm.Set("subject", subject).
		Set("out_trade_no", util.RandomString(32)).
		Set("total_amount", fmt.Sprintf("%.2f", price))

	// 创建订单
	aliRsp, err := ap.Client.TradePrecreate(ctx, bm)
	if err != nil {
		xlog.Errorf("client.TradePrecreate(), err:%v", err)
		return nil, err
	}

	if aliRsp.StatusCode != 200 {
		xlog.Errorf("aliRsp.StatusCode:%d", aliRsp.StatusCode)
		return nil, err
	}
	xlog.Warnf("aliRsp.QrCode:", aliRsp.QrCode)
	xlog.Warnf("aliRsp.OutTradeNo:", aliRsp.OutTradeNo)
	return aliRsp, nil
}
