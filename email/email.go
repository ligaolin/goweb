package email

import (
	"fmt"

	"gopkg.in/gomail.v2"
)

type EmailConfig struct {
	Smtp     string
	Port     int
	Email    string
	Password string
	FromName string
}

type Email struct {
	dialer *gomail.Dialer
	config *EmailConfig
}

func New(cfg *EmailConfig) (*Email, error) {
	if err := validateConfig(cfg); err != nil {
		return nil, err
	}

	return &Email{
		dialer: gomail.NewDialer(cfg.Smtp, cfg.Port, cfg.Email, cfg.Password),
		config: cfg,
	}, nil
}

func validateConfig(cfg *EmailConfig) error {
	if cfg.Smtp == "" {
		return fmt.Errorf("smtp服务器地址必须")
	}
	if cfg.Port <= 0 || cfg.Port > 65535 {
		return fmt.Errorf("无效的端口号")
	}
	if cfg.Email == "" {
		return fmt.Errorf("邮箱地址必须")
	}
	if cfg.Password == "" {
		return fmt.Errorf("邮箱密码必须")
	}
	return nil
}

func (e *Email) Send(to []string, subject, body string, opts ...Option) error {
	if err := e.validateSendParams(to, subject); err != nil {
		return err
	}

	options := newEmailOptions()
	for _, opt := range opts {
		opt(options)
	}

	m := gomail.NewMessage()

	from := e.config.Email
	if options.fromName != "" {
		m.SetHeader("From", m.FormatAddress(from, options.fromName))
	} else if e.config.FromName != "" {
		m.SetHeader("From", m.FormatAddress(from, e.config.FromName))
	} else {
		m.SetHeader("From", from)
	}

	m.SetHeader("To", to...)
	m.SetHeader("Subject", subject)

	if options.htmlBody != "" {
		if body != "" {
			m.SetBody("text/plain", body)
			m.AddAlternative("text/html", options.htmlBody)
		} else {
			m.SetBody("text/html", options.htmlBody)
		}
	} else {
		m.SetBody(options.contentType, body)
	}

	if len(options.cc) > 0 {
		m.SetHeader("Cc", options.cc...)
	}
	if len(options.bcc) > 0 {
		m.SetHeader("Bcc", options.bcc...)
	}

	for _, attachment := range options.attachments {
		m.Attach(attachment)
	}

	if err := e.dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("发送邮件失败: %w", err)
	}
	return nil
}

func (e *Email) validateSendParams(to []string, subject string) error {
	if len(to) == 0 {
		return fmt.Errorf("收件人不能为空")
	}
	if subject == "" {
		return fmt.Errorf("邮件主题不能为空")
	}
	return nil
}

func (e *Email) Close() error {
	return nil
}

type emailOptions struct {
	contentType string
	cc          []string
	bcc         []string
	attachments []string
	fromName    string
	htmlBody    string
}

func newEmailOptions() *emailOptions {
	return &emailOptions{
		contentType: "text/plain",
	}
}

type Option func(*emailOptions)

func WithHTML() Option {
	return func(o *emailOptions) {
		o.contentType = "text/html"
	}
}

func WithHTMLBody(html string) Option {
	return func(o *emailOptions) {
		o.htmlBody = html
	}
}

func WithCC(cc ...string) Option {
	return func(o *emailOptions) {
		o.cc = append(o.cc, cc...)
	}
}

func WithBCC(bcc ...string) Option {
	return func(o *emailOptions) {
		o.bcc = append(o.bcc, bcc...)
	}
}

func WithAttachment(file string) Option {
	return func(o *emailOptions) {
		o.attachments = append(o.attachments, file)
	}
}

func WithAttachments(files ...string) Option {
	return func(o *emailOptions) {
		o.attachments = append(o.attachments, files...)
	}
}

func WithFromName(name string) Option {
	return func(o *emailOptions) {
		o.fromName = name
	}
}
