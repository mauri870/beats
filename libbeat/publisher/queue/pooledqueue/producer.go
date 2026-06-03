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

package pooledqueue

import (
	"sync/atomic"

	"github.com/elastic/beats/v7/libbeat/publisher/queue"
)

// magazineSize (M) is the number of slot indices each producer pre-claims from
// the shared free channel in one bulk receive. The CPU layer's worst-case miss
// rate against the shared depot is bounded by 1/M regardless of workload
// (Bonwick & Adams §3.4). M=16 keeps the miss rate ≤6% while bounding the
// number of slots that can sit idle in a producer that stalls.
//
// Reference: Bonwick & Adams, "Magazines and Vmem: Extending the Slab
// Allocator to Many CPUs and Arbitrary Resources", USENIX 2001, §3 and §3.4.
const magazineSize = 16

// producer publishes events into one Queue. It acquires a slot from the
// shared pool's free list (blocking when the pool is empty), writes the event
// into that slot, and threads the slot onto its queue's per-pipeline FIFO.
//
// Each producer holds a magazine: a small pre-claimed stack of slot indices
// (Bonwick & Adams §3.1). Publish pops from the magazine without touching the
// shared pool.free channel; the channel is only accessed to bulk-refill when
// the magazine runs dry. This mirrors the hot-path described in §3.1c of the
// paper: "if the CPU's loaded magazine isn't empty, pop the top object and
// return it."
type producer[T any] struct {
	queue    *Queue[T]
	cfg      queue.ProducerConfig
	magazine []int // pre-claimed slot indices; drained LIFO before touching pool.free

	nextID atomic.Uint64
	closed atomic.Bool
}

// Publish adds an entry, blocking if the pool is empty until a slot is freed
// or the queue/pool is closed.
func (p *producer[T]) Publish(entry T) (queue.EntryID, bool) {
	if p.closed.Load() {
		return 0, false
	}
	slotIdx, ok := p.acquireSlot()
	if !ok {
		return 0, false
	}
	return p.fill(entry, slotIdx)
}

// acquireSlot returns a slot index, refilling the magazine from pool.free when
// it is empty. Returns false if the queue or pool is closed.
//
// The refill strategy is a simplified depot interaction (Bonwick & Adams §3.1):
// first attempt a non-blocking bulk-fill of up to magazineSize slots; if the
// depot is empty, fall back to a single blocking receive. The two phases are
// kept separate so no blocking can occur inside the non-blocking loop.
func (p *producer[T]) acquireSlot() (int, bool) {
	if len(p.magazine) > 0 {
		idx := p.magazine[len(p.magazine)-1]
		p.magazine = p.magazine[:len(p.magazine)-1]
		return idx, true
	}

	// Phase 1: non-blocking bulk-fill from the shared free channel (the depot).
	// A labeled break exits the for loop from inside the select's default case.
	p.magazine = p.magazine[:cap(p.magazine)] // reuse backing array
	filled := 0
	pool := p.queue.pool
fill:
	for filled < magazineSize {
		select {
		case idx := <-pool.free:
			p.magazine[filled] = idx
			filled++
		default:
			break fill
		}
	}

	if filled > 0 {
		p.magazine = p.magazine[:filled]
		idx := p.magazine[filled-1]
		p.magazine = p.magazine[:filled-1]
		return idx, true
	}

	// Phase 2: depot was empty — block until a slot is available or we close.
	p.magazine = p.magazine[:0]
	select {
	case idx := <-pool.free:
		return idx, true
	case <-p.queue.closeCh:
		return 0, false
	case <-pool.closed:
		return 0, false
	}
}

// TryPublish adds an entry only if a slot is immediately available; it never
// blocks.
func (p *producer[T]) TryPublish(entry T) (queue.EntryID, bool) {
	if p.closed.Load() {
		return 0, false
	}
	// Check the local magazine first before touching the shared channel.
	if len(p.magazine) > 0 {
		idx := p.magazine[len(p.magazine)-1]
		p.magazine = p.magazine[:len(p.magazine)-1]
		return p.fill(entry, idx)
	}
	var slotIdx int
	select {
	case slotIdx = <-p.queue.pool.free:
	default:
		return 0, false
	}
	return p.fill(entry, slotIdx)
}

// Close marks the producer as closed and returns any pre-claimed magazine slots
// to the pool (the depot) so they are available to other producers.
// This is the equivalent of returning a partially-loaded magazine to the depot
// on CPU offline (Bonwick & Adams §3.5).
//
// The sends to pool.free are non-blocking: pool.free has capacity equal to the
// total slot count, and a producer can hold at most magazineSize slots, so the
// channel must have room. A full channel would mean more slots are outstanding
// than exist, which is a bug — panic immediately rather than deadlock.
func (p *producer[T]) Close() {
	p.closed.Store(true)
	for _, idx := range p.magazine {
		select {
		case p.queue.pool.free <- idx:
		default:
			panic("pooledqueue: pool.free full on producer Close — slot accounting invariant violated")
		}
	}
	p.magazine = p.magazine[:0]
}

// fill stores entry in the given slot and threads it onto the queue's FIFO.
// If the queue is already closing, the slot is returned to the pool and the
// publish fails.
func (p *producer[T]) fill(entry T, slotIdx int) (queue.EntryID, bool) {
	id := p.nextID.Add(1)
	pool := p.queue.pool
	s := &pool.storage[slotIdx]
	s.event = entry
	s.next = -1
	s.producer = p
	s.producerID = id

	q := p.queue
	q.mu.Lock()
	if q.closing {
		// Queue closed between acquire and fill. Return the slot.
		var zero T
		s.event = zero
		s.producer = nil
		q.mu.Unlock()
		pool.free <- slotIdx
		return 0, false
	}
	if q.tail == -1 {
		q.head = slotIdx
	} else {
		pool.storage[q.tail].next = slotIdx
	}
	q.tail = slotIdx
	q.count++
	q.mu.Unlock()

	pool.observer.AddEvent(0)
	q.signal()
	return queue.EntryID(id), true
}
