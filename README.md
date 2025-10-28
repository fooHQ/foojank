# Foojank

Foojank is a prototype of a command and control framework written in Go.

Foojank uses NATS as a command-and-control server. NATS, a widely used message broker, is commonly used in IoT and cloud systems to facilitate communication between distributed services. NATS offers a persistence layer known as JetStream, enabling it to store messages on the server even when the receiver is offline. Additionally, NATS provides an object store that can be utilized for storing files.

Foojank leverages the NATS features to offer:

* Asynchronous or real-time communication with Agents over TCP or WebSockets.
* Server-based storage for file sharing and data exfiltration.
* JWT-based authentication and authorization.
* Full observability.
* Extensibility.

Foojank is currently compatible only with our prototype agent, [Vessel](https://github.com/foohq/vessel). However, we plan to implement support for integrating custom agents into the framework in the future.

## Installation

These steps are only suitable for a quick evaluation of the framework's capabilities or for developers. For an actual installation guide, please refer to [Foojank's manual]([https://foojank.com](https://foojank.com)).

### Requirements

* [Devbox]([https://www.jetify.com/devbox](https://www.jetify.com/devbox))

### Compatibility

* macOS
* Linux
* ~~Windows~~ is currently not supported by Devbox. This affects only the client's compatibility, not the agent's.

Build Foojank client:

```
$ git clone https://github.com/foohq/foojank
$ cd foojank/
$ devbox run build build-foojank-prod
```

Run Foojank client:

```
$ ./build/foojank
```

To set up a testing NATS server, follow the steps outlined in the server [README](./server/README.md).

## Usage

Foojank is controlled using a command-line client.

```
NAME:
   foojank - Command and control framework

USAGE:
   foojank [global options] [command [command options]]

VERSION:
   0.4.0

COMMANDS:
   account  Manage accounts
   agent    Manage agents
   job      Manage jobs
   storage  Manage storage
   config   Manage configuration

GLOBAL OPTIONS:
   --config-dir string  set path to a configuration directory
   --no-color           disable color output (default: false)
   --help, -h           show help
   --version, -v        print the version
```

## License

This software is distributed under the Apache License Version 2.0 found in the [LICENSE](./LICENSE) file.
