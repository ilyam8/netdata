<!--
title: "proc.plugin"
custom_edit_url: https://github.com/netdata/netdata/edit/master/collectors/proc.plugin/README.md
-->

# proc.plugin

- `/proc/net/dev` (all network interfaces for all their values)
- `/proc/diskstats` (all disks for all their values)
- `/proc/mdstat` (status of RAID arrays)
- `/proc/net/snmp` (total IPv4, TCP and UDP usage)
- `/proc/net/snmp6` (total IPv6 usage)
- `/proc/net/netstat` (more IPv4 usage)
- `/proc/net/wireless` (wireless extension)
- `/proc/net/stat/nf_conntrack` (connection tracking performance)
- `/proc/net/stat/synproxy` (synproxy performance)
- `/proc/net/ip_vs/stats` (IPVS connection statistics)
- `/proc/stat` (CPU utilization and attributes)
- `/proc/meminfo` (memory information)
- `/proc/vmstat` (system performance)
- `/proc/net/rpc/nfsd` (NFS server statistics for both v3 and v4 NFS servers)
- `/sys/fs/cgroup` (Control Groups - Linux Containers)
- `/proc/self/mountinfo` (mount points)
- `/proc/interrupts` (total and per core hardware interrupts)
- `/proc/softirqs` (total and per core software interrupts)
- `/proc/loadavg` (system load and total processes running)
- `/proc/pressure/{cpu,memory,io}` (pressure stall information)
- `/proc/sys/kernel/random/entropy_avail` (random numbers pool availability - used in cryptography)
- `/proc/spl/kstat/zfs/arcstats` (status of ZFS adaptive replacement cache)
- `/proc/spl/kstat/zfs/pool/state` (state of ZFS pools)
- `/sys/class/power_supply` (power supply properties)
- `/sys/class/infiniband` (infiniband interconnect)
- `ipc` (IPC semaphores and message queues)
- `ksm` Kernel Same-Page Merging performance (several files under `/sys/kernel/mm/ksm`).
- `netdata` (internal Netdata resources utilization)

---

## Metrics

