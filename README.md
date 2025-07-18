# ptcpdump

<div id="top"></div>

[![amd64-e2e](https://img.shields.io/github/actions/workflow/status/mozillazg/ptcpdump/test.yml?label=x86_64%20(amd64)%20e2e)](https://github.com/mozillazg/ptcpdump/actions/workflows/test.yml)
[![arm64-e2e](https://img.shields.io/circleci/build/gh/mozillazg/ptcpdump/master?label=aarch64%20(arm64)%20e2e)](https://app.circleci.com/pipelines/github/mozillazg/ptcpdump?branch=master)
[![Release](https://img.shields.io/github/v/release/mozillazg/ptcpdump)](https://github.com/mozillazg/ptcpdump/releases)
![Coveralls](https://img.shields.io/coverallsCoverage/github/mozillazg/ptcpdump?branch=master)
English | [中文](README.zh-CN.md)


ptcpdump is a tcpdump-compatible packet analyzer powered by eBPF,
automatically annotating packets with process/container/pod metadata when detectable.
Inspired by [jschwinger233/skbdump](https://github.com/jschwinger233/skbdump).

![](./docs/wireshark.png)

Table of Contents
=================

* [Features](#features)
* [Installation](#installation)
    * [Requirements](#requirements)
* [Usage](#usage)
    * [Example commands](#example-commands)
    * [Example output](#example-output)
    * [Running with Docker](#running-with-docker)
    * [Backend](#backend)
    * [Flags](#flags)
* [Compare with tcpdump](#compare-with-tcpdump)
* [Developing](#developing)
    * [Dependencies](#dependencies)
    * [Building](#building)


## Features

* 🔍 Process/container/pod-aware packet capture.
* 📦 Filter by: `--pid` (process), `--pname` (process name), `--container-id` (container), `--pod-name` (pod).
* 🎯 tcpdump-compatible flags (`-i`, `-w`, `-c`, `-s`, `-n`, `-C`, `-W`, `-A`, and more).
* 📜 Supports `pcap-filter(7)` syntax like tcpdump.
* 🌳 tcpdump-like output + process/container/pod context.
* 📑 Verbose mode shows detailed metadata for processes and containers/pods.
* 💾 PcapNG with embedded metadata (Wireshark-ready).
* 🌐 Cross-namespace capture (`--netns`).
* 🚀 Kernel-space BPF filtering (low overhead, reduces CPU usage).
* ⚡ Container runtime integration (Docker, containerd).


## Installation

You can download the statically linked executable for x86_64 and arm64 from the [releases page](https://github.com/mozillazg/ptcpdump/releases).


### Requirements

Linux kernel >= 5.2 (compiled with BPF and BTF support).

<details>

`ptcpdump` optionally requires debugfs. It has to be mounted in /sys/kernel/debug.
In case the folder is empty, it can be mounted with:

    mount -t debugfs none /sys/kernel/debug


The following kernel configuration is required. Building as Modules is also
possible.

| Option                    | Backend                   | Note                   |
|---------------------------|---------------------------|------------------------|
| CONFIG_BPF=y              | both                      | **Required**           |
| CONFIG_BPF_SYSCALL=y      | both                      | **Required**           |
| CONFIG_DEBUG_INFO=y       | both                      | **Required**           |
| CONFIG_DEBUG_INFO_BTF=y   | both                      | **Required**           |
| CONFIG_KPROBES=y          | both                      | **Required**           |
| CONFIG_KPROBE_EVENTS=y    | both                      | **Required**           |
| CONFIG_TRACEPOINTS=y      | both                      | **Required**           |
| CONFIG_PERF_EVENTS=y      | both                      | **Required**           |
| CONFIG_NET=y              | both                      | **Required**           |
| CONFIG_NET_SCHED=y        | tc                        | **Required**           |
| CONFIG_NET_CLS_BPF=y      | tc                        | **Required**           |
| CONFIG_NET_ACT_BPF=y      | tc                        | **Required**           |
| CONFIG_NET_SCH_INGRESS=y  | tc                        | **Required**           |
| CONFIG_CGROUPS=y          | cgroup-skb                | **Required**           |
| CONFIG_CGROUP_BPF=y       | cgroup-skb                | **Required**           |
| CONFIG_FILTER=y           | socket-filter             | **Required**           |
| CONFIG_BPF_TRAMPOLINE=y   | tp-btf                    | **Required**           |
| CONFIG_SECURITY=y         | both                      | Optional (Recommended) |
| CONFIG_BPF_TRAMPOLINE=y   | both                      | Optional (Recommended) |
| CONFIG_SOCK_CGROUP_DATA=y | both                      | Optional (Recommended) |
| CONFIG_BPF_JIT=y          | both                      | Optional (Recommended) |
| CONFIG_CGROUP_BPF=y       | tc, tp-btf, socket-filter | Optional (Recommended) |
| CONFIG_CGROUPS=y          | tc, tp-btf, socket-filter | Optional (Recommended) |

You can use `zgrep $OPTION /proc/config.gz` to validate whether an option is enabled.

</details>

<p align="right"><a href="#top">🔝</a></p>


## Usage

### Example commands

Filter like tcpdump:

    sudo ptcpdump -i eth0 tcp
    sudo ptcpdump -i eth0 -A -s 0 -n -v tcp and port 80 and host 10.10.1.1
    sudo ptcpdump -i any -s 0 -n -v -C 100MB -W 3 -w test.pcapng 'tcp and port 80 and host 10.10.1.1'
    sudo ptcpdump -i eth0 'tcp[tcpflags] & (tcp-syn|tcp-fin) != 0'

Multiple interfaces:

    sudo ptcpdump -i eth0 -i lo

Filter by process or user:

    sudo ptcpdump -i any --pid 1234 --pid 233 -f
    sudo ptcpdump -i any --pname curl
    sudo ptcpdump -i any --uid 1000

Capture by process via run target program:

    sudo ptcpdump -i any -- curl ubuntu.com

Filter by container or pod:

    sudo ptcpdump -i any --container-id 36f0310403b1
    sudo ptcpdump -i any --container-name test
    sudo ptcpdump -i any --pod-name test.default

Save data in PcapNG format:

    sudo ptcpdump -i any -w demo.pcapng
    sudo ptcpdump -i any -w - port 80 | tcpdump -n -r -
    sudo ptcpdump -i any -w - port 80 | tshark -r -


Capturing interfaces in other network namespaces:

    sudo ptcpdump -i lo --netns /run/netns/foo --netns /run/netns/bar
    sudo ptcpdump -i any --netns /run/netns/foobar
    sudo ptcpdump -i any --netns /proc/26/ns/net


<p align="right"><a href="#top">🔝</a></p>


### Example output


Default:

    09:32:09.718892 vethee2a302f wget.3553008 In IP 10.244.0.2.33426 > 139.178.84.217.80: Flags [S], seq 4113492822, win 64240, length 0, ParentProc [python3.834381], Container [test], Pod [test.default]
    09:32:09.718941 eth0 wget.3553008 Out IP 172.19.0.2.33426 > 139.178.84.217.80: Flags [S], seq 4113492822, win 64240, length 0, ParentProc [python3.834381], Container [test], Pod [test.default]

With `-q`:

    09:32:09.718892 vethee2a302f wget.3553008 In IP 10.244.0.2.33426 > 139.178.84.217.80: tcp 0, ParentProc [python3.834381], Container [test], Pod [test.default]
    09:32:09.718941 eth0 wget.3553008 Out IP 172.19.0.2.33426 > 139.178.84.217.80: tcp 0, ParentProc [python3.834381], Container [test], Pod [test.default]

With `-v`:

    13:44:41.529003 eth0 In IP (tos 0x4, ttl 45, id 45428, offset 0, flags [DF], proto TCP (6), length 52)
        139.178.84.217.443 > 172.19.0.2.42606: Flags [.], cksum 0x5284, seq 3173118145, ack 1385712707, win 118, options [nop,nop,TS val 134560683 ecr 1627716996], length 0
        Process (pid 553587, cmd /usr/bin/wget, args wget kernel.org)
        User (uid 1000)
        ParentProc (pid 553296, cmd /bin/sh, args sh)
        Container (name test, id d9028334568bf75a5a084963a8f98f78c56bba7f45f823b3780a135b71b91e95, image docker.io/library/alpine:3.18, labels {"io.cri-containerd.kind":"container","io.kubernetes.container.name":"test","io.kubernetes.pod.name":"test","io.kubernetes.pod.namespace":"default","io.kubernetes.pod.uid":"9e4bc54b-de48-4b1c-8b9e-54709f67ed0c"})
        Pod (name test, namespace default, UID 9e4bc54b-de48-4b1c-8b9e-54709f67ed0c, labels {"run":"test"}, annotations {"kubernetes.io/config.seen":"2024-07-21T12:41:00.460249620Z","kubernetes.io/config.source":"api"})

Using `--context` to limit context to include in the output:

<details>

    # --context=process
    09:32:09.718892 vethee2a302f wget.3553008 In IP 10.244.0.2.33426 > 139.178.84.217.80: Flags [S], seq 4113492822, win 64240, length 0
    
    # -v --context=process
    13:44:41.529003 eth0 In IP (tos 0x4, ttl 45, id 45428, offset 0, flags [DF], proto TCP (6), length 52)
        139.178.84.217.443 > 172.19.0.2.42606: Flags [.], cksum 0x5284, seq 3173118145, ack 1385712707, win 118, options [nop,nop,TS val 134560683 ecr 1627716996], length 0
        Process (pid 553587, cmd /usr/bin/wget, args wget kernel.org)
    
    # -v --context=process,parentproc,container,pod
    # or -v --context=process --context=parentproc --context=container --context=pod
    13:44:41.529003 eth0 In IP (tos 0x4, ttl 45, id 45428, offset 0, flags [DF], proto TCP (6), length 52)
        139.178.84.217.443 > 172.19.0.2.42606: Flags [.], cksum 0x5284, seq 3173118145, ack 1385712707, win 118, options [nop,nop,TS val 134560683 ecr 1627716996], length 0
        Process (pid 553587, cmd /usr/bin/wget, args wget kernel.org)
        ParentProc (pid 553296, cmd /bin/sh, args sh)
        Container (name test, id d9028334568bf75a5a084963a8f98f78c56bba7f45f823b3780a135b71b91e95, image docker.io/library/alpine:3.18, labels {"io.cri-containerd.kind":"container","io.kubernetes.container.name":"test","io.kubernetes.pod.name":"test","io.kubernetes.pod.namespace":"default","io.kubernetes.pod.uid":"9e4bc54b-de48-4b1c-8b9e-54709f67ed0c"})
        Pod (name test, namespace default, UID 9e4bc54b-de48-4b1c-8b9e-54709f67ed0c, labels {"run":"test"}, annotations {"kubernetes.io/config.seen":"2024-07-21T12:41:00.460249620Z","kubernetes.io/config.source":"api"})
</details>

With `-A`:

    14:44:34.457504 ens33 curl.205562 Out IP 10.0.2.15.39984 > 139.178.84.217.80: Flags [P.], seq 2722472188:2722472262, ack 892036871, win 64240, length 74, ParentProc [bash.180205]
    E..r.,@.@.o.
    .....T..0.P.E..5+g.P.......GET / HTTP/1.1
    Host: kernel.org
    User-Agent: curl/7.81.0
    Accept: */*
    

With `-x`:

    14:44:34.457504 ens33 curl.205562 Out IP 10.0.2.15.39984 > 139.178.84.217.80: Flags [P.], seq 2722472188:2722472262, ack 892036871, win 64240, length 74, ParentProc [bash.180205]
            0x0000:  4500 0072 de2c 4000 4006 6fbf 0a00 020f
            0x0010:  8bb2 54d9 9c30 0050 a245 a0fc 352b 6707
            0x0020:  5018 faf0 ecfe 0000 4745 5420 2f20 4854
            0x0030:  5450 2f31 2e31 0d0a 486f 7374 3a20 6b65
            0x0040:  726e 656c 2e6f 7267 0d0a 5573 6572 2d41
            0x0050:  6765 6e74 3a20 6375 726c 2f37 2e38 312e
            0x0060:  300d 0a41 6363 6570 743a 202a 2f2a 0d0a
            0x0070:  0d0a

With `-X`:

    14:44:34.457504 ens33 curl.205562 Out IP 10.0.2.15.39984 > 139.178.84.217.80: Flags [P.], seq 2722472188:2722472262, ack 892036871, win 64240, length 74, ParentProc [bash.180205]
            0x0000:  4500 0072 de2c 4000 4006 6fbf 0a00 020f  E..r.,@.@.o.....
            0x0010:  8bb2 54d9 9c30 0050 a245 a0fc 352b 6707  ..T..0.P.E..5+g.
            0x0020:  5018 faf0 ecfe 0000 4745 5420 2f20 4854  P.......GET / HT
            0x0030:  5450 2f31 2e31 0d0a 486f 7374 3a20 6b65  TP/1.1..Host: ke
            0x0040:  726e 656c 2e6f 7267 0d0a 5573 6572 2d41  rnel.org..User-A
            0x0050:  6765 6e74 3a20 6375 726c 2f37 2e38 312e  gent: curl/7.81.
            0x0060:  300d 0a41 6363 6570 743a 202a 2f2a 0d0a  0..Accept: */*..
            0x0070:  0d0a                                     ..


<p align="right"><a href="#top">🔝</a></p>


### Running with Docker

Docker images for `ptcpdump` are published at https://quay.io/repository/ptcpdump/ptcpdump.

    docker run --privileged --rm -t --net=host --pid=host \
      -v /sys/fs/cgroup:/sys/fs/cgroup:ro \
      -v /var/run:/var/run:ro \
      -v /run:/run:ro \
      quay.io/ptcpdump/ptcpdump:latest ptcpdump -i any -c 2 tcp

<p align="right"><a href="#top">🔝</a></p>


### Backend


ptcpdump supports specifying a particular eBPF technology for packet capture through the
`--backend` flag.

|                         | `tc`                      | `cgroup-skb`               | `socket-filter`               | `tp-btf`                |
|-------------------------|---------------------------|----------------------------|-------------------------------|-------------------------|
| eBPF Program Type       | `BPF_PROG_TYPE_SCHED_CLS` | `BPF_PROG_TYPE_CGROUP_SKB` | `BPF_PROG_TYPE_SOCKET_FILTER` | `BPF_PROG_TYPE_TRACING` |
| L2 data                 | ✅                         | ❌                          | ✅                             | ✅                       |
| Cross network namespace | ❌                         | ✅                          | ❌                             | ✅                       |
| Kernel version          | 5.2+                      | 5.2+                       | 5.4+                          | 5.5+                    |
| cgroup v2               | Recommended               | **Required**               | Recommended                   | Recommended             |


If this flag isn't specified, it defaults to `tc`.

<details>

* Running `curl http://1.1.1.1` on the host:

<details>

  * `--backend tc`:

        $ sudo ptcpdump -i any --backend tc host 1.1.1.1

        12:11:28.009276 ens33 curl.402475 Out IP 10.0.2.15.48448 > 1.1.1.1.80: Flags [S], seq 615672474, win 64240, options [mss 1460,sackOK,TS val 2168208063 ecr 0,nop,wscale 7], length 0, ParentProc [bash.321004]
        12:11:28.113779 ens33 curl.402475 In IP 1.1.1.1.80 > 10.0.2.15.48448: Flags [S.], seq 1810787293, ack 615672475, win 64240, options [mss 1460], length 0, ParentProc [bash.321004]
        12:11:28.113852 ens33 curl.402475 Out IP 10.0.2.15.48448 > 1.1.1.1.80: Flags [.], seq 615672475, ack 1810787294, win 64240, length 0, ParentProc [bash.321004]
        12:11:28.114216 ens33 curl.402475 Out IP 10.0.2.15.48448 > 1.1.1.1.80: Flags [P.], seq 615672475:615672545, ack 1810787294, win 64240, length 70, ParentProc [bash.321004]
        12:11:28.115383 ens33 curl.402475 In IP 1.1.1.1.80 > 10.0.2.15.48448: Flags [.], seq 1810787294, ack 615672545, win 64240, length 0, ParentProc [bash.321004]
        12:11:28.534486 ens33 curl.402475 In IP 1.1.1.1.80 > 10.0.2.15.48448: Flags [P.], seq 1810787294:1810787680, ack 615672545, win 64240, length 386, ParentProc [bash.321004]
        12:11:28.534751 ens33 curl.402475 Out IP 10.0.2.15.48448 > 1.1.1.1.80: Flags [.], seq 615672545, ack 1810787680, win 63854, length 0, ParentProc [bash.321004]
        12:11:28.536982 ens33 curl.402475 Out IP 10.0.2.15.48448 > 1.1.1.1.80: Flags [F.], seq 615672545, ack 1810787680, win 63854, length 0, ParentProc [bash.321004]
        12:11:28.538160 ens33 curl.402475 In IP 1.1.1.1.80 > 10.0.2.15.48448: Flags [.], seq 1810787680, ack 615672546, win 64239, length 0, ParentProc [bash.321004]
        12:11:28.642291 ens33 curl.402475 In IP 1.1.1.1.80 > 10.0.2.15.48448: Flags [FP.], seq 1810787680, ack 615672546, win 64239, length 0, ParentProc [bash.321004]
        12:11:28.642511 ens33 curl.402475 Out IP 10.0.2.15.48448 > 1.1.1.1.80: Flags [.], seq 615672546, ack 1810787681, win 63854, length 0, ParentProc [bash.321004]


  * `--backend cgroup-skb`:

        $ sudo ptcpdump -i any --backend cgroup-skb host 1.1.1.1

        12:11:28.009182 ens33 curl.402475 Out IP 10.0.2.15.48448 > 1.1.1.1.80: Flags [S], seq 615672474, win 64240, options [mss 146063 ecr 0,nop,wscale 7], length 0, Thread [curl.402475], ParentProc [bash.321004]
        12:11:28.113815 ens33 curl.402475 In IP 1.1.1.1.80 > 10.0.2.15.48448: Flags [S.], seq 1810787293, ack 615672475, win 64240, gth 0, ParentProc [bash.321004]
        12:11:28.113849 ens33 curl.402475 Out IP 10.0.2.15.48448 > 1.1.1.1.80: Flags [.], seq 615672475, ack 1810787294, win 64240, ash.321004]
        12:11:28.114212 ens33 curl.402475 Out IP 10.0.2.15.48448 > 1.1.1.1.80: Flags [P.], seq 615672475:615672545, ack 1810787294, hread [curl.402475], ParentProc [bash.321004]
        12:11:28.115409 ens33 curl.402475 In IP 1.1.1.1.80 > 10.0.2.15.48448: Flags [.], seq 1810787294, ack 615672545, win 64240, lsh.321004]
        12:11:28.534596 ens33 curl.402475 In IP 1.1.1.1.80 > 10.0.2.15.48448: Flags [P.], seq 1810787294:1810787680, ack 615672545, ParentProc [bash.321004]
        12:11:28.534738 ens33 curl.402475 Out IP 10.0.2.15.48448 > 1.1.1.1.80: Flags [.], seq 615672545, ack 1810787680, win 63854, ash.321004]
        12:11:28.536967 ens33 curl.402475 Out IP 10.0.2.15.48448 > 1.1.1.1.80: Flags [F.], seq 615672545, ack 1810787680, win 63854,.402475], ParentProc [bash.321004]
        12:11:28.538189 ens33 curl.402475 In IP 1.1.1.1.80 > 10.0.2.15.48448: Flags [.], seq 1810787680, ack 615672546, win 64239, lsh.321004]
        12:11:28.642419 ens33 curl.402475 Out IP 10.0.2.15.48448 > 1.1.1.1.80: Flags [.], seq 615672546, ack 1810787681, win 63854, ash.321004]


  * `--backend socket-filter`:

        $ sudo ptcpdump -i any --backend socket-filter host 1.1.1.1

        12:11:28.009426 ens33 curl.402475 Out IP 10.0.2.15.48448 > 1.1.1.1.80: Flags [S], seq 615672474, win 64240, options [mss 146063 ecr 0,nop,wscale 7], length 0, ParentProc [bash.321004]
        12:11:28.113762 ens33 curl.402475 In IP 1.1.1.1.80 > 10.0.2.15.48448: Flags [S.], seq 1810787293, ack 615672475, win 64240, gth 0, ParentProc [bash.321004]
        12:11:28.113861 ens33 curl.402475 Out IP 10.0.2.15.48448 > 1.1.1.1.80: Flags [.], seq 615672475, ack 1810787294, win 64240, ash.321004]
        12:11:28.114503 ens33 curl.402475 Out IP 10.0.2.15.48448 > 1.1.1.1.80: Flags [P.], seq 615672475:615672545, ack 1810787294, arentProc [bash.321004]
        12:11:28.115335 ens33 curl.402475 In IP 1.1.1.1.80 > 10.0.2.15.48448: Flags [.], seq 1810787294, ack 615672545, win 64240, lsh.321004]
        12:11:28.534424 ens33 curl.402475 In IP 1.1.1.1.80 > 10.0.2.15.48448: Flags [P.], seq 1810787294:1810787680, ack 615672545, ParentProc [bash.321004]
        12:11:28.534825 ens33 curl.402475 Out IP 10.0.2.15.48448 > 1.1.1.1.80: Flags [.], seq 615672545, ack 1810787680, win 63854, ash.321004]
        12:11:28.537088 ens33 curl.402475 Out IP 10.0.2.15.48448 > 1.1.1.1.80: Flags [F.], seq 615672545, ack 1810787680, win 63854,bash.321004]
        12:11:28.538153 ens33 curl.402475 In IP 1.1.1.1.80 > 10.0.2.15.48448: Flags [.], seq 1810787680, ack 615672546, win 64239, lsh.321004]
        12:11:28.642247 ens33 curl.402475 In IP 1.1.1.1.80 > 10.0.2.15.48448: Flags [FP.], seq 1810787680, ack 615672546, win 64239,bash.321004]
        12:11:28.642537 ens33 curl.402475 Out IP 10.0.2.15.48448 > 1.1.1.1.80: Flags [.], seq 615672546, ack 1810787681, win 63854, ash.321004]

  * `--backend tp-btf`:

        $ sudo ptcpdump -i any --backend tp-btf host 1.1.1.1

        12:11:28.009353 ens33 curl.402475 Out IP 10.0.2.15.48448 > 1.1.1.1.80: Flags [S], seq 615672474, win 64240, options [mss 146063 ecr 0,nop,wscale 7], length 0, ParentProc [bash.321004]
        12:11:28.113739 ens33 In IP 1.1.1.1.80 > 10.0.2.15.48448: Flags [S.], seq 1810787293, ack 615672475, win 64240, options [mss
        12:11:28.113857 ens33 curl.402475 Out IP 10.0.2.15.48448 > 1.1.1.1.80: Flags [.], seq 615672475, ack 1810787294, win 64240, ash.321004]
        12:11:28.114225 ens33 curl.402475 Out IP 10.0.2.15.48448 > 1.1.1.1.80: Flags [P.], seq 615672475:615672545, ack 1810787294, arentProc [bash.321004]
        12:11:28.115242 ens33 In IP 1.1.1.1.80 > 10.0.2.15.48448: Flags [.], seq 1810787294, ack 615672545, win 64240, length 0
        12:11:28.534245 ens33 In IP 1.1.1.1.80 > 10.0.2.15.48448: Flags [P.], seq 1810787294:1810787680, ack 615672545, win 64240, l
        12:11:28.534768 ens33 curl.402475 Out IP 10.0.2.15.48448 > 1.1.1.1.80: Flags [.], seq 615672545, ack 1810787680, win 63854, ash.321004]
        12:11:28.537038 ens33 curl.402475 Out IP 10.0.2.15.48448 > 1.1.1.1.80: Flags [F.], seq 615672545, ack 1810787680, win 63854,bash.321004]
        12:11:28.538129 ens33 In IP 1.1.1.1.80 > 10.0.2.15.48448: Flags [.], seq 1810787680, ack 615672546, win 64239, length 0
        12:11:28.642088 ens33 In IP 1.1.1.1.80 > 10.0.2.15.48448: Flags [FP.], seq 1810787680, ack 615672546, win 64239, length 0
        12:11:28.642523 ens33 curl.402475 Out IP 10.0.2.15.48448 > 1.1.1.1.80: Flags [.], seq 615672546, ack 1810787681, win 63854, ash.321004]

</details>


* Running `curl http://1.1.1.1` in a docker container:

<details>

  * `--backend tc`:

        $ sudo ptcpdump -i any --backend tc host 1.1.1.1

        12:20:31.336397 veth1d387b0 curl.405939 In IP 172.17.0.4.38670 > 1.1.1.1.80: Flags [S], seq 3064539219, win 64240, options [mss 1460,sackOK,TS val 1731159046 ecr 0,nop,wscale 7], length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.336533 docker0 curl.405939 In IP 172.17.0.4.38670 > 1.1.1.1.80: Flags [S], seq 3064539219, win 64240, options [mss 1460,sackOK,TS val 1731159046 ecr 0,nop,wscale 7], length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.336794 ens33 Out IP 10.0.2.15.38670 > 1.1.1.1.80: Flags [S], seq 3064539219, win 64240, options [mss 1460,sackOK,TS val 1731159046 ecr 0,nop,wscale 7], length 0
        12:20:31.468027 ens33 In IP 1.1.1.1.80 > 10.0.2.15.38670: Flags [S.], seq 488132001, ack 3064539220, win 64240, options [mss 1460], length 0
        12:20:31.467769 docker0 curl.405939 Out IP 1.1.1.1.80 > 172.17.0.4.38670: Flags [S.], seq 488132001, ack 3064539220, win 64240, options [mss 1460], length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.467781 veth1d387b0 curl.405939 Out IP 1.1.1.1.80 > 172.17.0.4.38670: Flags [S.], seq 488132001, ack 3064539220, win 64240, options [mss 1460], length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.468025 veth1d387b0 curl.405939 In IP 172.17.0.4.38670 > 1.1.1.1.80: Flags [.], seq 3064539220, ack 488132002, win 64240, length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.468042 docker0 curl.405939 In IP 172.17.0.4.38670 > 1.1.1.1.80: Flags [.], seq 3064539220, ack 488132002, win 64240, length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.468061 ens33 Out IP 10.0.2.15.38670 > 1.1.1.1.80: Flags [.], seq 3064539220, ack 488132002, win 64240, length 0
        12:20:31.468089 veth1d387b0 curl.405939 In IP 172.17.0.4.38670 > 1.1.1.1.80: Flags [P.], seq 3064539220:3064539291, ack 488132002, win 64240, length 71, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.468093 docker0 curl.405939 In IP 172.17.0.4.38670 > 1.1.1.1.80: Flags [P.], seq 3064539220:3064539291, ack 488132002, win 64240, length 71, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.468110 ens33 Out IP 10.0.2.15.38670 > 1.1.1.1.80: Flags [P.], seq 3064539220:3064539291, ack 488132002, win 64240, length 71
        12:20:31.468464 ens33 In IP 1.1.1.1.80 > 10.0.2.15.38670: Flags [.], seq 488132002, ack 3064539291, win 64240, length 0
        12:20:31.468535 docker0 curl.405939 Out IP 1.1.1.1.80 > 172.17.0.4.38670: Flags [.], seq 488132002, ack 3064539291, win 64240, length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.468558 veth1d387b0 curl.405939 Out IP 1.1.1.1.80 > 172.17.0.4.38670: Flags [.], seq 488132002, ack 3064539291, win 64240, length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.575461 ens33 In IP 1.1.1.1.80 > 10.0.2.15.38670: Flags [P.], seq 488132002:488132388, ack 3064539291, win 64240, length 386
        12:20:31.575576 docker0 curl.405939 Out IP 1.1.1.1.80 > 172.17.0.4.38670: Flags [P.], seq 488132002:488132388, ack 3064539291, win 64240, length 386, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.575613 veth1d387b0 curl.405939 Out IP 1.1.1.1.80 > 172.17.0.4.38670: Flags [P.], seq 488132002:488132388, ack 3064539291, win 64240, length 386, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.575877 veth1d387b0 curl.405939 In IP 172.17.0.4.38670 > 1.1.1.1.80: Flags [.], seq 3064539291, ack 488132388, win 63854, length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.575890 docker0 curl.405939 In IP 172.17.0.4.38670 > 1.1.1.1.80: Flags [.], seq 3064539291, ack 488132388, win 63854, length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.575916 ens33 Out IP 10.0.2.15.38670 > 1.1.1.1.80: Flags [.], seq 3064539291, ack 488132388, win 63854, length 0
        12:20:31.577079 veth1d387b0 curl.405939 In IP 172.17.0.4.38670 > 1.1.1.1.80: Flags [F.], seq 3064539291, ack 488132388, win 63854, length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.577107 docker0 curl.405939 In IP 172.17.0.4.38670 > 1.1.1.1.80: Flags [F.], seq 3064539291, ack 488132388, win 63854, length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.577146 ens33 Out IP 10.0.2.15.38670 > 1.1.1.1.80: Flags [F.], seq 3064539291, ack 488132388, win 63854, length 0
        12:20:31.577736 ens33 In IP 1.1.1.1.80 > 10.0.2.15.38670: Flags [.], seq 488132388, ack 3064539292, win 64239, length 0
        12:20:31.577761 docker0 curl.405939 Out IP 1.1.1.1.80 > 172.17.0.4.38670: Flags [.], seq 488132388, ack 3064539292, win 64239, length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.577773 veth1d387b0 curl.405939 Out IP 1.1.1.1.80 > 172.17.0.4.38670: Flags [.], seq 488132388, ack 3064539292, win 64239, length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.687029 ens33 In IP 1.1.1.1.80 > 10.0.2.15.38670: Flags [FP.], seq 488132388, ack 3064539292, win 64239, length 0
        12:20:31.687166 docker0 curl.405939 Out IP 1.1.1.1.80 > 172.17.0.4.38670: Flags [FP.], seq 488132388, ack 3064539292, win 64239, length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.687214 veth1d387b0 curl.405939 Out IP 1.1.1.1.80 > 172.17.0.4.38670: Flags [FP.], seq 488132388, ack 3064539292, win 64239, length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.687398 veth1d387b0 curl.405939 In IP 172.17.0.4.38670 > 1.1.1.1.80: Flags [.], seq 3064539292, ack 488132389, win 63854, length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.687413 docker0 curl.405939 In IP 172.17.0.4.38670 > 1.1.1.1.80: Flags [.], seq 3064539292, ack 488132389, win 63854, length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.687453 ens33 Out IP 10.0.2.15.38670 > 1.1.1.1.80: Flags [.], seq 3064539292, ack 488132389, win 63854, length 0

  * `--backend cgroup-skb`:

        $ sudo ptcpdump -i any --backend cgroup-skb host 1.1.1.1

        12:20:31.336108 45@4026533097 curl.405939 Out IP 172.17.0.4.38670 > 1.1.1.1.80: Flags [S], seq 3064539219, win 64240, options [mss 1460,sackOK,TS val 1731159046 ecr 0,nop,wscale 7], length 0, Thread [curl.405939], ParentProc [bash.405653], Container [musing_banach]
        12:20:31.467819 45@4026533097 curl.405939 In IP 1.1.1.1.80 > 172.17.0.4.38670: Flags [S.], seq 488132001, ack 3064539220, win 64240, options [mss 1460], length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.467876 45@4026533097 curl.405939 Out IP 172.17.0.4.38670 > 1.1.1.1.80: Flags [.], seq 3064539220, ack 488132002, win 64240, length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.468072 45@4026533097 curl.405939 Out IP 172.17.0.4.38670 > 1.1.1.1.80: Flags [P.], seq 3064539220:3064539291, ack 488132002, win 64240, length 71, Thread [curl.405939], ParentProc [bash.405653], Container [musing_banach]
        12:20:31.468681 45@4026533097 curl.405939 In IP 1.1.1.1.80 > 172.17.0.4.38670: Flags [.], seq 488132002, ack 3064539291, win 64240, length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.575750 45@4026533097 curl.405939 In IP 1.1.1.1.80 > 172.17.0.4.38670: Flags [P.], seq 488132002:488132388, ack 3064539291, win 64240, length 386, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.575848 45@4026533097 curl.405939 Out IP 172.17.0.4.38670 > 1.1.1.1.80: Flags [.], seq 3064539291, ack 488132388, win 63854, length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.576982 45@4026533097 curl.405939 Out IP 172.17.0.4.38670 > 1.1.1.1.80: Flags [F.], seq 3064539291, ack 488132388, win 63854, length 0, Thread [curl.405939], ParentProc [bash.405653], Container [musing_banach]
        12:20:31.577843 45@4026533097 curl.405939 In IP 1.1.1.1.80 > 172.17.0.4.38670: Flags [.], seq 488132388, ack 3064539292, win 64239, length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.687357 45@4026533097 curl.405939 Out IP 172.17.0.4.38670 > 1.1.1.1.80: Flags [.], seq 3064539292, ack 488132389, win 63854, length 0, ParentProc [bash.405653], Container [musing_banach]


  * `--backend socket-filter`:

        $ sudo ptcpdump -i any --backend socket-filter host 1.1.1.1

        12:20:31.336456 docker0 curl.405939 In IP 172.17.0.4.38670 > 1.1.1.1.80: Flags [S], seq 3064539219, win 64240, options [mss 1460,sackOK,TS val 1731159046 ecr 0,nop,wscale 7], length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.336818 ens33 Out IP 10.0.2.15.38670 > 1.1.1.1.80: Flags [S], seq 3064539219, win 64240, options [mss 1460,sackOK,TS val 1731159046 ecr 0,nop,wscale 7], length 0
        12:20:31.467700 ens33 In IP 1.1.1.1.80 > 10.0.2.15.38670: Flags [S.], seq 488132001, ack 3064539220, win 64240, options [mss 1460], length 0
        12:20:31.467776 docker0 curl.405939 Out IP 1.1.1.1.80 > 172.17.0.4.38670: Flags [S.], seq 488132001, ack 3064539220, win 64240, options [mss 1460], length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.467784 veth1d387b0 curl.405939 Out IP 1.1.1.1.80 > 172.17.0.4.38670: Flags [S.], seq 488132001, ack 3064539220, win 64240, options [mss 1460], length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.468030 docker0 curl.405939 In IP 172.17.0.4.38670 > 1.1.1.1.80: Flags [.], seq 3064539220, ack 488132002, win 64240, length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.468066 ens33 Out IP 10.0.2.15.38670 > 1.1.1.1.80: Flags [.], seq 3064539220, ack 488132002, win 64240, length 0
        12:20:31.468092 docker0 curl.405939 In IP 172.17.0.4.38670 > 1.1.1.1.80: Flags [P.], seq 3064539220:3064539291, ack 488132002, win 64240, length 71, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.468122 ens33 Out IP 10.0.2.15.38670 > 1.1.1.1.80: Flags [P.], seq 3064539220:3064539291, ack 488132002, win 64240, length 71
        12:20:31.468461 ens33 In IP 1.1.1.1.80 > 10.0.2.15.38670: Flags [.], seq 488132002, ack 3064539291, win 64240, length 0
        12:20:31.468552 docker0 curl.405939 Out IP 1.1.1.1.80 > 172.17.0.4.38670: Flags [.], seq 488132002, ack 3064539291, win 64240, length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.468565 veth1d387b0 curl.405939 Out IP 1.1.1.1.80 > 172.17.0.4.38670: Flags [.], seq 488132002, ack 3064539291, win 64240, length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.575416 ens33 In IP 1.1.1.1.80 > 10.0.2.15.38670: Flags [P.], seq 488132002:488132388, ack 3064539291, win 64240, length 386
        12:20:31.575601 docker0 curl.405939 Out IP 1.1.1.1.80 > 172.17.0.4.38670: Flags [P.], seq 488132002:488132388, ack 3064539291, win 64240, length 386, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.575623 veth1d387b0 curl.405939 Out IP 1.1.1.1.80 > 172.17.0.4.38670: Flags [P.], seq 488132002:488132388, ack 3064539291, win 64240, length 386, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.575889 docker0 curl.405939 In IP 172.17.0.4.38670 > 1.1.1.1.80: Flags [.], seq 3064539291, ack 488132388, win 63854, length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.575928 ens33 Out IP 10.0.2.15.38670 > 1.1.1.1.80: Flags [.], seq 3064539291, ack 488132388, win 63854, length 0
        12:20:31.577085 docker0 curl.405939 In IP 172.17.0.4.38670 > 1.1.1.1.80: Flags [F.], seq 3064539291, ack 488132388, win 63854, length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.577153 ens33 Out IP 10.0.2.15.38670 > 1.1.1.1.80: Flags [F.], seq 3064539291, ack 488132388, win 63854, length 0
        12:20:31.577733 ens33 In IP 1.1.1.1.80 > 10.0.2.15.38670: Flags [.], seq 488132388, ack 3064539292, win 64239, length 0
        12:20:31.577770 docker0 curl.405939 Out IP 1.1.1.1.80 > 172.17.0.4.38670: Flags [.], seq 488132388, ack 3064539292, win 64239, length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.577778 veth1d387b0 curl.405939 Out IP 1.1.1.1.80 > 172.17.0.4.38670: Flags [.], seq 488132388, ack 3064539292, win 64239, length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.687015 ens33 In IP 1.1.1.1.80 > 10.0.2.15.38670: Flags [FP.], seq 488132388, ack 3064539292, win 64239, length 0
        12:20:31.687206 docker0 curl.405939 Out IP 1.1.1.1.80 > 172.17.0.4.38670: Flags [FP.], seq 488132388, ack 3064539292, win 64239, length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.687223 veth1d387b0 curl.405939 Out IP 1.1.1.1.80 > 172.17.0.4.38670: Flags [FP.], seq 488132388, ack 3064539292, win 64239, length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.687409 docker0 curl.405939 In IP 172.17.0.4.38670 > 1.1.1.1.80: Flags [.], seq 3064539292, ack 488132389, win 63854, length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.687464 ens33 Out IP 10.0.2.15.38670 > 1.1.1.1.80: Flags [.], seq 3064539292, ack 488132389, win 63854, length 0

  * `--backend tp-btf`:

        $ sudo ptcpdump -i any --backend tp-btf host 1.1.1.1

        12:20:31.336316 eth0@4026533097 curl.405939 Out IP 172.17.0.4.38670 > 1.1.1.1.80: Flags [S], seq 3064539219, win 64240, options [mss 1460,sackOK,TS val 1731159046 ecr 0,nop,wscale 7], length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.336382 veth1d387b0 curl.405939 In IP 172.17.0.4.38670 > 1.1.1.1.80: Flags [S], seq 3064539219, win 64240, options [mss 1460,sackOK,TS val 1731159046 ecr 0,nop,wscale 7], length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.336443 docker0 curl.405939 In IP 172.17.0.4.38670 > 1.1.1.1.80: Flags [S], seq 3064539219, win 64240, options [mss 1460,sackOK,TS val 1731159046 ecr 0,nop,wscale 7], length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.336801 ens33 Out IP 10.0.2.15.38670 > 1.1.1.1.80: Flags [S], seq 3064539219, win 64240, options [mss 1460,sackOK,TS val 1731159046 ecr 0,nop,wscale 7], length 0
        12:20:31.467682 ens33 In IP 1.1.1.1.80 > 10.0.2.15.38670: Flags [S.], seq 488132001, ack 3064539220, win 64240, options [mss 1460], length 0
        12:20:31.467773 docker0 curl.405939 Out IP 1.1.1.1.80 > 172.17.0.4.38670: Flags [S.], seq 488132001, ack 3064539220, win 64240, options [mss 1460], length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.467783 veth1d387b0 curl.405939 Out IP 1.1.1.1.80 > 172.17.0.4.38670: Flags [S.], seq 488132001, ack 3064539220, win 64240, options [mss 1460], length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.467811 eth0@4026533097 curl.405939 In IP 1.1.1.1.80 > 172.17.0.4.38670: Flags [S.], seq 488132001, ack 3064539220, win 64240, options [mss 1460], length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.468005 eth0@4026533097 curl.405939 Out IP 172.17.0.4.38670 > 1.1.1.1.80: Flags [.], seq 3064539220, ack 488132002, win 64240, length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.468022 veth1d387b0 curl.405939 In IP 172.17.0.4.38670 > 1.1.1.1.80: Flags [.], seq 3064539220, ack 488132002, win 64240, length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.468029 docker0 curl.405939 In IP 172.17.0.4.38670 > 1.1.1.1.80: Flags [.], seq 3064539220, ack 488132002, win 64240, length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.468063 ens33 Out IP 10.0.2.15.38670 > 1.1.1.1.80: Flags [.], seq 3064539220, ack 488132002, win 64240, length 0
        12:20:31.468078 eth0@4026533097 curl.405939 Out IP 172.17.0.4.38670 > 1.1.1.1.80: Flags [P.], seq 3064539220:3064539291, ack 488132002, win 64240, length 71, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.468085 veth1d387b0 curl.405939 In IP 172.17.0.4.38670 > 1.1.1.1.80: Flags [P.], seq 3064539220:3064539291, ack 488132002, win 64240, length 71, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.468091 docker0 curl.405939 In IP 172.17.0.4.38670 > 1.1.1.1.80: Flags [P.], seq 3064539220:3064539291, ack 488132002, win 64240, length 71, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.468112 ens33 Out IP 10.0.2.15.38670 > 1.1.1.1.80: Flags [P.], seq 3064539220:3064539291, ack 488132002, win 64240, length 71
        12:20:31.468446 ens33 In IP 1.1.1.1.80 > 10.0.2.15.38670: Flags [.], seq 488132002, ack 3064539291, win 64240, length 0
        12:20:31.468543 docker0 curl.405939 Out IP 1.1.1.1.80 > 172.17.0.4.38670: Flags [.], seq 488132002, ack 3064539291, win 64240, length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.468562 veth1d387b0 curl.405939 Out IP 1.1.1.1.80 > 172.17.0.4.38670: Flags [.], seq 488132002, ack 3064539291, win 64240, length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.468668 eth0@4026533097 curl.405939 In IP 1.1.1.1.80 > 172.17.0.4.38670: Flags [.], seq 488132002, ack 3064539291, win 64240, length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.575358 ens33 In IP 1.1.1.1.80 > 10.0.2.15.38670: Flags [P.], seq 488132002:488132388, ack 3064539291, win 64240, length 386
        12:20:31.575586 docker0 curl.405939 Out IP 1.1.1.1.80 > 172.17.0.4.38670: Flags [P.], seq 488132002:488132388, ack 3064539291, win 64240, length 386, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.575617 veth1d387b0 curl.405939 Out IP 1.1.1.1.80 > 172.17.0.4.38670: Flags [P.], seq 488132002:488132388, ack 3064539291, win 64240, length 386, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.575732 eth0@4026533097 curl.405939 In IP 1.1.1.1.80 > 172.17.0.4.38670: Flags [P.], seq 488132002:488132388, ack 3064539291, win 64240, length 386, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.575855 eth0@4026533097 curl.405939 Out IP 172.17.0.4.38670 > 1.1.1.1.80: Flags [.], seq 3064539291, ack 488132388, win 63854, length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.575870 veth1d387b0 curl.405939 In IP 172.17.0.4.38670 > 1.1.1.1.80: Flags [.], seq 3064539291, ack 488132388, win 63854, length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.575883 docker0 curl.405939 In IP 172.17.0.4.38670 > 1.1.1.1.80: Flags [.], seq 3064539291, ack 488132388, win 63854, length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.575920 ens33 Out IP 10.0.2.15.38670 > 1.1.1.1.80: Flags [.], seq 3064539291, ack 488132388, win 63854, length 0
        12:20:31.577059 eth0@4026533097 curl.405939 Out IP 172.17.0.4.38670 > 1.1.1.1.80: Flags [F.], seq 3064539291, ack 488132388, win 63854, length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.577074 veth1d387b0 curl.405939 In IP 172.17.0.4.38670 > 1.1.1.1.80: Flags [F.], seq 3064539291, ack 488132388, win 63854, length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.577082 docker0 curl.405939 In IP 172.17.0.4.38670 > 1.1.1.1.80: Flags [F.], seq 3064539291, ack 488132388, win 63854, length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.577148 ens33 Out IP 10.0.2.15.38670 > 1.1.1.1.80: Flags [F.], seq 3064539291, ack 488132388, win 63854, length 0
        12:20:31.577704 ens33 In IP 1.1.1.1.80 > 10.0.2.15.38670: Flags [.], seq 488132388, ack 3064539292, win 64239, length 0
        12:20:31.577764 docker0 curl.405939 Out IP 1.1.1.1.80 > 172.17.0.4.38670: Flags [.], seq 488132388, ack 3064539292, win 64239, length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.577774 veth1d387b0 curl.405939 Out IP 1.1.1.1.80 > 172.17.0.4.38670: Flags [.], seq 488132388, ack 3064539292, win 64239, length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.577835 eth0@4026533097 curl.405939 In IP 1.1.1.1.80 > 172.17.0.4.38670: Flags [.], seq 488132388, ack 3064539292, win 64239, length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.686955 ens33 In IP 1.1.1.1.80 > 10.0.2.15.38670: Flags [FP.], seq 488132388, ack 3064539292, win 64239, length 0
        12:20:31.687183 docker0 curl.405939 Out IP 1.1.1.1.80 > 172.17.0.4.38670: Flags [FP.], seq 488132388, ack 3064539292, win 64239, length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.687218 veth1d387b0 curl.405939 Out IP 1.1.1.1.80 > 172.17.0.4.38670: Flags [FP.], seq 488132388, ack 3064539292, win 64239, length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.687316 eth0@4026533097 curl.405939 In IP 1.1.1.1.80 > 172.17.0.4.38670: Flags [FP.], seq 488132388, ack 3064539292, win 64239, length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.687369 eth0@4026533097 curl.405939 Out IP 172.17.0.4.38670 > 1.1.1.1.80: Flags [.], seq 3064539292, ack 488132389, win 63854, length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.687388 veth1d387b0 curl.405939 In IP 172.17.0.4.38670 > 1.1.1.1.80: Flags [.], seq 3064539292, ack 488132389, win 63854, length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.687404 docker0 curl.405939 In IP 172.17.0.4.38670 > 1.1.1.1.80: Flags [.], seq 3064539292, ack 488132389, win 63854, length 0, ParentProc [bash.405653], Container [musing_banach]
        12:20:31.687457 ens33 Out IP 10.0.2.15.38670 > 1.1.1.1.80: Flags [.], seq 3064539292, ack 488132389, win 63854, length 0

</details>


</details>

<p align="right"><a href="#top">🔝</a></p>


### Flags

<details>

    Usage:
      ptcpdump [flags] [expression] [-- command [args]]
    
    Examples:
      sudo ptcpdump -i any tcp
      sudo ptcpdump -i eth0 -i lo
      sudo ptcpdump -i eth0 --pid 1234 port 80 and host 10.10.1.1
      sudo ptcpdump -i any --pname curl -A
      sudo ptcpdump -i any --container-id 36f0310403b1
      sudo ptcpdump -i any --container-name test
      sudo ptcpdump -i any -- curl ubuntu.com
      sudo ptcpdump -i any -w ptcpdump.pcapng
      sudo ptcpdump -i any -w - | tcpdump -n -r -
      sudo ptcpdump -i any -w - | tshark -r -
      ptcpdump -r ptcpdump.pcapng
    
    Expression: see "man 7 pcap-filter"
    
    Flags:
          --backend string                               Specify the backend to use for capturing packets. Possible values are "tc", "cgroup-skb", "tp-btf" and "socket-filter" (default "tc")
          --container-id string                          Filter by container id (only TCP and UDP packets are supported)
          --container-name string                        Filter by container name (only TCP and UDP packets are supported)
          --containerd-address string                    Address of containerd service (default "/run/containerd/containerd.sock")
          --context strings                              Specify which context information to include in the output (default [process,thread,parentproc,user,container,pod])
          --count                                        Print only on stdout the packet count when reading capture file instead of parsing/printing the packets
          --cri-runtime-address string                   Address of CRI container runtime service (default: uses in order the first successful one of [/var/run/dockershim.sock, /var/run/cri-dockerd.sock, /run/crio/crio.sock, /run/containerd/containerd.sock])
          --delay-before-handle-packet-events duration   Delay some durations before handle packet events
      -Q, --direction string                             Choose send/receive direction for which packets should be captured. Possible values are 'in', 'out' and 'inout' (default "inout")
          --disable-reverse-match                        Disable reverse match for TCP and UDP packets.
          --docker-address string                        Address of Docker Engine service (default "/var/run/docker.sock")
          --embed-keylog-to-pcapng -- CMD [ARGS]         Write TLS Key Log file to this path (experimental: only support unstripped Go binary and must combined with -- CMD [ARGS])
          --event-chan-size uint                         Size of event chan (default 20)
          --exec-events-worker-number uint               Number of worker to handle exec events (default 50)
      -F, --expression-file string                       Use file as input for the filter expression. An additional expression given on the command line is ignored.
      -W, --file-count uint                              Used in conjunction with the -C option, this will limit the number of files created to the specified number, and begin overwriting files from the beginning, thus creating a 'rotating' buffer.
      -C, --file-size fileSize                           Before writing a raw packet to a savefile, check whether the file is currently larger than file_size and, if so, close the current savefile and open a new one. Savefiles after the first savefile will have the name specified with the -w flag, with a number after it, starting at 1 and continuing upward.
      -f, --follow-forks                                 Trace child processes as they are created by currently traced processes when filter by process
      -h, --help                                         help for ptcpdump
      -i, --interface strings                            Interfaces to capture (default [lo])
          --kernel-btf string                            specify kernel BTF file (default: uses in order the first successful one of [/sys/kernel/btf/vmlinux, /var/lib/ptcpdump/btf/vmlinux, /var/lib/ptcpdump/btf/vmlinux-$(uname -r), /var/lib/ptcpdump/btf/$(uname -r).btf, download BTF file from https://mirrors.openanolis.cn/coolbpf/btf/ and https://github.com/aquasecurity/btfhub-archive/]
      -D, --list-interfaces                              Print the list of the network interfaces available on the system
          --log-level string                             Set the logging level ("debug", "info", "warn", "error", "fatal") (default "warn")
          --micro                                        Shorthands for --time-stamp-precision=micro
          --nano                                         Shorthands for --time-stamp-precision=nano
          --netns strings                                Path to an network namespace file or name (default [/proc/self/ns/net])
      -n, --no-convert-addr count                        Don't convert addresses (i.e., host addresses, port numbers, etc.) to names
      -#, --number                                       Print an optional packet number at the beginning of the line
          --oneline                                      Print parsed packet output in a single line
          --pid uints                                    Filter by process IDs (only TCP and UDP packets are supported) (default [])
          --pname string                                 Filter by process name (only TCP and UDP packets are supported)
          --pod-name string                              Filter by pod name (format: NAME.NAMESPACE, only TCP and UDP packets are supported)
          --print                                        Print parsed packet output, even if the raw packets are being saved to a file with the -w flag
      -A, --print-data-in-ascii                          Print each packet (minus its link level header) in ASCII
      -x, --print-data-in-hex count                      When parsing and printing, in addition to printing the headers of each packet, print the data of each packet in hex
      -X, --print-data-in-hex-ascii count                When parsing and printing, in addition to printing the headers of each packet, print the data of each packet in hex and ASCII
      -t, --print-timestamp count                        control the format of the timestamp printed in the output
      -q, --quiet                                        Quiet output. Print less protocol information so output lines are shorter
      -r, --read-file string                             Read packets from file (which was created with the -w option). e.g. ptcpdump.pcapng
      -c, --receive-count uint                           Exit after receiving count packets
      -s, --snapshot-length uint32                       Snarf snaplen bytes of data from each packet rather than the default of 262144 bytes (default 262144)
          --time-stamp-precision string                  When capturing, set the time stamp precision for the capture to the format (default "micro")
          --uid uints                                    Filter by user IDs (only TCP and UDP packets are supported) (default [])
      -v, --verbose count                                When parsing and printing, produce (slightly more) verbose output
          --version                                      Print the ptcpdump and libpcap version strings and exit
      -w, --write-file string                            Write the raw packets to file rather than parsing and printing them out. They can later be printed with the -r option. Standard output is used if file is '-'. e.g. ptcpdump.pcapng
          --write-keylog-file -- CMD [ARGS]              Write TLS Key Log file to this path (experimental: only support unstripped Go binary and must combined with -- CMD [ARGS])


</details>

<p align="right"><a href="#top">🔝</a></p>


## Compare with tcpdump

| Options                                           | tcpdump | ptcpdump                 |
|---------------------------------------------------|---------|--------------------------|
| *expression*                                      | ✅       | ✅                        |
| -i *interface*, --interface=*interface*           | ✅       | ✅                        |
| -w *x.pcapng*                                     | ✅       | ✅ (with process info)    |
| -w *x.pcap*                                       | ✅       | ✅ (without process info) |
| -w *-*                                            | ✅       | ✅                        |
| -r *x.pcapng*, -r *x.pcap*                        | ✅       | ✅                        |
| -r *-*                                            | ✅       | ✅                        |
| --pid *process_id*                                |         | ✅                        |
| --pname *process_name*                            |         | ✅                        |
| --uid *user_id*                                   |         | ✅                        |
| --container-id *container_id*                     |         | ✅                        |
| --container-name *container_name*                 |         | ✅                        |
| --pod-name *pod_name.namespace*                   |         | ✅                        |
| -f, --follow-forks                                |         | ✅                        |
| -- *command [args]*                               |         | ✅                        |
| --netns *path_to_net_ns*                          |         | ✅                        |
| --print                                           | ✅       | ✅                        |
| -A                                                | ✅       | ✅                        |
| -B *bufer_size*, --buffer-size=*buffer_size*      | ✅       |                          |
| -c *count*                                        | ✅       | ✅                        |
| --count                                           | ✅       | ✅                        |
| -C *file_size                                     | ✅       | ✅                        |
| -d                                                | ✅       |                          |
| -dd                                               | ✅       |                          |
| -ddd                                              | ✅       |                          |
| -D, --list-interfaces                             | ✅       | ✅                        |
| -e                                                | ✅       |                          |
| -f                                                | ✅       | ⛔                        |
| -F *file*                                         | ✅       | ✅                        |
| -G *rotate_seconds*                               | ✅       |                          |
| -h, --help                                        | ✅       | ✅                        |
| -H                                                | ✅       |                          |
| -I, --monitor-mode                                | ✅       |                          |
| --immediate-mode                                  | ✅       |                          |
| -j *tstamp_type*, --time-stamp-type=*tstamp_type* | ✅       |                          |
| --time-stamp-precision=*tstamp_precision*         | ✅       | ✅                        |
| -J, --list-time-stamp-types                       | ✅       |                          |
| --micro                                           | ✅       | ✅                        |
| --nano                                            | ✅       | ✅                        |
| -K, --dont-verify-checksums                       | ✅       |                          |
| -l                                                | ✅       |                          |
| -L, --list-data-link-types                        | ✅       |                          |
| -m *module*                                       | ✅       |                          |
| -M *secret*                                       | ✅       |                          |
| -n                                                | ✅       | ✅                        |
| -N                                                | ✅       |                          |
| -#, --number                                      | ✅       | ✅                        |
| -O, --no-optimize                                 | ✅       |                          |
| -p, --no-promiscuous-mode                         | ✅       | ⛔                        |
| -q                                                | ✅       | ✅                        |
| -Q *direction*, --direction=*direction*           | ✅       | ✅                        |
| -S, --absolute-tcp-sequence-numbers               | ✅       |                          |
| -s *snaplen*, --snapshot-length=*snaplen*         | ✅       | ✅                        |
| -T *type*                                         | ✅       |                          |
| -t                                                | ✅       | ✅                        |
| -tt                                               | ✅       | ✅                        |
| -ttt                                              | ✅       | ✅                        |
| -tttt                                             | ✅       | ✅                        |
| -ttttt                                            | ✅       | ✅                        |
| -u                                                | ✅       |                          |
| -U, --packet-buffered                             | ✅       |                          |
| -y *datalinktype*, --linktype=*datalinktype*      | ✅       |                          |
| -v                                                | ✅       | ✅                        |
| -vv                                               | ✅       | ⭕                        |
| -vvv                                              | ✅       | ⭕                        |
| -V *file*                                         | ✅       |                          |
| --version                                         | ✅       | ✅                        |
| -W *filecont*                                     | ✅       | ✅                        |
| -x                                                | ✅       | ✅                        |
| -xx                                               | ✅       | ✅                        |
| -X                                                | ✅       | ✅                        |
| -XX                                               | ✅       | ✅                        |
| -z *postrotate-command*                           | ✅       |                          |
| -Z *user*, --relinquish-privileges=*user*         | ✅       |                          |

<p align="right"><a href="#top">🔝</a></p>



## Developing


### Dependencies

Ensure you have the following dependencies installed on your system:

* Go >= 1.23
* Clang/LLVM >= 14
* Bison >= 3.8
* Lex/Flex >= 2.6
* GCC
* GNU make
* autoconf
* libelf

On Debian/Ubuntu systems, you can install most of these with:

    sudo apt-get update
    sudo apt-get install -y build-essential clang llvm bison flex \
         make autoconf libelf-dev


### Building

Building `ptcpdump` involves two main steps: generating the eBPF bytecode
(optional if you don't modify the eBPF code) and compiling the Go application.

#### 1. Build the eBPF Programs (Optional)

This step compiles the C eBPF code located in the `bpf/` directory
into bytecode for multiple target architectures (amd64, arm64, arm).
The generated bytecode is then embedded into the final `ptcpdump` Go executable.
You only need to run this step if you modify the eBPF source code or
if the pre-generated files are missing.

*   **Native Build**: Requires Clang/LLVM installed locally.

        make build-bpf

*   **Using Docker**: Use this if you don't have the required C build toolchain
    locally. It utilizes a pre-configured development image.

        make build-bpf-via-docker


#### 2. Build the `ptcpdump` Executable

After ensuring the eBPF bytecode are generated
(or using the pre-generated ones), you can compile the main Go application.

* **Static Build (Recommended)**:
  This builds `ptcpdump` with all its dependencies,
  including a bundled version of `libpcap` (from the `lib/libpcap` submodule),
  statically linked into the final executable.
  This results in a portable binary with no external library dependencies.
  This is the default and recommended method.

    * **Native Build**: This command first ensures the bundled `libpcap` is compiled
      and then builds the static `ptcpdump` executable.

          # This first builds libpcap, then builds ptcpdump statically
          make build

    * **Using Docker**: Builds the static executable inside the development container.

          make build-via-docker

* **Dynamic Build**:
  This builds `ptcpdump` dynamically linked against the system's
  installed `libpcap` library. You must have the `libpcap` development package
  installed on your system beforehand (e.g., `libpcap-dev` on Debian/Ubuntu,
  `libpcap-devel` on Fedora/CentOS). 
  This results in a smaller executable but requires `libpcap` to be present
  on the target system.

       # Ensure libpcap development package is installed first
       # e.g., sudo apt-get install libpcap-dev
       make build-dynamic-link

The final executable will be created in the project's root directory
(e.g., `./ptcpdump`).

<p align="right"><a href="#top">🔝</a></p>
