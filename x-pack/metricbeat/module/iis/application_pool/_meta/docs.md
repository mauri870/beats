This is the application_pool metricset of the module iis.

This metricset allows users to retrieve relevant metrics for the application pools running on IIS.

Metric values are divided in several groups:

The `process` object contains System/Process counters like the the overall server and CPU usage for the IIS Worker Process and memory (currently used and available memory for the IIS Worker Process).

The `net_clr` object which returns ASP.NET error rate counter values. Users can specify the application pools they would like to monitor using the configuration option `application_pool.name`, else, all application pools are considered.


### Dashboard [_dashboard_27]

![metricbeat iis application pool overview](images/metricbeat-iis-application-pool-overview.png)
