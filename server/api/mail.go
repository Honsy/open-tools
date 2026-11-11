package api

import (
	"crypto/tls"
	"errors"
	"fmt"
	"mime"
	"net"
	"net/smtp"
	"strings"

	"opentools/config"
)

var (
	smtpHost     string
	smtpPort     string
	smtpUser     string
	smtpPassword string
	smtpFrom     string
)

func configureMail(cfg config.Config) {
	smtpHost = strings.TrimSpace(cfg.SMTPHost)
	smtpPort = strings.TrimSpace(cfg.SMTPPort)
	if smtpPort == "" {
		smtpPort = "465"
	}
	smtpUser = strings.TrimSpace(cfg.SMTPUser)
	smtpPassword = cfg.SMTPPassword
	smtpFrom = strings.TrimSpace(cfg.SMTPFrom)
}

func mailConfigured() bool {
	return smtpHost != "" && smtpFrom != "" && smtpUser != "" && smtpPassword != ""
}

func sendMail(to, subject, body string) error {
	if !mailConfigured() {
		return errors.New("邮件还没配好")
	}
	fromAddr := envelopeFrom(smtpFrom)
	headerFrom := smtpFrom
	if !strings.Contains(headerFrom, "<") {
		headerFrom = fromAddr
	}
	msg := []byte("From: " + headerFrom + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + mime.QEncoding.Encode("utf-8", subject) + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n" +
		"\r\n" +
		body)
	addr := net.JoinHostPort(smtpHost, smtpPort)
	auth := smtp.PlainAuth("", smtpUser, smtpPassword, smtpHost)
	if smtpPort == "465" {
		return sendMailTLS(addr, plainAuth{smtpUser, smtpPassword}, fromAddr, to, msg)
	}
	return smtp.SendMail(addr, auth, fromAddr, []string{to}, msg)
}

// plainAuth is for an already encrypted connection. smtp.PlainAuth refuses
// credentials unless the client itself negotiated STARTTLS.
type plainAuth struct {
	user, pass string
}

func (a plainAuth) Start(*smtp.ServerInfo) (string, []byte, error) {
	return "PLAIN", []byte("\x00" + a.user + "\x00" + a.pass), nil
}

func (a plainAuth) Next(_ []byte, more bool) ([]byte, error) {
	if more {
		return nil, errors.New("邮件服务器还在要验证")
	}
	return nil, nil
}

func sendMailTLS(addr string, auth smtp.Auth, from, to string, msg []byte) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: smtpHost})
	if err != nil {
		return err
	}
	defer conn.Close()
	client, err := smtp.NewClient(conn, smtpHost)
	if err != nil {
		return err
	}
	defer client.Close()
	if err := client.Auth(auth); err != nil {
		return err
	}
	if err := client.Mail(from); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return client.Quit()
}

func envelopeFrom(from string) string {
	from = strings.TrimSpace(from)
	if i := strings.LastIndex(from, "<"); i >= 0 && strings.HasSuffix(from, ">") {
		return strings.TrimSpace(from[i+1 : len(from)-1])
	}
	return from
}

func resetMailBody(link string) string {
	return fmt.Sprintf("你正在重置开物录的密码。\n\n打开下面的链接设置新密码，30 分钟内有效：\n%s\n\n如果不是你本人操作，忽略这封邮件。密码不会因此改变。\n", link)
}
