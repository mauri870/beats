// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package fbreceiver

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/consumererror"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"
)

// makeLogs returns a plog.Logs with n log records.
func makeLogs(n int) plog.Logs {
	ld := plog.NewLogs()
	sl := ld.ResourceLogs().AppendEmpty().ScopeLogs().AppendEmpty()
	for i := range n {
		sl.LogRecords().AppendEmpty().Body().SetStr(fmt.Sprintf("record-%d", i))
	}
	return ld
}

// fastRetryConfig returns a RetryConfig with very short intervals so unit
// tests complete in milliseconds rather than seconds.
func fastRetryConfig() RetryConfig {
	return RetryConfig{
		InitialInterval: 1 * time.Millisecond,
		MaxInterval:     10 * time.Millisecond,
	}
}

func TestRetryLogsConsumerRetriesOnTransientError(t *testing.T) {
	const failCount = 5
	var calls atomic.Int64

	next, err := consumer.NewLogs(func(_ context.Context, _ plog.Logs) error {
		if n := calls.Add(1); n <= failCount {
			return errors.New("transient error")
		}
		return nil
	})
	require.NoError(t, err)

	rc := newRetryLogsConsumer(fastRetryConfig(), zap.NewNop(), next)

	// A single ConsumeLogs call must retry internally until the downstream succeeds.
	require.NoError(t, rc.ConsumeLogs(t.Context(), plog.NewLogs()))
	assert.Greater(t, calls.Load(), int64(failCount),
		"consumer must have been called more than %d times", failCount)
}

func TestRetryLogsConsumerNeverRetriesPermanentError(t *testing.T) {
	var calls atomic.Int64

	next, err := consumer.NewLogs(func(_ context.Context, _ plog.Logs) error {
		calls.Add(1)
		return consumererror.NewPermanent(errors.New("permanent failure"))
	})
	require.NoError(t, err)

	rc := newRetryLogsConsumer(fastRetryConfig(), zap.NewNop(), next)

	err = rc.ConsumeLogs(t.Context(), plog.NewLogs())
	require.Error(t, err)
	assert.Equal(t, int64(1), calls.Load(), "permanent errors must never be retried")
}

func TestRetryLogsConsumerStopsOnContextCancel(t *testing.T) {
	var calls atomic.Int64

	next, err := consumer.NewLogs(func(_ context.Context, _ plog.Logs) error {
		calls.Add(1)
		return errors.New("always fails")
	})
	require.NoError(t, err)

	cfg := RetryConfig{
		InitialInterval: 50 * time.Millisecond,
		MaxInterval:     50 * time.Millisecond,
	}
	rc := newRetryLogsConsumer(cfg, zap.NewNop(), next)

	ctx, cancel := context.WithCancel(t.Context())
	// Cancel after the first call completes.
	go func() {
		for calls.Load() == 0 {
			time.Sleep(time.Millisecond)
		}
		cancel()
	}()

	err = rc.ConsumeLogs(ctx, plog.NewLogs())
	require.Error(t, err)
	require.ErrorIs(t, err, context.Canceled)
	assert.GreaterOrEqual(t, calls.Load(), int64(1))
}

func TestRetryLogsConsumerNarrowsRetryPayload(t *testing.T) {
	// Simulate a partial-batch failure: the downstream accepts the first call
	// but wraps only the second half of the records in consumererror.Logs.
	// The retry must be issued only with those failed records, not the full
	// original batch — mirroring the upstream consumerretry.NewLogs behaviour.
	const total = 4

	var receivedCounts []int
	var mu sync.Mutex

	next, err := consumer.NewLogs(func(_ context.Context, ld plog.Logs) error {
		n := ld.LogRecordCount()
		mu.Lock()
		receivedCounts = append(receivedCounts, n)
		mu.Unlock()

		if n == total {
			// First call: fail, returning only the last half as needing retry.
			partial := plog.NewLogs()
			sl := partial.ResourceLogs().AppendEmpty().ScopeLogs().AppendEmpty()
			src := ld.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords()
			for i := total / 2; i < src.Len(); i++ {
				src.At(i).CopyTo(sl.LogRecords().AppendEmpty())
			}
			return consumererror.NewLogs(errors.New("partial failure"), partial)
		}
		// Second call: succeed (narrowed payload of total/2 records).
		return nil
	})
	require.NoError(t, err)

	rc := newRetryLogsConsumer(fastRetryConfig(), zap.NewNop(), next)
	require.NoError(t, rc.ConsumeLogs(t.Context(), makeLogs(total)))

	mu.Lock()
	defer mu.Unlock()
	require.Len(t, receivedCounts, 2, "expected exactly 2 calls to the downstream consumer")
	assert.Equal(t, total, receivedCounts[0], "first call must receive the full batch")
	assert.Equal(t, total/2, receivedCounts[1], "retry must receive only the narrowed (failed) records")
}
