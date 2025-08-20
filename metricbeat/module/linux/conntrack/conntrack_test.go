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

package conntrack

import (
	"maps"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	mbtest "github.com/elastic/beats/v7/metricbeat/mb/testing"
	_ "github.com/elastic/beats/v7/metricbeat/module/linux"
	"github.com/elastic/elastic-agent-libs/mapstr"
)

func TestData(t *testing.T) {
	f := mbtest.NewReportingMetricSetV2Error(t, getConfig())
	err := mbtest.WriteEventsReporterV2Error(f, t, ".")
	if err != nil {
		t.Fatal("write", err)
	}
}

func TestFetch(t *testing.T) {
	f := mbtest.NewReportingMetricSetV2Error(t, getConfig())
	events, errs := mbtest.ReportingFetchV2Error(f)

	assert.Empty(t, errs)
	if !assert.NotEmpty(t, events) {
		t.FailNow()
	}

	testConn := mapstr.M{
		"drop":           uint64(0),
		"early_drop":     uint64(0),
		"entries":        uint64(16),
		"found":          uint64(0),
		"ignore":         uint64(3271028),
		"insert_failed":  uint64(0),
		"invalid":        uint64(122),
		"search_restart": uint64(3),
	}

	rawEvent := events[0].BeatEvent("linux", "conntrack").Fields["linux"].(mapstr.M)["conntrack"].(mapstr.M)["summary"]

	assert.Equal(t, testConn, rawEvent)
}

func BenchmarkFetch(b *testing.B) {
	f := mbtest.NewReportingMetricSetV2Error(b, getConfig())
	for range b.N {
		_, errs := mbtest.ReportingFetchV2Error(f)
		require.Empty(b, errs, "fetch should not return an error")
	}
}

func BenchmarkFetchNetlink(b *testing.B) {
	cfg := getConfig()
	cfg["hostfs"] = b.TempDir()
	// TODO: In order to run this benchmark it needs CAP_NET_ADMIN
	f := mbtest.NewReportingMetricSetV2Error(b, cfg)
	for range b.N {
		_, errs := mbtest.ReportingFetchV2Error(f)
		require.Empty(b, errs, "fetch should not return an error")
	}
}

func TestFetchNetlink(t *testing.T) {
	// hide /proc/net/stat/nf_conntrack file so it uses netlink
	cfg := getConfig()
	cfg["hostfs"] = t.TempDir()

	// TODO: In order to run this test it needs CAP_NET_ADMIN
	f := mbtest.NewReportingMetricSetV2Error(t, cfg)
	events, errs := mbtest.ReportingFetchV2Error(f)
	require.Empty(t, errs, "fetch should not return an error")

	require.NotEmpty(t, events)
	rawEvent := events[0].BeatEvent("linux", "conntrack").Fields["linux"].(mapstr.M)["conntrack"].(mapstr.M)["summary"].(mapstr.M)
	keys := maps.Keys(rawEvent)
	assert.Contains(t, keys, "drop")
	assert.Contains(t, keys, "early_drop")
	assert.Contains(t, keys, "entries")
	assert.Contains(t, keys, "found")
	assert.Contains(t, keys, "ignore")
	assert.Contains(t, keys, "insert_failed")
	assert.Contains(t, keys, "invalid")
	assert.Contains(t, keys, "search_restart")
}

func getConfig() map[string]interface{} {
	return map[string]interface{}{
		"module":     "linux",
		"metricsets": []string{"conntrack"},
		"hostfs":     "./_meta/testdata",
	}
}
