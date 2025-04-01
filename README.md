# ioping_prober
Prometheus style "ioping" prober.

![Example Graph](example-graph.png)

## Overview

This prober uses `ioping` command to run probes against a local file, directory or device and records the responses in Prometheus histogram metrics.

Code initially forked from https://github.com/SuperQ/smokeping_prober/

```
usage: ioping_prober [<flags>] <target>...

Flags:
  -h, --help              Show context-sensitive help (also try --help-long and --help-man).
      --web.listen-address=":9374"  
                          Address on which to expose metrics and web interface.
      --web.telemetry-path="/metrics"  
                          Path under which to expose metrics.
      --buckets="5e-05,0.0001,0.0002,0.0004,0.0008,0.0016,0.0032,0.0064,0.0128,0.0256,0.0512,0.1024,0.2048,0.4096,0.8192,1.6384,3.2768,6.5536,13.1072,26.2144"  
                          A comma delimited list of buckets to use
  -i, --ping.interval=1s  Ping interval duration
      --write             Write to target. Uses ioping -W and is safe for directory target.
      --unsafewrite       Unsafely write to target. Uses ioping -WWW and is destructive to file|device- read ioping manpage.
      --log.level="info"  Only log messages with the given severity or above. Valid levels: [debug, info, warn, error, fatal]
      --log.format="logger:stderr"  
                          Set the log target and format. Example: "logger:syslog?appname=bob&local=7" or "logger:stdout?json=true"
      --version           Show application version.

Args:
  <target>  List of target directory/file/device to ioping

```

## Requirements

`ioping` command is required on the local machine. This was tested against version `ioping 1.2.16.gf549dff`

## Building and running

Requires Go >= 1.11

```console
go get github.com/dalees/ioping_prober
```

Running the exporter is best suited for a VM with dedicated disks attached. Do not use `--unsafewrite` without using volumes that are dedicated to this test. In the example we have attached two ceph volumes, one spinning rust and one nvme. The differences can be seen on the example graph.

```
./ioping_prober --ping.interval=1s --log.level=debug /dev/vdb /dev/vdc --unsafewrite
```

## Metrics Exported

With the exporter running, Prometheus can be configured to [scrape the exporter target](example-scrapeconfig.yml) and ingest the current metrics.


 Metric Name                            | Type       | Description
----------------------------------------|------------|-------------------------------------------
 ioping\_measurements\_total            | Counter    | Counter of iopings made.
 ioping\_measurement\_duration\_seconds | Histogram  | Filesystem response duration.

## Dashboard

A Grafana dashboard is available in [example-dashboard.json](example-dashboard.json) that presents the above metrics in a heatmap.
