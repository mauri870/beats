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

package index

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/elastic/beats/v7/metricbeat/helper"
	"github.com/elastic/beats/v7/metricbeat/helper/elastic"
	"github.com/elastic/beats/v7/metricbeat/mb"
	"github.com/elastic/beats/v7/metricbeat/module/elasticsearch"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/elastic/elastic-agent-libs/mapstr"
)

// Based on https://github.com/elastic/elasticsearch/blob/master/x-pack/plugin/monitoring/src/main/java/org/elasticsearch/xpack/monitoring/collector/indices/IndexStatsMonitoringDoc.java#L127-L203
type stats struct {
	Indices map[string]Index `json:"indices"`
}

type Index struct {
	UUID      string    `json:"uuid"`
	Primaries primaries `json:"primaries"`
	Total     total     `json:"total"`

	Index          string     `json:"index"`
	Status         string     `json:"status"`
	TierPreference string     `json:"tier_preference,omitempty"`
	CreationDate   string     `json:"creation_date,omitempty"`
	Version        string     `json:"version,omitempty"`
	Shards         shardStats `json:"shards"`
}

type primaries struct {
	Docs struct {
		Count   int `json:"count"`
		Deleted int `json:"deleted"`
	} `json:"docs"`
	Indexing struct {
		IndexTotal           int `json:"index_total"`
		IndexTimeInMillis    int `json:"index_time_in_millis"`
		ThrottleTimeInMillis int `json:"throttle_time_in_millis"`
	} `json:"indexing"`
	Merges struct {
		TotalSizeInBytes int `json:"total_size_in_bytes"`
	} `json:"merges"`
	Segments struct {
		Count                     int `json:"count"`
		MemoryInBytes             int `json:"memory_in_bytes"`
		TermsMemoryInBytes        int `json:"terms_memory_in_bytes"`
		StoredFieldsMemoryInBytes int `json:"stored_fields_memory_in_bytes"`
		TermVectorsMemoryInBytes  int `json:"term_vectors_memory_in_bytes"`
		NormsMemoryInBytes        int `json:"norms_memory_in_bytes"`
		PointsMemoryInBytes       int `json:"points_memory_in_bytes"`
		DocValuesMemoryInBytes    int `json:"doc_values_memory_in_bytes"`
		IndexWriterMemoryInBytes  int `json:"index_writer_memory_in_bytes"`
		VersionMapMemoryInBytes   int `json:"version_map_memory_in_bytes"`
		FixedBitSetMemoryInBytes  int `json:"fixed_bit_set_memory_in_bytes"`
	} `json:"segments"`
	Store struct {
		SizeInBytes             int `json:"size_in_bytes"`
		TotalDataSetSizeInBytes int `json:"total_data_set_size_in_bytes"`
	} `json:"store"`
	Refresh struct {
		TotalTimeInMillis         int `json:"total_time_in_millis"`
		ExternalTotalTimeInMillis int `json:"external_total_time_in_millis"`
	} `json:"refresh"`
	QueryCache struct {
		MemorySizeInBytes int `json:"memory_size_in_bytes"`
		HitCount          int `json:"hit_count"`
		MissCount         int `json:"miss_count"`
	} `json:"query_cache"`
	RequestCache struct {
		MemorySizeInBytes int `json:"memory_size_in_bytes"`
		HitCount          int `json:"hit_count"`
		MissCount         int `json:"miss_count"`
		Evictions         int `json:"evictions"`
	} `json:"request_cache"`
	Search struct {
		QueryTotal        int `json:"query_total"`
		QueryTimeInMillis int `json:"query_time_in_millis"`
	} `json:"search"`
}

