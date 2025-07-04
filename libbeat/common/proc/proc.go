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

//go:build !windows

package proc

import (
	"os"
	"syscall"
)

// Job is noop on Unix
type Job int

// JobObject is a global instance of Job. noop on Unix
var JobObject Job

// StopCmd sends SIGINT to the process
func StopCmd(p *os.Process) error {
	return p.Signal(syscall.SIGINT)
}

// CreateJobObject returns a job object.
func CreateJobObject() (pj Job, err error) {
	return pj, err
}

// NewJob is noop on unix
func NewJob() (Job, error) {
	return 0, nil
}

// Close is noop on unix
func (job Job) Close() error {
	return nil
}

// Assign is noop on unix
func (job Job) Assign(p *os.Process) error {
	return nil
}
