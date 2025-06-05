package oteltest

import (
	"fmt"
	"testing"
	"time"

	"github.com/elastic/beats/v7/x-pack/filebeat/fbreceiver"
	"github.com/elastic/beats/v7/x-pack/metricbeat/mbreceiver"
	"github.com/elastic/elastic-agent-libs/mapstr"
	"github.com/open-telemetry/opentelemetry-collector-contrib/testbed/testbed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/connector"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/debugexporter"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
	"go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/otelcol"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/multierr"
	"gopkg.in/yaml.v3"
)

// components returns the factories that the test otel collector will support.
func components() (otelcol.Factories, error) {
	var errs error

	extensions, err := otelcol.MakeFactoryMap[extension.Factory]()
	errs = multierr.Append(errs, err)

	receivers, err := otelcol.MakeFactoryMap[receiver.Factory](
		fbreceiver.NewFactory(),
		mbreceiver.NewFactory(),
	)
	errs = multierr.Append(errs, err)

	exporters, err := otelcol.MakeFactoryMap[exporter.Factory](
		debugexporter.NewFactory(),
		otlpexporter.NewFactory(),
	)
	errs = multierr.Append(errs, err)

	processors, err := otelcol.MakeFactoryMap[processor.Factory]()
	errs = multierr.Append(errs, err)

	connectors, err := otelcol.MakeFactoryMap[connector.Factory]()
	errs = multierr.Append(errs, err)

	factories := otelcol.Factories{
		Extensions: extensions,
		Receivers:  receivers,
		Processors: processors,
		Exporters:  exporters,
		Connectors: connectors,
	}

	return factories, errs
}

func TestFBReceiverTestbedCollector(t *testing.T) {
	name := "filebeatreceiver/1"
	cfg := map[string]any{
		name: map[string]any{
			"filebeat": map[string]any{
				"inputs": []map[string]any{
					{
						"type":    "benchmark",
						"enabled": true,
						"message": "abc123",
						"eps":     1,
					},
				},
			},
			"output": map[string]any{
				"otelconsumer": map[string]any{},
			},
			"logging": map[string]any{
				"level": "debug",
				"selectors": []string{
					"*",
				},
			},
			"path.home": t.TempDir(),
		},
	}
	assertReceiverTestdbedCollector(t, name, cfg)
}

func TestMBReceiverTestbedCollector(t *testing.T) {
	name := "metricbeatreceiver/1"
	cfg := map[string]any{
		name: map[string]any{
			"metricbeat": map[string]any{
				"modules": []map[string]any{
					{
						"module":     "system",
						"enabled":    true,
						"period":     "1s",
						"processes":  []string{".*"},
						"metricsets": []string{"cpu"},
					},
				},
			},
			"output": map[string]any{
				"otelconsumer": map[string]any{},
			},
			"logging": map[string]any{
				"level": "debug",
				"selectors": []string{
					"*",
				},
			},
		},
	}
	assertReceiverTestdbedCollector(t, name, cfg)
}

func assertReceiverTestdbedCollector(t *testing.T, name string, cfg mapstr.M) {
	factories, err := components()
	assert.NoError(t, err)
	runner := testbed.NewInProcessCollector(factories)

	// TODO: Provider and sender don't really matter for this test, as the beats receiver will produce the data.
	// I couldn't figure out out how to use a dummy provider and sender, so the test ends up failing because it expects data to be delivered.
	options := testbed.LoadOptions{
		DataItemsPerSecond: 1_000,
		Parallel:           1,
		ItemsPerBatch:      100,
	}
	provider := testbed.NewPerfTestDataProvider(options)
	sender := testbed.NewOTLPMetricDataSender(testbed.DefaultHost, 0)

	// the otlp exporter will receive the logs so we can make assertions
	// Port here does not matter as the data is send by otelconsumer through the collector pipeline.
	exporter := testbed.NewOTLPDataReceiver(60001)

	rawCfg := `receivers:
  %s
exporters:%v
  debug:
    verbosity: detailed
    sampling_initial: 100
processors: {}
service:
  telemetry:
    logs:
      level: "DEBUG"
      development: true
      encoding: "json"
  pipelines:
    logs:
      receivers: [%s]
      processors: []
      exporters: [%v, debug]
`

	receiversYAML, err := yaml.Marshal(cfg)
	require.NoError(t, err)
	yamlCfg := fmt.Sprintf(rawCfg, receiversYAML, exporter.GenConfigYAMLStr(), name, exporter.ProtocolName())
	t.Logf("YAML Config:\n%s", yamlCfg)
	configCleanup, cfgErr := runner.PrepareConfig(t, yamlCfg)
	defer configCleanup()
	assert.NoError(t, cfgErr)
	assert.NotNil(t, configCleanup)

	tc := testbed.NewTestCase(t,
		provider,
		sender,
		exporter,
		runner,
		&testbed.PerfTestValidator{},
		&testbed.CorrectnessResults{},
		// testbed.WithResourceLimits(testbed.ResourceSpec{ExpectedMaxCPU: 10, ExpectedMaxRAM: 120}),
	)
	defer tc.Stop()

	tc.EnableRecording()

	tc.StartBackend()
	tc.StartAgent()
	defer tc.StopAgent()

	// TODO: Needs investigation. This works, but the test is expecting the generated data to be delivered and it will never will, so it ends up failing the test.
	var logs []plog.Logs
	assert.Eventually(t, func() bool {
		logs = tc.MockBackend.GetReceivedLogs()
		t.Logf("Waiting for logs to be received, logs count: %d", len(logs))
		return len(logs) > 0
	}, 30*time.Second, 100*time.Millisecond, "should have received logs")

	assert.True(t, tc.AgentLogsContains("Generated new processors"), "should have logged global processors loaded")
}
