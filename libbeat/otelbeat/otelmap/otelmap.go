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

// Package otelmap provides utilities for converting between beats and otel map types.
package otelmap

import (
	"fmt"
	"time"

	"github.com/elastic/elastic-agent-libs/mapstr"

	"go.opentelemetry.io/collector/pdata/pcommon"
)

// ToMapstr converts a [pcommon.Map] to a [mapstr.M].
func ToMapstr(m pcommon.Map) mapstr.M {
	return m.AsRaw()
}

// FromMapstr converts a [mapstr.M] to a [pcommon.Map].
func FromMapstr(m mapstr.M) pcommon.Map {
	out := pcommon.NewMap()
	for k, v := range m {
		switch x := v.(type) {
		case []string:
			ss := out.PutEmptySlice(k)
			for _, sv := range x {
				anyToPcommonValue(sv).CopyTo(ss.AppendEmpty())
			}
		case []int:
			ss := out.PutEmptySlice(k)
			for _, sv := range x {
				anyToPcommonValue(sv).CopyTo(ss.AppendEmpty())
			}
		case []int8:
			ss := out.PutEmptySlice(k)
			for _, sv := range x {
				anyToPcommonValue(sv).CopyTo(ss.AppendEmpty())
			}
		case []int16:
			ss := out.PutEmptySlice(k)
			for _, sv := range x {
				anyToPcommonValue(sv).CopyTo(ss.AppendEmpty())
			}
		case []int32:
			ss := out.PutEmptySlice(k)
			for _, sv := range x {
				anyToPcommonValue(sv).CopyTo(ss.AppendEmpty())
			}
		case []int64:
			ss := out.PutEmptySlice(k)
			for _, sv := range x {
				anyToPcommonValue(sv).CopyTo(ss.AppendEmpty())
			}
		case []uint:
			ss := out.PutEmptySlice(k)
			for _, sv := range x {
				anyToPcommonValue(sv).CopyTo(ss.AppendEmpty())
			}
		case []uint8:
			ss := out.PutEmptySlice(k)
			for _, sv := range x {
				anyToPcommonValue(sv).CopyTo(ss.AppendEmpty())
			}
		case []uint16:
			ss := out.PutEmptySlice(k)
			for _, sv := range x {
				anyToPcommonValue(sv).CopyTo(ss.AppendEmpty())
			}
		case []uint32:
			ss := out.PutEmptySlice(k)
			for _, sv := range x {
				anyToPcommonValue(sv).CopyTo(ss.AppendEmpty())
			}
		case []uint64:
			ss := out.PutEmptySlice(k)
			for _, sv := range x {
				anyToPcommonValue(sv).CopyTo(ss.AppendEmpty())
			}
		case []float32:
			ss := out.PutEmptySlice(k)
			for _, sv := range x {
				anyToPcommonValue(sv).CopyTo(ss.AppendEmpty())
			}
		case []float64:
			ss := out.PutEmptySlice(k)
			for _, sv := range x {
				anyToPcommonValue(sv).CopyTo(ss.AppendEmpty())
			}
		case []bool:
			ss := out.PutEmptySlice(k)
			for _, sv := range x {
				anyToPcommonValue(sv).CopyTo(ss.AppendEmpty())
			}
		case []any:
			ss := out.PutEmptySlice(k)
			for _, sv := range x {
				anyToPcommonValue(sv).CopyTo(ss.AppendEmpty())
			}
		case []time.Time:
			ss := out.PutEmptySlice(k)
			for _, sv := range x {
				anyToPcommonValue(sv).CopyTo(ss.AppendEmpty())
			}
		case []mapstr.M:
			dest := out.PutEmptySlice(k)
			for _, i := range v.([]mapstr.M) {
				FromMapstr(i).CopyTo(dest.AppendEmpty().SetEmptyMap())
			}
		case mapstr.M:
			FromMapstr(x).CopyTo(out.PutEmptyMap(k))
		default:
			anyToPcommonValue(x).CopyTo(out.PutEmpty(k))
		}
	}
	return out
}

func toSlice(v []any) {
	println(v)
}

func anyToPcommonValue(v any) pcommon.Value {
	switch x := v.(type) {
	case string:
		return pcommon.NewValueStr(x)
	case int:
		return pcommon.NewValueInt(int64(x))
	case int8:
		return pcommon.NewValueInt(int64(x))
	case int16:
		return pcommon.NewValueInt(int64(x))
	case int32:
		return pcommon.NewValueInt(int64(x))
	case int64:
		return pcommon.NewValueInt(x)
	case uint:
		return pcommon.NewValueInt(int64(x))
	case uint8:
		return pcommon.NewValueInt(int64(x))
	case uint16:
		return pcommon.NewValueInt(int64(x))
	case uint32:
		return pcommon.NewValueInt(int64(x))
	case uint64:
		return pcommon.NewValueInt(int64(x))
	case float32:
		return pcommon.NewValueDouble(float64(x))
	case float64:
		return pcommon.NewValueDouble(float64(x))
	case bool:
		return pcommon.NewValueBool(x)
	case time.Time:
		return pcommon.NewValueStr(x.UTC().Format("2006-01-02T15:04:05.000Z"))
	default:
		return pcommon.NewValueStr(fmt.Sprintf("unknown type: %T", x))
	}
}
