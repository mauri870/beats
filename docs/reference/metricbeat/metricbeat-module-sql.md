---
mapped_pages:
  - https://www.elastic.co/guide/en/beats/metricbeat/current/metricbeat-module-sql.html
---

% This file is generated! See scripts/docs_collector.py

# SQL module [metricbeat-module-sql]

The SQL module allows you to execute custom queries against an SQL database and store the results in {{es}}. It also enables the development of various SQL metrics integrations, using SQL query as input.

This module supports the databases that you can monitor with {{metricbeat}}, including:

* PostgreSQL
* MySQL
* Oracle
* Microsoft SQL
* CockroachDB

To enable the module, run:

```shell
metricbeat module enable sql
```

After enabling the module, open `modules.d/sql.yml` and set the required fields:

`driver`
:   The driver can be any driver that has a {{metricbeat}} module, such as `mssql` or `postgres`.

`fetch_from_all_databases`
:   Expects either `true` or `false` and it is by default set to `false`. Marking as `true` will enable execution `sql_queries` or `sql_query` for all databases in a server. Currently only `mssql` driver supports this feature. For other drivers, if enabled, "fetch from all databases feature is not supported for driver: <driver_name>" error would be logged.

`raw_data.enabled`
:   Expects either `true` or `false` and it is by default set to `false`. Marking as `true` will generate event results in a new field format.

Use `sql_queries` or `sql_query` depending on the use-case.

`sql_queries`
:   List of queries to execute. `query` and `response_format` fields are repeated to get multiple query inputs.

    `query`
    :   Single SQL query.

    `response_format`
    :   Either `variables` or `table`.

        `variables`
        :   Expects a two-column table that looks like a key-value result. The left column is considered a key and the right column is the value. This mode generates a single event on each fetch operation.

        `table`
        :   Expects any number of columns. This mode generates a single event for each row.
  
`ssl` configuration
:   Use `ssl.certificate_authorities`, `ssl.certificate`, `ssl.key`, and `ssl.verification_mode` params. See SSL configuration section. Supported drivers are `mysql`, `postgres`, and `mssql`.


`sql_query` (`Backward Compatibility`)
:   Single query you want to run. Also, provide corresponding `sql_response_format` (value: `variables` or `table`) similar to `sql_queries`'s `response_format`.


## Example [_example_4]

Examples of configurations in `sql.yml` to connect with supported databases are mentioned below.


### Example: Capture Innodb-related metrics [_example_capture_innodb_related_metrics]

This `sql.yml` configuration shows how to capture Innodb-related metrics that result from the query `SHOW GLOBAL STATUS LIKE 'Innodb_system%'` in a MySQL database:

```yaml
- module: sql
  metricsets:
    - query
  period: 10s
  hosts: ["root:root@tcp(localhost:3306)/ps"]

  driver: "mysql"
  sql_query: "SHOW GLOBAL STATUS LIKE 'Innodb_system%'"
  sql_response_format: variables
```

The `SHOW GLOBAL STATUS` query results in this table:

| Variable_name | Value |
| --- | --- |
| Innodb_system_rows_deleted | 0 |
| Innodb_system_rows_inserted | 0 |
| Innodb_system_rows_read | 5062 |
| Innodb_system_rows_updated | 315 |

Results are grouped by type in the result event for convenient mapping in {{es}}. For example, `strings` values are grouped into `sql.strings`, `numeric` into `sql.numeric`, and so on.

The example shown earlier generates this event:

```json
{
  "@timestamp": "2020-06-09T15:09:14.407Z",
  "@metadata": {
    "beat": "metricbeat",
    "type": "_doc",
    "version": "8.0.0"
  },
  "service": {
    "address": "172.18.0.2:3306",
    "type": "sql"
  },
  "event": {
    "dataset": "sql.query",
    "module": "sql",
    "duration": 1272810
  },
  "sql": {
    "driver": "mysql",
    "query": "SHOW GLOBAL STATUS LIKE 'Innodb_system%'",
    "metrics": {
      "numeric": {
        "innodb_system_rows_updated": 315,
        "innodb_system_rows_deleted": 0,
        "innodb_system_rows_inserted": 0,
        "innodb_system_rows_read": 5062
      }
    }
  },
  "metricset": {
    "name": "query",
    "period": 10000
  },
  "ecs": {
    "version": "1.5.0"
  },
  "host": {
    "name": "elastic"
  },
  "agent": {
    "name": "elastic",
    "type": "metricbeat",
    "version": "8.0.0",
    "ephemeral_id": "488431bd-bd3c-4442-ad51-0c50eb555787",
    "id": "670ef211-87f0-4f38-8beb-655c377f1629"
  }
}
```


