---
title: Documentation
---

Welcome to Foojank documentation!

## What is Foojank?

Foojank is a command and control (C2) framework that aims to provide operators with a reliable tool for red/purple team
engagements. Foojank capabilities are extendable with custom modules that provide extraneous functionality
similar to what libraries do, and scripts which run in the builtin scripting engine inside Foojank agent and utilize the
loaded modules to implement high-level commands for operators.

Foojank targets all major operating systems - GNU/Linux, Microsoft Windows, and macOS.

## Features

- **Malleable Scripting Engine**: Foojank's agent, named **vessel**, contains a builtin scripting engine which enables operators to execute
custom scripts. In fact, the agent is an empty vessel (hence the name) which has no builtin commands at all. All the functionality
is implemented as scripts executing in the context of a running agent with standard input and output connected over the network.

- **Powerful Backend**: Foojank's C2 server is built on top of NATS server. NATS is a powerful message broker which provides, among other things, 
communication based on Request-Reply pattern between connected nodes, stream-like view of published data and persistent consumers.
Foojank utilizes NATS server to facilitate communication between agents and
clients over a WebSocket connection, as well as NATS's ObjectStore to implement a shared storage.

- **Virtual Filesystems**: Foojank has a concept of virtual filesystems which provide filesystem-like API to facilitate communication over various internet
protocols or system resources, such as SSH, SMB, NATS, Windows Registry, and more. Custom virtual filesystems can be implemented as well.

- **Simplicity and Local-First Principle**: Foojank tasks are performed with a single command-line tool. Tool's ergonomics were highly inspired by Docker but with intention to keep the
number of commands low. Most workflows were designed with the Local-First principle in mind to make debugging of custom changes feel natural.

- **Implemented in Go**: Go is a cross-platform, compiled, statically typed programming language developed by Google. Some of
  the strengths of Go is its relative simplicity, fast compilation, out-of-box cross-compilation, and statically linked binaries.

- **Self-contained**: All dependencies required by the framework are versioned with [devbox](), which uses Nix to manage the dependencies.
Thanks to Nix, multiple versions of the same software can coexist without a conflict or need to change your
operating system.
