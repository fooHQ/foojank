# Foojank

Foojank is an open-source command-and-control (C2) framework built on NATS, designed for ethical hackers conducting red team and purple team engagements. It provides a scalable, observable, and extensible C2 platform for adversary simulation and security testing.

**Features**

* Customizable C2 transport over TCP and WebSockets.
* Built-in file storage for payload distribution and data exfiltration.
* Extensible support for custom and third-party agents.
* Observability, including visibility into agent activity and C2 messaging.

Foojank is currently compatible only with our prototype agent, [Vessel](https://github.com/foohq/vessel). However, we
plan to implement support for integrating custom agents into the framework in the future.

#### Platform Compatibility

|                | Linux | macOS | Windows |
|----------------|:-----:|:-----:|:-------:|
| Vessel (Agent) |   ✅   |   ✅   |    ✅    |
| Client         |   ✅   |   ✅   |    ❌    |
| Server         |   ✅   |   ✅   |    ✅    |

**Note:** Client is not supported on Windows due to being incompatible with Devbox.

## Installation

### Server

```
$ curl -fsSL https://github.com/fooHQ/foojank/releases/latest/download/server.sh | bash
```

### Client

```
$ curl -fsSL https://github.com/fooHQ/foojank/releases/latest/download/client.sh | bash
```

### Installation from Source

```
$ git clone https://github.com/foohq/foojank
$ cd foojank/
$ devbox run build-foojank-prod
# install ./build/foojank /usr/local/bin/foojank
```

## Usage

Foojank is controlled using a command-line client.

```
NAME:
   foojank - Command and control framework

USAGE:
   foojank [global options] [command [command options]]

VERSION:
   0.4.2

COMMANDS:
   account  Manage accounts
   agent    Manage agents
   config   Manage configuration
   job      Manage jobs
   profile  Manage profiles
   storage  Manage storage

GLOBAL OPTIONS:
   --no-color     disable color output
   --help, -h     show help
   --version, -v  print the version
```

## License

This software is distributed under the Apache License Version 2.0 found in the [LICENSE](./LICENSE) file.