### Example: Query PostgreSQL and generate a "table" result [_example_query_postgresql_and_generate_a_table_result]

This `sql.yml` configuration shows how to query PostgreSQL and generate a "table" result. This configuration generates a single event for each row returned:

```yaml
- module: sql
  metricsets:
    - query
  period: 10s
  hosts: ["postgres://postgres:postgres@localhost:5432/stuff?sslmode=disable"]

  driver: "postgres"
  sql_query: "SELECT datid, datname, blks_read, blks_hit, tup_returned, tup_fetched, stats_reset FROM pg_stat_database"
  sql_response_format: table
```

The SELECT query results in this table:

| datid | datname | blks_read | blks_hit | tup_returned | tup_fetched | stats_reset |
| --- | --- | --- | --- | --- | --- | --- |
| 69448 | stuff | 8652 | 205976 | 1484625 | 53218 | 2020-06-07 22:50:12 |
| 13408 | postgres | 0 | 0 | 0 | 0 |  |
| 13407 | template0 | 0 | 0 | 0 | 0 |  |

Because the table contains three rows, three events are generated, one event for each row. For example, this event is created for the first row:

```json
{
  "@timestamp": "2020-06-09T14:47:35.481Z",
  "@metadata": {
    "beat": "metricbeat",
    "type": "_doc",
    "version": "8.0.0"
  },
  "service": {
    "address": "localhost:5432",
    "type": "sql"
  },
  "ecs": {
    "version": "1.5.0"
  },
  "host": {
    "name": "elastic"
  },
  "agent": {
    "type": "metricbeat",
    "version": "8.0.0",
    "ephemeral_id": "1bffe66d-a1ae-4ed6-985a-fd48548a1971",
    "id": "670ef211-87f0-4f38-8beb-655c377f1629",
    "name": "elastic"
  },
  "sql": {
    "metrics": {
      "numeric": {
        "tup_fetched": 53350,
        "datid": 69448,
        "blks_read": 8652,
        "blks_hit": 206501,
        "tup_returned": 1.491873e+06
      },
      "string": {
        "stats_reset": "2020-06-07T20:50:12.632975Z",
        "datname": "stuff"
      }
    },
    "driver": "postgres",
    "query": "SELECT datid, datname, blks_read, blks_hit, tup_returned, tup_fetched, stats_reset FROM pg_stat_database"
  },
  "event": {
    "dataset": "sql.query",
    "module": "sql",
    "duration": 14076705
  },
  "metricset": {
    "name": "query",
    "period": 10000
  }
}
```


### Example: Get the buffer catch hit ratio in Oracle [_example_get_the_buffer_catch_hit_ratio_in_oracle]

This `sql.yml` configuration shows how to get the buffer cache hit ratio:

```yaml
- module: sql
  metricsets:
    - query
  period: 10s
  hosts: ["oracle://sys:password@172.17.0.3:1521/ORCLPDB1.localdomain?sysdba=1"]

  driver: "oracle"
  sql_query: 'SELECT name, physical_reads, db_block_gets, consistent_gets, 1 - (physical_reads / (db_block_gets + consistent_gets)) "Hit Ratio" FROM V$BUFFER_POOL_STATISTICS'
  sql_response_format: table
```

The example generates this event:

