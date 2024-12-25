# foojank

foojank is a cross-platform post-exploitation and command & control (C2) framework written in Go. The framework implements three main components: agent, server, and client.

## Agent

Agent provides persistence in a target system and establishes a communication over a TCP
or WebSocket TLS-encrypted connection. The capabilities of the agent are implemented
as scripts, written in [Risor](https://risor.io) scripting language. The syntax is similar to Go, but it has the convenience of other dynamic languages.
The scripts are sent by a client over the network and executed in a built-in scripting engine on the agent.
The engine can be further extended with custom modules written in Go.

## Server

Unlike other C2 frameworks, foojank does not have its own server implementation. Instead, it makes use of the [NATS](https://nats.io) server.
What is a NATS server? A NATS server is many things, but for the framework it is a message broker and an object store.

As a message broker, NATS implements a request/reply pattern around application-defined "subjects", which the connected
clients use to exchange messages. In the case of the framework, an agent subscribes itself to a predefined "subject" and waits
for a request to arrive. When it does, the agent takes care of it and sends back a reply.

As an object store, NATS can be used as a storage for large files. In the framework, the primary function is to facilitate the
exchange of scripts and other files between a client and an agent.

The benefit of using NATS is not merely practical for its features. Utilizing an open-source, multipurpose technology makes
it harder for defenders to discover your infrastructure by scanning the internet, as the server should be indistinguishable
from a regular NATS server.

## Client

The client is a command-line utility through which the entire framework is controlled. The client is capable of configuring
the framework, building and controlling agents, and so on.