type total struct {
	Docs struct {
		Count   int `json:"count"`
		Deleted int `json:"deleted"`
	} `json:"docs"`
	FieldData struct {
		MemorySizeInBytes int `json:"memory_size_in_bytes"`
		Evictions         int `json:"evictions"`
	} `json:"fielddata"`
	Indexing struct {
		IndexTotal           int `json:"index_total"`
		IndexTimeInMillis    int `json:"index_time_in_millis"`
		ThrottleTimeInMillis int `json:"throttle_time_in_millis"`
	} `json:"indexing"`
	Bulk   *bulkStats `json:"bulk,omitempty"`
	Merges struct {
		TotalSizeInBytes int `json:"total_size_in_bytes"`
	} `json:"merges"`
	Search struct {
		QueryTotal        int `json:"query_total"`
		QueryTimeInMillis int `json:"query_time_in_millis"`
	} `json:"search"`
	Segments struct {
		Count                     int `json:"count"`
		MemoryInBytes             int `json:"memory_in_bytes"`
		TermsMemoryInBytes        int `json:"terms_memory_in_bytes"`
		StoredFieldsMemoryInBytes int `json:"stored_fields_memory_in_bytes"`
		TermVectorsMemoryInBytes  int `json:"term_vectors_memory_in_bytes"`
		NormsMemoryInBytes        int `json:"norms_memory_in_bytes"`
		PointsMemoryInBytes       int `json:"points_memory_in_bytes"`
		DocValuesMemoryInBytes    int `json:"doc_values_memory_in_bytes"`
		IndexWriterMemoryInBytes  int `json:"index_writer_memory_in_bytes"`
		VersionMapMemoryInBytes   int `json:"version_map_memory_in_bytes"`
		FixedBitSetMemoryInBytes  int `json:"fixed_bit_set_memory_in_bytes"`
	} `json:"segments"`
	Store struct {
		SizeInBytes             int `json:"size_in_bytes"`
		TotalDataSetSizeInBytes int `json:"total_data_set_size_in_bytes"`
	} `json:"store"`
	Refresh struct {
		TotalTimeInMillis         int `json:"total_time_in_millis"`
		ExternalTotalTimeInMillis int `json:"external_total_time_in_millis"`
	} `json:"refresh"`
	QueryCache struct {
		MemorySizeInBytes int `json:"memory_size_in_bytes"`
		HitCount          int `json:"hit_count"`
		MissCount         int `json:"miss_count"`
		Evictions         int `json:"evictions"`
	} `json:"query_cache"`
	RequestCache struct {
		MemorySizeInBytes int `json:"memory_size_in_bytes"`
		HitCount          int `json:"hit_count"`
		MissCount         int `json:"miss_count"`
		Evictions         int `json:"evictions"`
	} `json:"request_cache"`
}

type shardStats struct {
	Total     int `json:"total"`
	Primaries int `json:"primaries"`
	Replicas  int `json:"-"`

	ActiveTotal     int `json:"-"`
	ActivePrimaries int `json:"-"`
	ActiveReplicas  int `json:"-"`

	UnassignedTotal     int `json:"-"`
	UnassignedPrimaries int `json:"-"`
	UnassignedReplicas  int `json:"-"`

	Initializing int `json:"-"`
	Relocating   int `json:"-"`
}

type bulkStats struct {
	TotalOperations   int `json:"total_operations"`
	TotalTimeInMillis int `json:"total_time_in_millis"`
	TotalSizeInBytes  int `json:"total_size_in_bytes"`
	AvgTimeInMillis   int `json:"avg_time_in_millis"`
	AvgSizeInBytes    int `json:"avg_size_in_bytes"`
}

func eventsMapping(
	r mb.ReporterV2,
	httpClient *helper.HTTP,
	info elasticsearch.Info,
	content []byte,
	isXpack bool,
	log *logp.Logger) error {
	clusterStateMetrics := []string{"routing_table"}
	clusterStateFilterPaths := []string{"routing_table"}
	clusterState, err := elasticsearch.GetClusterState(httpClient, httpClient.GetURI(), clusterStateMetrics, clusterStateFilterPaths)
	if err != nil {
		return fmt.Errorf("failure retrieving cluster state from Elasticsearch: %w", err)
	}

	indicesSettingsPattern := "*,-.*"
	indicesSettingsFilterPaths := []string{"*.settings.index.creation_date", "*.settings.index.**._tier_preference", "*.settings.index.version.created"}
	indicesSettings, err := elasticsearch.GetIndexSettings(httpClient, httpClient.GetURI(), indicesSettingsPattern, indicesSettingsFilterPaths)
	if err != nil {
		return fmt.Errorf("failure retrieving index settings from Elasticsearch: %w", err)
	}

	// Under some very rare circumstances, an index in the stats response might not have an entry in the settings or cluster state.
	// This can happen if the index got deleted between the time the settings and cluster state were retrieved and the time the stats were retrieved.

	var indicesStats stats
	if err := parseAPIResponse(content, &indicesStats); err != nil {
		return fmt.Errorf("failure parsing Indices Stats Elasticsearch API response: %w", err)
	}

	var errs []error
	for name := range indicesStats.Indices {
		event := mb.Event{
			ModuleFields: mapstr.M{},
		}
		idx := indicesStats.Indices[name]
		idx.Index = name

		err = addClusterStateFields(&idx, clusterState)
		if err != nil {
			// We can't continue processing this index, so we skip it.
			errs = append(errs, fmt.Errorf("failure adding cluster state fields: %w", err))
			continue
		}

		err = addIndexSettings(&idx, indicesSettings)
		if err != nil {
			// Failure to add index settings is sometimes expected and won't be breaking,
			// so we log it as debug and carry on with regular processing.
			log.Named("elasticsearch.index").Debugf("failure adding index settings: %v", err)
		}

		event.ModuleFields.Put("cluster.id", info.ClusterID)
		event.ModuleFields.Put("cluster.name", info.ClusterName)

		// Convert struct to common.Mapstr by passing it to JSON first so we can store the data in the root of the
		// metricset level
		indexBytes, err := json.Marshal(idx)
		if err != nil {
			errs = append(errs, fmt.Errorf("failure trying to convert metrics results to JSON: %w", err))
			continue
		}
		var indexOutput mapstr.M
		if err = json.Unmarshal(indexBytes, &indexOutput); err != nil {
			errs = append(errs, fmt.Errorf("failure trying to convert JSON metrics back to mapstr: %w", err))
			continue
		}

		event.MetricSetFields = indexOutput
		event.MetricSetFields.Put("name", name)
		delete(event.MetricSetFields, "index")

		// xpack.enabled in config using standalone metricbeat writes to `.monitoring` instead of `metricbeat-*`
		// When using Agent, the index name is overwritten anyways.
		if isXpack {
			index := elastic.MakeXPackMonitoringIndexName(elastic.Elasticsearch)
			event.Index = index
		}

		r.Event(event)
	}

	return errors.Join(errs...)
}

