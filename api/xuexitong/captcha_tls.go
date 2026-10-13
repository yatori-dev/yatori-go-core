package xuexitong

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	utls "github.com/refraction-networking/utls"
	"golang.org/x/net/http2"
)

// captchaFingerprintTimeout 滑块check接口请求超时
const captchaFingerprintTimeout = 30 * time.Second

// 客户端模式: 超星captcha域名经实测只协商HTTP/1.1(浏览器同款行为为
// ClientHello携带["h2","http/1.1"]，服务器选http/1.1后在同一连接上跑HTTP/1.1)
type captchaFingerprintMode int

const (
	// fingerprintAuto HTTP/1.1连接(完整Chrome ClientHello)；服务器若协商h2则报错交由h2模式兜底
	fingerprintAuto captchaFingerprintMode = iota
	// fingerprintHTTP2 强制HTTP/2(要求协商到h2)
	fingerprintHTTP2
)

// newCaptchaChromeClient 创建伪装Chrome TLS指纹的HTTP客户端。
// 超星滑块 check/verification/result 接口已启用TLS指纹风控，普通Go/OpenSSL
// 指纹会被拒(verification error)，必须使用浏览器指纹才能通过。
func newCaptchaChromeClient(proxyIP string, mode captchaFingerprintMode) *http.Client {
	var proxyURL *url.URL
	if proxyIP != "" {
		if parsed, err := url.Parse("http://" + proxyIP); err == nil {
			proxyURL = parsed
		}
	}

	if mode == fingerprintHTTP2 {
		transport := &http2.Transport{
			DialTLSContext: func(ctx context.Context, _ string, addr string, _ *tls.Config) (net.Conn, error) {
				return dialChromeUTLS(ctx, addr, proxyURL, true)
			},
		}
		return &http.Client{Transport: transport, Timeout: captchaFingerprintTimeout}
	}

	transport := &http.Transport{
		DialTLSContext: func(ctx context.Context, _ string, addr string) (net.Conn, error) {
			return dialChromeUTLS(ctx, addr, proxyURL, false)
		},
	}
	return &http.Client{Transport: transport, Timeout: captchaFingerprintTimeout}
}

// dialChromeUTLS 完成 TCP(可选代理CONNECT) + uTLS Chrome指纹握手。
// requireH2=true 时要求服务器协商到 h2；false 时若协商到 h2 则报错(该连接无法跑HTTP/1.1)
func dialChromeUTLS(ctx context.Context, addr string, proxyURL *url.URL, requireH2 bool) (net.Conn, error) {
	dialer := &net.Dialer{Timeout: 20 * time.Second}
	dialAddr := addr
	if proxyURL != nil {
		dialAddr = proxyURL.Host
	}
	rawConn, err := dialer.DialContext(ctx, "tcp", dialAddr)
	if err != nil {
		return nil, err
	}
	if proxyURL != nil {
		if err := httpProxyConnect(rawConn, addr); err != nil {
			_ = rawConn.Close()
			return nil, err
		}
	}

	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
	}
	config := &utls.Config{
		ServerName: host,
		MinVersion: tls.VersionTLS12,
	}
	uconn := utls.UClient(rawConn, config, utls.HelloChrome_120)
	if err := uconn.HandshakeContext(ctx); err != nil {
		_ = rawConn.Close()
		return nil, err
	}
	negotiated := uconn.ConnectionState().NegotiatedProtocol
	if requireH2 && negotiated != "h2" {
		_ = uconn.Close()
		return nil, fmt.Errorf("HTTP/2 未协商成功: %q", negotiated)
	}
	if !requireH2 && negotiated == "h2" {
		_ = uconn.Close()
		return nil, fmt.Errorf("服务器协商到HTTP/2，请使用HTTP/2模式")
	}
	return uconn, nil
}

// httpProxyConnect 通过HTTP代理向目标地址发起CONNECT隧道
func httpProxyConnect(conn net.Conn, targetAddr string) error {
	req := &http.Request{
		Method: http.MethodConnect,
		URL:    &url.URL{Opaque: targetAddr},
		Host:   targetAddr,
		Header: make(http.Header),
	}
	if err := req.Write(conn); err != nil {
		return err
	}
	reader := bufio.NewReader(conn)
	resp, err := http.ReadResponse(reader, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("代理CONNECT失败: %s", resp.Status)
	}
	return nil
}

// captchaFingerprintGet 使用Chrome指纹客户端发起GET请求
func captchaFingerprintGet(urlStr string, headers map[string]string, proxyIP string, mode captchaFingerprintMode) ([]byte, *http.Response, error) {
	client := newCaptchaChromeClient(proxyIP, mode)
	defer client.CloseIdleConnections()

	req, err := http.NewRequest(http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, nil, err
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}
	return body, resp, nil
}
