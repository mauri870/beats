This is the partition metricset of the Kafka module.

## Configuration [_configuration_2]

As the partition metricset fetches the data from the complete Kafka cluster, only one connection host has to be defined. Currently if multiple hosts are defined, the data is fetched multiple times. Support for multiple initial connections host is planned to be added in future releases.


## Metricset [_metricset]

The current implementation of the partition metricset fetches the data for all leader partitions. Data for the replicas is not available yet.

