package main

import (
	"time"

	"github.com/cenkalti/backoff/v4"
	"go.uber.org/zap"
)

const (
	retryDefaultInitInterval    = 100 * time.Millisecond  // 默认第一次重试前等待100ms
	retryDefaultMaxTime         = 2000 * time.Millisecond // 默认重试过程总耗时30秒
	retryDefaultDefaultMaxRetry = 3                       // 默认重试3次
)

// 重试
func retry(callback func() error, initInterval /*第一次重试前要等待多久*/, maxTime /*重试过程一共有多少时间*/ time.Duration,
	maxRetry /*允许重试多少次*/ uint64, notifies /*错误回调*/ ...func(err error, d time.Duration)) error {

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = maxTime
	b.InitialInterval = initInterval
	b.RandomizationFactor = 0.5
	b.Multiplier = 1.5
	bm := backoff.WithMaxRetries(b, maxRetry)
	return backoff.RetryNotify(callback, bm, func(err error, duration time.Duration) {
		for _, fn := range append(notifies, defaultNotify) {
			fn(err, duration)
		}
	})
}

// 默认失败重试前的通知钩子函数
func defaultNotify(err error, d time.Duration) {
	log.Info("方法发生重试", zap.Duration("after", d), zap.Error(err))
}
