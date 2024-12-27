// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package fbreceiver

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/receivertest"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestNewReceiver(t *testing.T) {
	config := Config{
		Beatconfig: map[string]interface{}{
			"filebeat": map[string]interface{}{
				"inputs": []map[string]interface{}{
					{
						"type":    "benchmark",
						"enabled": true,
						"message": "test",
						"count":   1,
					},
				},
			},
			"output": map[string]interface{}{
				"otelconsumer": map[string]interface{}{},
			},
			"logging": map[string]interface{}{
				"level": "debug",
				"selectors": []string{
					"*",
				},
			},
			"path.home": t.TempDir(),
		},
	}

	var zapLogs bytes.Buffer
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(&zapLogs),
		zapcore.DebugLevel)

	receiverSettings := receiver.Settings{}
	receiverSettings.Logger = zap.New(core)

	var countLogs int
	logConsumer, err := consumer.NewLogs(func(ctx context.Context, ld plog.Logs) error {
		countLogs = countLogs + ld.LogRecordCount()
		return nil
	})
	assert.NoError(t, err, "Error creating log consumer")

	r, err := NewFactory().CreateLogsReceiver(context.Background(), receiverSettings, &config, logConsumer)
	assert.NoError(t, err, "Error creating log receiver")
	err = r.Start(context.Background(), componenttest.NewNopHost())
	assert.NoError(t, err, "Error starting filebeatreceiver")

	ch := make(chan bool, 1)
	timer := time.NewTimer(120 * time.Second)
	defer timer.Stop()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for tick := ticker.C; ; {
		select {
		case <-timer.C:
			t.Fatalf("consumed logs didn't increase\nCount: %d\nLogs: %s\n", countLogs, zapLogs.String())
		case <-tick:
			tick = nil
			go func() { ch <- countLogs > 0 }()
		case v := <-ch:
			if v {
				goto found
			}
			tick = ticker.C
		}
	}
found:
	err = r.Shutdown(context.Background())
	assert.NoError(t, err, "Error shutting down filebeatreceiver")
}

type logGenerator struct {
	t           *testing.T
	tmpDir      string
	f           *os.File
	filePattern string
	sequenceNum int64
}

func newLogGenerator(t *testing.T, tmpDir string) *logGenerator {
	f, err := os.CreateTemp(tmpDir, "input-*.log")
	require.NoError(t, err)
	return &logGenerator{
		t:           t,
		f:           f,
		filePattern: "input-*.log",
	}
}

func (g *logGenerator) Start() {
	g.sequenceNum = 0
}

func (g *logGenerator) Stop() {}

func (g *logGenerator) Generate() []receivertest.UniqueIDAttrVal {
	// Generate may be called concurrently.
	id := receivertest.UniqueIDAttrVal(strconv.FormatInt(atomic.AddInt64(&g.sequenceNum, 1), 10))

	_, err := io.WriteString(g.f, fmt.Sprintf(`{"id": "%s", "message": "log message"}`, id))
	require.NoError(g.t, err, "failed to write log line to file")
	_, err = g.f.Write([]byte("\n"))
	require.NoError(g.t, err, "failed to write newline to file")
	require.NoError(g.t, g.f.Sync(), "failed to sync log file")

	// And return the ids for bookkeeping by the test.
	return []receivertest.UniqueIDAttrVal{id}
}

func TestConsumeContract(t *testing.T) {
	tmpDir := t.TempDir()
	const logsPerTest = 1

	gen := newLogGenerator(t, tmpDir)
	inputFile := gen.f.Name()

	cfg := &Config{
		Beatconfig: map[string]interface{}{
			"filebeat": map[string]interface{}{
				"inputs": []map[string]interface{}{
					{
						"type":    "filestream",
						"id":      "filestream-test",
						"enabled": true,
						"paths": []string{
							inputFile,
						},
						"queue.mem.flush.timeout": "0s",
						"file_identity.native":    map[string]interface{}{},
						"prospector": map[string]interface{}{
							"scanner": map[string]interface{}{
								"fingerprint.enabled": false,
							},
						},
						"parsers": []map[string]interface{}{
							{
								"ndjson": map[string]interface{}{
									"document_id": "id",
								},
							},
						},
					},
				},
			},
			"output": map[string]interface{}{
				"otelconsumer": map[string]interface{}{},
			},
			"logging": map[string]interface{}{
				"level": "debug",
				"selectors": []string{
					"*",
				},
			},
			"path.home": tmpDir,
		},
	}

	// Run the contract checker. This will trigger test failures if any problems are found.
	receivertest.CheckConsumeContract(receivertest.CheckConsumeContractParams{
		T:        t,
		Factory:  NewFactory(),
		DataType: component.DataTypeLogs,
		GetLogRecordID: func(record plog.LogRecord) (pcommon.Value, bool) {
			body, exists := record.Body().Map().Get("@metadata")
			if !exists {
				return pcommon.NewValueEmpty(), exists
			}

			return body.Map().Get("_id")
		},
		Config:        cfg,
		Generator:     gen,
		GenerateCount: logsPerTest,
	})
}
