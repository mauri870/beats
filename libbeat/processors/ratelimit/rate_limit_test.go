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

package ratelimit

import (
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/require"

	"github.com/elastic/beats/v7/libbeat/beat"
	conf "github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/elastic/elastic-agent-libs/logp/logptest"
	"github.com/elastic/elastic-agent-libs/mapstr"
)

func TestNew(t *testing.T) {
	cases := map[string]struct {
		config mapstr.M
		err    string
	}{
		"default": {
			mapstr.M{},
			"",
		},
		"unknown_algo": {
			mapstr.M{
				"algorithm": mapstr.M{
					"foobar": mapstr.M{},
				},
			},
			"rate limiting algorithm 'foobar' not implemented",
		},
	}

	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			config := conf.MustNewConfigFrom(test.config)
			_, err := new(config, logptest.NewTestingLogger(t, ""))
			if test.err == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err, test.err)
			}
		})
	}
}

func TestRateLimit(t *testing.T) {
	var inEvents []beat.Event
	for i := 1; i <= 6; i++ {
		event := beat.Event{
			Timestamp: time.Now(),
			Fields: mapstr.M{
				"event_number": i,
			},
		}
		inEvents = append(inEvents, event)
	}

	withField := func(in beat.Event, key string, value interface{}) beat.Event {
		out := in
		out.Fields = in.Fields.Clone()

		out.Fields.Put(key, value)
		return out
	}

	cases := map[string]struct {
		config    mapstr.M
		inEvents  []beat.Event
		delay     time.Duration
		outEvents []beat.Event
	}{
		"rate_0": {
			config:    mapstr.M{},
			inEvents:  inEvents,
			outEvents: []beat.Event{},
		},
		"rate_1_per_min": {
			config: mapstr.M{
				"limit": "1/m",
			},
			inEvents:  inEvents,
			outEvents: inEvents[0:1],
		},
		"rate_2_per_min": {
			config: mapstr.M{
				"limit": "2/m",
			},
			inEvents:  inEvents,
			outEvents: inEvents[0:2],
		},
		"rate_6_per_min": {
			config: mapstr.M{
				"limit": "6/m",
			},
			inEvents:  inEvents,
			outEvents: inEvents,
		},
		"rate_2_per_sec": {
			config: mapstr.M{
				"limit": "2/s",
			},
			delay:     200 * time.Millisecond,
			inEvents:  inEvents,
			outEvents: []beat.Event{inEvents[0], inEvents[1], inEvents[3], inEvents[5]},
		},
		"with_fields": {
			config: mapstr.M{
				"limit":  "1/s",
				"fields": []string{"foo"},
			},
			delay: 400 * time.Millisecond,
			inEvents: []beat.Event{
				withField(inEvents[0], "foo", "bar"),
				withField(inEvents[1], "foo", "bar"),
				inEvents[2],
				withField(inEvents[3], "foo", "seger"),
			},
			outEvents: []beat.Event{
				withField(inEvents[0], "foo", "bar"),
				inEvents[2],
				withField(inEvents[3], "foo", "seger"),
			},
		},
		"with_burst": {
			config: mapstr.M{
				"limit":            "2/s",
				"burst_multiplier": 2,
			},
			delay:     400 * time.Millisecond,
			inEvents:  inEvents,
			outEvents: inEvents,
		},
	}

	for name, test := range cases {
		t.Run(name, func(t *testing.T) {
			p, err := new(conf.MustNewConfigFrom(test.config), logptest.NewTestingLogger(t, ""))
			require.NoError(t, err)

			fakeClock := clockwork.NewFakeClock()

			p.(*rateLimit).setClock(fakeClock)

			out := make([]beat.Event, 0)
			for _, in := range test.inEvents {
				inCopy := in
				inCopy.Fields = in.Fields.Clone()

				o, err := p.Run(&inCopy)
				require.NoError(t, err)
				if o != nil {
					out = append(out, *o)
				}
				fakeClock.Advance(test.delay)
			}

			require.Equal(t, test.outEvents, out)
		})
	}
}

func TestAllocs(t *testing.T) {
	p, err := new(conf.MustNewConfigFrom(mapstr.M{
		"limit": "100/s",
	}), logp.NewNopLogger())
	require.NoError(t, err)
	event := beat.Event{Fields: mapstr.M{"field": 1}}

	allocs := testing.AllocsPerRun(1000, func() {
		p.Run(&event) //nolint:errcheck // ignore
	})
	if allocs > 0 {
		t.Errorf("allocs = %v; want 0", allocs)
	}
}

func BenchmarkRateLimit(b *testing.B) {
	p, err := new(conf.MustNewConfigFrom(mapstr.M{
		"limit": "100/s",
	}), logp.NewNopLogger())
	require.NoError(b, err)
	event := beat.Event{Fields: mapstr.M{"field": 1}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.Run(&event) //nolint:errcheck // ignore
	}
}
