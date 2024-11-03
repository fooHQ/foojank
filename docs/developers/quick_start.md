# Quick start

If you are considering contributing code or want to help writing the documentation, this guide is for you.
For the purposes of this guide we will assume you are already familiar with GNU/Linux. Developing on Windows is certainly
possible but out of scope of this guide. All code examples are tailored for Debian.

## Install the dependencies

```shell
$ sudo apt install git make gcc golang upx
```

## Clone the repository

```shell
$ git clone https://github.com/foohq/foojank
# or via ssh:
$ git clone git@github.com:foohq/foojank.git
```

## Build the executables

Building executables is done via Makefile. Each executable has its own Makefile target that can be invoked by running `make` command. Built executables are stored in `build/`
directory in the root of the repository.

### Vessel

Vessel is foojank's agent. Vessel executable can be built in different ways depending on the intended use. 

- **Development build** - builds the executable keeping all symbols and debug information in the resulting file. This option is ideal for testing/development purposes but expect the file to be large.
- **Production build** - builds the executable without symbol table and debug information. This option is ideal for production use as the resulting file is much smaller but without debugging information it may be more difficult to diagnose runtime errors.
- **Small build** - this is a production build but the resulting binary file is compressed with `upx` packer using lzma. This option is ideal in places where smaller is better, however AntiVirus software may find the use of `upx` suspicious.

```bash
$ make build/vessel/dev

$ make build/vessel/prod

$ make build/vessel/small
```

For convenience purposes Makefile contains `run/vessel` target which builds and runs the agent with a single command. Useful if you are making changes to the code and want to test them quickly.

```bash
$ make run/vessel
```

### Client

Foojank client is a command line tool that is used to command and control connected agents.

```bash
$ make build/client
```

### Cross compiling executables

You can cross compile foojank binaries by changing `GOOS` and `GOARCH` environment variables. See Go's official documentation for all [supported values](https://go.dev/doc/install/source#environment).

!> `upx` is not compatible with all compilation targets.

```bash
$ make build/vessel/small GOOS=windows
$ make build/vessel/prod GOOS=darwin
```

## Generate code

Some code in the repository is generated. A Makefile target `generate` is used to regenerate all code in the repository.

### Cap'n Proto files

Cap'n Proto is a lightweight binary data serialization format. Foojank uses it to serialize API request/response data. Capn'n Proto uses their own schema language
to describe the messages and their format. These schema files are feed into the `go-capnp` generator, which generates the Go code. All protocol related files and code are stored in `proto/` directory with schema files in `proto/capnp/`.

Generated code is committed in the repository, therefore you don't usually need to install any additional dependencies, that is unless you are changing the proto files and want to generate a new code.
If that's the case, make sure you have all the required dependencies installed first.

```bash
$ sudo apt install capnproto
$ git clone https://github.com/capnproto/go-capnp
$ go install capnproto.org/go/capnp/v3/capnpc-go@latest
```

After changing the schema files generate the Go code.

```bash
$ make generate
```