| Metric                                                           |      Scope       |                                                                                                                                           Dimensions                                                                                                                                           |     Units      |
|------------------------------------------------------------------|:----------------:|:----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------:|:--------------:|
| system.cpu.utilization.perc                                      |      global      |                                                                                                            guest_nice, guest, steal, softirq, irq, user, system, nice, iowait, idle                                                                                                            |   percentage   |
| system.cpu.pressure.some.perc                                    |      global      |                                                                                                                                      10sec, 60sec, 300sec                                                                                                                                      |   percentage   |
| system.cpu.pressure.some.time                                    |      global      |                                                                                                                                              time                                                                                                                                              |    seconds     |
| system.cpu.pressure.full.perc                                    |      global      |                                                                                                                                      10sec, 60sec, 300sec                                                                                                                                      |   percentage   |
| system.cpu.pressure.full.time                                    |      global      |                                                                                                                                              time                                                                                                                                              |    seconds     |
| system.cpu.context_switches.rate                                 |      global      |                                                                                                                                            switches                                                                                                                                            |   switches/s   |
| system.cpu.interrupts.rate                                       |      global      |                                                                                                                                           interrupts                                                                                                                                           |  interrupts/s  |
| system.cpu.interrupts.per_device.rate                            |      global      |                                                                                                                                 <i>a dimension per device</i>                                                                                                                                  |  interrupts/s  |
| system.cpu.softirqs.per_tasklet.rate                             |      global      |                                                                                                                                 <i>a dimension per tasklet</i>                                                                                                                                 |   softirqs/s   |
| system.cpu.core.utilization.perc                                 |       core       |                                                                                                            guest_nice, guest, steal, softirq, irq, user, system, nice, iowait, idle                                                                                                            |   percentage   |
| system.cpu.core.speed.frequency.num                              |       core       |                                                                                                                                              freq                                                                                                                                              |       Hz       |
| system.cpu.core.speed.throttling.rate                            |       core       |                                                                                                                                         core, package                                                                                                                                          |    events/s    |
| system.cpu.core.cstate.perc                                      |       core       |                                                                                                                                 <i>a dimension per c-state</i>                                                                                                                                 |   percentage   |
| system.cpu.core.interrupts.per_device.rate                       |       core       |                                                                                                                                 <i>a dimension per device</i>                                                                                                                                  |  interrupts/s  |
| system.cpu.core.softirqs.per_tasklet.rate                        |       core       |                                                                                                                                 <i>a dimension per tasklet</i>                                                                                                                                 |   softirqs/s   |
| system.memory.ram.usage.size                                     |      global      |                                                                                                                                  free, used, cached, buffers                                                                                                                                   |     bytes      |
| system.memory.available.size                                     |      global      |                                                                                                                                           available                                                                                                                                            |     bytes      |
| system.memory.committed.size                                     |      global      |                                                                                                                                          Committed_AS                                                                                                                                          |     bytes      |
| system.memory.kernel.size                                        |      global      |                                                                                                                       Slab, KernelStack, PageTables, VmallocUsed, Percpu                                                                                                                       |     bytes      |
| system.memory.pressure.some.perc                                 |      global      |                                                                                                                                      10sec, 60sec, 300sec                                                                                                                                      |   percentage   |
| system.memory.pressure.some.time                                 |      global      |                                                                                                                                              time                                                                                                                                              |    seconds     |
| system.memory.pressure.full.perc                                 |      global      |                                                                                                                                      10sec, 60sec, 300sec                                                                                                                                      |   percentage   |
| system.memory.pressure.full.time                                 |      global      |                                                                                                                                              time                                                                                                                                              |    seconds     |
| system.memory.mgmt.paging.swap.usage.size                        |      global      |                                                                                                                                           free, used                                                                                                                                           |     bytes      |
| system.memory.mgmt.paging.swap.io.rate                           |      global      |                                                                                                                                            in, out                                                                                                                                             |    bytes/s     |
| system.memory.mgmt.paging.io.rate                                |      global      |                                                                                                                                            in, out                                                                                                                                             |    bytes/s     |
| system.memory.mgmt.paging.faults.rate                            |      global      |                                                                                                                                          minor, major                                                                                                                                          |    faults/s    |
| system.memory.mgmt.paging.writeback.size                         |      global      |                                                                                                                     Dirty, Writeback, FuseWriteback, NfsWriteback, Bounce                                                                                                                      |     bytes      |
| system.memory.mgmt.ksm.size                                      |      global      |                                                                                                                              shared, unshared, sharing, volatile                                                                                                                               |     bytes      |
| system.memory.mgmt.ksm.saved.perc                                |      global      |                                                                                                                                             saved                                                                                                                                              |   percentage   |
| system.memory.mgmt.slab.size                                     |      global      |                                                                                                                                   reclaimable, unreclaimable                                                                                                                                   |     bytes      |
| system.memory.mgmt.hugepages.size                                |      global      |                                                                                                                                 free, used, surplus, reserved                                                                                                                                  |     bytes      |
| system.memory.mgmt.transparent_hugepages.size                    |      global      |                                                                                                                                 free, used, surplus, reserved                                                                                                                                  |     bytes      |
| system.memory.mgmt.numa.events.rate                              |      global      |                                                                                         local, foreign, interleave, otherpte_updates, huge_pte_updates, hint_faults, hint_faults_local, pages_migrated                                                                                         |    events/s    |
| system.memory.mgmt.numa.node.events.rate                         |    numa node     |                                                                                                                          hit, miss, local, foreign, interleave, other                                                                                                                          |    events/s    |
| system.memory.mgmt.numa.node.zone.type.per_page.size             | node, zone, type |                                                                                                                                <i>a dimension per page size</i>                                                                                                                                |     bytes      |
| system.memory.zram.size                                          |      global      |                                                                                                                                      compressed, metadata                                                                                                                                      |     bytes      |
| system.memory.zram.saved.size                                    |      global      |                                                                                                                                       savings, original                                                                                                                                        |     bytes      |
| system.memory.zram.saved.perc                                    |      global      |                                                                                                                                       savings, original                                                                                                                                        |   percentage   |
| system.memory.ecc.mc.errors.correctable.rate                     |  mem controller  |                                                                                                                                             errors                                                                                                                                             |    errors/s    |
| system.memory.ecc.mc.errors.uncorrectable.rate                   |  mem controller  |                                                                                                                                             errors                                                                                                                                             |    errors/s    |
| system.network.traffic.devices.rate                              |      global      |                                                                                                                                         received, sent                                                                                                                                         |     bits/s     |
| system.network.traffic.ipv4.rate                                 |      global      |                                                                                                                                         received, sent                                                                                                                                         |     bits/s     |
| system.network.traffic.ipv6.rate                                 |      global      |                                                                                                                                         received, sent                                                                                                                                         |     bits/s     |
| system.network.traffic.multicast.ipv4.rate                       |      global      |                                                                                                                                         received, sent                                                                                                                                         |     bits/s     |
| system.network.traffic.multicast.ipv6.rate                       |      global      |                                                                                                                                         received, sent                                                                                                                                         |     bits/s     |
| system.network.traffic.broadcast.ipv4.rate                       |      global      |                                                                                                                                         received, sent                                                                                                                                         |     bits/s     |
| system.network.traffic.broadcast.ipv6.rate                       |      global      |                                                                                                                                         received, sent                                                                                                                                         |     bits/s     |
| system.network.packets.ipv4.rate                                 |      global      |                                                                                                                              received, sent, forwarded, delivered                                                                                                                              |   packets/s    |
| system.network.packets.ipv6.rate                                 |      global      |                                                                                                                              received, sent, forwarded, delivered                                                                                                                              |   packets/s    |
| system.network.packets.multicast.ipv4.rate                       |      global      |                                                                                                                                         received, sent                                                                                                                                         |   packets/s    |
| system.network.packets.multicast.ipv6.rate                       |      global      |                                                                                                                                         received, sent                                                                                                                                         |   packets/s    |
| system.network.packets.broadcast.ipv4.rate                       |      global      |                                                                                                                                         received, sent                                                                                                                                         |   packets/s    |
| system.network.errors.ipv4.rate                                  |      global      |                                                                                   InDiscards, OutDiscards, InHdrErrors, OutNoRoutes, InAddrErrors, InUnknownProtos, InNoRoutes, InTruncatedPkt, InCsumErrors                                                                                   |   packets/s    |
| system.network.errors.ipv6.rate                                  |      global      |                                                                                 InDiscards, OutDiscards, InHdrErrors, InNoRoutes, OutNoRoutes, InAddrErrors, InUnknownProtos, InTooBigErrors, InTruncatedPkts                                                                                  |   packets/s    |
| system.network.device.traffic.rate                               |  network device  |                                                                                                                                         received, sent                                                                                                                                         |     bits/s     |
| system.network.device.packets.rate                               |  network device  |                                                                                                                                         received, sent                                                                                                                                         |   packets/s    |
| system.network.device.packets.multicast.rate                     |  network device  |                                                                                                                                            received                                                                                                                                            |   packets/s    |
| system.network.device.drops.rate                                 |  network device  |                                                                                                                                       inbound, outbound                                                                                                                                        |    drops/s     |
| system.network.device.errors.rate                                |  network device  |                                                                                                                                       inbound, outbound                                                                                                                                        |    errors/s    |
| system.network.device.errors.collisions.rate                     |  network device  |                                                                                                                                           collisions                                                                                                                                           |  collisions/s  |
| system.network.device.errors.frame.rate                          |  network device  |                                                                                                                                             errors                                                                                                                                             |    errors/s    |
| system.network.device.errors.carrier.rate                        |  network device  |                                                                                                                                             errors                                                                                                                                             |    errors/s    |
| system.network.device.errors.fifo.rate                           |  network device  |                                                                                                                                       inbound, outbound                                                                                                                                        |    errors/s    |
| system.network.device.compressed.rate                            |  network device  |                                                                                                                                         received, sent                                                                                                                                         |   packets/s    |
| system.network.device.speed.num                                  |  network device  |                                                                                                                                             speed                                                                                                                                              |     bits/s     |
| system.network.device.duplex.state                               |  network device  |                                                                                                                                      half, full, unknown                                                                                                                                       |     state      |
| system.network.device.operstate.state                            |  network device  |                                                                                                                up, down, notpresent, lowerlayerdown, testing, dormant, unknown                                                                                                                 |     state      |
| system.network.device.carrier.state                              |  network device  |                                                                                                                                            up, down                                                                                                                                            |     state      |
| system.network.device.mtu.size                                   |  network device  |                                                                                                                                              mtu                                                                                                                                               |     octets     |
| system.network.wireless.device.status.num                        |  network device  |                                                                                                                                             status                                                                                                                                             |     status     |
| system.network.wireless.device.link_quality.num                  |  network device  |                                                                                                                                          link_quality                                                                                                                                          |    quality     |
| system.network.wireless.device.signal_level.num                  |  network device  |                                                                                                                                          signal_level                                                                                                                                          |      dBm       |
| system.network.wireless.device.noise_level.num                   |  network device  |                                                                                                                                          noise_level                                                                                                                                           |      dBm       |
| system.network.wireless.device.discards.rate                     |  network device  |                                                                                                                                 nwid, crypt, frag, retry, misc                                                                                                                                 |   packets/s    |
| system.network.wireless.device.beacons.loss.rate                 |  network device  |                                                                                                                                             missed                                                                                                                                             |   beacons/s    |
| system.network.infiniband.port.traffic.rate                      |       port       |                                                                                                                                         Received, Sent                                                                                                                                         |     bits/s     |
| system.network.infiniband.port.packets.rate                      |       port       |                                                                                                                                         Received, Sent                                                                                                                                         |   packets/s    |
| system.network.infiniband.port.packets.multicast.rate            |       port       |                                                                                                                                     Mcast rcvd, Mcast sent                                                                                                                                     |   packets/s    |
| system.network.infiniband.port.packets.unicast.rate              |       port       |                                                                                                                                     Ucast rcvd, Ucast sent                                                                                                                                     |   packets/s    |
| system.network.infiniband.port.errors.rate                       |       port       |           Pkts_malformated, Pkts_rcvd_discarded, Pkts_sent_discarded, Tick_Wait_to_send, Pkts_missed_resource, Buffer_overrun, Link_Downed, Link_recovered, Link_integrity_err, Link_minor_errors, Pkts_rcvd_with_EBP, Pkts_rcvd_discarded_by_switch, Pkts_sent_discarded_by_switch            |    errors/s    |
| system.network.protocol.tcp.packets.rate                         |      global      |                                                                                                                                         received, sent                                                                                                                                         |   packets/s    |
| system.network.protocol.tcp.errors.rate                          |      global      |                                                                                                                               InErrs, InCsumErrors, RetransSegs                                                                                                                                |   packets/s    |
| system.network.protocol.tcp.sockets.ipv4.count                   |      global      |                                                                                                                                 alloc, orphan, inuse, timewait                                                                                                                                 |    sockets     |
| system.network.protocol.tcp.sockets.ipv6.count                   |      global      |                                                                                                                                             inuse                                                                                                                                              |    sockets     |
| system.network.protocol.tcp.sockets.memory.ipv4.size             |      global      |                                                                                                                                           allocated                                                                                                                                            |     bytes      |
| system.network.protocol.tcp.memory_pressure.rate                 |      global      |                                                                                                                                           pressures                                                                                                                                            |    events/s    |
| system.network.protocol.tcp.conn_aborts.rate                     |      global      |                                                                                                                     baddata, userclosed, nomemory, timeout, linger, failed                                                                                                                     | connections/s  |
| system.network.protocol.tcp.reorders.rate                        |      global      |                                                                                                                                  timestamp, sack, fack, reno                                                                                                                                   |   packets/s    |
| system.network.protocol.tcp.out_of_order.rate                    |      global      |                                                                                                                                inqueue, dropped, merged, pruned                                                                                                                                |   packets/s    |
| system.network.protocol.tcp.syn_cookies.rate                     |      global      |                                                                                                                                     received, sent, failed                                                                                                                                     |   packets/s    |
| system.network.protocol.tcp.syn_queue.issues.rate                |      global      |                                                                                                                                         drops, cookies                                                                                                                                         |   packets/s    |
| system.network.protocol.tcp.accept_queue.issues.rate             |      global      |                                                                                                                                        overflows, drops                                                                                                                                        |   packets/s    |
| system.network.protocol.tcp.opens.rate                           |      global      |                                                                                                                                        active, passive                                                                                                                                         | connections/s  |
| system.network.protocol.tcp.handshake.rate                       |      global      |                                                                                                                         EstabResets, OutRsts, AttemptFails, SynRetrans                                                                                                                         |    events/s    |
| system.network.protocol.sctp.transitions.rate                    |      global      |                                                                                                                               active, passive, aborted, shutdown                                                                                                                               | transitions/s  |
| system.network.protocol.sctp.packets.rate                        |      global      |                                                                                                                                         received, sent                                                                                                                                         |   packets/s    |
| system.network.protocol.sctp.errors.rate                         |      global      |                                                                                                                                       invalid, checksum                                                                                                                                        |   packets/s    |
| system.network.protocol.sctp.fragmentation.rate                  |      global      |                                                                                                                                    reassembled, fragmented                                                                                                                                     |   packets/s    |
| system.network.protocol.sctp.chunks.rate                         |      global      |                                                                                                                   InCtrl, InOrder, InUnorder, OutCtrl, OutOrder, OutUnorder                                                                                                                    |    chunks/s    |
| system.network.protocol.udp.packets.ipv4.rate                    |      global      |                                                                                                                                         received, sent                                                                                                                                         |   packets/s    |
| system.network.protocol.udp.packets.ipv6.rate                    |      global      |                                                                                                                                         received, sent                                                                                                                                         |   packets/s    |
| system.network.protocol.udp.errors.ipv4.rate                     |      global      |                                                                                                           RcvbufErrors, SndbufErrors, InErrors, NoPorts, InCsumErrors, IgnoredMulti                                                                                                            |    errors/s    |
| system.network.protocol.udp.errors.ipv6.rate                     |      global      |                                                                                                           RcvbufErrors, SndbufErrors, InErrors, NoPorts, InCsumErrors, IgnoredMulti                                                                                                            |    events/s    |
| system.network.protocol.udp.sockets.ipv4.count                   |      global      |                                                                                                                                             inuse                                                                                                                                              |    sockets     |
| system.network.protocol.udp.sockets.ipv6.count                   |      global      |                                                                                                                                             inuse                                                                                                                                              |    sockets     |
| system.network.protocol.udp.sockets.memory.ipv4.size             |      global      |                                                                                                                                           allocated                                                                                                                                            |     bytes      |
| system.network.protocol.udplite.packets.ipv4.rate                |      global      |                                                                                                                                         received, sent                                                                                                                                         |   packets/s    |
| system.network.protocol.udplite.packets.ipv6.rate                |      global      |                                                                                                                                         received, sent                                                                                                                                         |   packets/s    |
| system.network.protocol.udplite.errors.ipv4.rate                 |      global      |                                                                                                           RcvbufErrors, SndbufErrors, InErrors, NoPorts, InCsumErrors, IgnoredMulti                                                                                                            |    errors/s    |
| system.network.protocol.udplite.errors.ipv6.rate                 |      global      |                                                                                                                  RcvbufErrors, SndbufErrors, InErrors, NoPorts, InCsumErrors                                                                                                                   |   packets/s    |
| system.network.protocol.udplite.sockets.ipv4.count               |      global      |                                                                                                                                             inuse                                                                                                                                              |    sockets     |
| system.network.protocol.udplite.sockets.ipv6.count               |      global      |                                                                                                                                             inuse                                                                                                                                              |    sockets     |
| system.network.protocol.icmp.packets.ipv4.rate                   |      global      |                                                                                                                                         received, sent                                                                                                                                         |   packets/s    |
| system.network.protocol.icmp.packets.ipv6.rate                   |      global      |                                                                                                                                         received, sent                                                                                                                                         |   packets/s    |
| system.network.protocol.icmp.errors.ipv4.rate                    |      global      |                                                                                                                               InErrors, OutErrors, InCsumErrors                                                                                                                                |   packets/s    |
| system.network.protocol.icmp.errors.ipv6.rate                    |      global      |                                                                  InErrors, OutErrors, InCsumErrors, InDestUnreachs, InPktTooBigs, InTimeExcds, InParmProblems, OutDestUnreachs, OutPktTooBigs, OutTimeExcds, OutParmProblems                                                                   |    errors/s    |
| system.network.protocol.icmp.messages.ipv4.rate                  |      global      | InEchoReps, OutEchoReps, InDestUnreachs, OutDestUnreachs, InRedirects, OutRedirects, InEchos, OutEchos, InRouterAdvert, OutRouterAdvert, InRouterSelect, OutRouterSelect, InTimeExcds, OutTimeExcds, InParmProbs, OutParmProbs, InTimestamps, OutTimestamps, InTimestampReps, OutTimestampReps |   packets/s    |
| system.network.protocol.icmp.messages.ipv6.rate                  |      global      |                                                                                         InType1, InType128, InType129, InType136, OutType1, OutType128, OutType129, OutType133, OutType135, OutType143                                                                                         |   messages/s   |
| system.network.protocol.icmp.redirects.ipv6.rate                 |      global      |                                                                                                                                         received, sent                                                                                                                                         |  redirects/s   |
| system.network.protocol.icmp.echos.ipv6.rate                     |      global      |                                                                                                                        InEchos, OutEchos, InEchoReplies, OutEchoReplies                                                                                                                        |   messages/s   |
| system.network.protocol.icmp.mld.v1.messages.rate                |      global      |                                                                                                         InQueries, OutQueries, InResponses, OutResponses, InReductions, OutReductions                                                                                                          |   messages/s   |
| system.network.protocol.icmp.mld.v2.messages.rate                |      global      |                                                                                                                                         received, sent                                                                                                                                         |   reports/s    |
| system.network.protocol.icmp.ndp.messages.router.rate            |      global      |                                                                                                                  InSolicits, OutSolicits, InAdvertisements, OutAdvertisements                                                                                                                  |   messages/s   |
| system.network.protocol.icmp.ndp.messages.neighbour.rate         |      global      |                                                                                                                  InSolicits, OutSolicits, InAdvertisements, OutAdvertisements                                                                                                                  |   messages/s   |
| system.network.protocol.ecn.packets.ipv4.rate                    |      global      |                                                                                                                                   CEP, NoECTP, ECTP0, ECTP1                                                                                                                                    |   packets/s    |
| system.network.protocol.ecn.packets.ipv6.rate                    |      global      |                                                                                                                         InNoECTPkts, InECT1Pkts, InECT0Pkts, InCEPkts                                                                                                                          |   packets/s    |
| system.network.protocol.raw.sockets.ipv4.count                   |      global      |                                                                                                                                             inuse                                                                                                                                              |    sockets     |
| system.network.protocol.raw.sockets.ipv6.count                   |      global      |                                                                                                                                             inuse                                                                                                                                              |    sockets     |
| system.network.fragments.reassembly.packets.ipv4.rate            |      global      |                                                                                                                                        ok, failed, all                                                                                                                                         |   packets/s    |
| system.network.fragments.reassembly.packets.ipv6.rate            |      global      |                                                                                                                                    ok, failed, timeout, all                                                                                                                                    |   packets/s    |
| system.network.fragments.reassembly.hashtable.entries.ipv4.count |      global      |                                                                                                                                             inuse                                                                                                                                              |    entries     |
| system.network.fragments.reassembly.hashtable.entries.ipv6.count |      global      |                                                                                                                                             inuse                                                                                                                                              |    entries     |
| system.network.fragments.reassembly.memory.ipv4.size             |      global      |                                                                                                                                             inuse                                                                                                                                              |     bytes      |
| system.network.fragments.reassembly.memory.ipv6.size             |      global      |                                                                                                                                             inuse                                                                                                                                              |     bytes      |
| system.network.fragments.fragmentation.packets.ipv4.rate         |      global      |                                                                                                                                      ok, failed, created                                                                                                                                       |   packets/s    |
| system.network.fragments.fragmentation.packets.ipv6.rate         |      global      |                                                                                                                                        ok, failed, all                                                                                                                                         |   packets/s    |
| system.network.softirq.received.rate                             |      global      |                                                                                                                                            received                                                                                                                                            |    frames/s    |
| system.network.softirq.dropped.rate                              |      global      |                                                                                                                                            dropped                                                                                                                                             |    frames/s    |
| system.network.softirq.squeezed.rate                             |      global      |                                                                                                                                            squeezed                                                                                                                                            |    events/s    |
| system.network.softirq.received_rps.rate                         |      global      |                                                                                                                                              rps                                                                                                                                               |    events/s    |
| system.network.softirq.flow_limit.rate                           |      global      |                                                                                                                                           flow_limit                                                                                                                                           |    events/s    |
| system.network.softirq.core.received.rate                        |       core       |                                                                                                                                            received                                                                                                                                            |    frames/s    |
| system.network.softirq.core.dropped.rate                         |       core       |                                                                                                                                            dropped                                                                                                                                             |    frames/s    |
| system.network.softirq.core.squeezed.rate                        |       core       |                                                                                                                                            squeezed                                                                                                                                            |    events/s    |
| system.network.softirq.core.received_rps.rate                    |       core       |                                                                                                                                              rps                                                                                                                                               |    events/s    |
| system.network.softirq.core.flow_limit.rate                      |       core       |                                                                                                                                           flow_limit                                                                                                                                           |    events/s    |
| system.network.netfilter.conntrack.entries.count                 |      global      |                                                                                                                                          connections                                                                                                                                           |  connections   |
| system.network.netfilter.conntrack.new.rate                      |      global      |                                                                                                                                              new                                                                                                                                               |   entries/s    |
| system.network.netfilter.conntrack.ignore.rate                   |      global      |                                                                                                                                             ignore                                                                                                                                             |   packets/s    |
| system.network.netfilter.conntrack.invalid.rate                  |      global      |                                                                                                                                            invalid                                                                                                                                             |   packets/s    |
| system.network.netfilter.conntrack.changes.rate                  |      global      |                                                                                                                                 inserted, deleted, delete_list                                                                                                                                 |   changes/s    |
| system.network.netfilter.conntrack.expectations.rate             |      global      |                                                                                                                                     created, deleted, new                                                                                                                                      | expectations/s |
| system.network.netfilter.conntrack.lookups.rate                  |      global      |                                                                                                                                   searched, restarted, found                                                                                                                                   |   searches/s   |
| system.network.netfilter.conntrack.errors.rate                   |      global      |                                                                                                                          icmp_error, insert_failed, drop, early_drop                                                                                                                           |    events/s    |
| system.network.netfilter.synproxy.syn.rate                       |      global      |                                                                                                                                            received                                                                                                                                            |   packets/s    |
| system.network.netfilter.synproxy.connections.rate               |      global      |                                                                                                                                            reopened                                                                                                                                            | connections/s  |
| system.network.netfilter.synproxy.cookies.rate                   |      global      |                                                                                                                                  valid, invalid, retransmits                                                                                                                                   |   cookies/s    |
| system.network.ipvs.connections.rate                             |      global      |                                                                                                                                            created                                                                                                                                             | connections/s  |
| system.network.ipvs.packets.rate                                 |      global      |                                                                                                                                         received, sent                                                                                                                                         |    packets     |
| system.network.ipvs.traffic.rate                                 |      global      |                                                                                                                                         received, sent                                                                                                                                         |     bits/s     |
| system.storage.io.rate                                           |      global      |                                                                                                                                            in, out                                                                                                                                             |    bytes/s     |
| system.storage.pressure.some.perc                                |      global      |                                                                                                                                      10sec, 60sec, 300sec                                                                                                                                      |   percentage   |
| system.storage.pressure.some.time                                |      global      |                                                                                                                                              time                                                                                                                                              |    seconds     |
| system.storage.pressure.full.perc                                |      global      |                                                                                                                                      10sec, 60sec, 300sec                                                                                                                                      |   percentage   |
| system.storage.pressure.full.time                                |      global      |                                                                                                                                              time                                                                                                                                              |    seconds     |
| system.storage.device.io.rate                                    |       disk       |                                                                                                                                         reads, writes                                                                                                                                          |    bytes/s     |
| system.storage.device.iops.rate                                  |       disk       |                                                                                                                                reads, writes, discards, flushes                                                                                                                                |  operations/s  |
| system.storage.device.iops.time                                  |       disk       |                                                                                                                                reads, writes, discards, flushes                                                                                                                                |    seconds     |
| system.storage.device.iops.queued.count                          |       disk       |                                                                                                                                           operations                                                                                                                                           |   operations   |
| system.storage.device.discards.rate                              |       disk       |                                                                                                                                            discards                                                                                                                                            |    bytes/s     |
| system.storage.device.backlog.time                               |       disk       |                                                                                                                                            backlog                                                                                                                                             |    seconds     |
| system.storage.device.busy.time                                  |       disk       |                                                                                                                                              busy                                                                                                                                              |    seconds     |
| system.storage.device.iops.completion.time                       |       disk       |                                                                                                                                          read, write                                                                                                                                           |    seconds     |
| system.storage.device.iops.service.time                          |       disk       |                                                                                                                                           operation                                                                                                                                            |    seconds     |
| system.storage.device.iops.size                                  |       disk       |                                                                                                                                      read, write, discard                                                                                                                                      |     bytes      |
| system.storage.device.bcache.allocation.perc                     |       disk       |                                                                                                                           unused, dirty, clean, metadata, undefined                                                                                                                            |   percentage   |
| system.storage.device.bcache.hit.perc                            |       disk       |                                                                                                                                    5min, 1hour, 1day, ever                                                                                                                                     |   percentage   |
| system.storage.device.bcache.writeback.io.rate                   |       disk       |                                                                                                                                           writeback                                                                                                                                            |    bytes/s     |
| system.storage.device.bcache.dirty.size                          |       disk       |                                                                                                                                             dirty                                                                                                                                              |     bytes      |
| system.storage.device.bcache.available.perc                      |       disk       |                                                                                                                                             avail                                                                                                                                              |   percentage   |
| system.storage.device.bcache.read_races.rate                     |       disk       |                                                                                                                                              read                                                                                                                                              |    races/s     |
| system.storage.device.bcache.errors.rate                         |       disk       |                                                                                                                                             errors                                                                                                                                             |    errors/s    |
| system.storage.device.bcache.iops.rate                           |       disk       |                                                                                                                              hits, misses, collisions, readaheads                                                                                                                              |  operations/s  |
| system.storage.device.bcache.bypass_ops.rate                     |       disk       |                                                                                                                                          hits, misses                                                                                                                                          |  operations/s  |
| system.storage.md.disks.count                                    |     md array     |                                                                                                                                          inuse, down                                                                                                                                           |     disks      |
| system.storage.md.mismatches.count                               |     md array     |                                                                                                                                           mismatches                                                                                                                                           |   mismatches   |
| system.storage.md.activity.state                                 |     md array     |                                                                                                                                check, resync, recovery, reshape                                                                                                                                |     state      |
| system.storage.md.activity.progress.perc                         |     md array     |                                                                                                                                            progress                                                                                                                                            |    percent     |
| system.storage.md.activity.completion.time                       |     md array     |                                                                                                                                           completion                                                                                                                                           |    seconds     |
| system.storage.md.activity.io.rate                               |     md array     |                                                                                                                                            activity                                                                                                                                            |    bytes/s     |
| system.storage.md.nonredundant.state                             |     md array     |                                                                                                                                     available, unavailable                                                                                                                                     |     state      |
| system.storage.fs.zfs.pool.state                                 |       pool       |                                                                                                                      online, degraded, faulted, offline, removed, unavail                                                                                                                      |     state      |
| system.storage.fs.zfs.io.rate                                    |      global      |                                                                                                                                          read, write                                                                                                                                           |    bytes/s     |
| system.storage.fs.zfs.hits.perc                                  |      global      |                                                                                                                                          hits, misses                                                                                                                                          |   percentage   |
| system.storage.fs.zfs.hits.rate                                  |      global      |                                                                                                                                          hits, misses                                                                                                                                          |    events/s    |
| system.storage.fs.zfs.demand.data.hits.perc                      |      global      |                                                                                                                                          hits, misses                                                                                                                                          |   percentage   |
| system.storage.fs.zfs.demand.metadata.hits.rate                  |      global      |                                                                                                                                          hits, misses                                                                                                                                          |    events/s    |
| system.storage.fs.zfs.prefetch.data.hits.perc                    |      global      |                                                                                                                                          hits, misses                                                                                                                                          |   percentage   |
| system.storage.fs.zfs.prefetch.metadata.hits.rate                |      global      |                                                                                                                                          hits, misses                                                                                                                                          |    events/s    |
| system.storage.fs.zfs.l2.hits.perc                               |      global      |                                                                                                                                          hits, misses                                                                                                                                          |   percentage   |
| system.storage.fs.zfs.l2.hits.rate                               |      global      |                                                                                                                                          hits, misses                                                                                                                                          |    events/s    |
| system.storage.fs.zfs.list.hits.rate                             |      global      |                                                                                                                                 mfu, mfu_ghost, mru, mru_ghost                                                                                                                                 |     hits/s     |
| system.storage.fs.zfs.arc_size.perc                              |      global      |                                                                                                                                        recent, frequent                                                                                                                                        |   percentage   |
| system.storage.fs.zfs.memory.ops.rate                            |      global      |                                                                                                                                  direct, throttled, indirect                                                                                                                                   |  operations/s  |
| system.storage.fs.zfs.eviction.skip.rate                         |      global      |                                                                                                                                              skip                                                                                                                                              |     skip/s     |
| system.storage.fs.zfs.eviction.delete.rate                       |      global      |                                                                                                                                             delete                                                                                                                                             |   deletes/s    |
| system.storage.fs.zfs.eviction.mutex_miss.rate                   |      global      |                                                                                                                                              miss                                                                                                                                              |    misses/s    |
| system.storage.fs.zfs.hash.collisions.rate                       |      global      |                                                                                                                                           collisions                                                                                                                                           |  collisions/s  |
| system.storage.fs.zfs.hash.elements.max.count                    |      global      |                                                                                                                                              max                                                                                                                                               |    elements    |
| system.storage.fs.zfs.hash.elements.count                        |      global      |                                                                                                                                            current                                                                                                                                             |    elements    |
| system.storage.fs.zfs.hash.chain.elements.max.count              |      global      |                                                                                                                                              max                                                                                                                                               |    elements    |
| system.storage.fs.zfs.hash.chains.count                          |      global      |                                                                                                                                            current                                                                                                                                             |     chains     |
| system.storage.fs.btrfs.disk.size                                |       disk       |                                                                                                          unallocated, data_free, data_used, meta_free, meta_used, sys_free, sys_used                                                                                                           |     bytes      |
| system.storage.fs.btrfs.allocated.size                           |       disk       |                                                                                                         data_free, data_used, meta_free, meta_used, meta_reserved, sys_free, sys_used                                                                                                          |     bytes      |
| system.storage.fs.nfs.client.packets.rate                        |      global      |                                                                                                                                            udp, tcp                                                                                                                                            |   packets/s    |
| system.storage.fs.nfs.client.calls.rate                          |      global      |                                                                                                                                             calls                                                                                                                                              |    calls/s     |
| system.storage.fs.nfs.client.calls.v2.per_call.rate              |      global      |                                                                                                                                  <i>a dimension per call</i>                                                                                                                                   |    calls/s     |
| system.storage.fs.nfs.client.calls.v3.per_call.rate              |      global      |                                                                                                                                  <i>a dimension per call</i>                                                                                                                                   |    calls/s     |
| system.storage.fs.nfs.client.calls.v4.per_call.rate              |      global      |                                                                                                                                  <i>a dimension per call</i>                                                                                                                                   |    calls/s     |
| system.storage.fs.nfs.client.retransmits.rate                    |      global      |                                                                                                                                          retransmits                                                                                                                                           |    calls/s     |
| system.storage.fs.nfs.client.auth_refresh.rate                   |      global      |                                                                                                                                          auth_refresh                                                                                                                                          |    calls/s     |
| system.storage.fs.nfs.server.replycache.reads.rate               |      global      |                                                                                                                                     hits, misses, nocache                                                                                                                                      |    reads/s     |
| system.storage.fs.nfs.server.filehandles.rate                    |      global      |                                                                                                                                             stale                                                                                                                                              |   handles/s    |
| system.storage.fs.nfs.server.io.rate                             |      global      |                                                                                                                                          read, write                                                                                                                                           |    bytes/s     |
| system.storage.fs.nfs.server.threads.count                       |      global      |                                                                                                                                            threads                                                                                                                                             |    threads     |
| system.storage.fs.nfs.server.readahead.perc                      |      global      |                                                                                                                                        10%-100%, misses                                                                                                                                        |   percentage   |
| system.storage.fs.nfs.server.packets.rate                        |      global      |                                                                                                                                            udp, tcp                                                                                                                                            |   packets/s    |
| system.storage.fs.nfs.server.calls.rate                          |      global      |                                                                                                                                             calls                                                                                                                                              |    calls/s     |
| system.storage.fs.nfs.server.calls.v2.per_call.rate              |      global      |                                                                                                                                  <i>a dimension per call</i>                                                                                                                                   |    calls/s     |
| system.storage.fs.nfs.server.calls.v3.per_call.rate              |      global      |                                                                                                                                  <i>a dimension per call</i>                                                                                                                                   |    calls/s     |
| system.storage.fs.nfs.server.calls.v4.per_call.rate              |      global      |                                                                                                                                  <i>a dimension per call</i>                                                                                                                                   |    calls/s     |
| system.storage.fs.nfs.server.ops.v4.per_op.rate                  |      global      |                                                                                                                                <i>a dimension per operation</i>                                                                                                                                |  operations/s  |
| system.storage.fs.nfs.server.errors.rate                         |      global      |                                                                                                                                      bad_format, bad_auth                                                                                                                                      |    calls/s     |
| system.load.num                                                  |      global      |                                                                                                                                       1min, 5min, 15min                                                                                                                                        |      load      |
| system.os.processes.active.count                                 |      global      |                                                                                                                                             active                                                                                                                                             |   processes    |
| system.os.processes.forks.rate                                   |      global      |                                                                                                                                             forks                                                                                                                                              |  processes/s   |
| system.os.oom_kills.rate                                         |      global      |                                                                                                                                             kills                                                                                                                                              |    kills/s     |
| system.os.sockets.count                                          |      global      |                                                                                                                                              used                                                                                                                                              |    sockets     |
| system.os.ipc.sysv.semaphore_sets.semaphores.count               |      global      |                                                                                                                                           semaphores                                                                                                                                           |   semaphores   |
| system.os.ipc.sysv.semaphore_sets.arrays.count                   |      global      |                                                                                                                                             arrays                                                                                                                                             |     arrays     |
| system.os.ipc.sysv.message_queues.queue.messages.count           |      queue       |                                                                                                                                            messages                                                                                                                                            |    messages    |
| system.os.ipc.sysv.message_queues.queue.size                     |      queue       |                                                                                                                                              size                                                                                                                                              |     bytes      |
| system.os.ipc.sysv.shared_memory.segments.count                  |      global      |                                                                                                                                            segments                                                                                                                                            |    segments    |
| system.os.ipc.sysv.shared_memory.size                            |      global      |                                                                                                                                              size                                                                                                                                              |     bytes      |
| system.os.entropy.bits.count                                     |      global      |                                                                                                                                            entropy                                                                                                                                             |      bits      |
| system.os.uptime.time                                            |      global      |                                                                                                                                             uptime                                                                                                                                             |    seconds     |
| system.powersupply.capacity.perc                                 |   power supply   |                                                                                                                                            capacity                                                                                                                                            |   percentage   |
| system.powersupply.charge.num                                    |   power supply   |                                                                                                                                             charge                                                                                                                                             |       Ah       |
| system.powersupply.energy.num                                    |   power supply   |                                                                                                                                             energy                                                                                                                                             |       Wh       |
| system.powersupply.voltage.num                                   |   power supply   |                                                                                                                                            voltage                                                                                                                                             |       V        |

