// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package cel

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/elastic/elastic-agent-libs/transport/httpcommon"
	"github.com/elastic/mito/lib"
)

const defaultMaxExecutions = 1000

// config is the top-level configuration for a cel input.
type config struct {
	// Interval is the period interval between runs of the input.
	Interval time.Duration `config:"interval" validate:"required"`

	// Program is the CEL program to be run for each polling.
	Program string `config:"program" validate:"required"`
	// MaxExecutions is the maximum number of times a single
	// periodic CEL execution loop may repeat due to a true
	// "want_more" field. If it is nil a sensible default is
	// used.
	MaxExecutions *int `config:"max_executions"`
	// Regexps is the set of regular expression to be made
	// available to the program.
	Regexps map[string]string `config:"regexp"`
	// XSDs is the set of XSD type hint definitions to be
	// made available for XML parsing.
	XSDs map[string]string `config:"xsd"`
	// State is the initial state to be provided to the
	// program. If it has a cursor field, that field will
	// be overwritten by any stored cursor, but will be
	// available if no stored cursor exists.
	State map[string]interface{} `config:"state"`
	// Redact is the debug log state redaction configuration.
	Redact *redact `config:"redact"`

	// AllowedEnvironment is the set of env vars made
	// visible to an executing CEL evaluation.
	AllowedEnvironment []string `config:"allowed_environment"`

	// Auth is the authentication config for connection to an HTTP
	// API endpoint.
	Auth authConfig `config:"auth"`

	// Resource is the configuration for establishing an
	// HTTP request or for locating a local resource.
	Resource *ResourceConfig `config:"resource" validate:"required"`

	// FailureDump configures failure dump behaviour.
	FailureDump *dumpConfig `config:"failure_dump"`

	// RecordCoverage indicates whether a program should
	// record and log execution coverage.
	RecordCoverage bool `config:"record_coverage"`
}

type redact struct {
	// Fields indicates which fields to apply redaction to prior
	// to logging.
	Fields []string `config:"fields"`
	// Delete indicates that fields should be completely deleted
	// before logging rather than redaction with a "*".
	Delete bool `config:"delete"`
}

// dumpConfig configures the CEL program to retain
// the full evaluation state using the cel.OptTrackState
// option. The state is written to a file in the path if
// the evaluation fails.
type dumpConfig struct {
	Enabled  *bool  `config:"enabled"`
	Filename string `config:"filename"`
}

func (t *dumpConfig) enabled() bool {
	return t != nil && (t.Enabled == nil || *t.Enabled)
}

func (c config) Validate() error {
	if c.Interval <= 0 {
		return errors.New("interval must be greater than 0")
	}
	if c.MaxExecutions != nil && *c.MaxExecutions <= 0 {
		return fmt.Errorf("invalid maximum number of executions: %d <= 0", *c.MaxExecutions)
	}
	_, err := regexpsFromConfig(c)
	if err != nil {
		return fmt.Errorf("failed to check regular expressions: %w", err)
	}
	// TODO: Consider just building the program here to avoid this wasted work.
	var patterns map[string]*regexp.Regexp
	if len(c.Regexps) != 0 {
		patterns = map[string]*regexp.Regexp{".": nil}
	}
	wantDump := c.FailureDump.enabled() && c.FailureDump.Filename != ""
	_, _, _, err = newProgram(context.Background(), c.Program, root, nil, &http.Client{}, lib.HTTPOptions{}, patterns, c.XSDs, logp.L().Named("input.cel"), nil, wantDump, false)
	if err != nil {
		return fmt.Errorf("failed to check program: %w", err)
	}
	return nil
}

func defaultConfig() config {
	maxExecutions := defaultMaxExecutions
	maxAttempts := 5
	waitMin := time.Second
	waitMax := time.Minute
	transport := httpcommon.DefaultHTTPTransportSettings()
	transport.Timeout = 30 * time.Second

	return config{
		MaxExecutions: &maxExecutions,
		Interval:      time.Minute,
		Resource: &ResourceConfig{
			Retry: retryConfig{
				MaxAttempts: &maxAttempts,
				WaitMin:     &waitMin,
				WaitMax:     &waitMax,
			},
			RedirectForwardHeaders: false,
			RedirectMaxRedirects:   10,
			Transport:              transport,
		},
	}
}

type retryConfig struct {
	MaxAttempts *int           `config:"max_attempts"`
	WaitMin     *time.Duration `config:"wait_min"`
	WaitMax     *time.Duration `config:"wait_max"`
}

