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

package beater

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/elastic/elastic-agent-libs/monitoring"
)

func createTestCounter() *eventCounter {
	reg := monitoring.NewRegistry()
	return &eventCounter{
		added: monitoring.NewUint(reg, "added"),
		done:  monitoring.NewUint(reg, "done"),
		count: monitoring.NewInt(reg, "count"),
	}
}

func TestEventCounter(t *testing.T) {
	t.Run("add", func(t *testing.T) {
		counter := createTestCounter()
		counter.Add(5)

		assert.Equal(t, uint64(5), counter.added.Get())
		assert.Equal(t, int64(5), counter.count.Get())
	})

	t.Run("done", func(t *testing.T) {
		counter := createTestCounter()
		counter.Add(3)
		initialAdded := counter.added.Get()
		initialCount := counter.count.Get()

		for range 3 {
			counter.Done()
		}

		counter.Wait()

		assert.Equal(t, initialAdded, counter.added.Get()) // added should not change
		assert.Equal(t, uint64(3), counter.done.Get())
		assert.Equal(t, initialCount-3, counter.count.Get())
	})

	t.Run("wait", func(t *testing.T) {
		counter := createTestCounter()
		counter.Add(1)

		doneCh := make(chan struct{})
		go func() {
			counter.Wait()
			close(doneCh)
		}()
		select {
		case <-doneCh:
			t.Fatal("Wait returned before Done was called")
		case <-time.After(100 * time.Millisecond):
			// expected timeout
		}

		counter.Done()

		select {
		case <-doneCh:
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Wait did not return after Done was called")
		}
	})

	t.Run("concurrent", func(t *testing.T) {
		counter := createTestCounter()
		var wg sync.WaitGroup

		numGoroutines := 10
		incrementsPerGoroutine := 1000

		// Add
		for range numGoroutines {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for range incrementsPerGoroutine {
					counter.Add(1)
				}
			}()
		}

		wg.Wait()

		// Done
		for range numGoroutines {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for range incrementsPerGoroutine {
					counter.Done()
				}
			}()
		}

		wg.Wait()
		counter.Wait()

		assert.Equal(t, uint64(numGoroutines*incrementsPerGoroutine), counter.added.Get())
		assert.Equal(t, uint64(numGoroutines*incrementsPerGoroutine), counter.done.Get())
		assert.Equal(t, int64(0), counter.count.Get())
	})
}