```json
{
  "@timestamp": "2020-06-09T15:41:02.200Z",
  "@metadata": {
    "beat": "metricbeat",
    "type": "_doc",
    "version": "8.0.0"
  },
  "sql": {
    "metrics": {
      "numeric": {
        "hit ratio": 0.9742963357937117,
        "physical_reads": 17161,
        "db_block_gets": 122221,
        "consistent_gets": 545427
      },
      "string": {
        "name": "DEFAULT"
      }
    },
    "driver": "oracle",
    "query": "SELECT name, physical_reads, db_block_gets, consistent_gets, 1 - (physical_reads / (db_block_gets + consistent_gets)) \"Hit Ratio\" FROM V$BUFFER_POOL_STATISTICS"
  },
  "metricset": {
    "period": 10000,
    "name": "query"
  },
  "service": {
    "address": "172.17.0.3:1521",
    "type": "sql"
  },
  "event": {
    "dataset": "sql.query",
    "module": "sql",
    "duration": 39233704
  },
  "ecs": {
    "version": "1.5.0"
  },
  "host": {
    "name": "elastic"
  },
  "agent": {
    "id": "670ef211-87f0-4f38-8beb-655c377f1629",
    "name": "elastic",
    "type": "metricbeat",
    "version": "8.0.0",
    "ephemeral_id": "49e00060-0fa4-4b34-80f1-446881f7a788"
  }
}
```


### Example: Get the buffer cache hit ratio for MSSQL [_example_get_the_buffer_cache_hit_ratio_for_mssql]

This `sql.yml` configuration gets the buffer cache hit ratio:

```yaml
- module: sql
  metricsets:
    - query
  period: 10s
  hosts: ["sqlserver://SA:password@localhost"]

  driver: "mssql"
  sql_query: 'SELECT * FROM sys.dm_db_log_space_usage'
  sql_response_format: table
```

The example generates this event:

```json
{
  "@timestamp": "2020-06-09T15:39:14.421Z",
  "@metadata": {
    "beat": "metricbeat",
    "type": "_doc",
    "version": "8.0.0"
  },
  "sql": {
    "driver": "mssql",
    "query": "SELECT * FROM sys.dm_db_log_space_usage",
    "metrics": {
      "numeric": {
        "log_space_in_bytes_since_last_backup": 524288,
        "database_id": 1,
        "total_log_size_in_bytes": 2.08896e+06,
        "used_log_space_in_bytes": 954368,
        "used_log_space_in_percent": 45.686275482177734
      }
    }
  },
  "event": {
    "dataset": "sql.query",
    "module": "sql",
    "duration": 40750570
  }
}
```


### Example: Launch two or more queries. [_example_launch_two_or_more_queries]

To launch two or more queries, specify the full configuration for each query. For example:

```yaml
- module: sql
  metricsets:
    - query
  period: 10s
  hosts: ["postgres://postgres:postgres@localhost:5432/stuff?sslmode=disable"]
  driver: "postgres"
  raw_data.enabled: true

  sql_queries:
    - query: "SELECT datid, datname, blks_read, blks_hit, tup_returned, tup_fetched, stats_reset FROM pg_stat_database"
      response_format: table

    - query: "SELECT datname, datid FROM pg_stat_database;"
      response_format: variables
```

The example generates this event: The response event is generated in new format by enabling the flag `raw_data.enabled`.

```json
{
  "@timestamp": "2022-05-13T12:47:32.071Z",
  "@metadata": {
    "beat": "metricbeat",
    "type": "_doc",
    "version": "8.3.0"
  },
  "event": {
    "dataset": "sql.query",
    "module": "sql",
    "duration": 114468667
  },
  "metricset": {
    "name": "query",
    "period": 10000
  },
  "service": {
    "address": "localhost:55656",
    "type": "sql"
  },
  "sql": {
    "driver": "postgres",
    "query": "SELECT datid, datname, blks_read, blks_hit, tup_returned, tup_fetched, stats_reset FROM pg_stat_database",
    "metrics": {
      "blks_hit": 6360,
      "tup_returned": 2225,
      "tup_fetched": 1458,
      "datid": 13394,
      "datname": "template0",
      "blks_read": 33
    }
  },
  "ecs": {
    "version": "8.0.0"
  },
  "host": {
    "name": "mps"
  },
  "agent": {
    "type": "metricbeat",
    "version": "8.3.0",
    "ephemeral_id": "8decc9eb-5ea5-47d8-8a22-fac507a5521b",
    "id": "6bbf5058-afed-44c6-aa05-775ee14a2da4",
    "name": "mps"
  }
}
```

The example generates this event: By disabling the flag `raw_data.enabled`, which is the old format.