## TODO

| Metric | Scope | Dimensions | Units |
|--------|:-----:|:----------:|:-----:|


## Metrics

<details>
<summary>Metrics</summary>

| Metric                                  |     Scope      |                                                                                                                                           Dimensions                                                                                                                                           |         Units          |
|-----------------------------------------|:--------------:|:----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------:|:----------------------:|
| system.ipc_semaphores                   |     global     |                                                                                                                                           semaphores                                                                                                                                           |       semaphores       |
| system.ipc_semaphore_arrays             |     global     |                                                                                                                                             arrays                                                                                                                                             |         arrays         |
| system.message_queue_messages           |     global     |                                                                                                                                  <i>a dimension per queue</i>                                                                                                                                  |        messages        |
| system.message_queue_bytes              |     global     |                                                                                                                                  <i>a dimension per queue</i>                                                                                                                                  |         bytes          |
| system.shared_memory_segments           |     global     |                                                                                                                                            segments                                                                                                                                            |        segments        |
| system.shared_memory_bytes              |     global     |                                                                                                                                             bytes                                                                                                                                              |         bytes          |
| system.cpu                              |     global     |                                                                                                            guest_nice, guest, steal, softirq, irq, user, system, nice, iowait, idle                                                                                                            |       percentage       |
| system.io                               |     global     |                                                                                                                                            in, out                                                                                                                                             |         KiB/s          |
| system.net                              |     global     |                                                                                                                                         received, sent                                                                                                                                         |       kilobits/s       |
| system.ip                               |     global     |                                                                                                                                         received, sent                                                                                                                                         |       kilobits/s       |
| system.ipv6                             |     global     |                                                                                                                                         received, sent                                                                                                                                         |       kilobits/s       |
| system.load                             |     global     |                                                                                                                                      load1, load5, load15                                                                                                                                      |          load          |
| system.active_processes                 |     global     |                                                                                                                                             active                                                                                                                                             |       processes        |
| system.entropy                          |     global     |                                                                                                                                            entropy                                                                                                                                             |        entropy         |
| system.uptime                           |     global     |                                                                                                                                             uptime                                                                                                                                             |        seconds         |
| system.ram                              |     global     |                                                                                                                                  free, used, cached, buffers                                                                                                                                   |          MiB           |
| system.swap                             |     global     |                                                                                                                                           free, used                                                                                                                                           |          MiB           |
| system.swapio                           |     global     |                                                                                                                                            in, out                                                                                                                                             |         KiB/s          |
| system.pgpgio                           |     global     |                                                                                                                                            in, out                                                                                                                                             |         KiB/s          |
| system.pgfaults                         |     global     |                                                                                                                                          minor, major                                                                                                                                          |        faults/s        |
| system.intr                             |     global     |                                                                                                                                           interrupts                                                                                                                                           |      interrupts/s      |
| system.ctxt                             |     global     |                                                                                                                                            switches                                                                                                                                            |   context switches/s   |
| system.forks                            |     global     |                                                                                                                                            started                                                                                                                                             |      processes/s       |
| system.processes                        |     global     |                                                                                                                                        running, blocked                                                                                                                                        |       processes        |
| system.softnet_stat                     |     global     |                                                                                                                  processed, dropped, squeezed, received_rps, flow_limit_count                                                                                                                  |        events/s        |
| system.softirqs                         |     global     |                                                                                                                                 <i>a dimension per tasklet</i>                                                                                                                                 |       softirqs/s       |
| system.cpu_some_pressure                |     global     |                                                                                                                                   some_10, some_30, some_60                                                                                                                                    |       percentage       |
| system.cpu_some_pressure_stall_time     |     global     |                                                                                                                                              time                                                                                                                                              |           ms           |
| system.cpu_full_pressure                |     global     |                                                                                                                                   some_10, some_30, some_60                                                                                                                                    |       percentage       |
| system.cpu_full_pressure_stall_time     |     global     |                                                                                                                                              time                                                                                                                                              |           ms           |
| system.io_some_pressure                 |     global     |                                                                                                                                   some_10, some_30, some_60                                                                                                                                    |       percentage       |
| system.io_some_pressure_stall_time      |     global     |                                                                                                                                              time                                                                                                                                              |           ms           |
| system.io_full_pressure                 |     global     |                                                                                                                                   some_10, some_30, some_60                                                                                                                                    |       percentage       |
| system.io_full_pressure_stall_time      |     global     |                                                                                                                                              time                                                                                                                                              |           ms           |
| system.memory_some_pressure             |     global     |                                                                                                                                   some_10, some_30, some_60                                                                                                                                    |       percentage       |
| system.memory_some_pressure_stall_time  |     global     |                                                                                                                                              time                                                                                                                                              |           ms           |
| system.memory_full_pressure             |     global     |                                                                                                                                   some_10, some_30, some_60                                                                                                                                    |       percentage       |
| system.memory_full_pressure_stall_time  |     global     |                                                                                                                                              time                                                                                                                                              |           ms           |
| mem.hwcorrupt                           |     global     |                                                                                                                                       HardwareCorrupted                                                                                                                                        |          MiB           |
| mem.committed                           |     global     |                                                                                                                                          Committed_AS                                                                                                                                          |          MiB           |
| mem.writeback                           |     global     |                                                                                                                     Dirty, Writeback, FuseWriteback, NfsWriteback, Bounce                                                                                                                      |          MiB           |
| mem.kernel                              |     global     |                                                                                                                       Slab, KernelStack, PageTables, VmallocUsed, Percpu                                                                                                                       |          MiB           |
| mem.slab                                |     global     |                                                                                                                                   reclaimable, unreclaimable                                                                                                                                   |          MiB           |
| mem.hugepages                           |     global     |                                                                                                                                 free, used, surplus, reserved                                                                                                                                  |          MiB           |
| mem.transparent_hugepages               |     global     |                                                                                                                                 free, used, surplus, reserved                                                                                                                                  |          MiB           |
| mem.available                           |     global     |                                                                                                                                        anonymous, shmem                                                                                                                                        |          MiB           |
| mem.pagetype_global                     |     global     |                                                                                                                                  <i>a dimension per node</i>                                                                                                                                   |           B            |
| mem.pagetype                            |     global     |                                                                                                                                  <i>a dimension per node</i>                                                                                                                                   |           B            |
| mem.oom_kill                            |     global     |                                                                                                                                             kills                                                                                                                                              |        kills/s         |
| mem.numa                                |     global     |                                                                                         local, foreign, interleave, otherpte_updates, huge_pte_updates, hint_faults, hint_faults_local, pages_migrated                                                                                         |        events/s        |
| mem.<node_name>                         |   numa node    |                                                                                                                          hit, miss, local, foreign, interleave, other                                                                                                                          |        events/s        |
| mem.zram_usage                          |     global     |                                                                                                                                  compressed, metadata, device                                                                                                                                  |          MiB           |
| mem.zram_savings                        |     global     |                                                                                                                                   savings, original, device                                                                                                                                    |          MiB           |
| mem.zram_ratio                          |     global     |                                                                                                                                         ratio, device                                                                                                                                          |         ratio          |
| mem.zram_efficiency                     |     global     |                                                                                                                                        percent, device                                                                                                                                         |       percentage       |
| mem.ecc_ce                              |     global     |                                                                                                                              <i>a dimension per (?) instance</i>                                                                                                                               |         errors         |
| mem.ecc_ue                              |     global     |                                                                                                                              <i>a dimension per (?) instance</i>                                                                                                                               |         errors         |
| mem.ksm                                 |     global     |                                                                                                                              shared, unshared, sharing, volatile                                                                                                                               |          MiB           |
| mem.ksm_savings                         |     global     |                                                                                                                                        savings, offered                                                                                                                                        |          MiB           |
| mem.ksm_ratios                          |     global     |                                                                                                                                            savings                                                                                                                                             |       percentage       |
| system.interrupts                       |     global     |                                                                                                                                 <i>a dimension per device</i>                                                                                                                                  |      interrupts/s      |
| cpu.cpu                                 |      core      |                                                                                                            guest_nice, guest, steal, softirq, irq, user, system, nice, iowait, idle                                                                                                            |       percentage       |
| cpuidle.cpu_cstate_residency_time       |      core      |                                                                                                                                 <i>a dimension per c-state</i>                                                                                                                                 |       percentage       |
| cpu.core_throttling                     |     global     |                                                                                                                                  <i>a dimension per core</i>                                                                                                                                   |        events/s        |
| cpu.package_throttling                  |     global     |                                                                                                                                  <i>a dimension per core</i>                                                                                                                                   |        events/s        |
| cpufreq.cpufreq                         |     global     |                                                                                                                                  <i>a dimension per core</i>                                                                                                                                   |          MHz           |
| cpu.interrupts                          |      core      |                                                                                                                                 <i>a dimension per device</i>                                                                                                                                  |      interrupts/s      |
| cpu.softnet_stat                        |      core      |                                                                                                                  processed, dropped, squeezed, received_rps, flow_limit_count                                                                                                                  |        events/s        |
| cpu.softirqs                            |      core      |                                                                                                                                 <i>a dimension per tasklet</i>                                                                                                                                 |       softirqs/s       |
| disk.bcache_cache_alloc                 |      disk      |                                                                                                                           unused, dirty, clean, metadata, undefined                                                                                                                            |       percentage       |
| disk.bcache_hit_ratio                   |      disk      |                                                                                                                                    5min, 1hour, 1day, ever                                                                                                                                     |       percentage       |
| disk.bcache_rates                       |      disk      |                                                                                                                                      congested, writeback                                                                                                                                      |         KiB/s          |
| disk.bcache_size                        |      disk      |                                                                                                                                             dirty                                                                                                                                              |          MiB           |
| disk.bcache_usage                       |      disk      |                                                                                                                                             avail                                                                                                                                              |       percentage       |
| disk.bcache_cache_read_races            |      disk      |                                                                                                                                         races, errors                                                                                                                                          |      operations/s      |
| disk.bcache                             |      disk      |                                                                                                                              hits, misses, collisions, readaheads                                                                                                                              |      operations/s      |
| disk.bcache_bypass                      |      disk      |                                                                                                                                          hits, misses                                                                                                                                          |      operations/s      |
| disk.io                                 |      disk      |                                                                                                                                         reads, writes                                                                                                                                          |         KiB/s          |
| disk_ext.io                             |      disk      |                                                                                                                                            discards                                                                                                                                            |         KiB/s          |
| disk.ops                                |      disk      |                                                                                                                                         reads, writes                                                                                                                                          |        disk.ops        |
| disk_ext.ops                            |      disk      |                                                                                                                                       discards, flushes                                                                                                                                        |      operations/s      |
| disk.qops                               |      disk      |                                                                                                                                           operations                                                                                                                                           |       operations       |
| disk.backlog                            |      disk      |                                                                                                                                            backlog                                                                                                                                             |      milliseconds      |
| disk.busy                               |      disk      |                                                                                                                                              busy                                                                                                                                              |      milliseconds      |
| disk.util                               |      disk      |                                                                                                                                          utilization                                                                                                                                           |   % of time working    |
| disk.mops                               |      disk      |                                                                                                                                         reads, writes                                                                                                                                          |  merged operations/s   |
| disk_ext.mops                           |      disk      |                                                                                                                                            discards                                                                                                                                            |  merged operations/s   |
| disk.iotime                             |      disk      |                                                                                                                                         reads, writes                                                                                                                                          |     milliseconds/s     |
| disk_ext.iotime                         |      disk      |                                                                                                                                       discards, flushes                                                                                                                                        |     milliseconds/s     |
| disk.await                              |      disk      |                                                                                                                                         reads, writes                                                                                                                                          | milliseconds/operation |
| disk.avgsz                              |      disk      |                                                                                                                                         reads, writes                                                                                                                                          |     KiB/operation      |
| disk_ext.avgsz                          |      disk      |                                                                                                                                            discards                                                                                                                                            |     KiB/operation      |
| disk.svctm                              |      disk      |                                                                                                                                             svctm                                                                                                                                              | milliseconds/operation |
| md.health                               |      disk      |                                                                                                                                <i>a dimension per md array</i>                                                                                                                                 |      failed disks      |
| md.disks                                |    md array    |                                                                                                                                          inuse, down                                                                                                                                           |         disks          |
| md.mismatch_cnt                         |    md array    |                                                                                                                                             count                                                                                                                                              | unsynchronized blocks  |
| md.status                               |    md array    |                                                                                                                                check, resync, recovery, reshape                                                                                                                                |        percent         |
| md.expected_time_until_operation_finish |    md array    |                                                                                                                                           finish_in                                                                                                                                            |        seconds         |
| md.operation_speed                      |    md array    |                                                                                                                                             speed                                                                                                                                              |         KiB/s          |
| md.nonredundant                         |    md array    |                                                                                                                                           available                                                                                                                                            |        boolean         |
| net.net                                 | network device |                                                                                                                                         received, sent                                                                                                                                         |       kilobits/s       |
| net.compressed                          | network device |                                                                                                                                         received, sent                                                                                                                                         |       packets/s        |
| net.drops                               | network device |                                                                                                                                       inbound, outbound                                                                                                                                        |        drops/s         |
| net.errors                              | network device |                                                                                                                                       inbound, outbound                                                                                                                                        |        errors/s        |
| net.events                              | network device |                                                                                                                                  frames, collisions, carrier                                                                                                                                   |        events/s        |
| net.fifo                                | network device |                                                                                                                                       receive, transmit                                                                                                                                        |         errors         |
| net.packets                             | network device |                                                                                                                                   received, sent, multicast                                                                                                                                    |       packets/s        |
| net.speed                               | network device |                                                                                                                                             speed                                                                                                                                              |       kilobits/s       |
| net.duplex                              | network device |                                                                                                                                      half, full, unknown                                                                                                                                       |         state          |
| net.operstate                           | network device |                                                                                                                up, down, notpresent, lowerlayerdown, testing, dormant, unknown                                                                                                                 |         state          |
| net.carrier                             | network device |                                                                                                                                            up, down                                                                                                                                            |         state          |
| net.mtu                                 | network device |                                                                                                                                              mtu                                                                                                                                               |         octets         |
| wireless.status                         | network device |                                                                                                                                             status                                                                                                                                             |         status         |
| wireless.link_quality                   | network device |                                                                                                                                          link_quality                                                                                                                                          |         value          |
| wireless.signal_level                   | network device |                                                                                                                                          signal_level                                                                                                                                          |          dBm           |
| wireless.noise_level                    | network device |                                                                                                                                          noise_level                                                                                                                                           |          dBm           |
| wireless.discarded_packets              | network device |                                                                                                                                 nwid, crypt, frag, retry, misc                                                                                                                                 |          dBm           |
| wireless.missed_beacons                 | network device |                                                                                                                                         missed_beacons                                                                                                                                         |        frames/s        |
| ipvs.sockets                            |     global     |                                                                                                                                          connections                                                                                                                                           |     connections/s      |
| ipvs.packets                            |     global     |                                                                                                                                         received, sent                                                                                                                                         |        packets         |
| ipvs.net                                |     global     |                                                                                                                                         received, sent                                                                                                                                         |       kilobits/s       |
| ipvs.net                                |     global     |                                                                                                                                         received, sent                                                                                                                                         |       kilobits/s       |
| ip.inerrors                             |     global     |                                                                                                                                 noroutes, truncated, checksum                                                                                                                                  |       packets/s        |
| ip.mcast                                |     global     |                                                                                                                                         received, sent                                                                                                                                         |       kilobits/s       |
| ip.bcast                                |     global     |                                                                                                                                         received, sent                                                                                                                                         |       kilobits/s       |
| ip.mcastpkts                            |     global     |                                                                                                                                         received, sent                                                                                                                                         |       packets/s        |
| ip.bcastpkts                            |     global     |                                                                                                                                         received, sent                                                                                                                                         |       packets/s        |
| ip.ecnpkts                              |     global     |                                                                                                                                   CEP, NoECTP, ECTP0, ECTP1                                                                                                                                    |       packets/s        |
| ip.tcpmemorypressures                   |     global     |                                                                                                                                           pressures                                                                                                                                            |        events/s        |
| ip.tcpconnaborts                        |     global     |                                                                                                                     baddata, userclosed, nomemory, timeout, linger, failed                                                                                                                     |     connections/s      |
| ip.tcpreorders                          |     global     |                                                                                                                                  timestamp, sack, fack, reno                                                                                                                                   |       packets/s        |
| ip.tcpofo                               |     global     |                                                                                                                                inqueue, dropped, merged, pruned                                                                                                                                |       packets/s        |
| ip.tcpsyncookies                        |     global     |                                                                                                                                     received, sent, failed                                                                                                                                     |       packets/s        |
| ip.tcp_syn_queue                        |     global     |                                                                                                                                         drops, cookies                                                                                                                                         |       packets/s        |
| ip.tcp_accept_queue                     |     global     |                                                                                                                                        overflows, drops                                                                                                                                        |       packets/s        |
| ipv4.packets                            |     global     |                                                                                                                              received, sent, forwarded, delivered                                                                                                                              |       packets/s        |
| ipv4.fragsout                           |     global     |                                                                                                                                      ok, failed, created                                                                                                                                       |       packets/s        |
| ipv4.fragsin                            |     global     |                                                                                                                                        ok, failed, all                                                                                                                                         |       packets/s        |
| ipv4.errors                             |     global     |                                                                                                        InDiscards, OutDiscards, InHdrErrors, OutNoRoutes, InAddrErrors, InUnknownProtos                                                                                                        |       packets/s        |
| ipv4.icmp                               |     global     |                                                                                                                                         received, sent                                                                                                                                         |       packets/s        |
| ipv4.icmp_errors                        |     global     |                                                                                                                               InErrors, OutErrors, InCsumErrors                                                                                                                                |       packets/s        |
| ipv4.icmpmsg                            |     global     | InEchoReps, OutEchoReps, InDestUnreachs, OutDestUnreachs, InRedirects, OutRedirects, InEchos, OutEchos, InRouterAdvert, OutRouterAdvert, InRouterSelect, OutRouterSelect, InTimeExcds, OutTimeExcds, InParmProbs, OutParmProbs, InTimestamps, OutTimestamps, InTimestampReps, OutTimestampReps |       packets/s        |
| ipv4.tcppackets                         |     global     |                                                                                                                                         received, sent                                                                                                                                         |       packets/s        |
| ipv4.tcperrors                          |     global     |                                                                                                                               InErrs, InCsumErrors, RetransSegs                                                                                                                                |       packets/s        |
| ipv4.tcpopens                           |     global     |                                                                                                                                        active, passive                                                                                                                                         |     connections/s      |
| ipv4.tcphandshake                       |     global     |                                                                                                                         EstabResets, OutRsts, AttemptFails, SynRetrans                                                                                                                         |        events/s        |
| ipv4.udppackets                         |     global     |                                                                                                                                         received, sent                                                                                                                                         |       packets/s        |
| ipv4.udperrors                          |     global     |                                                                                                           RcvbufErrors, SndbufErrors, InErrors, NoPorts, InCsumErrors, IgnoredMulti                                                                                                            |        events/s        |
| ipv4.udplite                            |     global     |                                                                                                                                         received, sent                                                                                                                                         |       packets/s        |
| ipv4.udplite_errors                     |     global     |                                                                                                           RcvbufErrors, SndbufErrors, InErrors, NoPorts, InCsumErrors, IgnoredMulti                                                                                                            |       packets/s        |
| ipv4.sockstat_sockets                   |     global     |                                                                                                                                              used                                                                                                                                              |        sockets         |
| ipv4.sockstat_tcp_sockets               |     global     |                                                                                                                                 alloc, orphan, inuse, timewait                                                                                                                                 |        sockets         |
| ipv4.sockstat_tcp_mem                   |     global     |                                                                                                                                              mem                                                                                                                                               |          KiB           |
| ipv4.sockstat_udp_sockets               |     global     |                                                                                                                                             inuse                                                                                                                                              |        sockets         |
| ipv4.sockstat_udp_mem                   |     global     |                                                                                                                                              mem                                                                                                                                               |          KiB           |
| ipv4.sockstat_udplite_sockets           |     global     |                                                                                                                                             inuse                                                                                                                                              |        sockets         |
| ipv4.sockstat_raw_sockets               |     global     |                                                                                                                                             inuse                                                                                                                                              |        sockets         |
| ipv4.sockstat_frag_sockets              |     global     |                                                                                                                                             inuse                                                                                                                                              |       fragments        |
| ipv4.sockstat_frag_mem                  |     global     |                                                                                                                                             inuse                                                                                                                                              |          KiB           |
| ipv6.packets                            |     global     |                                                                                                                              received, sent, forwarded, delivers                                                                                                                               |       packets/s        |
| ipv6.fragsout                           |     global     |                                                                                                                                        ok, failed, all                                                                                                                                         |       packets/s        |
| ipv6.fragsin                            |     global     |                                                                                                                                    ok, failed, timeout, all                                                                                                                                    |       packets/s        |
| ipv6.errors                             |     global     |                                                                                 InDiscards, OutDiscards, InHdrErrors, InNoRoutes, OutNoRoutes, InAddrErrors, InUnknownProtos, InTooBigErrors, InTruncatedPkts                                                                                  |       packets/s        |
| ipv6.udppackets                         |     global     |                                                                                                                                         received, sent                                                                                                                                         |       packets/s        |
| ipv6.udperrors                          |     global     |                                                                                                           RcvbufErrors, SndbufErrors, InErrors, NoPorts, InCsumErrors, IgnoredMulti                                                                                                            |        events/s        |
| ipv6.udplitepackets                     |     global     |                                                                                                                                         received, sent                                                                                                                                         |       packets/s        |
| ipv6.udpliteerrors                      |     global     |                                                                                                                  RcvbufErrors, SndbufErrors, InErrors, NoPorts, InCsumErrors                                                                                                                   |       packets/s        |
| ipv6.mcast                              |     global     |                                                                                                                                         received, sent                                                                                                                                         |       kilobits/s       |
| ipv6.bcast                              |     global     |                                                                                                                                         received, sent                                                                                                                                         |       kilobits/s       |
| ipv6.mcastpkts                          |     global     |                                                                                                                                         received, sent                                                                                                                                         |       packets/s        |
| ipv6.icmp                               |     global     |                                                                                                                                         received, sent                                                                                                                                         |       messages/s       |
| ipv6.icmpredir                          |     global     |                                                                                                                                         received, sent                                                                                                                                         |      redirects/s       |
| ipv6.icmperrors                         |     global     |                                                                  InErrors, OutErrors, InCsumErrors, InDestUnreachs, InPktTooBigs, InTimeExcds, InParmProblems, OutDestUnreachs, OutPktTooBigs, OutTimeExcds, OutParmProblems                                                                   |        errors/s        |
| ipv6.icmpechos                          |     global     |                                                                                                                        InEchos, OutEchos, InEchoReplies, OutEchoReplies                                                                                                                        |       messages/s       |
| ipv6.groupmemb                          |     global     |                                                                                                         InQueries, OutQueries, InResponses, OutResponses, InReductions, OutReductions                                                                                                          |       messages/s       |
| ipv6.icmprouter                         |     global     |                                                                                                                  InSolicits, OutSolicits, InAdvertisements, OutAdvertisements                                                                                                                  |       messages/s       |
| ipv6.icmpneighbor                       |     global     |                                                                                                                  InSolicits, OutSolicits, InAdvertisements, OutAdvertisements                                                                                                                  |       messages/s       |
| ipv6.icmpmldv2                          |     global     |                                                                                                                                         received, sent                                                                                                                                         |       reports/s        |
| ipv6.icmptypes                          |     global     |                                                                                         InType1, InType128, InType129, InType136, OutType1, OutType128, OutType129, OutType133, OutType135, OutType143                                                                                         |       messages/s       |
| ipv6.ect                                |     global     |                                                                                                                         InNoECTPkts, InECT1Pkts, InECT0Pkts, InCEPkts                                                                                                                          |       packets/s        |
| ipv6.sockstat6_tcp_sockets              |     global     |                                                                                                                                             inuse                                                                                                                                              |        sockets         |
| ipv6.sockstat6_udp_sockets              |     global     |                                                                                                                                             inuse                                                                                                                                              |        sockets         |
| ipv6.sockstat6_udplite_sockets          |     global     |                                                                                                                                             inuse                                                                                                                                              |        sockets         |
| ipv6.sockstat6_raw_sockets              |     global     |                                                                                                                                             inuse                                                                                                                                              |        sockets         |
| ipv6.sockstat6_frag_sockets             |     global     |                                                                                                                                             inuse                                                                                                                                              |        sockets         |
| nfs.net                                 |     global     |                                                                                                                                            udp, tcp                                                                                                                                            |      operations/s      |
| nfs.rpc                                 |     global     |                                                                                                                                calls, retransmits, auth_refresh                                                                                                                                |        calls/s         |
| nfs.proc2                               |     global     |                                                                                                                                  <i>a dimension per call</i>                                                                                                                                   |        calls/s         |
| nfs.proc3                               |     global     |                                                                                                                                  <i>a dimension per call</i>                                                                                                                                   |        calls/s         |
| nfs.proc4                               |     global     |                                                                                                                                  <i>a dimension per call</i>                                                                                                                                   |        calls/s         |
| nfsd.readcache                          |     global     |                                                                                                                                     hits, misses, nocache                                                                                                                                      |        reads/s         |
| nfsd.filehandles                        |     global     |                                                                                                                                             stale                                                                                                                                              |       handles/s        |
| nfsd.io                                 |     global     |                                                                                                                                          read, write                                                                                                                                           |      kilobytes/s       |
| nfsd.threads                            |     global     |                                                                                                                                            threads                                                                                                                                             |        threads         |
| nfsd.readahead                          |     global     |                                                                                                                                        10%-100%, misses                                                                                                                                        |       percentage       |
| nfsd.net                                |     global     |                                                                                                                                            udp, tcp                                                                                                                                            |       packets/s        |
| nfsd.rpc                                |     global     |                                                                                                                                  calls, bad_format, bad_auth                                                                                                                                   |        calls/s         |
| nfsd.proc2                              |     global     |                                                                                                                                  <i>a dimension per call</i>                                                                                                                                   |        calls/s         |
| nfsd.proc3                              |     global     |                                                                                                                                  <i>a dimension per call</i>                                                                                                                                   |        calls/s         |
| nfsd.proc4                              |     global     |                                                                                                                                  <i>a dimension per call</i>                                                                                                                                   |        calls/s         |
| nfsd.proc4ops                           |     global     |                                                                                                                                <i>a dimension per operation</i>                                                                                                                                |      operations/s      |
| sctp.transitions                        |     global     |                                                                                                                               active, passive, aborted, shutdown                                                                                                                               |     transitions/s      |
| sctp.packets                            |     global     |                                                                                                                                         received, sent                                                                                                                                         |       packets/s        |
| sctp.packet_errors                      |     global     |                                                                                                                                       invalid, checksum                                                                                                                                        |       packets/s        |
| sctp.fragmentation                      |     global     |                                                                                                                                    reassembled, fragmented                                                                                                                                     |       packets/s        |
| sctp.chunks                             |     global     |                                                                                                                   InCtrl, InOrder, InUnorder, OutCtrl, OutOrder, OutUnorder                                                                                                                    |        chunks/s        |
| netfilter.conntrack_sockets             |     global     |                                                                                                                                          connections                                                                                                                                           |   active connections   |
| netfilter.conntrack_new                 |     global     |                                                                                                                                      new, ignore, invalid                                                                                                                                      |     connections/s      |
| netfilter.conntrack_changes             |     global     |                                                                                                                                 inserted, deleted, delete_list                                                                                                                                 |       changes/s        |
| netfilter.conntrack_expect              |     global     |                                                                                                                                     created, deleted, new                                                                                                                                      |     expectations/s     |
| netfilter.conntrack_search              |     global     |                                                                                                                                   searched, restarted, found                                                                                                                                   |       searches/s       |
| netfilter.conntrack_errors              |     global     |                                                                                                                          icmp_error, insert_failed, drop, early_drop                                                                                                                           |        events/s        |
| netfilter.synproxy_syn_received         |     global     |                                                                                                                                            received                                                                                                                                            |       packets/s        |
| netfilter.synproxy_conn_reopened        |     global     |                                                                                                                                            received                                                                                                                                            |     connections/s      |
| netfilter.synproxy_cookies              |     global     |                                                                                                                                  valid, invalid, retransmits                                                                                                                                   |       cookies/s        |
| btrfs.disk                              |      disk      |                                                                                                          unallocated, data_free, data_used, meta_free, meta_used, sys_free, sys_used                                                                                                           |          MiB           |
| btrfs.data                              |      disk      |                                                                                                                                           free, used                                                                                                                                           |          MiB           |
| btrfs.system                            |      disk      |                                                                                                                                           free, used                                                                                                                                           |          MiB           |
| btrfs.metadata                          |      disk      |                                                                                                                                      free, used, reserved                                                                                                                                      |          MiB           |
| zfspool.state                           |      pool      |                                                                                                                      online, degraded, faulted, offline, removed, unavail                                                                                                                      |        boolean         |
| zfs.reads                               |     global     |                                                                                                                                arc, demand, prefetch, metadata                                                                                                                                 |        reads/s         |
| zfs.bytes                               |     global     |                                                                                                                                          read, write                                                                                                                                           |         KiB/s          |
| zfs.hits                                |     global     |                                                                                                                                          hits, misses                                                                                                                                          |       percentage       |
| zfs.dhits                               |     global     |                                                                                                                                          hits, misses                                                                                                                                          |       percentage       |
| zfs.phits                               |     global     |                                                                                                                                          hits, misses                                                                                                                                          |       percentage       |
| zfs.mhits                               |     global     |                                                                                                                                          hits, misses                                                                                                                                          |       percentage       |
| zfs.l2hits                              |     global     |                                                                                                                                          hits, misses                                                                                                                                          |       percentage       |
| zfs.list_hits                           |     global     |                                                                                                                                          hits, misses                                                                                                                                          |       percentage       |
| zfs.arc_size_breakdown                  |     global     |                                                                                                                                        recent, frequent                                                                                                                                        |       percentage       |
| zfs.memory_ops                          |     global     |                                                                                                                                  direct, throttled, indirect                                                                                                                                   |      operations/s      |
| zfs.important_ops                       |     global     |                                                                                                                        evict_skip, deleted, mutex_miss, hash_collisions                                                                                                                        |      operations/s      |
| zfs.actual_hits                         |     global     |                                                                                                                                          hits, misses                                                                                                                                          |       percentage       |
| zfs.demand_data_hits                    |     global     |                                                                                                                                          hits, misses                                                                                                                                          |       percentage       |
| zfs.prefetch_data_hits                  |     global     |                                                                                                                                          hits, misses                                                                                                                                          |       percentage       |
| zfs.hash_elements                       |     global     |                                                                                                                                          current, max                                                                                                                                          |        elements        |
| zfs.hash_chains                         |     global     |                                                                                                                                          current, max                                                                                                                                          |         chains         |
| ib.bytes                                |      port      |                                                                                                                                         Received, Sent                                                                                                                                         |       kilobits/s       |
| ib.packets                              |      port      |                                                                                                                 Received, Sent, Mcast rcvd, Mcast sent, Ucast rcvd, Ucast sent                                                                                                                 |       packets/s        |
| ib.errors                               |      port      |           Pkts_malformated, Pkts_rcvd_discarded, Pkts_sent_discarded, Tick_Wait_to_send, Pkts_missed_resource, Buffer_overrun, Link_Downed, Link_recovered, Link_integrity_err, Link_minor_errors, Pkts_rcvd_with_EBP, Pkts_rcvd_discarded_by_switch, Pkts_sent_discarded_by_switch            |        errors/s        |
| powersupply.capacity                    |    battery     |                                                                                                                                            capacity                                                                                                                                            |       percentage       |
| powersupply.charge                      |    battery     |                                                                                                                          empty_design, empty, now, full, full_design                                                                                                                           |           Ah           |
| powersupply.energy                      |    battery     |                                                                                                                          empty_design, empty, now, full, full_design                                                                                                                           |           Wh           |
| powersupply.voltage                     |    battery     |                                                                                                                             min_design, min, now, max, max_design                                                                                                                              |           V            |

