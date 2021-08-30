# containerd-healthcheck

A daemon that connects to containerd and performs health checks to know when to restart a container. It allows you to register async health checks and the task of the container itself, and provides a restart mechanism based on the check configuration.

## Rationale

[containerd](https://containerd.io/) is an abstraction of kernel features that provide a relatively high-level container interface with limited feature set and typically perform low-level tasks for running a container. Due to its nature of being small and focused, containerd doesn't support health checks like higher-level components as Docker and Kubernetes do. Therefore, this project solves this problem by providing a simple way to let the system know, given a configuration, if a task running with containerd is whether working or not working. If a task is not working, then it "restart" the task - stop and start a new one.

## Installation

On Linux, you can install using go get:

```
go get github.com/vouch-opensource/containerd-healthcheck
```

Also, you can run it directly by using its Docker image:

```
docker run -v /run/containerd/containerd.sock:/run/containerd/containerd.sock vouchio/containerd-healthcheck
```

Or you can manually download the binary from the [github releases page](https://github.com/vouch-opensource/containerd-healthcheck/releases).

## Usage

### Arguments

```
Usage of ./containerd-healthcheck:
  -a, --addr string     HTTP address for prometheus endpoint (default ":9434")
  -c, --config string   Path to configuration file (default "config.yml")
  -e, --env string      Application environment (default "development")
  -v, --version         Print app version
````

### Configuration

The `config.yml` is the required configuration file for the `containerd-healthcheck`. It will read the `config.yml` file in the current working directory or specified with the `--config` option to be used by the daemon.

Example

```yaml
containerd:
  socket: /run/containerd/containerd.sock
  namespace: default
checks:
  - container_task: example-api
    http:
      url: 127.0.0.1:8080/health
      method: GET
      expected_body: "OK"
      expected_status: 200
    execution_period: 2
    initial_delay: 2
    threshold: 3
    timeout: 5
```

#### `containerd`

This section specifies the required information to establish a connection with containerd

* `socket`: containerd socket path to be used to establish a connection between the client and containerd over GRPC
* `namespace`: namespace of containerd which the container resources - tasks, images, snapshots - are located

#### `checks`

The section is a list of health checks to be performed by the `containerd-healthcheck`

* `container_task:` task name of the container, mostly the same as container
* `execution_period`: check interval to be executed (in seconds)
* `initial_delay`: time to delay first check execution (in seconds)
* `restart_delay`: time to sleep after restarting a task (in seconds)
* `threshold`: the number of consecutive of check failures required before considering a target unhealthy and then marked to be restarted
* `timeout`: Timeout used for the request (in seconds)
* `http.url`: URL to be called by the check
* `http.method`: HTTP method
* `http.expected_body`: Operates as a basic 'body should contains < string >'
* `http.expected_status`: Expected response status code

## Metrics

By default, `containerd-healthcheck` daemon serves a prometheus http endpoint with built-in metrics provided by [go-sundheit](https://github.com/AppsFlyer/go-sundheit) library; also it includes the total number of restarts per task running on containerd. Once the daemon is running, it can be accessed by the address `127.0.0.1:9434/metrics`. A custom address can be defined from `--addr` argument as well.

## License ##

```
Copyright (c) Vouch, Inc. All rights reserved. The use and
distribution terms for this software are covered by the Eclipse
Public License 1.0 (http://opensource.org/licenses/eclipse-1.0.php)
which can be found in the file epl-v10.html at the root of this
distribution. By using this software in any fashion, you are
agreeing to be bound by the terms of this license. You must
not remove this notice, or any other, from this software.
```
