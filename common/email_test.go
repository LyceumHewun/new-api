package common

import (
	"net"
	"strconv"
	"testing"
	"time"
)

func TestSendEmailTimesOutWaitingForSMTPServer(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		time.Sleep(200 * time.Millisecond)
	}()

	host, portValue, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		t.Fatalf("split host port: %v", err)
	}
	port, err := strconv.Atoi(portValue)
	if err != nil {
		t.Fatalf("parse port: %v", err)
	}

	originalSMTPServer := SMTPServer
	originalSMTPPort := SMTPPort
	originalSMTPSSLEnabled := SMTPSSLEnabled
	originalSMTPForceAuthLogin := SMTPForceAuthLogin
	originalSMTPAccount := SMTPAccount
	originalSMTPFrom := SMTPFrom
	originalSMTPToken := SMTPToken
	originalEmailSendTimeout := emailSendTimeout
	defer func() {
		SMTPServer = originalSMTPServer
		SMTPPort = originalSMTPPort
		SMTPSSLEnabled = originalSMTPSSLEnabled
		SMTPForceAuthLogin = originalSMTPForceAuthLogin
		SMTPAccount = originalSMTPAccount
		SMTPFrom = originalSMTPFrom
		SMTPToken = originalSMTPToken
		emailSendTimeout = originalEmailSendTimeout
	}()

	SMTPServer = host
	SMTPPort = port
	SMTPSSLEnabled = false
	SMTPForceAuthLogin = false
	SMTPAccount = "sender@example.com"
	SMTPFrom = "sender@example.com"
	SMTPToken = "token"
	emailSendTimeout = 20 * time.Millisecond

	start := time.Now()
	err = SendEmail("subject", "receiver@example.com", "<p>hello</p>")
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected timeout error")
	}
	if elapsed >= 150*time.Millisecond {
		t.Fatalf("SendEmail took %s, expected SMTP timeout to stop before gateway timeout", elapsed)
	}
}
