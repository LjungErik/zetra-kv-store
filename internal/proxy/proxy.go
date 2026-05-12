package proxy

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Proxy interface {
	ProxyRequest(ctx *gin.Context, proxyToAddr string) error
}

type proxy struct {
	useTLS bool
}

type Option func(*proxy)

func WithUseTLS(useTLS bool) Option {
	return func(p *proxy) {
		p.useTLS = useTLS
	}
}

var _ Proxy = (*proxy)(nil)

func NewProxy(options ...Option) *proxy {
	p := &proxy{
		useTLS: false,
	}

	for _, opt := range options {
		opt(p)
	}

	return p
}

func (p *proxy) ProxyRequest(ctx *gin.Context, proxyToAddr string) error {
	req := ctx.Request
	rw := ctx.Writer
	slog.Debug("Proxy forwarding request", "remote_address", req.RemoteAddr, "method", req.Method, "url", req.URL)

	client := &http.Client{}

	url := fmt.Sprintf("%s://%s%s", p.getProtocol(), proxyToAddr, req.RequestURI)

	slog.Info("sending proxy request", "url", url)

	proxyReq, err := http.NewRequestWithContext(req.Context(), req.Method, url, req.Body)
	if err != nil {
		return fmt.Errorf("failed to create proxy request")
	}

	proxyReq.Header = req.Header.Clone()

	if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err != nil {
		appendXForwardHeader(proxyReq.Header, clientIP)
	}

	resp, err := client.Do(proxyReq)
	if err != nil {
		return fmt.Errorf("failed to proxy request: %w", err)
	}

	defer resp.Body.Close()

	slog.Info("Well this worked, just wonder why")

	copyHeader(rw.Header(), resp.Header)
	rw.WriteHeader(resp.StatusCode)
	if _, err := io.Copy(rw, resp.Body); err != nil {
		return fmt.Errorf("failed to write to response writer: %w", err)
	}

	return nil
}

func appendXForwardHeader(header http.Header, host string) {
	header.Add("X-Forwarded-For", host)
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func (p *proxy) getProtocol() string {
	if p.useTLS {
		return "https"
	}

	return "http"
}