func parseAPIResponse(content []byte, indicesStats *stats) error {
	return json.Unmarshal(content, indicesStats)
}

// Fields added here are based on same fields being added by internal collection in
// https://github.com/elastic/elasticsearch/blob/master/x-pack/plugin/monitoring/src/main/java/org/elasticsearch/xpack/monitoring/collector/indices/IndexStatsMonitoringDoc.java#L62-L124
func addClusterStateFields(idx *Index, clusterState mapstr.M) error {
	indexRoutingTable, err := getClusterStateMetricForIndex(clusterState, idx.Index, "routing_table")
	if err != nil {
		return fmt.Errorf("failed to get index routing table from cluster state: %w", err)
	}

	shards, err := getShardsFromRoutingTable(indexRoutingTable)
	if err != nil {
		return fmt.Errorf("failed to get shards from routing table: %w", err)
	}

	// "index_stats.version.created", <--- don't think this is being used in the UI, so can we skip it?
	// "index_stats.version.upgraded", <--- don't think this is being used in the UI, so can we skip it?

	status, err := getIndexStatus(shards)
	if err != nil {
		return fmt.Errorf("failed to get index status: %w", err)
	}
	idx.Status = status

	shardStats, err := getIndexShardStats(shards)
	if err != nil {
		return fmt.Errorf("failed to get index shard stats: %w", err)
	}
	idx.Shards = *shardStats
	return nil
}

func addIndexSettings(idx *Index, indicesSettings mapstr.M) error {

	// Recover the index settings for our specific index
	indexSettingsValue, err := indicesSettings.GetValue(idx.Index)
	if err != nil {
		return fmt.Errorf("failed to get index settings for index %s: %w", idx.Index, err)
	}

	indexSettings, ok := indexSettingsValue.(map[string]interface{})
	if !ok {
		return fmt.Errorf("index settings is not a map for index: %s", idx.Index)
	}

	indexCreationDate, err := getIndexSettingForIndex(indexSettings, idx.Index, "index.creation_date")
	if err != nil {
		return fmt.Errorf("failed to get index creation date: %w", err)
	}

	idx.CreationDate = indexCreationDate

	indexTierPreference, err := getIndexSettingForIndex(indexSettings, idx.Index, "index.routing.allocation.require._tier_preference")
	if err != nil {
		indexTierPreference, err = getIndexSettingForIndex(indexSettings, idx.Index, "index.routing.allocation.include._tier_preference")
		if err != nil {
			return fmt.Errorf("failed to get index tier preference: %w", err)
		}
	}

	idx.TierPreference = indexTierPreference

	indexVersion, err := getIndexSettingForIndex(indexSettings, idx.Index, "index.version.created")
	if err != nil {
		return fmt.Errorf("failed to get index version: %w", err)
	}

	idx.Version = indexVersion

	return nil
}

func getIndexSettingForIndex(indexSettings mapstr.M, index, settingKey string) (string, error) {
	fieldKey := "settings." + settingKey
	value, err := indexSettings.GetValue(fieldKey)
	if err != nil {
		return "", fmt.Errorf("'"+fieldKey+"': %w", err)
	}

	setting, ok := value.(string)
	if !ok {
		return "", elastic.MakeErrorForMissingField(fieldKey, elastic.Elasticsearch)
	}
	return setting, nil
}

