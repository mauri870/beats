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

	"github.com/elastic/beats/v7/libbeat/otelbeat/oteltest"
	"github.com/elastic/elastic-agent-libs/mapstr"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pipeline"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/receivertest"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
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

	oteltest.CheckReceivers(oteltest.CheckReceiversParams{
		T: t,
		Receivers: []oteltest.ReceiverConfig{
			{
				Name:    "r1",
				Config:  &config,
				Factory: NewFactory(),
			},
		},
		AssertFunc: func(t *assert.CollectT, logs map[string][]mapstr.M, zapLogs *observer.ObservedLogs) {
			_ = zapLogs
			assert.Len(t, logs["r1"], 1)
		},
	})
}

func TestReceiverDefaultProcessors(t *testing.T) {
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

	oteltest.CheckReceivers(oteltest.CheckReceiversParams{
		T: t,
		Receivers: []oteltest.ReceiverConfig{
			{
				Name:    "r1",
				Config:  &config,
				Factory: NewFactory(),
			},
		},
		AssertFunc: func(t *assert.CollectT, logs map[string][]mapstr.M, zapLogs *observer.ObservedLogs) {
			require.Len(t, logs["r1"], 1)

			processorsLoaded := zapLogs.FilterMessageSnippet("Generated new processors").
				FilterMessageSnippet("add_host_metadata").
				FilterMessageSnippet("add_cloud_metadata").
				FilterMessageSnippet("add_docker_metadata").
				FilterMessageSnippet("add_kubernetes_metadata").
				Len() == 1
			require.True(t, processorsLoaded, "processors not loaded")
			// Check that add_host_metadata works, other processors are not guaranteed to add fields in all environments
			require.Contains(t, logs["r1"][0].Flatten(), "host.architecture")
		},
	})
}

func BenchmarkFactory(b *testing.B) {
	tmpDir := b.TempDir()

	cfg := &Config{
		Beatconfig: map[string]interface{}{
			"filebeat": map[string]interface{}{
				"inputs": []map[string]interface{}{
					{
						"type":    "benchmark",
						"enabled": true,
						"message": "test",
						"count":   10,
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

	var zapLogs bytes.Buffer
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.Lock(zapcore.AddSync(&zapLogs)),
		zapcore.DebugLevel)

	receiverSettings := receiver.Settings{}
	receiverSettings.Logger = zap.New(core)

	factory := NewFactory()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := factory.CreateLogs(context.Background(), receiverSettings, cfg, nil)
		require.NoError(b, err)
	}
}

func TestMultipleReceivers(t *testing.T) {
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

	factory := NewFactory()
	oteltest.CheckReceivers(oteltest.CheckReceiversParams{
		T: t,
		Receivers: []oteltest.ReceiverConfig{
			{
				Name:    "r1",
				Config:  &config,
				Factory: factory,
			},
			{
				Name:    "r2",
				Config:  &config,
				Factory: factory,
			},
		},
		AssertFunc: func(t *assert.CollectT, logs map[string][]mapstr.M, zapLogs *observer.ObservedLogs) {
			_ = zapLogs
			require.Len(t, logs["r1"], 1)
			require.Len(t, logs["r2"], 1)
		},
	})
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

	os.Setenv("OTELCONSUMER_RECEIVERTEST", "true")

	cfg := &Config{
		Beatconfig: map[string]interface{}{
			"queue.mem.flush.timeout": "0s",
			"filebeat": map[string]interface{}{
				"inputs": []map[string]interface{}{
					{
						"type":    "filestream",
						"id":      "filestream-test",
						"enabled": true,
						"paths": []string{
							inputFile,
						},
						"file_identity.native": map[string]interface{}{},
						"prospector": map[string]interface{}{
							"scanner": map[string]interface{}{
								"fingerprint.enabled": false,
								"check_interval":      "0.1s",
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
			"path.logs": tmpDir,
		},
	}

	// Run the contract checker. This will trigger test failures if any problems are found.
	receivertest.CheckConsumeContract(receivertest.CheckConsumeContractParams{
		T:             t,
		Factory:       NewFactory(),
		Signal:        pipeline.SignalLogs,
		Config:        cfg,
		Generator:     gen,
		GenerateCount: logsPerTest,
	})
}
