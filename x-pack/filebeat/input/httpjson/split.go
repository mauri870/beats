// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package httpjson

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/elastic/beats/v7/libbeat/management/status"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/elastic/elastic-agent-libs/mapstr"
)

var (
	errEmptyField          = errors.New("the requested field is empty")
	errEmptyRootField      = errors.New("the requested root field is empty")
	errExpectedSplitArr    = errors.New("split was expecting field to be an array")
	errExpectedSplitObj    = errors.New("split was expecting field to be an object")
	errExpectedSplitString = errors.New("split was expecting field to be a string")
	errUnknownSplitType    = errors.New("unknown split type")
)

// split is a split processor chain element. Split processing is executed
// by applying elements of the chain's linked list to an input until completed
// or an error state is encountered.
type split struct {
	targetInfo       targetInfo
	kind             string
	transforms       []basicTransform
	child            *split
	keepParent       bool
	ignoreEmptyValue bool
	keyField         string
	isRoot           bool
	delimiter        string
	status           status.StatusReporter
	log              *logp.Logger
}

// newSplitResponse returns a new split based on the provided config and
// logging to the provided logger, tagging the split as the root of the chain.
func newSplitResponse(cfg *splitConfig, stat status.StatusReporter, log *logp.Logger) (*split, error) {
	if cfg == nil {
		return nil, nil
	}

	split, err := newSplit(cfg, stat, log)
	if err != nil {
		return nil, err
	}
	// We want to be able to identify which split is the root of the chain.
	split.isRoot = true
	return split, nil
}

// newSplit returns a new split based on the provided config and
// logging to the provided logger.
func newSplit(c *splitConfig, stat status.StatusReporter, log *logp.Logger) (*split, error) {
	ti, err := getTargetInfo(c.Target)
	if err != nil {
		return nil, err
	}

	if ti.Type != targetBody {
		return nil, fmt.Errorf("invalid target type: %s", ti.Type)
	}

	ts, err := newBasicTransformsFromConfig(registeredTransforms, c.Transforms, responseNamespace, stat, log)
	if err != nil {
		return nil, err
	}

	var s *split
	if c.Split != nil {
		s, err = newSplit(c.Split, stat, log)
		if err != nil {
			return nil, err
		}
	}

	return &split{
		targetInfo:       ti,
		kind:             c.Type,
		keepParent:       c.KeepParent,
		ignoreEmptyValue: c.IgnoreEmptyValue,
		keyField:         c.KeyField,
		delimiter:        c.DelimiterString,
		transforms:       ts,
		child:            s,
		status:           stat,
		log:              log,
	}, nil
}

// run runs the split operation on the contents of resp, processing successive
// split results on via h. ctx is passed to transforms that are called during
// the split.
func (s *split) run(ctx context.Context, trCtx *transformContext, resp transformable, h handler) error {
	root := resp.body()
	return s.split(ctx, trCtx, root, h)
}

