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
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elastic/beats/v7/metricbeat/mb"
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
		mbtest.ReportingFetchV2Error(f)
	}
}

func TestFetchConntrackModuleNotLoaded(t *testing.T) {
	// Create a temporary directory to simulate a missing /proc/net/stat/nf_conntrack file
	tmpDir := t.TempDir()
	assert.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "proc/net/stat"), 0755))
	c := getConfig()
	c["hostfs"] = tmpDir

	f := mbtest.NewReportingMetricSetV2Error(t, c)
	events, errs := mbtest.ReportingFetchV2Error(f)

	require.Len(t, errs, 1)
	err := errors.Join(errs...)
	assert.ErrorAs(t, err, &mb.PartialMetricsError{})
	assert.Contains(t, err.Error(), "error fetching conntrack stats: nf_conntrack kernel module not loaded")
	require.Empty(t, events)
}

func getConfig() map[string]interface{} {
	return map[string]interface{}{
		"module":     "linux",
		"metricsets": []string{"conntrack"},
		"hostfs":     "./_meta/testdata",
	}
}

func TestParseConntrackCLIStats(t *testing.T) {
	// 'conntrack --stats' running as a normal user should fallback to obsolete procfs fields
	t.Run("user", func(t *testing.T) {
		in := `cpu=0           entries=42 clashres=0 found=0 new=0 invalid=0 ignore=3271028 delete=0 chainlength=0 insert=0 insert_failed=0 drop=0 early_drop=0 icmp_error=0 expect_new=0 expect_create=0 expect_delete=0 search_restart=0
cpu=1           entries=109 clashres=0 found=0 new=0 invalid=122 ignore=0 delete=0 chainlength=0 insert=0 insert_failed=0 drop=0 early_drop=0 icmp_error=0 expect_new=0 expect_create=0 expect_delete=0 search_restart=3
	`

		stats, err := parseConntrackCLIStats(in)
		require.NoError(t, err)
		assert.Len(t, stats, 2)

		// CPU 0
		assert.Equal(t, uint64(0), stats[0].Drop)
		assert.Equal(t, uint64(0), stats[0].EarlyDrop)
		assert.Equal(t, uint64(42), stats[0].Entries)
		assert.Equal(t, uint64(0), stats[0].Found)
		assert.Equal(t, uint64(3271028), stats[0].Ignore)
		assert.Equal(t, uint64(0), stats[0].InsertFailed)
		assert.Equal(t, uint64(0), stats[0].Invalid)
		assert.Equal(t, uint64(0), stats[0].SearchRestart)

		// CPU 1
		assert.Equal(t, uint64(0), stats[1].Drop)
		assert.Equal(t, uint64(0), stats[1].EarlyDrop)
		assert.Equal(t, uint64(109), stats[1].Entries)
		assert.Equal(t, uint64(0), stats[1].Found)
		assert.Equal(t, uint64(0), stats[1].Ignore)
		assert.Equal(t, uint64(0), stats[1].InsertFailed)
		assert.Equal(t, uint64(122), stats[1].Invalid)
		assert.Equal(t, uint64(3), stats[1].SearchRestart)
	})

	// 'conntrack --stats' running as root uses netfilter fields
	t.Run("root", func(t *testing.T) {
		in := `cpu=0           found=0 invalid=109 insert=0 insert_failed=0 drop=0 early_drop=0 error=0 search_restart=0 clash_resolve=0 chaintoolong=0
cpu=1           found=0 invalid=109 insert=0 insert_failed=0 drop=0 early_drop=0 error=0 search_restart=0 clash_resolve=0 chaintoolong=0
	`

		stats, err := parseConntrackCLIStats(in)
		require.NoError(t, err)
		assert.Len(t, stats, 2)

		// CPU 0
		assert.Equal(t, uint64(0), stats[0].Drop)
		assert.Equal(t, uint64(0), stats[0].EarlyDrop)
		assert.Equal(t, uint64(0), stats[0].Entries)
		assert.Equal(t, uint64(0), stats[0].Found)
		assert.Equal(t, uint64(0), stats[0].Ignore)
		assert.Equal(t, uint64(0), stats[0].InsertFailed)
		assert.Equal(t, uint64(109), stats[0].Invalid)
		assert.Equal(t, uint64(0), stats[0].SearchRestart)

		// CPU 1
		assert.Equal(t, uint64(0), stats[1].Drop)
		assert.Equal(t, uint64(0), stats[1].EarlyDrop)
		assert.Equal(t, uint64(0), stats[1].Entries)
		assert.Equal(t, uint64(0), stats[1].Found)
		assert.Equal(t, uint64(0), stats[1].Ignore)
		assert.Equal(t, uint64(0), stats[1].InsertFailed)
		assert.Equal(t, uint64(109), stats[1].Invalid)
		assert.Equal(t, uint64(0), stats[1].SearchRestart)
	})
}
