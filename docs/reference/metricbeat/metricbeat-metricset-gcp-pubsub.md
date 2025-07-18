---
mapped_pages:
  - https://www.elastic.co/guide/en/beats/metricbeat/current/metricbeat-metricset-gcp-pubsub.html
---

% This file is generated! See scripts/docs_collector.py

# Google Cloud Platform pubsub metricset [metricbeat-metricset-gcp-pubsub]

PubSub metricsetf fetches metrics from [Pub/Sub](https://cloud.google.com/pubsub/) topics and subscriptions in Google Cloud Platform.

The `pubsub` metricset contains all GA stage metrics exported from the [Stackdriver API](https://cloud.google.com/monitoring/api/metrics_gcp#gcp-pubsub). The field names are aligned to [Beats naming conventions](/extend/event-conventions.md) with minor modifications to their GCP metrics name counterpart.

No special permissions are needed apart from the ones detailed in the module section of the docs.

This is a default metricset. If the host module is unconfigured, this metricset is enabled by default.

## Fields [_fields]

For a description of each field in the metricset, see the [exported fields](/reference/metricbeat/exported-fields-gcp.md) section.

Here is an example document generated by this metricset:

```json
{
    "@timestamp": "2017-10-12T08:05:34.853Z",
    "cloud": {
        "account": {
            "id": "elastic-observability"
        },
        "provider": "gcp"
    },
    "event": {
        "dataset": "gcp.pubsub",
        "duration": 115000,
        "module": "gcp"
    },
    "gcp": {
        "labels": {
            "resource": {
                "subscription_id": "test-subscription-1"
            }
        },
        "pubsub": {
            "subscription": {
                "backlog_bytes": {
                    "value": 0
                }
            }
        }
    },
    "metricset": {
        "name": "pubsub",
        "period": 10000
    },
    "service": {
        "type": "gcp"
    }
}
```
