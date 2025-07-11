// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

//go:build darwin || freebsd || linux || windows || aix

package process

import (
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/elastic/beats/v7/metricbeat/mb"
	mbtest "github.com/elastic/beats/v7/metricbeat/mb/testing"
	_ "github.com/elastic/beats/v7/metricbeat/module/system"
	"github.com/elastic/elastic-agent-system-metrics/metric/system/process"
)

func TestFetch(t *testing.T) {

	f := mbtest.NewReportingMetricSetV2Error(t, getConfig())
	events, errs := mbtest.ReportingFetchV2Error(f)
	for _, err := range errs {
		assert.ErrorIsf(t, err, process.NonFatalErr{}, "Expected non-fatal error, got %v", err)
	}
	assert.NotEmpty(t, events)

	time.Sleep(2 * time.Second)

	events, errs = mbtest.ReportingFetchV2Error(f)
	for _, err := range errs {
		assert.ErrorIsf(t, err, process.NonFatalErr{}, "Expected non-fatal error, got %v", err)
	}
	assert.NotEmpty(t, events)

	t.Logf("fetched %d events, showing events[0]:", len(events))
	t.Logf("%s/%s event: %+v", f.Module().Name(), f.Name(),
		events[0].BeatEvent("system", "process").Fields.StringToPrint())
}

func TestFetchDegradeOnPartial(t *testing.T) {
	if runtime.GOOS == "windows" && os.Getenv("CI") == "true" {
		t.Skip("Skip: CI run on windows. It is run as admin, but the test requires to run as non-admin")
	}
	if runtime.GOOS != "windows" && os.Getuid() == 0 {
		t.Skip("Skip: running as root on non-windows, but the test requires to run as non-root")
	}

	config := getConfig()
	config["degrade_on_partial"] = true

	f := mbtest.NewReportingMetricSetV2Error(t, config)

	var errs []error
	_, errs = mbtest.ReportingFetchV2Error(f)
	assert.NotEmpty(t, errs, "expected at least one error, got none")

	for _, err := range errs {
		assert.NotErrorIsf(t, err, &mb.PartialMetricsError{},
			"Expected non-fatal error, got %v", err)
	}
}

func TestFetchSinglePid(t *testing.T) {

	cfg := getConfig()
	cfg["process.pid"] = os.Getpid()

	f := mbtest.NewReportingMetricSetV2Error(t, cfg)
	events, errs := mbtest.ReportingFetchV2Error(f)
	assert.Empty(t, errs)
	assert.NotEmpty(t, events)
	assert.Equal(t, os.Getpid(), events[0].RootFields["process"].(map[string]interface{})["pid"])
	assert.NotEmpty(t, events[0].MetricSetFields["cpu"])
}

func TestData(t *testing.T) {
	f := mbtest.NewReportingMetricSetV2Error(t, getConfig())

	// Do a first fetch to have percentages
	mbtest.ReportingFetchV2Error(f)
	time.Sleep(10 * time.Second)

	err := mbtest.WriteEventsReporterV2Error(f, t, ".")
	if err != nil {
		t.Fatal("write", err)
	}
}

func getConfig() map[string]interface{} {
	return map[string]interface{}{
		"module":                        "system",
		"metricsets":                    []string{"process"},
		"processes":                     []string{".*"}, // in case we want a prettier looking example for data.json
		"process.cgroups.enabled":       true,
		"process.include_cpu_ticks":     true,
		"process.cmdline.cache.enabled": true,
		"process.include_top_n":         process.IncludeTopConfig{Enabled: true, ByCPU: 5},
	}
}