```json
{
  "@timestamp": "2022-05-13T13:09:19.599Z",
  "@metadata": {
    "beat": "metricbeat",
    "type": "_doc",
    "version": "8.3.0"
  },
  "event": {
    "dataset": "sql.query",
    "module": "sql",
    "duration": 77509917
  },
"service": {
    "address": "localhost:55656",
    "type": "sql"
  },
  "metricset": {
    "name": "query",
    "period": 10000
  },

  "sql": {
    "driver": "postgres",
    "query": "SELECT datid, datname, blks_read, blks_hit, tup_returned, tup_fetched, stats_reset FROM pg_stat_database",
    "metrics": {
      "string": {
        "stats_reset": "2022-05-13T12:02:33.825483Z"
      },
      "numeric": {
        "blks_hit": 6360,
        "tup_returned": 2225,
        "tup_fetched": 1458,
        "datid": 0,
        "blks_read": 33
      }
    }
  },
  "ecs": {
    "version": "8.0.0"
  },
  "host": {
        "name": "mps"
    },
  "agent": {
    "version": "8.3.0",
    "ephemeral_id": "bc09584b-62db-4b45-bfe9-6b7e8e982361",
    "id": "6bbf5058-afed-44c6-aa05-775ee14a2da4",
    "name": "mps",
    "type": "metricbeat"
  }
}
```


### Example: Merge multiple queries into a single event. [_example_merge_multiple_queries_into_a_single_event]

Multiple queries will create multiple events, one for each query.  It may be preferable to create a single event by combining the metrics together in a single event.

This feature can be enabled using the `merge_results` config.

However, such a merge is possible only if the table queries are merged, each produces a single row.

For example:

```yaml
- module: sql
  metricsets:
    - query
  period: 10s
  hosts: ["postgres://postgres:postgres@localhost?sslmode=disable"]

  driver: "postgres"
  raw_data.enabled: true
  merge_results: true
  sql_queries:
    - query: "SELECT blks_hit,blks_read FROM pg_stat_database limit 1;"
      response_format: table
    - query: "select checkpoints_timed,checkpoints_req from pg_stat_bgwriter;"
      response_format: table
```

This creates a combined event as below, where `blks_hit`, `blks_read`, `checkpoints_timed` and `checkpoints_req` are part of same event.

```json
{
  "@timestamp": "2022-07-21T07:07:06.747Z",
  "agent": {
    "name": "MBP-2",
    "type": "metricbeat",
    "version": "8.4.0",
    "ephemeral_id": "b0867287-e56a-492f-b421-0ac870c426f9",
    "id": "3fe7b378-6f9e-4ca3-9aa1-067c4a6866e5"
  },
  "metricset": {
    "period": 10000,
    "name": "query"
  },
  "service": {
    "type": "sql",
    "address": "localhost"
  },
  "sql": {
    "metrics": {
      "blks_read": 21,
      "checkpoints_req": 1,
      "checkpoints_timed": 66,
      "blks_hit": 7592
    },
    "driver": "postgres"
  },
  "event": {
    "module": "sql",
    "duration": 18883084,
    "dataset": "sql.query"
  }
}
```


### Example: Execute given queries for all database(s) present in a server [_example_execute_given_queries_for_all_databases_present_in_a_server]

Assuming a user could have 100s of databases on their server and then it becomes cumbersome to add them manually to the query. If `fetch_from_all_databases` is set to `true` then SQL module would fetch the databases names automatically and prefix the database selector statement to the queries so that the queries can run against the database provided.

Currently, this feature only works with `mssql` driver. For example:

```yaml
- module: sql
  metricsets:
    - query
  period: 50s
  hosts: ["sqlserver://<user>:<password>@<host>"]
  raw_data.enabled: true

  fetch_from_all_databases: true

  driver: "mssql"
  sql_queries:
    - query: SELECT DB_NAME() AS 'database_name';
      response_format: table
```

For an mssql instance, by default only four databases are present namely — `master`, `model`, `msdb`, `tempdb`. So, if `fetch_from_all_databases` is enabled then query `SELECT DB_NAME() AS 'database_name'` runs for each one of them i.e., there would be in total 4 documents (one each for 4 databases) for every scrape.

