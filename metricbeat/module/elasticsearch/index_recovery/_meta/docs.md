This is the index_recovery metricset of the module Elasticsearch.

By default data about all indices are fetched. To gather data about indices that are under active recovery only, set `index_recovery.active_only: true`:

```yaml
- module: elasticsearch
  metricsets:
    - index_recovery
  hosts: ["localhost:9200"]
  index_recovery.active_only: false
```
