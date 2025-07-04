This is the webserver metricset of the module iis.

This metricset allows users to retrieve relevant metrics from IIS.

Metric values are divided in several groups:

The `process` object contains System/Process counters like the the overall server and CPU usage for the IIS Worker Processes and memory (currently used and available memory for the IIS Worker Processes).

The `network` object contains the IIS Performance counters like: Web Service: Bytes Received/Sec (helpful to track to identify potential spikes in traffic), Web Service: Bytes Sent/Sec (helpful to track to identify potential spikes in traffic), Web Service: Current Connections (through experience with their apps app, users can identify what is a normal value for this) and others.

The `cache` object contains metrics from the user mode cache and output cache.

The `asp_net` and `asp_net_application` contain asp.net related performance counter values.


### Dashboard [_dashboard_28]

![metricbeat iis webserver overview](images/metricbeat-iis-webserver-overview.png)

![metricbeat iis webserver process](images/metricbeat-iis-webserver-process.png)
