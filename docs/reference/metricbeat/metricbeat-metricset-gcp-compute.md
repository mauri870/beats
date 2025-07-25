---
mapped_pages:
  - https://www.elastic.co/guide/en/beats/metricbeat/current/metricbeat-metricset-gcp-compute.html
---

% This file is generated! See scripts/docs_collector.py

# Google Cloud Platform compute metricset [metricbeat-metricset-gcp-compute]

Compute metricset to fetch metrics from [Compute Engine](https://cloud.google.com/compute/) Virtual Machines in Google Cloud Platform. No Monitoring or Logging agent is required in your instances to use this metricset.

The `compute` metricset contains all metrics exported from the [Stackdriver API](https://cloud.google.com/monitoring/api/metrics_gcp#gcp-compute). The field names are aligned to [Beats naming conventions](/extend/event-conventions.md) with minor modifications to their GCP metrics name counterpart.

Extra labels and metadata are also extracted using the [Compute API](https://cloud.google.com/compute/docs/reference/rest/v1/instances/get). This is enough to get most of the info associated with a metric like Compute labels and metadata and metric specific Labels.


## Labels [_labels]

Here is a list of labels collected by `compute` metricset depending on the type of metric being collected:

* `instance_name`: The name of the VM instance. Collected with:

    * `gcp.instance.firewall.*`
    * `gcp.instance.cpu.*`
    * `gcp.instance.disk.*`
    * `gcp.instance.memory.*`
    * `gcp.instance.network.*`
    * `gcp.instance.uptime`

* `device_name`: The name of the disk device. Collected with:

    * `gcp.instance.disk.*`

* `storage_type`: The storage type: `pd-standard`, `pd-ssd`, or `local-ssd`. Collected with:

    * `gcp.instance.disk.*`

* `device_type`: The disk type: `ephemeral` or `permanent`. Collected with:

    * `gcp.instance.disk.*`

* `loadBalanced`: Whether traffic was sent from an L3 loadbalanced IP address assigned to the VM. Traffic that is externally routed from the VM’s standard internal or external IP address, such as L7 loadbalanced traffic, is not considered to be loadbalanced in this metric. Collected with:

    * `gcp.instance.network.*`

This is a default metricset. If the host module is unconfigured, this metricset is enabled by default.

## Fields [_fields]

For a description of each field in the metricset, see the [exported fields](/reference/metricbeat/exported-fields-gcp.md) section.

Here is an example document generated by this metricset:

```json
{
    "@timestamp": "2017-10-12T08:05:34.853Z",
    "cloud": {
        "account": {
            "id": "elastic-observability",
            "name": "elastic-observability"
        },
        "instance": {
            "id": "1113015278728017638",
            "name": "gke-apm-ci-k8s-cluster-pool-2-e8852348-58mx"
        },
        "provider": "gcp",
        "availability_zone": "us-central1-a",
        "region": "us-central1"
    },
    "event": {
        "dataset": "gcp.compute",
        "duration": 115000,
        "module": "gcp"
    },
    "gcp": {
        "compute": {
            "instance": {
                "disk": {
                    "read_bytes_count": {
                        "value": 989296850
                    },
                    "read_ops_count": {
                        "value": 14862
                    },
                    "write_bytes_count": {
                        "value": 1323095
                    },
                    "write_ops_count": {
                        "value": 105
                    }
                }
            }
        },
        "labels": {
            "metrics": {
                "device_name": "gke-apm-ci-k8s-cluster-pool-2-e8852348-58mx",
                "device_type": "permanent",
                "storage_type": "pd-standard"
            }
        }
    },
    "host": {
        "disk": {
            "read": {
                "bytes": 989296850
            },
            "write": {
                "bytes": 1323095
            }
        },
        "id": "1113015278728017638",
        "name": "gke-apm-ci-k8s-cluster-pool-2-e8852348-58mx"
    },
    "metricset": {
        "name": "compute",
        "period": 10000
    },
    "service": {
        "type": "gcp"
    }
}
```