</details>

## Monitoring Disks

> Live demo of disk monitoring at: **[http://london.netdata.rocks](https://registry.my-netdata.io/#menu_disk)**

Performance monitoring for Linux disks is quite complicated. The main reason is the plethora of disk technologies
available. There are a lot of different hardware disk technologies, but there are even more **virtual disk**
technologies that can provide additional storage features.

Hopefully, the Linux kernel provides many metrics that can provide deep insights of what our disks our doing. The kernel
measures all these metrics on all layers of storage: **virtual disks**, **physical disks** and **partitions of disks**.

### Monitored disk metrics

- **I/O bandwidth/s (kb/s)**
  The amount of data transferred from and to the disk.
- **Amount of discarded data (kb/s)**
- **I/O operations/s**
  The number of I/O operations completed.
- **Extended I/O operations/s**
  The number of extended I/O operations completed.
- **Queued I/O operations**
  The number of currently queued I/O operations. For traditional disks that execute commands one after another, one of
  them is being run by the disk and the rest are just waiting in a queue.
- **Backlog size (time in ms)**
  The expected duration of the currently queued I/O operations.
- **Utilization (time percentage)**
  The percentage of time the disk was busy with something. This is a very interesting metric, since for most disks, that
  execute commands sequentially, **this is the key indication of congestion**. A sequential disk that is 100% of the
  available time busy, has no time to do anything more, so even if the bandwidth or the number of operations executed by
  the disk is low, its capacity has been reached.
  Of course, for newer disk technologies (like fusion cards) that are capable to execute multiple commands in parallel,
  this metric is just meaningless.
- **Average I/O operation time (ms)**
  The average time for I/O requests issued to the device to be served. This includes the time spent by the requests in
  queue and the time spent servicing them.
- **Average I/O operation time for extended operations (ms)**
  The average time for extended I/O requests issued to the device to be served. This includes the time spent by the
  requests in queue and the time spent servicing them.
- **Average I/O operation size (kb)**
  The average amount of data of the completed I/O operations.
- **Average amount of discarded data (kb)**
  The average amount of data of the completed discard operations.
- **Average Service Time (ms)**
  The average service time for completed I/O operations. This metric is calculated using the total busy time of the disk
  and the number of completed operations. If the disk is able to execute multiple parallel operations the reporting
  average service time will be misleading.
- **Average Service Time for extended I/O operations (ms)**
  The average service time for completed extended I/O operations.
- **Merged I/O operations/s**
  The Linux kernel is capable of merging I/O operations. So, if two requests to read data from the disk are adjacent,
  the Linux kernel may merge them to one before giving them to disk. This metric measures the number of operations that
  have been merged by the Linux kernel.
- **Merged discard operations/s**
- **Total I/O time**
  The sum of the duration of all completed I/O operations. This number can exceed the interval if the disk is able to
  execute multiple I/O operations in parallel.
- **Space usage**
  For mounted disks, Netdata will provide a chart for their space, with 3 dimensions:
    1. free
    2. used
    3. reserved for root
- **inode usage**
  For mounted disks, Netdata will provide a chart for their inodes (number of file and directories), with 3 dimensions:
    1. free
    2. used
    3. reserved for root

### disk names

Netdata will automatically set the name of disks on the dashboard, from the mount point they are mounted, of course only
when they are mounted. Changes in mount points are not currently detected (you will have to restart Netdata to change
the name of the disk). To use disk IDs provided by `/dev/disk/by-id`, the `name disks by id` option should be enabled.
The `preferred disk ids` simple pattern allows choosing disk IDs to be used in the first place.

### performance metrics

By default, Netdata will enable monitoring metrics only when they are not zero. If they are constantly zero they are
ignored. Metrics that will start having values, after Netdata is started, will be detected and charts will be
automatically added to the dashboard (a refresh of the dashboard is needed for them to appear though). Set `yes` for a
chart instead of `auto` to enable it permanently. You can also set the `enable zero metrics` option to `yes` in
the `[global]` section which enables charts with zero metrics for all internal Netdata plugins.

Netdata categorizes all block devices in 3 categories:

1. physical disks (i.e. block devices that do not have child devices and are not partitions)
2. virtual disks (i.e. block devices that have child devices - like RAID devices)
3. disk partitions (i.e. block devices that are part of a physical disk)

Performance metrics are enabled by default for all disk devices, except partitions and not-mounted virtual disks. Of
course, you can enable/disable monitoring any block device by editing the Netdata configuration file.

### Netdata configuration

You can get the running Netdata configuration using this:

```sh
cd /etc/netdata
curl "http://localhost:19999/netdata.conf" >netdata.conf.new
mv netdata.conf.new netdata.conf
```

Then edit `netdata.conf` and find the following section. This is the basic plugin configuration.

```
[plugin:proc:/proc/diskstats]
  # enable new disks detected at runtime = yes
  # performance metrics for physical disks = auto
  # performance metrics for virtual disks = auto
  # performance metrics for partitions = no
  # bandwidth for all disks = auto
  # operations for all disks = auto
  # merged operations for all disks = auto
  # i/o time for all disks = auto
  # queued operations for all disks = auto
  # utilization percentage for all disks = auto
  # extended operations for all disks = auto
  # backlog for all disks = auto
  # bcache for all disks = auto
  # bcache priority stats update every = 0
  # remove charts of removed disks = yes
  # path to get block device = /sys/block/%s
  # path to get block device bcache = /sys/block/%s/bcache
  # path to get virtual block device = /sys/devices/virtual/block/%s
  # path to get block device infos = /sys/dev/block/%lu:%lu/%s
  # path to device mapper = /dev/mapper
  # path to /dev/disk/by-label = /dev/disk/by-label
  # path to /dev/disk/by-id = /dev/disk/by-id
  # path to /dev/vx/dsk = /dev/vx/dsk
  # name disks by id = no
  # preferred disk ids = *
  # exclude disks = loop* ram*
  # filename to monitor = /proc/diskstats
  # performance metrics for disks with major 8 = yes
```

For each virtual disk, physical disk and partition you will have a section like this:

```
[plugin:proc:/proc/diskstats:sda]
	# enable = yes
	# enable performance metrics = auto
	# bandwidth = auto
	# operations = auto
	# merged operations = auto
	# i/o time = auto
	# queued operations = auto
	# utilization percentage = auto
    # extended operations = auto
	# backlog = auto
```

For all configuration options:

- `auto` = enable monitoring if the collected values are not zero
- `yes` = enable monitoring
- `no` = disable monitoring

Of course, to set options, you will have to uncomment them. The comments show the internal defaults.

After saving `/etc/netdata/netdata.conf`, restart your Netdata to apply them.

#### Disabling performance metrics for individual device and to multiple devices by device type

You can pretty easy disable performance metrics for individual device, for ex.:

```
[plugin:proc:/proc/diskstats:sda]
	enable performance metrics = no
```

But sometimes you need disable performance metrics for all devices with the same type, to do it you need to figure out
device type from `/proc/diskstats` for ex.:

```
   7       0 loop0 1651 0 3452 168 0 0 0 0 0 8 168
   7       1 loop1 4955 0 11924 880 0 0 0 0 0 64 880
   7       2 loop2 36 0 216 4 0 0 0 0 0 4 4
   7       6 loop6 0 0 0 0 0 0 0 0 0 0 0
   7       7 loop7 0 0 0 0 0 0 0 0 0 0 0
 251       2 zram2 27487 0 219896 188 79953 0 639624 1640 0 1828 1828
 251       3 zram3 27348 0 218784 152 79952 0 639616 1960 0 2060 2104
```

All zram devices starts with `251` number and all loop devices starts with `7`.
So, to disable performance metrics for all loop devices you could add `performance metrics for disks with major 7 = no`
to `[plugin:proc:/proc/diskstats]` section.

```
[plugin:proc:/proc/diskstats]
       performance metrics for disks with major 7 = no
```

## Monitoring RAID arrays

### Monitored RAID array metrics

1. **Health** Number of failed disks in every array (aggregate chart).
2. **Disks stats**
    - total (number of devices array ideally would have)
    - inuse (number of devices currently are in use)
3. **Mismatch count**
    - unsynchronized blocks
4. **Current status**
    - resync in percent
    - recovery in percent
    - reshape in percent
    - check in percent
5. **Operation status** (if resync/recovery/reshape/check is active)
    - finish in minutes
    - speed in megabytes/s
6. **Nonredundant array availability**

#### configuration

```
[plugin:proc:/proc/mdstat]
  # faulty devices = yes
  # nonredundant arrays availability = yes
  # mismatch count = auto
  # disk stats = yes
  # operation status = yes
  # make charts obsolete = yes
  # filename to monitor = /proc/mdstat
  # mismatch_cnt filename to monitor = /sys/block/%s/md/mismatch_cnt
```

## Monitoring CPUs

The `/proc/stat` module monitors CPU utilization, interrupts, context switches, processes started/running, thermal
throttling, frequency, and idle states. It gathers this information from multiple files.

If your system has more than 50 processors (`physical processors * cores per processor * threads per core`), the Agent
automatically disables CPU thermal throttling, frequency, and idle state charts. To override this default, see the next
section on configuration.

### Configuration

The settings for monitoring CPUs is in the `[plugin:proc:/proc/stat]` of your `netdata.conf` file.

The `keep per core files open` option lets you reduce the number of file operations on multiple files.

If your system has more than 50 processors and you would like to see the CPU thermal throttling, frequency, and idle
state charts that are automatically disabled, you can set the following boolean options in the
`[plugin:proc:/proc/stat]` section.

```conf
    keep per core files open = yes
    keep cpuidle files open = yes
    core_throttle_count = yes
    package_throttle_count = yes
    cpu frequency = yes
    cpu idle states = yes
```

### CPU frequency

The module shows the current CPU frequency as set by the `cpufreq` kernel
module.

**Requirement:**
You need to have `CONFIG_CPU_FREQ` and (optionally) `CONFIG_CPU_FREQ_STAT`
enabled in your kernel.

`cpufreq` interface provides two different ways of getting the information
through `/sys/devices/system/cpu/cpu*/cpufreq/scaling_cur_freq`
and `/sys/devices/system/cpu/cpu*/cpufreq/stats/time_in_state` files. The latter is more accurate so it is preferred in
the module. `scaling_cur_freq` represents only the current CPU frequency, and doesn't account for any state changes
which happen between updates. The module switches back and forth between these two methods if governor is changed.

It produces one chart with multiple lines (one line per core).

#### configuration

`scaling_cur_freq filename to monitor` and `time_in_state filename to monitor` in the `[plugin:proc:/proc/stat]`
configuration section

### CPU idle states

The module monitors the usage of CPU idle states.

**Requirement:**
Your kernel needs to have `CONFIG_CPU_IDLE` enabled.

It produces one stacked chart per CPU, showing the percentage of time spent in
each state.

#### configuration

`schedstat filename to monitor`, `cpuidle name filename to monitor`, and `cpuidle time filename to monitor` in
the `[plugin:proc:/proc/stat]` configuration section

## Monitoring memory

### Monitored memory metrics

- Amount of memory swapped in/out
- Amount of memory paged from/to disk
- Number of memory page faults
- Number of out of memory kills
- Number of NUMA events

### Configuration

```conf
[plugin:proc:/proc/vmstat]
	filename to monitor = /proc/vmstat
	swap i/o = auto
	disk i/o = yes
	memory page faults = yes
	out of memory kills = yes
	system-wide numa metric summary = auto
```

## Monitoring Network Interfaces

### Monitored network interface metrics

- **Physical Network Interfaces Aggregated Bandwidth (kilobits/s)**
  The amount of data received and sent through all physical interfaces in the system. This is the source of data for the
  Net Inbound and Net Outbound dials in the System Overview section.

- **Bandwidth (kilobits/s)**
  The amount of data received and sent through the interface.

- **Packets (packets/s)**
  The number of packets received, packets sent, and multicast packets transmitted through the interface.

- **Interface Errors (errors/s)**
  The number of errors for the inbound and outbound traffic on the interface.

- **Interface Drops (drops/s)**
  The number of packets dropped for the inbound and outbound traffic on the interface.

- **Interface FIFO Buffer Errors (errors/s)**
  The number of FIFO buffer errors encountered while receiving and transmitting data through the interface.

- **Compressed Packets (packets/s)**
  The number of compressed packets transmitted or received by the device driver.

- **Network Interface Events (events/s)**
  The number of packet framing errors, collisions detected on the interface, and carrier losses detected by the device
  driver.

By default, Netdata will enable monitoring metrics only when they are not zero. If they are constantly zero they are
ignored. Metrics that will start having values, after Netdata is started, will be detected and charts will be
automatically added to the dashboard (a refresh of the dashboard is needed for them to appear though).

### Monitoring wireless network interfaces

The settings for monitoring wireless is in the `[plugin:proc:/proc/net/wireless]` section of your `netdata.conf` file.

```conf
    status for all interfaces = yes
    quality for all interfaces = yes
    discarded packets for all interfaces = yes
    missed beacon for all interface = yes
```

You can set the following values for each configuration option:

- `auto` = enable monitoring if the collected values are not zero
- `yes` = enable monitoring
- `no` = disable monitoring

#### Monitored wireless interface metrics

- **Status**
  The current state of the interface. This is a device-dependent option.

- **Link**    
  Overall quality of the link.

- **Level**
  Received signal strength (RSSI), which indicates how strong the received signal is.

- **Noise**
  Background noise level.

- **Discarded packets**
  Number of packets received with a different NWID or ESSID (`nwid`), unable to decrypt (`crypt`), hardware was not able
  to properly re-assemble the link layer fragments (`frag`), packets failed to deliver (`retry`), and packets lost in
  relation with specific wireless operations (`misc`).

- **Missed beacon**    
  Number of periodic beacons from the cell or the access point the interface has missed.

#### Wireless configuration

#### alarms

There are several alarms defined in `health.d/net.conf`.

The tricky ones are `inbound packets dropped` and `inbound packets dropped ratio`. They have quite a strict policy so
that they warn users about possible issues. These alarms can be annoying for some network configurations. It is
especially true for some bonding configurations if an interface is a child or a bonding interface itself. If it is
expected to have a certain number of drops on an interface for a certain network configuration, a separate alarm with
different triggering thresholds can be created or the existing one can be disabled for this specific interface. It can
be done with the help of the [families](/health/REFERENCE.md#alarm-line-families) line in the alarm configuration. For
example, if you want to disable the `inbound packets dropped` alarm for `eth0`, set `families: !eth0 *` in the alarm
definition for `template: inbound_packets_dropped`.

#### configuration

Module configuration:

```
[plugin:proc:/proc/net/dev]
  # filename to monitor = /proc/net/dev
  # path to get virtual interfaces = /sys/devices/virtual/net/%s
  # path to get net device speed = /sys/class/net/%s/speed
  # enable new interfaces detected at runtime = auto
  # bandwidth for all interfaces = auto
  # packets for all interfaces = auto
  # errors for all interfaces = auto
  # drops for all interfaces = auto
  # fifo for all interfaces = auto
  # compressed packets for all interfaces = auto
  # frames, collisions, carrier counters for all interfaces = auto
  # disable by default interfaces matching = lo fireqos* *-ifb
  # refresh interface speed every seconds = 10
```

Per interface configuration:

```
[plugin:proc:/proc/net/dev:enp0s3]
  # enabled = yes
  # virtual = no
  # bandwidth = auto
  # packets = auto
  # errors = auto
  # drops = auto
  # fifo = auto
  # compressed = auto
  # events = auto
```

## Linux Anti-DDoS

![image6](https://cloud.githubusercontent.com/assets/2662304/14253733/53550b16-fa95-11e5-8d9d-4ed171df4735.gif)

---

SYNPROXY is a TCP SYN packets proxy. It can be used to protect any TCP server (like a web server) from SYN floods and
similar DDos attacks.

SYNPROXY is a netfilter module, in the Linux kernel (since version 3.12). It is optimized to handle millions of packets
per second utilizing all CPUs available without any concurrency locking between the connections.

The net effect of this, is that the real servers will not notice any change during the attack. The valid TCP connections
will pass through and served, while the attack will be stopped at the firewall.

Netdata does not enable SYNPROXY. It just uses the SYNPROXY metrics exposed by your kernel, so you will first need to
configure it. The hard way is to run iptables SYNPROXY commands directly on the console. An easier way is to
use [FireHOL](https://firehol.org/), which, is a firewall manager for iptables. FireHOL can configure SYNPROXY using the
following setup guides:

- **[Working with SYNPROXY](https://github.com/firehol/firehol/wiki/Working-with-SYNPROXY)**
- **[Working with SYNPROXY and traps](https://github.com/firehol/firehol/wiki/Working-with-SYNPROXY-and-traps)**

### Real-time monitoring of Linux Anti-DDoS

Netdata is able to monitor in real-time (per second updates) the operation of the Linux Anti-DDoS protection.

It visualizes 4 charts:

1. TCP SYN Packets received on ports operated by SYNPROXY
2. TCP Cookies (valid, invalid, retransmits)
3. Connections Reopened
4. Entries used

Example image:

![ddos](https://cloud.githubusercontent.com/assets/2662304/14398891/6016e3fc-fdf0-11e5-942b-55de6a52cb66.gif)

See Linux Anti-DDoS in action
at: **[Netdata demo site (with SYNPROXY enabled)](https://registry.my-netdata.io/#menu_netfilter_submenu_synproxy)**

## Linux power supply

This module monitors various metrics reported by power supply drivers
on Linux. This allows tracking and alerting on things like remaining
battery capacity.

Depending on the underlying driver, it may provide the following charts
and metrics:

1. Capacity: The power supply capacity expressed as a percentage.

    - capacity_now

2. Charge: The charge for the power supply, expressed as amphours.

    - charge_full_design
    - charge_full
    - charge_now
    - charge_empty
    - charge_empty_design

3. Energy: The energy for the power supply, expressed as watthours.

    - energy_full_design
    - energy_full
    - energy_now
    - energy_empty
    - energy_empty_design

4. Voltage: The voltage for the power supply, expressed as volts.

    - voltage_max_design
    - voltage_max
    - voltage_now
    - voltage_min
    - voltage_min_design

#### configuration

```
[plugin:proc:/sys/class/power_supply]
  # battery capacity = yes
  # battery charge = no
  # battery energy = no
  # power supply voltage = no
  # keep files open = auto
  # directory to monitor = /sys/class/power_supply
```

#### notes

- Most drivers provide at least the first chart. Battery powered ACPI
  compliant systems (like most laptops) provide all but the third, but do
  not provide all the metrics for each chart.

- Current, energy, and voltages are reported with a *very* high precision
  by the power_supply framework. Usually, this is far higher than the
  actual hardware supports reporting, so expect to see changes in these
  charts jump instead of scaling smoothly.

- If `max` or `full` attribute is defined by the driver, but not a
  corresponding `min` or `empty` attribute, then Netdata will still provide
  the corresponding `min` or `empty`, which will then always read as zero.
  This way, alerts which match on these will still work.

## Infiniband interconnect

This module monitors every active Infiniband port. It provides generic counters statistics, and per-vendor hw-counters (
if vendor is supported).

### Monitored interface metrics

Each port will have its counters metrics monitored, grouped in the following charts:

- **Bandwidth usage**
  Sent/Received data, in KB/s

- **Packets Statistics**
  Sent/Received packets, in 3 categories: total, unicast and multicast.

- **Errors Statistics**
  Many errors counters are provided, presenting statistics for:
    - Packets: malformed, sent/received discarded by card/switch, missing resource
    - Link: downed, recovered, integrity error, minor error
    - Other events: Tick Wait to send, buffer overrun

If your vendor is supported, you'll also get HW-Counters statistics. These being vendor specific, please refer to their
documentation.

-

Mellanox: [see statistics documentation](https://community.mellanox.com/s/article/understanding-mlx5-linux-counters-and-status-parameters)

### configuration

Default configuration will monitor only enabled infiniband ports, and refresh newly activated or created ports every 30
seconds

```
[plugin:proc:/sys/class/infiniband]
  # dirname to monitor = /sys/class/infiniband
  # bandwidth counters = yes
  # packets counters = yes
  # errors counters = yes
  # hardware packets counters = auto
  # hardware errors counters = auto
  # monitor only ports being active = auto
  # disable by default interfaces matching = 
  # refresh ports state every seconds = 30
```

## IPC

### Monitored IPC metrics

- **number of messages in message queues**
- **amount of memory used by message queues**
- **number of semaphores**
- **number of semaphore arrays**
- **number of shared memory segments**
- **amount of memory used by shared memory segments**

As far as the message queue charts are dynamic, sane limits are applied for the number of dimensions per chart (the
limit is configurable).

### configuration

```
[plugin:proc:ipc]
  # message queues = yes
  # semaphore totals = yes
  # shared memory totals = yes
  # msg filename to monitor = /proc/sysvipc/msg
  # shm filename to monitor = /proc/sysvipc/shm
  # max dimensions in memory allowed = 50
```