```json
{
    "@timestamp": "2023-07-16T22:05:26.976Z",
    "@metadata": {
        "beat": "metricbeat",
        "type": "_doc",
        "version": "8.10.0"
    },
    "service": {
        "type": "sql",
        "address": "localhost"
    },
    "event": {
        "dataset": "sql.query",
        "module": "sql",
        "duration": 40346375
    },
    "metricset": {
        "name": "query",
        "period": 50000
    },
    "sql": {
        "metrics": {
            "database_name": "master"
        },
        "driver": "mssql",
        "query": "USE [master]; SELECT DB_NAME() AS 'database_name';"
    },
    "host": {
        "os": {
            "type": "macos",
            "platform": "darwin",
            "version": "13.3.1",
            "family": "darwin",
            "name": "macOS",
            "kernel": "<redacted>",
            "build": "<redacted>"
        },
        "name": "<redacted>",
        "id": "<redacted>",
        "ip": [
            "<redacted>"
        ],
        "mac": [
            "<redacted>"
        ],
        "hostname": "<redacted>",
        "architecture": "arm64"
    },
    "agent": {
        "name": "<redacted>",
        "type": "metricbeat",
        "version": "8.10.0",
        "ephemeral_id": "<redacted>",
        "id": "<redacted>"
    },
    "ecs": {
        "version": "8.0.0"
    }
}
{
    "@timestamp": "2023-07-16T22:05:26.976Z",
    "@metadata": {
        "beat": "metricbeat",
        "type": "_doc",
        "version": "8.10.0"
    },
    "agent": {
        "ephemeral_id": "<redacted>",
        "id": "<redacted>",
        "name": "<redacted>",
        "type": "metricbeat",
        "version": "8.10.0"
    },
    "event": {
        "module": "sql",
        "duration": 43147875,
        "dataset": "sql.query"
    },
    "metricset": {
        "period": 50000,
        "name": "query"
    },
    "service": {
        "address": "localhost",
        "type": "sql"
    },
    "sql": {
        "metrics": {
            "database_name": "tempdb"
        },
        "driver": "mssql",
        "query": "USE [tempdb]; SELECT DB_NAME() AS 'database_name';"
    },
    "ecs": {
        "version": "8.0.0"
    },
    "host": {
        "name": "<redacted>",
        "architecture": "arm64",
        "os": {
            "platform": "darwin",
            "version": "13.3.1",
            "family": "darwin",
            "name": "macOS",
            "kernel": "<redacted>",
            "build": "<redacted>",
            "type": "macos"
        },
        "id": "<redacted>",
        "ip": [
            "<redacted>"
        ],
        "mac": [
            "<redacted>"
        ],
        "hostname": "<redacted>"
    }
}
{
    "@timestamp": "2023-07-16T22:05:26.976Z",
    "@metadata": {
        "beat": "metricbeat",
        "type": "_doc",
        "version": "8.10.0"
    },
    "host": {
        "os": {
            "build": "<redacted>",
            "type": "macos",
            "platform": "darwin",
            "version": "13.3.1",
            "family": "darwin",
            "name": "macOS",
            "kernel": "<redacted>"
        },
        "id": "<redacted>",
        "ip": [
            "<redacted>"
        ],
        "mac": [
            "<redacted>"
        ],
        "hostname": "<redacted>",
        "name": "<redacted>",
        "architecture": "arm64"
    },
    "agent": {
        "ephemeral_id": "<redacted>",
        "id": "<redacted>",
        "name": "<redacted>",
        "type": "metricbeat",
        "version": "8.10.0"
    },
    "service": {
        "address": "localhost",
        "type": "sql"
    },
    "sql": {
        "metrics": {
            "database_name": "model"
        },
        "driver": "mssql",
        "query": "USE [model]; SELECT DB_NAME() AS 'database_name';"
    },
    "event": {
        "dataset": "sql.query",
        "module": "sql",
        "duration": 46623125
    },
    "metricset": {
        "name": "query",
        "period": 50000
    },
    "ecs": {
        "version": "8.0.0"
    }
}
{
    "@timestamp": "2023-07-16T22:05:26.976Z",
    "@metadata": {
        "beat": "metricbeat",
        "type": "_doc",
        "version": "8.10.0"
    },
    "host": {
        "architecture": "arm64",
        "os": {
            "kernel": "<redacted>",
            "build": "<redacted>",
            "type": "macos",
            "platform": "darwin",
            "version": "13.3.1",
            "family": "darwin",
            "name": "macOS"
        },
        "name": "<redacted>",
        "id": "<redacted>",
        "ip": [
            "<redacted>"
        ],
        "mac": [
            "<redacted>"
        ],
        "hostname": "<redacted>"
    },
    "agent": {
        "type": "metricbeat",
        "version": "8.10.0",
        "ephemeral_id": "<redacted>",
        "id": "<redacted>",
        "name": "<redacted>"
    },
    "event": {
        "dataset": "sql.query",
        "module": "sql",
        "duration": 49649250
    },
    "metricset": {
        "name": "query",
        "period": 50000
    },
    "service": {
        "address": "localhost",
        "type": "sql"
    },
    "sql": {
        "metrics": {
            "database_name": "msdb"
        },
        "driver": "mssql",
        "query": "USE [msdb]; SELECT DB_NAME() AS 'database_name';"
    },
    "ecs": {
        "version": "8.0.0"
    }
}
```

