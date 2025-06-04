package oteltest

import (
	"testing"
	"time"

	"github.com/elastic/beats/v7/x-pack/filebeat/fbreceiver"
	"github.com/elastic/beats/v7/x-pack/metricbeat/mbreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/testbed/testbed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/connector"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/debugexporter"
	"go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/otelcol"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/multierr"
	"gopkg.in/yaml.v3"
)

func Components() (otelcol.Factories, error) {
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

	return factories, nil
}

func TestNewInProcessPipeline(t *testing.T) {
	factories, err := Components()
	assert.NoError(t, err)
	runner := testbed.NewInProcessCollector(factories)

	receivers := map[string]any{
		"filebeatreceiver": map[string]any{
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

	cfg := map[string]any{
		"receivers": receivers,
		"exporters": map[string]any{
			"debug": map[string]any{
				"verbosity":        "detailed",
				"sampling_initial": 100,
			},
		},
		"processors": map[string]any{},
		"service": map[string]any{
			"pipelines": map[string]any{
				"logs": map[string]any{
					"receivers":  []string{"filebeatreceiver"},
					"processors": []string{},
					"exporters":  []string{"debug"},
				},
			},
		},
	}

	ymlCfg, err := yaml.Marshal(cfg)
	require.NoError(t, err)
	t.Logf("YAML Config:\n%s", string(ymlCfg))
	configCleanup, cfgErr := runner.PrepareConfig(t, string(ymlCfg))
	defer configCleanup()
	assert.NoError(t, cfgErr)
	assert.NotNil(t, configCleanup)
	args := testbed.StartParams{Name: "test"}
	defer func() {
		time.Sleep(20 * time.Second) // Change this to wait for logs
		_, err := runner.Stop()
		require.NoError(t, err)
	}()
	assert.NoError(t, runner.Start(args))
}
