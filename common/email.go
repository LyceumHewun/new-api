package common

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"net/smtp"
	"slices"
	"strings"
	"time"
)

var emailSendTimeout = 10 * time.Second

func generateMessageID() (string, error) {
	split := strings.Split(SMTPFrom, "@")
	if len(split) < 2 {
		return "", fmt.Errorf("invalid SMTP account")
	}
	domain := strings.Split(SMTPFrom, "@")[1]
	return fmt.Sprintf("<%d.%s@%s>", time.Now().UnixNano(), GetRandomString(12), domain), nil
}

func shouldUseSMTPLoginAuth() bool {
	if SMTPForceAuthLogin {
		return true
	}
	return isOutlookServer(SMTPAccount) || slices.Contains(EmailLoginAuthServerList, SMTPServer)
}

func getSMTPAuth() smtp.Auth {
	if shouldUseSMTPLoginAuth() {
		return LoginAuth(SMTPAccount, SMTPToken)
	}
	return smtp.PlainAuth("", SMTPAccount, SMTPToken, SMTPServer)
}

func SendEmail(subject string, receiver string, content string) error {
	if SMTPFrom == "" { // for compatibility
		SMTPFrom = SMTPAccount
	}
	id, err2 := generateMessageID()
	if err2 != nil {
		return err2
	}
	if SMTPServer == "" && SMTPAccount == "" {
		return fmt.Errorf("SMTP 服务器未配置")
	}
	encodedSubject := fmt.Sprintf("=?UTF-8?B?%s?=", base64.StdEncoding.EncodeToString([]byte(subject)))
	mail := []byte(fmt.Sprintf("To: %s\r\n"+
		"From: %s <%s>\r\n"+
		"Subject: %s\r\n"+
		"Date: %s\r\n"+
		"Message-ID: %s\r\n"+ // 添加 Message-ID 头
		"Content-Type: text/html; charset=UTF-8\r\n\r\n%s\r\n",
		receiver, SystemName, SMTPFrom, encodedSubject, time.Now().Format(time.RFC1123Z), id, content))
	auth := getSMTPAuth()
	addr := fmt.Sprintf("%s:%d", SMTPServer, SMTPPort)
	to := strings.Split(receiver, ";")
	var err error
	deadline := time.Now().Add(emailSendTimeout)
	if SMTPPort == 465 || SMTPSSLEnabled {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         SMTPServer,
		}
		dialer := &net.Dialer{Timeout: emailSendTimeout}
		conn, err := tls.DialWithDialer(dialer, "tcp", addr, tlsConfig)
		if err != nil {
			return err
		}
		_ = conn.SetDeadline(deadline)
		client, err := smtp.NewClient(conn, SMTPServer)
		if err != nil {
			_ = conn.Close()
			return err
		}
		err = sendMailWithClient(client, auth, SMTPFrom, to, mail, false, true, SMTPServer)
	} else {
		err = sendMailWithTimeout(addr, auth, SMTPFrom, to, mail, deadline, SMTPServer)
	}
	if err != nil {
		SysError(fmt.Sprintf("failed to send email to %s: %v", receiver, err))
	}
	return err
}

func sendMailWithTimeout(addr string, auth smtp.Auth, from string, to []string, mail []byte, deadline time.Time, serverName string) error {
	timeout := time.Until(deadline)
	if timeout <= 0 {
		return fmt.Errorf("SMTP send timeout")
	}
	dialer := &net.Dialer{Timeout: timeout}
	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return err
	}
	_ = conn.SetDeadline(deadline)
	client, err := smtp.NewClient(conn, serverName)
	if err != nil {
		_ = conn.Close()
		return err
	}
	return sendMailWithClient(client, auth, from, to, mail, true, false, serverName)
}

func sendMailWithClient(client *smtp.Client, auth smtp.Auth, from string, to []string, mail []byte, startTLS bool, requireAuth bool, serverName string) error {
	defer client.Close()
	if startTLS {
		if ok, _ := client.Extension("STARTTLS"); ok {
			config := &tls.Config{ServerName: serverName}
			if err := client.StartTLS(config); err != nil {
				return err
			}
		}
	}
	if auth != nil {
		if ok, _ := client.Extension("AUTH"); ok || requireAuth {
			if err := client.Auth(auth); err != nil {
				return err
			}
		}
	}
	if err := client.Mail(from); err != nil {
		return err
	}
	for _, receiver := range to {
		if err := client.Rcpt(receiver); err != nil {
			return err
		}
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err = w.Write(mail); err != nil {
		_ = w.Close()
		return err
	}
	if err = w.Close(); err != nil {
		return err
	}
	return client.Quit()
}
