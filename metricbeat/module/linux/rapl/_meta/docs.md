::::{warning}
This functionality is in beta and is subject to change. The design and code is less mature than official GA features and is being provided as-is with no warranties. Beta features are not subject to the support SLA of official GA features.
::::


The RAPL metricset reports power usage in Watts as reported by Intel’s RAPL (Running Average Power Limit) API. In practice, RAPL is used for setting power usage limits on a CPU, but it can also be used for monitoring power usage.

RAPL is available on most Intel CPUs in some form since Sandy Bridge, however, this implementation requires a linux distro that exposes intel MSRs (model-specific registers) via `/dev/cpu/[CPU]/msr`. You can check to see if your linux distro and CPU support RAPL by running

```
sudo rdmsr 0x606
```

If that command returns a hex value such as `a1003`, RAPL is supported.

RAPL divides a CPU into four domains, each one corresponding to a component within the hardware itself: `Package`, `DRAM`, `PP0` (usually the collective power usage of the cores), and `PP1` (usually the integrated GPU). Not all packages are available on all CPUs, and the RAPL metricset will report all RAPL domains that it finds.


## msr-safe usage [_msr_safe_usage]

This metricset also supports [msr-safe](https://github.com/LLNL/msr-safe), which allows a user to access an MSR device without root. The `rapl.use_msr_safe` flag in the linux module config will enable usage of the `/dev/cpu/[CPU]/msr_safe` device.

For existing msr-safe installs, the following allowlist will open all RAPL MSRs for reading:

```
0x00000606 0x0000000000000000 # "SMSR_RAPL_POWER_UNIT"

0x00000610 0x0000000000000000 # "SMSR_PACKAGE_POWER_LIMIT"
0x00000611 0x0000000000000000 # "SMSR_PACKAGE_ENERGY_STATUS"
0x00000613 0x0000000000000000 # "SMSR_PACKAGE_PERF_STATUS"
0x00000614 0x0000000000000000 # "SMSR_PACKAGE_POWER_INFO"

0x00000618 0x0000000000000000 # "SMSR_DRAM_POWER_LIMIT"
0x00000619 0x0000000000000000 # "SMSR_DRAM_ENERGY_STATUS"
0x0000061b 0x0000000000000000 # "SMSR_DRAM_PERF_STATUS"
0x0000061c 0x0000000000000000 # "SMSR_DRAM_POWER_INFO"

0x00000638 0x0000000000000000 # "SMSR_PP0_POWER_LIMIT"
0x00000639 0x0000000000000000 # "SMSR_PP0_ENERGY_STATUS"
0x0000063a 0x0000000000000000 # "SMSR_PP0_POLICY"
0x0000063b 0x0000000000000000 # "SMSR_PP0_PERF_STATUS"

0x00000640 0x0000000000000000 # "SMSR_PP1_POWER_LIMIT"
0x00000641 0x0000000000000000 # "SMSR_PP1_ENERGY_STATUS"
0x00000642 0x0000000000000000 # "SMSR_PP1_POLICY"
```
