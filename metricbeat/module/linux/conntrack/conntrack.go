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
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/prometheus/procfs"

	"github.com/elastic/beats/v7/libbeat/common/cfgwarn"
	"github.com/elastic/beats/v7/metricbeat/mb"
	"github.com/elastic/elastic-agent-libs/mapstr"
	"github.com/elastic/elastic-agent-system-metrics/metric/system/resolve"
)

// init registers the MetricSet with the central registry as soon as the program
// starts. The New function will be called later to instantiate an instance of
// the MetricSet for each host defined in the module's configuration. After the
// MetricSet has been created then Fetch will begin to be called periodically.
func init() {
	mb.Registry.MustAddMetricSet("linux", "conntrack", New)
}

type fetchFunc func() ([]procfs.ConntrackStatEntry, error)

// MetricSet holds any configuration or state information. It must implement
// the mb.MetricSet interface. And this is best achieved by embedding
// mb.BaseMetricSet because it implements all of the required mb.MetricSet
// interface methods except for Fetch.
type MetricSet struct {
	mb.BaseMetricSet
	mod       resolve.Resolver
	fetchFunc fetchFunc
}

// New creates a new instance of the MetricSet. New is responsible for unpacking
// any MetricSet specific configuration options if there are any.
func New(base mb.BaseMetricSet) (mb.MetricSet, error) {
	base.Logger().Warn(cfgwarn.Beta("The linux conntrack metricset is beta."))

	sys, ok := base.Module().(resolve.Resolver)
	if !ok {
		return nil, fmt.Errorf("unexpected module type: %T", base.Module())
	}

	var fetchFunc fetchFunc
	if _, err := exec.LookPath("conntrack"); err == nil {
		base.Logger().Info("Using conntrack(8) to fetch conntrack metrics")
		fetchFunc = fetchConntrackCLIMetrics
	} else {
		base.Logger().Info("Using procfs to fetch conntrack metrics")
		fetchFunc = fetchProcFSConntrackMetrics
	}

	return &MetricSet{
		BaseMetricSet: base,
		mod:           sys,
		fetchFunc:     fetchFunc,
	}, nil
}

// Fetch methods implements the data gathering and data conversion to the right
// format. It publishes the event which is then forwarded to the output. In case
// of an error set the Error field of mb.Event or simply call report.Error().
func (m *MetricSet) Fetch(report mb.ReporterV2) error {
	conntrackStats, err := m.fetchFunc()
	if err != nil {
		return fmt.Errorf("error fetching conntrack stats: %w", err)
	}

	summedEvents := procfs.ConntrackStatEntry{}
	for _, conn := range conntrackStats {
		summedEvents.Entries += conn.Entries
		summedEvents.Found += conn.Found
		summedEvents.Invalid += conn.Invalid
		summedEvents.Ignore += conn.Ignore
		summedEvents.InsertFailed += conn.InsertFailed
		summedEvents.Drop += conn.Drop
		summedEvents.EarlyDrop += conn.EarlyDrop
		summedEvents.SearchRestart += conn.SearchRestart
	}

	report.Event(mb.Event{
		MetricSetFields: mapstr.M{
			"summary": mapstr.M{
				"entries":        summedEvents.Entries,
				"found":          summedEvents.Found,
				"invalid":        summedEvents.Invalid,
				"ignore":         summedEvents.Ignore,
				"insert_failed":  summedEvents.InsertFailed,
				"drop":           summedEvents.Drop,
				"early_drop":     summedEvents.EarlyDrop,
				"search_restart": summedEvents.SearchRestart,
			},
		},
	})

	return nil
}

func fetchProcFSConntrackMetrics() ([]procfs.ConntrackStatEntry, error) {
	newFS, err := procfs.NewFS("/proc")
	if err != nil {
		return nil, fmt.Errorf("error creating new Host FS at /proc: %w", err)
	}
	conntrackStats, err := newFS.ConntrackStat()
	if err != nil {
		if os.IsNotExist(err) {
			err = mb.PartialMetricsError{Err: fmt.Errorf("nf_conntrack kernel module not loaded: %w", err)}
		}
		return nil, err
	}
	return conntrackStats, nil
}

func fetchConntrackCLIMetrics() ([]procfs.ConntrackStatEntry, error) {
	cmd := exec.Command("conntrack", "--stats")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	out, err := cmd.Output()
	outStr := strings.TrimSpace(string(out))
	if err != nil {
		if strings.HasPrefix(outStr, "cpu=") {
			// Without sudo, conntrack falls back to procfs but logs a permission error and exit code 1.
			// It reports valid stats to stdout.
		} else {
			return nil, fmt.Errorf("error invoking 'conntrack --stats': %w: %s", err, stderr.String())
		}
	}

	stats, err := parseConntrackCLIStats(outStr)
	if err != nil {
		return nil, fmt.Errorf("error parsing conntrack stats: %w", err)
	}

	// compute total entries
	if len(stats) > 0 && stats[0].Entries == 0 {
		cmd := exec.Command("conntrack", "--count")
		out, err = cmd.Output()
		outStr := strings.TrimSpace(string(out))
		if err != nil {
			if outStr == "" {
				return nil, fmt.Errorf("error invoking 'conntrack --count': %w: %s", err, stderr.String())
			}
		}

		count, err := strconv.ParseUint(outStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing conntrack entries count: %w", err)
		}
		stats[0].Entries = count
	}

	return stats, nil
}

func parseConntrackCLIStats(data string) ([]procfs.ConntrackStatEntry, error) {
	lines := strings.Split(data, "\n")
	statsEntries := make([]procfs.ConntrackStatEntry, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "cpu=") {
			continue
		}

		fields := strings.Fields(line)
		entry := procfs.ConntrackStatEntry{}

		for _, field := range fields {
			kv := strings.SplitN(field, "=", 2)
			if len(kv) != 2 {
				continue
			}
			key, valStr := kv[0], kv[1]
			val, err := strconv.ParseUint(valStr, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid integer for %s: %w", key, err)
			}

			switch key {
			case "entries":
				entry.Entries = val
			case "found":
				entry.Found = val
			case "invalid":
				entry.Invalid = val
			case "ignore":
				entry.Ignore = val
			case "insert_failed":
				entry.InsertFailed = val
			case "drop":
				entry.Drop = val
			case "early_drop":
				entry.EarlyDrop = val
			case "search_restart":
				entry.SearchRestart = val
			}
		}

		statsEntries = append(statsEntries, entry)
	}

	return statsEntries, nil
}