// split recursively executes the split processor chain.
func (s *split) split(ctx context.Context, trCtx *transformContext, root mapstr.M, h handler) error {

	v, err := root.GetValue(s.targetInfo.Name)
	if err != nil && err != mapstr.ErrKeyNotFound { //nolint:errorlint // mapstr.ErrKeyNotFound is never wrapped by GetValue.
		return err
	}

	if v == nil {
		if s.ignoreEmptyValue {
			if s.child != nil {
				return s.child.split(ctx, trCtx, root, h)
			}
			if s.keepParent {
				h.handleEvent(ctx, root)
			}
			return nil
		}
		if s.isRoot {
			if s.keepParent {
				h.handleEvent(ctx, root)
				return errEmptyField
			}
			return errEmptyRootField
		}
		h.handleEvent(ctx, root)
		return errEmptyField
	}

	switch s.kind {
	case "", splitTypeArr:
		varr, ok := v.([]interface{})
		if !ok {
			return errExpectedSplitArr
		}

		if len(varr) == 0 {
			if s.ignoreEmptyValue {
				if s.child != nil {
					return s.child.split(ctx, trCtx, root, h)
				}
				if s.keepParent {
					h.handleEvent(ctx, root)
				}
				return nil
			}
			if s.isRoot {
				h.handleEvent(ctx, root)
				return errEmptyRootField
			}
			h.handleEvent(ctx, root)
			return errEmptyField
		}

		for _, e := range varr {
			err := s.processMessage(ctx, trCtx, root, s.targetInfo.Name, e, h)
			if err != nil {
				s.log.Debug(err)
			}
		}

		return nil
	case splitTypeMap:
		vmap, ok := toMapStr(v, s.targetInfo.Name)
		if !ok {
			return errExpectedSplitObj
		}

		if len(vmap) == 0 {
			if s.ignoreEmptyValue {
				if s.child != nil {
					return s.child.split(ctx, trCtx, root, h)
				}
				if s.keepParent {
					h.handleEvent(ctx, root)
				}
				return nil
			}
			if s.isRoot {
				return errEmptyRootField
			}
			h.handleEvent(ctx, root)
			return errEmptyField
		}

		for k, e := range vmap {
			if err := s.processMessage(ctx, trCtx, root, k, e, h); err != nil {
				s.log.Debug(err)
			}
		}

		return nil
	case splitTypeString:
		vstr, ok := v.(string)
		if !ok {
			return errExpectedSplitString
		}

		if len(vstr) == 0 {
			if s.ignoreEmptyValue {
				if s.child != nil {
					return s.child.split(ctx, trCtx, root, h)
				}
				return nil
			}
			if s.isRoot {
				return errEmptyRootField
			}
			h.handleEvent(ctx, root)
			return errEmptyField
		}
		for _, substr := range strings.Split(vstr, s.delimiter) {
			if err := s.processMessageSplitString(ctx, trCtx, root, substr, h); err != nil {
				s.log.Debug(err)
			}
		}

		return nil
	}

	return errUnknownSplitType
}

// processMessage processes an array or map split result value, v, via h after performing
// any necessary transformations. If key is "", the value is an element of an array.
func (s *split) processMessage(ctx context.Context, trCtx *transformContext, root mapstr.M, key string, v interface{}, h handler) error {
	obj, ok := toMapStr(v, s.targetInfo.Name)
	if !ok {
		return errExpectedSplitObj
	}
	if s.keyField != "" && key != "" {
		_, _ = obj.Put(s.keyField, key)
	}

	clone := root.Clone()
	if s.keepParent {
		_, _ = clone.Put(s.targetInfo.Name, v)
	} else {
		clone = obj
	}

	tr := transformable{}
	tr.setBody(clone)

	var err error
	for _, t := range s.transforms {
		tr, err = t.run(trCtx, tr)
		if err != nil {
			return err
		}
	}

	if s.child != nil {
		return s.child.split(ctx, trCtx, clone, h)
	}

	h.handleEvent(ctx, clone)

	return nil
}

func toMapStr(v interface{}, key string) (mapstr.M, bool) {
	if v == nil {
		return mapstr.M{}, false
	}
	switch t := v.(type) {
	case mapstr.M:
		return t, true
	case map[string]interface{}:
		return mapstr.M(t), true
	case string, []bool, []int, []string, []interface{}:
		return mapstr.M{key: t}, true
	}
	return mapstr.M{}, false
}

// sendMessage processes a string split result value, v, via h after performing any
// necessary transformations. If key is "", the value is an element of an array.
func (s *split) processMessageSplitString(ctx context.Context, trCtx *transformContext, root mapstr.M, v string, h handler) error {
	clone := root.Clone()
	_, _ = clone.Put(s.targetInfo.Name, v)

	tr := transformable{}
	tr.setBody(clone)

	var err error
	for _, t := range s.transforms {
		tr, err = t.run(trCtx, tr)
		if err != nil {
			return err
		}
	}

	if s.child != nil {
		return s.child.split(ctx, trCtx, clone, h)
	}

	h.handleEvent(ctx, clone)

	return nil
}
