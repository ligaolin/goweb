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
	AppID               string `json:"app_id"`
	PrivateKey          string `json:"private_key"`
	AppPublicCert       string `json:"app_public_cert"`
	AlipayRootCert      string `json:"alipay_root_cert"`
	AlipayCertPublicKey string `json:"alipay_cert_public_key"`
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

	client, err := alipay.NewClientV3(cfg.AppID, string(privateKey), true)
	if err != nil {
		xlog.Error(err)
		return nil, err
	}

	client.DebugSwitch = gopay.DebugOn

	appPublicCert, err := os.ReadFile(cfg.AppPublicCert)
	if err != nil {
		xlog.Error(err)
		return nil, err
	}

	alipayRootCert, err := os.ReadFile(cfg.AlipayRootCert)
	if err != nil {
		xlog.Error(err)
		return nil, err
	}

	alipayCertPublicKey, err := os.ReadFile(cfg.AlipayCertPublicKey)
	if err != nil {
		xlog.Error(err)
		return nil, err
	}

	err = client.SetCert(appPublicCert, alipayRootCert, alipayCertPublicKey)
	return &AliPay{Client: client, Config: cfg}, err
}

func (ap *AliPay) NativePay(ctx context.Context, subject string, price float32) (*alipay.TradePrecreateRsp, error) {
	bm := make(gopay.BodyMap)
	bm.Set("subject", subject).
		Set("out_trade_no", util.RandomString(32)).
		Set("total_amount", fmt.Sprintf("%.2f", price))

	aliRsp, err := ap.Client.TradePrecreate(ctx, bm)
	if err != nil {
		xlog.Errorf("client.TradePrecreate(): %v", err)
		return nil, err
	}

	if aliRsp.StatusCode != 200 {
		xlog.Errorf("aliRsp.StatusCode: %d", aliRsp.StatusCode)
		return nil, fmt.Errorf("支付宝创建订单失败，状态码: %d", aliRsp.StatusCode)
	}

	xlog.Warnf("aliRsp.QrCode: %s, OutTradeNo: %s", aliRsp.QrCode, aliRsp.OutTradeNo)
	return aliRsp, nil
}
