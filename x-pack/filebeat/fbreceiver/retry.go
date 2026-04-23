// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package fbreceiver

import (
	"context"
	"errors"
	"fmt"
	"time"

	libbeatbackoff "github.com/elastic/beats/v7/libbeat/common/backoff"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/consumererror"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"
)

// RetryConfig controls how the receiver retries ConsumeLogs calls that return
// a retryable error from the downstream pipeline. Retries are infinite by
// design to match Filebeat's guaranteed-delivery semantics.
type RetryConfig struct {
	InitialInterval time.Duration `mapstructure:"initial_interval"`
	MaxInterval     time.Duration `mapstructure:"max_interval"`
}

func defaultRetryConfig() RetryConfig {
	return RetryConfig{
		InitialInterval: 1 * time.Second,
		MaxInterval:     30 * time.Second,
	}
}

type retryLogsConsumer struct {
	consumer.Logs
	cfg    RetryConfig
	logger *zap.Logger
}

func newRetryLogsConsumer(cfg RetryConfig, logger *zap.Logger, next consumer.Logs) consumer.Logs {
	return &retryLogsConsumer{
		Logs:   next,
		cfg:    cfg,
		logger: logger,
	}
}

func (r *retryLogsConsumer) ConsumeLogs(ctx context.Context, logs plog.Logs) error {
	expBackoff := libbeatbackoff.NewExpBackoff(ctx.Done(), r.cfg.InitialInterval, r.cfg.MaxInterval)

	for attempt := 1; ; attempt++ {
		err := r.Logs.ConsumeLogs(ctx, logs)
		if err == nil {
			return nil
		}

		if consumererror.IsPermanent(err) {
			r.logger.Error("ConsumeLogs() failed with permanent error. Dropping data.",
				zap.Error(err),
				zap.Int("dropped_items", logs.LogRecordCount()),
			)
			return err
		}

		// Narrow the retry payload: if the error wraps a consumererror.Logs,
		// it contains only the records that failed. Retrying the full original
		// batch would re-send records already accepted by the downstream.
		var retryableLogs consumererror.Logs
		if errors.As(err, &retryableLogs) {
			logs = retryableLogs.Data()
		}

		r.logger.Debug("ConsumeLogs() failed. Retrying.",
			zap.Error(err),
			zap.Int("attempt", attempt),
		)

		if !expBackoff.Wait() {
			return fmt.Errorf("context cancelled during retry: %w", ctx.Err())
		}
	}
}
