---
mapped_pages:
  - https://www.elastic.co/guide/en/beats/metricbeat/current/metricbeat-metricset-docker-network_summary.html
---

% This file is generated! See scripts/docs_collector.py

# Docker network_summary metricset [metricbeat-metricset-docker-network_summary]

::::{warning}
This functionality is in beta and is subject to change. The design and code is less mature than official GA features and is being provided as-is with no warranties. Beta features are not subject to the support SLA of official GA features.
::::


The `network_summary` metricset collects detailed network metrics from the processes associated with a container. These stats come from `/proc/[PID]/net`, and are summed across the different namespaces found across the PIDs.

Because this metricset will try to access network counters from procfs, it is only available on linux.

## Fields [_fields]

For a description of each field in the metricset, see the [exported fields](/reference/metricbeat/exported-fields-docker.md) section.

Here is an example document generated by this metricset:

```json
{
    "@timestamp": "2017-10-12T08:05:34.853Z",
    "container": {
        "id": "b1c3118fb2087e499c421aa20e28847ad4c185be6c1864166b1a03cc2affaf80",
        "image": {
            "name": "docker.elastic.co/beats/elastic-agent:7.13.0-SNAPSHOT"
        },
        "name": "elastic-package-stack_elastic-agent_1",
        "runtime": "docker"
    },
    "docker": {
        "container": {
            "labels": {
                "com_docker_compose_config-hash": "fdad7f67fafe5a242b9b9d27bff78bef4b96c683a2bb1412d91c32fb5981f2b3",
                "com_docker_compose_container-number": "1",
                "com_docker_compose_oneoff": "False",
                "com_docker_compose_project": "elastic-package-stack",
                "com_docker_compose_project_config_files": "/home/alexk/.elastic-package/stack/openport/snapshot.yml",
                "com_docker_compose_project_working_dir": "/home/alexk/.elastic-package/stack/openport",
                "com_docker_compose_service": "elastic-agent",
                "com_docker_compose_version": "1.27.4",
                "description": "Agent manages other beats based on configuration provided.",
                "io_k8s_description": "Agent manages other beats based on configuration provided.",
                "io_k8s_display-name": "Elastic-Agent image",
                "license": "Elastic License",
                "maintainer": "infra@elastic.co",
                "name": "elastic-agent",
                "org_label-schema_build-date": "2021-04-21T06:45:43Z",
                "org_label-schema_license": "Elastic License",
                "org_label-schema_name": "elastic-agent",
                "org_label-schema_schema-version": "1.0",
                "org_label-schema_url": "https://www.elastic.co/beats/elastic-agent",
                "org_label-schema_vcs-ref": "c52c43963ca8416dc92c7d3dbf6cb6e89dd00acf",
                "org_label-schema_vcs-url": "github.com/elastic/beats/v7",
                "org_label-schema_vendor": "Elastic",
                "org_label-schema_version": "7.13.0-SNAPSHOT",
                "org_opencontainers_image_created": "2021-04-21T06:45:43Z",
                "org_opencontainers_image_licenses": "Elastic License",
                "org_opencontainers_image_title": "Elastic-Agent",
                "org_opencontainers_image_vendor": "Elastic",
                "release": "1",
                "summary": "elastic-agent",
                "url": "https://www.elastic.co/beats/elastic-agent",
                "vendor": "Elastic",
                "version": "7.13.0-SNAPSHOT"
            }
        },
        "network_summary": {
            "icmp": {
                "InAddrMaskReps": 0,
                "InAddrMasks": 0,
                "InCsumErrors": 0,
                "InDestUnreachs": 0,
                "InEchoReps": 0,
                "InEchos": 0,
                "InErrors": 0,
                "InMsgs": 0,
                "InParmProbs": 0,
                "InRedirects": 0,
                "InSrcQuenchs": 0,
                "InTimeExcds": 0,
                "InTimestampReps": 0,
                "InTimestamps": 0,
                "OutAddrMaskReps": 0,
                "OutAddrMasks": 0,
                "OutDestUnreachs": 0,
                "OutEchoReps": 0,
                "OutEchos": 0,
                "OutErrors": 0,
                "OutMsgs": 0,
                "OutParmProbs": 0,
                "OutRedirects": 0,
                "OutSrcQuenchs": 0,
                "OutTimeExcds": 0,
                "OutTimestampReps": 0,
                "OutTimestamps": 0
            },
            "ip": {
                "DefaultTTL": 64,
                "ForwDatagrams": 0,
                "Forwarding": 1,
                "FragCreates": 0,
                "FragFails": 0,
                "FragOKs": 0,
                "InAddrErrors": 0,
                "InBcastOctets": 0,
                "InBcastPkts": 0,
                "InCEPkts": 0,
                "InCsumErrors": 0,
                "InDelivers": 1246323,
                "InDiscards": 0,
                "InECT0Pkts": 0,
                "InECT1Pkts": 0,
                "InHdrErrors": 0,
                "InMcastOctets": 0,
                "InMcastPkts": 0,
                "InNoECTPkts": 1246337,
                "InNoRoutes": 0,
                "InOctets": 172967508,
                "InReceives": 1246323,
                "InTruncatedPkts": 0,
                "InUnknownProtos": 0,
                "OutBcastOctets": 0,
                "OutBcastPkts": 0,
                "OutDiscards": 0,
                "OutMcastOctets": 0,
                "OutMcastPkts": 0,
                "OutNoRoutes": 0,
                "OutOctets": 2783525210,
                "OutRequests": 1550542,
                "ReasmFails": 0,
                "ReasmOKs": 0,
                "ReasmOverlaps": 0,
                "ReasmReqds": 0,
                "ReasmTimeout": 0
            },
            "namespace": {
                "ID": 4026532414,
                "pid": 2240728
            },
            "tcp": {
                "ActiveOpens": 742,
                "ArpFilter": 0,
                "AttemptFails": 0,
                "BusyPollRxPackets": 0,
                "CurrEstab": 12,
                "DelayedACKLocked": 0,
                "DelayedACKLost": 1,
                "DelayedACKs": 23466,
                "EmbryonicRsts": 0,
                "EstabResets": 0,
                "IPReversePathFilter": 0,
                "InCsumErrors": 0,
                "InErrs": 0,
                "InSegs": 1243467,
                "ListenDrops": 0,
                "ListenOverflows": 0,
                "LockDroppedIcmps": 0,
                "MaxConn": -1,
                "OfoPruned": 0,
                "OutOfWindowIcmps": 0,
                "OutRsts": 709,
                "OutSegs": 2936174,
                "PAWSActive": 0,
                "PAWSEstab": 19,
                "PFMemallocDrop": 0,
                "PassiveOpens": 4,
                "PruneCalled": 0,
                "RcvPruned": 0,
                "RetransSegs": 1613,
                "RtoAlgorithm": 1,
                "RtoMax": 120000,
                "RtoMin": 200,
                "SyncookiesFailed": 0,
                "SyncookiesRecv": 0,
                "SyncookiesSent": 0,
                "TCPACKSkippedChallenge": 0,
                "TCPACKSkippedFinWait2": 0,
                "TCPACKSkippedPAWS": 0,
                "TCPACKSkippedSeq": 0,
                "TCPACKSkippedSynRecv": 0,
                "TCPACKSkippedTimeWait": 0,
                "TCPAbortFailed": 0,
                "TCPAbortOnClose": 0,
                "TCPAbortOnData": 705,
                "TCPAbortOnLinger": 0,
                "TCPAbortOnMemory": 0,
                "TCPAbortOnTimeout": 0,
                "TCPAckCompressed": 0,
                "TCPAutoCorking": 763,
                "TCPBacklogCoalesce": 1358,
                "TCPBacklogDrop": 0,
                "TCPChallengeACK": 0,
                "TCPDSACKIgnoredDubious": 0,
                "TCPDSACKIgnoredNoUndo": 1498,
                "TCPDSACKIgnoredOld": 0,
                "TCPDSACKOfoRecv": 0,
                "TCPDSACKOfoSent": 0,
                "TCPDSACKOldSent": 1,
                "TCPDSACKRecv": 1586,
                "TCPDSACKRecvSegs": 1589,
                "TCPDSACKUndo": 2,
                "TCPDeferAcceptDrop": 0,
                "TCPDelivered": 2275375,
                "TCPDeliveredCE": 0,
                "TCPFastOpenActive": 0,
                "TCPFastOpenActiveFail": 0,
                "TCPFastOpenBlackhole": 0,
                "TCPFastOpenCookieReqd": 0,
                "TCPFastOpenListenOverflow": 0,
                "TCPFastOpenPassive": 0,
                "TCPFastOpenPassiveAltKey": 0,
                "TCPFastOpenPassiveFail": 0,
                "TCPFastRetrans": 7,
                "TCPFromZeroWindowAdv": 0,
                "TCPFullUndo": 1,
                "TCPHPAcks": 537452,
                "TCPHPHits": 144087,
                "TCPHystartDelayCwnd": 0,
                "TCPHystartDelayDetect": 0,
                "TCPHystartTrainCwnd": 36,
                "TCPHystartTrainDetect": 2,
                "TCPKeepAlive": 171573,
                "TCPLossFailures": 0,
                "TCPLossProbeRecovery": 0,
                "TCPLossProbes": 2183,
                "TCPLossUndo": 0,
                "TCPLostRetransmit": 0,
                "TCPMD5Failure": 0,
                "TCPMD5NotFound": 0,
                "TCPMD5Unexpected": 0,
                "TCPMTUPFail": 0,
                "TCPMTUPSuccess": 0,
                "TCPMemoryPressures": 0,
                "TCPMemoryPressuresChrono": 0,
                "TCPMinTTLDrop": 0,
                "TCPOFODrop": 0,
                "TCPOFOMerge": 0,
                "TCPOFOQueue": 0,
                "TCPOrigDataSent": 2273068,
                "TCPPartialUndo": 1,
                "TCPPureAcks": 215589,
                "TCPRcvCoalesce": 10,
                "TCPRcvCollapsed": 0,
                "TCPRcvQDrop": 0,
                "TCPRenoFailures": 0,
                "TCPRenoRecovery": 0,
                "TCPRenoRecoveryFail": 0,
                "TCPRenoReorder": 0,
                "TCPReqQFullDoCookies": 0,
                "TCPReqQFullDrop": 0,
                "TCPRetransFail": 0,
                "TCPSACKDiscard": 0,
                "TCPSACKReneging": 0,
                "TCPSACKReorder": 256,
                "TCPSYNChallenge": 0,
                "TCPSackFailures": 0,
                "TCPSackMerged": 5,
                "TCPSackRecovery": 4,
                "TCPSackRecoveryFail": 0,
                "TCPSackShiftFallback": 160,
                "TCPSackShifted": 0,
                "TCPSlowStartRetrans": 0,
                "TCPSpuriousRTOs": 0,
                "TCPSpuriousRtxHostQueues": 2,
                "TCPSynRetrans": 24,
                "TCPTSReorder": 1,
                "TCPTimeWaitOverflow": 0,
                "TCPTimeouts": 24,
                "TCPToZeroWindowAdv": 0,
                "TCPWantZeroWindowAdv": 0,
                "TCPWinProbe": 0,
                "TCPWqueueTooBig": 0,
                "TCPZeroWindowDrop": 0,
                "TW": 3,
                "TWKilled": 0,
                "TWRecycled": 0,
                "TcpDuplicateDataRehash": 0,
                "TcpTimeoutRehash": 24
            },
            "udp": {
                "IgnoredMulti": 0,
                "InCsumErrors": 0,
                "InDatagrams": 2856,
                "InErrors": 0,
                "NoPorts": 0,
                "OutDatagrams": 2856,
                "RcvbufErrors": 0,
                "SndbufErrors": 0
            },
            "udp_lite": {
                "IgnoredMulti": 0,
                "InCsumErrors": 0,
                "InDatagrams": 0,
                "InErrors": 0,
                "NoPorts": 0,
                "OutDatagrams": 0,
                "RcvbufErrors": 0,
                "SndbufErrors": 0
            }
        }
    },
    "event": {
        "dataset": "docker.network_summary",
        "duration": 115000,
        "module": "docker"
    },
    "metricset": {
        "name": "network_summary",
        "period": 10000
    },
    "service": {
        "address": "unix:///var/run/docker.sock",
        "type": "docker"
    }
}
```
