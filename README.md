# Foojank

Foojank is a prototype command-and-control (C2) framework that uses NATS for C2 communications.

NATS is a widely used message broker in IoT and cloud systems to facilitate communication between geographically
distributed services. NATS allows passing messages between connected services and offers a persistence layer known as
JetStream, enabling it to store messages on the server even when the receiver is offline. Additionally, NATS provides an
object store that can be utilized for storing files.

Foojank leverages the NATS features to offer:

* Asynchronous or real-time communication with Agents over TCP or WebSockets.
* Server-based storage for file sharing and data exfiltration.
* JWT-based authentication and authorization.
* Full observability.
* Extensibility.

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
$ curl -fsSL https://github.com/fooHQ/foojank/releases/latest/download/server.sh | sh
```

### Client

```
$ curl -fsSL https://github.com/fooHQ/foojank/releases/latest/download/client.sh | sh
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
