---
mapped_pages:
  - https://www.elastic.co/guide/en/beats/metricbeat/current/metricbeat-metricset-vsphere-datastore.html
---

% This file is generated! See scripts/docs_collector.py

# vSphere datastore metricset [metricbeat-metricset-vsphere-datastore]

This is the `datastore` metricset of the vSphere module.

This is a default metricset. If the host module is unconfigured, this metricset is enabled by default.

## Fields [_fields]

For a description of each field in the metricset, see the [exported fields](/reference/metricbeat/exported-fields-vsphere.md) section.

Here is an example document generated by this metricset:

```json
{
    "@timestamp": "2017-10-12T08:05:34.853Z",
    "event": {
        "dataset": "vsphere.datastore",
        "duration": 115000,
        "module": "vsphere"
    },
    "metricset": {
        "name": "datastore",
        "period": 10000
    },
    "service": {
        "address": "127.0.0.1:33365",
        "type": "vsphere"
    },
    "vsphere": {
        "datastore": {
            "host": {
                "count": 1,
                "names": [
                    "DC3_H0"
                ]
            },
            "status": "green",
            "vm": {
                "count": 6,
                "names": [
                    "DC3_H0_VM0"
                ]
            },
            "read": {
                "bytes": 0
            },
            "write": {
                "bytes": 337000
            },
            "disk": {
                "capacity": {
                    "usage": {
                        "bytes": 520505786368
                    },
                    "bytes": 1610344300544
                },
                "provisioned": {
                    "bytes": 520505786368
                }
            },
            "capacity": {
                "free": {
                    "bytes": 37120094208
                },
                "total": {
                    "bytes": 74686664704
                },
                "used": {
                    "bytes": 37566570496,
                    "pct": 0.502988996026061
                }
            },
            "fstype": "local",
            "name": "LocalDS_0",
            "id": "datastore-0"
        }
    }
}
```
