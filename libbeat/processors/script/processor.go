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

package script

import (
	"fmt"
	"strings"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/processors"
	"github.com/elastic/beats/v7/libbeat/processors/script/javascript"
	"github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"

	// Register javascript modules with the processor.
	_ "github.com/elastic/beats/v7/libbeat/processors/script/javascript/module"
)

func init() {
	processors.RegisterPlugin("script", New)
}

// New constructs a new script processor.
func New(c *config.C, log *logp.Logger) (beat.Processor, error) {
	var config = struct {
		Lang string `config:"lang" validate:"required"`
	}{}
	if err := c.Unpack(&config); err != nil {
		return nil, err
	}

	switch strings.ToLower(config.Lang) {
	case "javascript", "js":
		return javascript.New(c, log)
	default:
		return nil, fmt.Errorf("script type must be declared (e.g. type: javascript)")
	}
}
