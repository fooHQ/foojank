# Foojank

[![Version](https://img.shields.io/github/v/release/fooHQ/foojank)](https://github.com/fooHQ/foojank/releases/latest)

Foojank is a C2 framework designed for ethical hackers conducting red team and purple team engagements.
Built on top of [NATS](https://nats.io), Foojank is extensible, observable, and scalable.

## Features

* Command-line client interface that can be integrated in scripts and AI agents.
* Built-in data storage for payload distribution and data exfiltration.
* Multi-user support with configurable roles and privileges.
* Extensibility – support for custom protocols and third-party agents via Adapters.
* Observability – visibility into operator and agent activity and C2 messaging.
* Scalability – support for large-scale deployments with distributed architecture.

## Architecture

```mermaid
flowchart LR
    subgraph CL["Clients"]
        direction RL
        Alice["Alice"]
        Bob["Bob"]
    end

    Daemon["Daemon"]

    subgraph GW["Gateways"]
        direction TB
        GW1["Gateway A"]
        AD1["Adapter A"]
        AD2["Adapter B"]
        GW2["Gateway B"]
        AD3["Adapter C"]
    end

    subgraph AG["Agents"]
        direction TB
        A1["Agent A1"]
        A2["Agent A2"]
        B1["Agent B1"]
        C1["Agent C1"]
    end

    %% NATS-based communication
    Alice <-->|NATS| Daemon
    Bob <-->|NATS| Daemon
    Daemon <-->|NATS| GW1
    Daemon <-->|NATS| GW2

    %% Gateway-specific protocols
    GW1 <-->|RPC| AD1
    GW1 <-->|RPC| AD2
    GW2 <-->|RPC| AD3
    
    AD1 <-->|Protocol A| A1
    AD1 <-->|Protocol A| A2
    AD2 <-->|Protocol B| B1
    AD3 <-->|Protocol C| C1

    %% Styling: solid blue = NATS, dashed orange = gateway protocol
    linkStyle 0,1,2,3 stroke:#2b7bb9,stroke-width:2px
    linkStyle 4,5,6 stroke:#2BB92B,stroke-width:2px
    linkStyle 7,8,9,10 stroke:#e08a1e,stroke-width:2px,stroke-dasharray:5 3
```

### Client

TODO

### Daemon

TODO

### Gateway

TODO

### Adapter

Adapter allows clients to communicate with agents using different protocols or third-party systems.

| Name                                                                    | Description                                                                                          | Status |
|-------------------------------------------------------------------------|------------------------------------------------------------------------------------------------------|--------|
| [Azure Queue Storage](https://github.com/foohq/foojank-azqueue-gateway) | Enables client-to-agent messaging over Azure Queue Storage for reliable, asynchronous communication. | WIP    |

### Agent

| Name                                      | Language | Platform               | Status |
|-------------------------------------------|----------|------------------------|--------|
| [Vessel](https://github.com/foohq/vessel) | Go       | Windows, Linux, Darwin | WIP    |

## Try it

TODO

## License

This software is distributed under the Apache License Version 2.0 found in the [LICENSE](./LICENSE) file.
