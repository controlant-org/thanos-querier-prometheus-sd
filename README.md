# thanos-querier-prometheus-sd
A simple service to watch Kubernetes Prometheus custom resource instances and write the gathered targets to a service
discovery file for the Thanos querier.

## Overview
This service currently has a very simple, rudimentary implementation. It lists all services in its Kubernetes cluster
that contain a specific label key-value pairing (`operated-prometheus=true`), which is set by the Prometheus operator.
From each service it builds a DNSSRV record and adds that to a target list to be written out as a YAML file for
ingestion by a Thanos Querier. This service will then wait for a specified interval, after which it will relist and
rewrite the file.

The intended use case here is to run this as a sidecar container that writes out the result file to a shared volume with
a Thanos Querier configured to read additional targets from the shared volume.

## Usage
At present there are only two CLI flags that can be set: the wait interval duration, and the output file location.

| CLI Flag      | Default                 | Description                                                              |
|---------------|-------------------------|--------------------------------------------------------------------------|
| `interval`    | `10000`                 | The number of milliseconds to wait before regenerating the targets file. |
| `output-file` | `/tmp/tqsd/result.yaml` | The filesystem location of the output targets file.                      |