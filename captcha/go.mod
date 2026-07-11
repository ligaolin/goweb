module github.com/ligaolin/goweb/captcha

go 1.26.4

replace (
	github.com/ligaolin/goweb/cache => ../cache
	github.com/ligaolin/goweb/email => ../email
	github.com/ligaolin/goweb/sdk/ali => ../sdk/ali
)

require (
	github.com/alibabacloud-go/dysmsapi-20170525/v5 v5.6.0
	github.com/alibabacloud-go/tea v1.5.2
	github.com/alibabacloud-go/tea-utils/v2 v2.0.9
	github.com/ligaolin/goweb/cache v0.0.0-00010101000000-000000000000
	github.com/ligaolin/goweb/email v0.0.0-00010101000000-000000000000
	github.com/ligaolin/goweb/sdk/ali v0.0.0-00010101000000-000000000000
	github.com/mojocn/base64Captcha v1.3.6
)

require (
	github.com/alibabacloud-go/alibabacloud-gateway-spi v0.0.5 // indirect
	github.com/alibabacloud-go/darabonba-openapi/v2 v2.2.3 // indirect
	github.com/alibabacloud-go/debug v1.0.1 // indirect
	github.com/aliyun/credentials-go v1.4.5 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/clbanning/mxj/v2 v2.7.0 // indirect
	github.com/go-pay/crypto v0.0.1 // indirect
	github.com/go-pay/gopay v1.5.122 // indirect
	github.com/go-pay/util v0.0.4 // indirect
	github.com/go-pay/xlog v0.0.3 // indirect
	github.com/go-pay/xtime v0.0.2 // indirect
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/google/btree v1.1.3 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/redis/go-redis/v9 v9.21.0 // indirect
	github.com/tjfoc/gmsm v1.4.1 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	golang.org/x/image v0.13.0 // indirect
	golang.org/x/net v0.26.0 // indirect
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
)
