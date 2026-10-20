# Foojank

[![Version](https://img.shields.io/github/v/release/fooHQ/foojank)](https://github.com/fooHQ/foojank/releases/latest)

> [!NOTE]
> This README describes an upcoming release of Foojank that is still in development. See Releases for older versions.

## Overview

Foojank is a C2 framework designed for ethical hackers conducting red team and purple team engagements.
Built on top of [NATS](https://nats.io), Foojank is extensible, observable, and scalable.

### Features

* Command-line client interface that can be integrated in scripts and AI agents.
* Built-in data storage for payload distribution and data exfiltration.
* Multi-user support with configurable roles and privileges.
* Extensibility – support for custom protocols and third-party agents via Gateways.
* Observability – visibility into operator and agent activity and C2 messaging.
* Scalability – support for large-scale deployments with distributed architecture.

### Gateways

Gateways allow clients to communicate with agents using different protocols or third-party systems.

| Name                                                                    | Description                                                                                          | Status |
|-------------------------------------------------------------------------|------------------------------------------------------------------------------------------------------|--------|
| [Azure Queue Storage](https://github.com/foohq/foojank-azqueue-gateway) | Enables client-to-agent messaging over Azure Queue Storage for reliable, asynchronous communication. | WIP    |

### Agents

| Name                                      | Language | Platform               | Status |
|-------------------------------------------|----------|------------------------|--------|
| [Vessel](https://github.com/foohq/vessel) | Go       | Windows, Linux, Darwin | WIP    |

### Architecture

```mermaid
flowchart RL
    subgraph CL["Clients"]
        direction RL
        subgraph PC1["John's computer"]
            John["John"]
        end
        subgraph PC2["Alice's computer"]
            Alice["Alice"]
        end
        subgraph PBob["Bob's computer"]
            Bob["Bob"]
        end
    end

    Daemon["Daemon"]

    subgraph GW["Gateways"]
        direction TB
        GW1["Gateway A"]
        GW2["Gateway B"]
        GW3["Gateway C"]
    end

    subgraph AG["Agents"]
        direction TB
        A1["Agent A1"]
        A2["Agent A2"]
        B1["Agent B1"]
        C1["Agent C1"]
    end

    %% NATS-based communication
    John <-->|NATS| Daemon
    Alice <-->|NATS| Daemon
    Bob <-->|NATS| Daemon
    Daemon <-->|NATS| GW1
    Daemon <-->|NATS| GW2
    Daemon <-->|NATS| GW3

    %% Gateway-specific protocols
    GW1 <-->|Protocol A| A1
    GW1 <-->|Protocol A| A2
    GW2 <-->|Protocol B| B1
    GW3 <-->|Protocol C| C1

    %% Styling: solid blue = NATS, dashed orange = gateway protocol
    linkStyle 0,1,2,3,4,5 stroke:#2b7bb9,stroke-width:2px
    linkStyle 6,7,8,9 stroke:#e08a1e,stroke-width:2px,stroke-dasharray:5 3
```

## Try it

TODO

## License

This software is distributed under the Apache License Version 2.0 found in the [LICENSE](./LICENSE) file.