func (c retryConfig) Validate() error {
	switch {
	case c.MaxAttempts != nil && *c.MaxAttempts <= 0:
		return errors.New("max_attempts must be greater than zero")
	case c.WaitMin != nil && *c.WaitMin <= 0:
		return errors.New("wait_min must be greater than zero")
	case c.WaitMax != nil && *c.WaitMax <= 0:
		return errors.New("wait_max must be greater than zero")
	}
	return nil
}

func (c retryConfig) getMaxAttempts() int {
	if c.MaxAttempts == nil {
		return 0
	}
	return *c.MaxAttempts
}

func (c retryConfig) getWaitMin() time.Duration {
	if c.WaitMin == nil {
		return 0
	}
	return *c.WaitMin
}

func (c retryConfig) getWaitMax() time.Duration {
	if c.WaitMax == nil {
		return 0
	}
	return *c.WaitMax
}

type rateLimitConfig struct {
	Limit *float64 `config:"limit"`
	Burst *int     `config:"burst"`
}

func (c rateLimitConfig) Validate() error {
	if c.Limit != nil && *c.Limit <= 0 {
		return errors.New("limit must be greater than zero")
	}
	if c.Limit == nil && c.Burst != nil && *c.Burst <= 0 {
		return errors.New("burst must be greater than zero if limit is not specified")
	}
	return nil
}

type keepAlive struct {
	Disable             *bool         `config:"disable"`
	MaxIdleConns        int           `config:"max_idle_connections"`
	MaxIdleConnsPerHost int           `config:"max_idle_connections_per_host"` // If zero, http.DefaultMaxIdleConnsPerHost is the value used by http.Transport.
	IdleConnTimeout     time.Duration `config:"idle_connection_timeout"`
}

func (c keepAlive) Validate() error {
	if c.Disable == nil || *c.Disable {
		return nil
	}
	if c.MaxIdleConns < 0 {
		return errors.New("max_idle_connections must not be negative")
	}
	if c.MaxIdleConnsPerHost < 0 {
		return errors.New("max_idle_connections_per_host must not be negative")
	}
	if c.IdleConnTimeout < 0 {
		return errors.New("idle_connection_timeout must not be negative")
	}
	return nil
}

func (c keepAlive) settings() httpcommon.WithKeepaliveSettings {
	return httpcommon.WithKeepaliveSettings{
		Disable:             c.Disable == nil || *c.Disable,
		MaxIdleConns:        c.MaxIdleConns,
		MaxIdleConnsPerHost: c.MaxIdleConnsPerHost,
		IdleConnTimeout:     c.IdleConnTimeout,
	}
}

type ResourceConfig struct {
	URL                    *urlConfig       `config:"url" validate:"required"`
	Headers                http.Header      `config:"headers"`
	Retry                  retryConfig      `config:"retry"`
	RedirectForwardHeaders bool             `config:"redirect.forward_headers"`
	RedirectHeadersBanList []string         `config:"redirect.headers_ban_list"`
	RedirectMaxRedirects   int              `config:"redirect.max_redirects"`
	MaxBodySize            int64            `config:"max_body_size"`
	RateLimit              *rateLimitConfig `config:"rate_limit"`
	KeepAlive              keepAlive        `config:"keep_alive"`

	Transport httpcommon.HTTPTransportSettings `config:",inline"`

	Tracer *tracerConfig `config:"tracer"`
}

type tracerConfig struct {
	Enabled           *bool `config:"enabled"`
	lumberjack.Logger `config:",inline"`
}

func (t *tracerConfig) enabled() bool {
	return t != nil && (t.Enabled == nil || *t.Enabled)
}

type urlConfig struct {
	*url.URL
}

func (u *urlConfig) Unpack(in string) error {
	parsed, err := url.Parse(in)
	if err != nil {
		return err
	}
	if parsed.Scheme == "file" {
		// This may not work correctly on Windows because it will leave a slash
		// prefix on resulting absolute path. This is go.dev/issue/6027 and
		// related. Python, Node and Java all do the same thing. Clients on
		// Windows will need to handle paths themselves.
		//
		// file:///C:\path\to\file -> /C:\file\to\path
		parsed.Scheme = ""
	}
	u.URL = parsed
	return nil
}

func (c *ResourceConfig) Validate() error {
	if c.Tracer == nil {
		return nil
	}
	if c.Tracer.Filename == "" {
		return errors.New("request tracer must have a filename if used")
	}
	if c.Tracer.MaxSize == 0 {
		// By default Lumberjack caps file sizes at 100MB which
		// is excessive for a debugging logger, so default to 1MB
		// which is the minimum.
		c.Tracer.MaxSize = 1
	}
	return nil
}