func getClusterStateMetricForIndex(clusterState mapstr.M, index, metricKey string) (mapstr.M, error) {
	fieldKey := metricKey + ".indices." + index
	value, err := clusterState.GetValue(fieldKey)
	if err != nil {
		return nil, fmt.Errorf("'"+fieldKey+"': %w", err)
	}

	metric, ok := value.(map[string]interface{})
	if !ok {
		return nil, elastic.MakeErrorForMissingField(fieldKey, elastic.Elasticsearch)
	}
	return mapstr.M(metric), nil
}

func getIndexStatus(shards map[string]interface{}) (string, error) {
	if len(shards) == 0 {
		// No shards, index is red
		return "red", nil
	}

	areAllPrimariesStarted := true
	areAllReplicasStarted := true

	for indexName, indexShard := range shards {
		is, ok := indexShard.([]interface{})
		if !ok {
			return "", fmt.Errorf("shards is not an array")
		}

		for shardIdx, shard := range is {
			s, ok := shard.(map[string]interface{})
			if !ok {
				return "", fmt.Errorf("%v.shards[%v] is not a map", indexName, shardIdx)
			}

			shard := mapstr.M(s)

			isPrimary, ok := shard["primary"].(bool)
			if !ok {
				return "", fmt.Errorf("%v.shards[%v].primary is not a boolean", indexName, shardIdx)
			}

			state, ok := shard["state"].(string)
			if !ok {
				return "", fmt.Errorf("%v.shards[%v].state is not a string", indexName, shardIdx)
			}

			if isPrimary {
				areAllPrimariesStarted = areAllPrimariesStarted && (state == "STARTED")
			} else {
				areAllReplicasStarted = areAllReplicasStarted && (state == "STARTED")
			}
		}
	}

	if areAllPrimariesStarted && areAllReplicasStarted {
		return "green", nil
	}

	if areAllPrimariesStarted && !areAllReplicasStarted {
		return "yellow", nil
	}

	return "red", nil
}

func getIndexShardStats(shards mapstr.M) (*shardStats, error) {
	primaries := 0
	replicas := 0

	activePrimaries := 0
	activeReplicas := 0

	unassignedPrimaries := 0
	unassignedReplicas := 0

	initializing := 0
	relocating := 0

	for indexName, indexShard := range shards {
		is, ok := indexShard.([]interface{})
		if !ok {
			return nil, fmt.Errorf("shards is not an array")
		}

		for shardIdx, shard := range is {
			s, ok := shard.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("%v.shards[%v] is not a map", indexName, shardIdx)
			}

			shard := mapstr.M(s)

			isPrimary, ok := shard["primary"].(bool)
			if !ok {
				return nil, fmt.Errorf("%v.shards[%v].primary is not a boolean", indexName, shardIdx)
			}

			state, ok := shard["state"].(string)
			if !ok {
				return nil, fmt.Errorf("%v.shards[%v].state is not a string", indexName, shardIdx)
			}

			if isPrimary {
				primaries++
				switch state {
				case "STARTED":
					activePrimaries++
				case "UNASSIGNED":
					unassignedPrimaries++
				}
			} else {
				replicas++
				switch state {
				case "STARTED":
					activeReplicas++
				case "UNASSIGNED":
					unassignedReplicas++
				}
			}

			switch state {
			case "INITIALIZING":
				initializing++
			case "RELOCATING":
				relocating++
			}
		}
	}

	return &shardStats{
		Total:               primaries + replicas,
		Primaries:           primaries,
		Replicas:            replicas,
		ActiveTotal:         activePrimaries + activeReplicas,
		ActivePrimaries:     activePrimaries,
		ActiveReplicas:      activeReplicas,
		UnassignedTotal:     unassignedPrimaries + unassignedReplicas,
		UnassignedPrimaries: unassignedPrimaries,
		UnassignedReplicas:  unassignedReplicas,
		Initializing:        initializing,
		Relocating:          relocating,
	}, nil
}

func getShardsFromRoutingTable(indexRoutingTable mapstr.M) (map[string]interface{}, error) {
	s, err := indexRoutingTable.GetValue("shards")
	if err != nil {
		return nil, err
	}

	shards, ok := s.(map[string]interface{})
	if !ok {
		return nil, elastic.MakeErrorForMissingField("shards", elastic.Elasticsearch)
	}

	return shards, nil
}