### Host Setup

Some drivers require additional configuration to work. Find here instructions for these drivers.

### Oracle Database Connection Pre-requisites

To get connected with the Oracle Database `ORACLE_SID`, `ORACLE_BASE`, `ORACLE_HOME` environment variables should be set.

For example: Let us consider Oracle Database 21c installation using RPM manually by following [this](https://docs.oracle.com/en/database/oracle/oracle-database/21/ladbi/running-rpm-packages-to-install-oracle-database.html) link, environment variables should be set as follows:

```bash
export ORACLE_BASE=/opt/oracle/oradata
export ORACLE_HOME=/opt/oracle/product/21c/dbhome_1
```

Also, add `ORACLE_HOME/bin` to the `PATH` environment variable. 

#### Oracle Instant Client Installation

Oracle Instant Client enables the development and deployment of applications that connect to the Oracle Database. The Instant Client libraries provide the necessary network connectivity and advanced data features to make full use of the Oracle Database. If you have an OCI Oracle server which comes with these libraries pre-installed, you don't need a separate client installation.

The OCI library installs a few Client Shared Libraries that must be referenced on the machine where Metricbeat is installed. Please follow [this](https://docs.oracle.com/en/database/oracle/oracle-database/21/lacli/install-instant-client-using-zip.html#GUID-D3DCB4FB-D3CA-4C25-BE48-3A1FB5A22E84) link for OCI Instant Client set up. The OCI Instant Client is available with the Oracle Universal Installer, RPM file or ZIP file. Download links can be found at here(https://www.oracle.com/database/technologies/instant-client/downloads.html).

##### Enable Oracle Listener

The Oracle listener is a service that runs on the database host and receives requests from Oracle clients. Make sure that [listener](https://docs.oracle.com/cd/B19306_01/network.102/b14213/lsnrctl.htm) should be running. 
To check if the listener is running or not, run: 

```bash
lsnrctl STATUS
```

If the listener is not running, use the command to start:

```bash
lsnrctl START
```

Then, Metricbeat can be launched.

##### Host Configuration for Oracle

The following types of host configuration are supported:

1. An old-style Oracle connection string, for backwards compatibility:

    a. `hosts: ["user/pass@0.0.0.0:1521/ORCLPDB1.localdomain"]`

    b. `hosts: ["user/password@0.0.0.0:1521/ORCLPDB1.localdomain as sysdba"]`

2. DSN configuration as a URL:

    a. `hosts: ["oracle://user:pass@0.0.0.0:1521/ORCLPDB1.localdomain?sysdba=1"]`

3. DSN configuration as a logfmt-encoded parameter list:

    a. `hosts: ['user="user" password="pass" connectString="0.0.0.0:1521/ORCLPDB1.localdomain"']`
    
    b. `hosts: ['user="user" password="password" connectString="host:port/service_name" sysdba=true']`

DSN host configuration is the recommended configuration type as it supports the use of special characters in the password.

In a URL any special characters should be URL encoded.

In the logfmt-encoded DSN format, if the password contains a backslash character (`\`), it must be escaped with another backslash. For example, if the password is `my\_password`, it must be written as `my\\_password`.

The username and password to connect to the database can be provided as values to the `username` and `password` keys of `sql.yml`.

```yaml
- module: sql
  metricsets:
    - query
  period: 10s
  driver: "oracle"
  enabled: true
  hosts: ['user="" password="" connectString="0.0.0.0:1521/ORCLCDB.localdomain" sysdba=true']
  username: sys
  password: password
  sql_queries: 
  - query: SELECT METRIC_NAME, VALUE FROM V$SYSMETRIC WHERE GROUP_ID = 2 and METRIC_NAME LIKE '%'
    response_format: variables 
```
----

### SSL Setup

The SSL configuration is driver-specific. Different drivers interpret parameters not in the same way. Subset of the [params](https://www.elastic.co/docs/reference/beats/metricbeat/configuration-ssl#ssl-client-config) is supported.

Currently, there are two ways to make SSL connections to the databases:

- Set `ssl.*` configuration parameters.
- Don't set any `ssl.*` configuration parameters and supply all SSL parameters in the connection string in `hosts`. Example: `postgres://postgres:mysecretpassword@localhost:5432?sslmode=verify-full&sslcert=%2Fpath%2Fto%2Fcert.pem&sslkey=%2Fpath%2Fto%2Fkey.pem&sslrootcert=%2Fpath%2Fto%2Fca.pem`

#### Limitations

The module supports SSL with `mysql`, `mssql`, and `postgres` drivers. 

When any `ssl.*` parameters are set, only URL-formatted connection strings are accepted, like `"postgres://myuser:mypassword@localhost:5432/mydb"`, not like `"user=myuser password=mypassword dbname=mydb"`.

##### `mysql` driver

Params supported: `ssl.verification_mode`, `ssl.certificate`, `ssl.key`, `ssl.certificate_authorities`.

The certificates can be passed both as file paths and as certificate content ([embedding certificate example](https://www.elastic.co/docs/reference/beats/metricbeat/configuration-ssl#client-certificate-authorities)).

##### `postgres` driver

Params supported: `ssl.verification_mode`, `ssl.certificate`, `ssl.key`, `ssl.certificate_authorities`.

Only one certificate can be passed to `ssl.certificate_authorities` parameter.
The certificates can be passed only as file paths. The files have to be present in the environment where the metricbeat is running.

The `ssl.verification_mode` is translated as following:

- `full` -> `verify-full`

- `strict` -> `verify-full`

- `certificate` -> `verify-ca`

- `none` -> `require`

##### `mssql` driver

Params supported: `ssl.verification_mode`, `ssl.certificate_authorities`.

Only one certificate can be passed to `ssl.certificate_authorities` parameter.
The certificates can be passed only as file paths. The files have to be present in the environment where the metricbeat is running.

If `ssl.verification_mode` is set to `None`, `TrustServerCertificate` will be set to `true`, otherwise it is `false`


## Example configuration [_example_configuration]

The SQL module supports the standard configuration options that are described in [Modules](/reference/metricbeat/configuration-metricbeat.md). Here is an example configuration:

```yaml
metricbeat.modules:
- module: sql
  metricsets:
    - query
  period: 10s
  hosts: ["postgres://postgres:mysecretpassword@localhost:5432"]
  # Example of using SSL parameters manually in the Postgres connection string (with ssl.* parameters unset). The Postgres SSL parameters "sslmode", "sslcert", "sslkey", and "sslrootcert" are passed in the connection string with slashes "/" being url-encoded to "%2F"
  # hosts: ["postgres://postgres:mysecretpassword@localhost:5432?sslmode=verify-full&sslcert=%2Fpath%2Fto%2Fcert.pem&sslkey=%2Fpath%2Fto%2Fkey.pem&sslrootcert=%2Fpath%2Fto%2Fca.pem"]
  # Example for SQL server
  # hosts: ["sqlserver://myuser:mypassword@localhost:1433?TrustServerCertificate=false&certificate=%2Fpath%2Fto%2Fca.pem&database=mydb&encrypt=true"]


  driver: "postgres"
  sql_query: "select now()"
  sql_response_format: table

  # List of root certificates for SSL/TLS server verification
  # ssl.certificate_authorities: ["/path/to/ca.pem"]

  # Certificate for SSL/TLS client authentication
  # ssl.certificate: "/path/to/client-cert.pem"

  # Client certificate key file
  # ssl.key: "/path/to/client-key.pem"

  # Controls the verification of server certificate
  # ssl.verification_mode: full
```


## Metricsets [_metricsets]

The following metricsets are available:

* [query](/reference/metricbeat/metricbeat-metricset-sql-query.md)
