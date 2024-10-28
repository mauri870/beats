// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package fbreceiver

import (
	"bytes"
	"context"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/elasticsearchexporter"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/consumererror"
	"go.opentelemetry.io/collector/exporter/exportertest"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pipeline"
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

	r, err := createReceiver(context.Background(), receiverSettings, &config, logConsumer)
	assert.NoErrorf(t, err, "Error creating receiver. Logs:\n %s", zapLogs.String())
	err = r.Start(context.Background(), nil)
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

type exampleReceiver struct {
	nextConsumer consumer.Logs
}

func (s *exampleReceiver) Start(context.Context, component.Host) error {
	return nil
}

func (s *exampleReceiver) Shutdown(context.Context) error {
	return nil
}

func (s *exampleReceiver) Receive(data plog.Logs) {
	// This very simple implementation demonstrates how a single items receiving should happen.
	for {
		err := s.nextConsumer.ConsumeLogs(context.Background(), data)
		if err != nil {
			// The next consumer returned an error.
			if !consumererror.IsPermanent(err) {
				// It is not a permanent error, so we must retry sending it again. In network-based
				// receivers instead we can ask our sender to re-retry the same data again later.
				// We may also pause here a bit if we don't want to hammer the next consumer.
				continue
			}
		}
		// If we are hear either the ConsumeLogs returned success or it returned a permanent error.
		// In either case we don't need to retry the same data, we are done.
		return
	}
}

// A config for exampleReceiver.
type exampleReceiverConfig struct {
	generator *exampleLogGenerator
}

// A generator that can send data to exampleReceiver.
type exampleLogGenerator struct {
	t           *testing.T
	receiver    *exampleReceiver
	sequenceNum int64
}

func (g *exampleLogGenerator) Start() {
	g.sequenceNum = 0
}

func (g *exampleLogGenerator) Stop() {}

func (g *exampleLogGenerator) Generate() []receivertest.UniqueIDAttrVal {
	// Make sure the id is atomically incremented. Generate() may be called concurrently.
	id := receivertest.UniqueIDAttrVal(strconv.FormatInt(atomic.AddInt64(&g.sequenceNum, 1), 10))

	data := receivertest.CreateOneLogWithID(id)

	// Send the generated data to the receiver.
	g.receiver.Receive(data)

	// And return the ids for bookkeeping by the test.
	return []receivertest.UniqueIDAttrVal{id}
}

func newExampleFactory() receiver.Factory {
	return receiver.NewFactory(
		component.MustNewType("example_receiver"),
		func() component.Config {
			return &exampleReceiverConfig{}
		},
		receiver.WithLogs(createLog, component.StabilityLevelBeta),
	)
}

func createLog(
	_ context.Context,
	_ receiver.Settings,
	cfg component.Config,
	consumer consumer.Logs,
) (receiver.Logs, error) {
	rcv := &exampleReceiver{nextConsumer: consumer}
	cfg.(*exampleReceiverConfig).generator.receiver = rcv
	return rcv, nil
}

func TestConsumeContractReceiver(t *testing.T) {
	generator := &exampleLogGenerator{t: t}

	cfg := Config{
		Beatconfig: map[string]interface{}{
			"filebeat": map[string]interface{}{
				"inputs": []map[string]interface{}{
					{
						"type":    "benchmark",
						"enabled": true,
						"message": "test",
						"count":   100,
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

	receivertest.CheckConsumeContract(receivertest.CheckConsumeContractParams{
		T:         t,
		Factory:   NewFactory(),
		Signal:    pipeline.SignalLogs,
		Config:    &cfg,
		Generator: generator,
	})
}

func TestConsumeContract(t *testing.T) {
	cfg := Config{
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
			"path.home": t.TempDir(),
		},
	}

	exporterCfg := elasticsearchexporter.Config{}
	exporterCfg.Mapping.Mode = "bodymap"
	exporterCfg.Endpoint = "http://localhost:9200"

	exportertest.CheckConsumeContract(exportertest.CheckConsumeContractParams{
		T:                    t,
		NumberOfTestElements: 10,
		ExporterFactory:      elasticsearchexporter.NewFactory(),
		Signal:               pipeline.SignalLogs,
		ExporterConfig:       &exporterCfg,
		ReceiverFactory:      NewFactory(),
		ReceiverConfig:       &cfg,
	})
}
