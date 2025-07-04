The System module allows you to monitor your servers. Because the System module always applies to the local server, the `hosts` config option is not needed.

The default metricsets are `cpu`, `load`, `memory`, `network`, `process`, `process_summary`, `socket_summary`, `filesystem`, `fsstat`, and `uptime`. To disable a default metricset, comment it out in the `modules.d/system.yml` configuration file. If *all* metricsets are commented out and the System module is enabled, Metricbeat uses the default metricsets.

Note that certain metricsets may access `/proc` to gather process information, and the resulting `ptrace_may_access()` call by the kernel to check for permissions can be blocked by [AppArmor and other LSM software](https://gitlab.com/apparmor/apparmor/wikis/TechnicalDoc_Proc_and_ptrace), even though the System module doesn’t use `ptrace` directly.

::::{admonition} How and when metrics are collected
:class: tip

Certain metrics monitored by the System module require multiple values to be collected. For example, the `system.process.cpu.total.norm.pct` field reports the percentage of CPU time spent by the process since the last event. For this percentage to be determined, the process needs to appear at least twice so that a performance delta can be calculated.

Note that in some cases a field like this may be missing from the System module metricset if the process has not been available long enough to be included in two periods of metric collection.

::::



## Dashboard [_dashboard_42] 

The System module comes with a predefined dashboard. For example:

:::{image} images/metricbeat_system_dashboard.png
:alt: metricbeat system dashboard
:::


## Required permissions [_required_permissions] 

The System metricsets collect different kinds of metric data, which may require dedicated permissions to be fetched. For security reasons it’s advised to grant the lowest possible permissions. This section justifies which permissions must be present for particular metricsets.

Please notice that modern Linux implementations divide the privileges traditionally associated with superuser into distinct units, known as capabilities, which can be independently enabled and disabled. Capabilities are a per-thread attribute.


### cpu [_cpu] 

CPU statistics (idle, irq, user, system, iowait, softirq, cores, nice, steal, total) should be available without elevated permissions.


### load [_load] 

CPU load data (1 min, 5 min, 15 min, cores) should be available without elevated permissions.


### memory [_memory] 

Memory statistics (swap, total, used, free, actual) should be available without elevated permissions.


### network [_network] 

Network metrics for interfaces (in, out, errors, dropped, bytes, packets) should be available without elevated permissions.


### process [_process] 

Process execution data (state, memory, cpu, cmdline) should be available for an authorized user.

If the beats process is running as less privileged user, it may not be able to read process data belonging to other users. The issue should be reported in application logs:

```
2019-12-23T13:32:06.457+0100    DEBUG   [processes]     process/process.go:475  Skip process pid=235: error getting process state for pid=235: Could not read process info for pid 23
```


### process_summary [_process_summary] 

General process summary (unknown, dead, total, sleeping, running, idle, stopped, zombie) should be available without elevated permissions. Please notice that if the process data belongs to the other users, it will be counted as unknown value (no error will be reported in application logs).


### socket_summary [_socket_summary] 

Used sockets summary (TCP, UDP, count, listening, established, wait, etc.) should be available without elevated permissions.


### entropy [_entropy] 

Entropy data (available, pool size) requires access to the `/proc/sys/kernel/random` path. Otherwise an error will be reported.


### core [_core] 

Usage statistics for each CPU core (idle, irq, user, system, iowait, softirq, cores, nice, steal, total) should be available without elevated permissions.


### diskio [_diskio] 

Disk IO metrics (io, read, write) should be available without elevated permissions.


### socket [_socket] 

Events for each new TCP socket should be available for an authorized user.

If the beats process is running as less privileged user, it may not be able to view socket data belonging to other users.


### service [_service] 

Systemd service data (memory, tasks, states) should be available for an authorized user.

If the beats process is running as less privileged user, it may not be able to read process data belonging to other users. The issue should be reported in application logs:

```
2020-01-02T08:19:50.635Z	INFO	module/wrapper.go:252	Error fetching data for metricset system.service: error getting list of running units: Rejected send message, 2 matched rules; type="method_call", sender=":1.35" (uid=1000 pid=4429 comm="./metricbeat -d * -e ") interface="org.freedesktop.systemd1.Manager" member="ListUnitsByPatterns" error name="(unset)" requested_reply="0" destination="org.freedesktop.systemd1" (uid=0 pid=1 comm="/usr/lib/systemd/systemd --switched-root --system ")
```


### filesystem [_filesystem_2] 

Filesystem metrics data (total, available, type, mount point, files, free, used) should be available without elevated permissions.


### fsstat [_fsstat] 

Fsstat metrics data (total size, free, total, used count) should be available without elevated permissions.


### uptime [_uptime] 

Uptime metrics data (duration) should be available without elevated permissions.


### raid [_raid] 

RAID metrics data (block, disks) requires access to the `/sys/block` mount point and all referenced devices. Otherwise an error will be reported.
