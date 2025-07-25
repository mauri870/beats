The Redis `key` metricset collects information about Redis keys.

For each key matching one of the configured patterns, an event is sent to Elasticsearch with information about this key, what includes the type, its length when available, and its ttl.

Patterns are configured as a list containing these fields:

* `pattern` (required): pattern for key names, as accepted by the Redis `KEYS` or `SCAN` commands.
* `limit` (optional): safeguard when using patterns with wildcards to avoid collecting too many keys (Default: 0, no limit)
* `keyspace` (optional): Identifier of the database to use to look for the keys (Default: 0)

For example the following configuration will collect information about all keys whose name starts with `pipeline-*`, with a limit of 20 keys.

```yaml
- module: redis
  metricsets: ['key']
  key.patterns:
    - pattern: 'pipeline-*'
      limit: 20
```


## Dashboard [_dashboard_39]

The Redis key metricset comes with a predefined dashboard. For example:

![metricbeat redis key dashboard](images/metricbeat_redis_key_dashboard.png)
