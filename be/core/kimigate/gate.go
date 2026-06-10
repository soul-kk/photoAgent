package kimigate

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"
)

// ErrTooManyConcurrent 并发槽位等待超时。
var ErrTooManyConcurrent = errors.New("kimi: too many concurrent requests")

// Gate 限制 Kimi 上游并发并提供调优过的 HTTP 客户端。
type Gate struct {
	sem          chan struct{}
	client       *http.Client
	timeout      time.Duration
	queueTimeout time.Duration
}

// Options 从配置构造 Gate。
type Options struct {
	MaxConcurrent  int
	TimeoutSec     int
	QueueWaitSec   int
}

func New(opts Options) *Gate {
	maxC := opts.MaxConcurrent
	if maxC <= 0 {
		maxC = 5
	}
	timeoutSec := opts.TimeoutSec
	if timeoutSec <= 0 {
		timeoutSec = 300
	}
	queueSec := opts.QueueWaitSec
	if queueSec <= 0 {
		queueSec = 30
	}
	timeout := time.Duration(timeoutSec) * time.Second
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   20,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: timeout,
	}
	return &Gate{
		sem: make(chan struct{}, maxC),
		client: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
		timeout:      timeout,
		queueTimeout: time.Duration(queueSec) * time.Second,
	}
}

func (g *Gate) HTTPClient() *http.Client {
	if g == nil {
		return http.DefaultClient
	}
	return g.client
}

func (g *Gate) Timeout() time.Duration {
	if g == nil {
		return 300 * time.Second
	}
	return g.timeout
}

// Acquire 获取并发槽位；等待超过 queueTimeout 返回 ErrTooManyConcurrent。
func (g *Gate) Acquire(ctx context.Context) error {
	if g == nil {
		return nil
	}
	timer := time.NewTimer(g.queueTimeout)
	defer timer.Stop()
	select {
	case g.sem <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return ErrTooManyConcurrent
	}
}

func (g *Gate) Release() {
	if g == nil {
		return
	}
	select {
	case <-g.sem:
	default:
	}
}

// UpstreamContext 非流式：客户端断开也不取消上游，避免网关先断导致半失败。
func (g *Gate) UpstreamContext(parent context.Context) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}
	return context.WithTimeout(context.WithoutCancel(parent), g.Timeout())
}

// StreamUpstreamContext 流式：客户端断开时取消上游，释放连接与配额。
func (g *Gate) StreamUpstreamContext(parent context.Context) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}
	return context.WithTimeout(parent, g.Timeout())
}
