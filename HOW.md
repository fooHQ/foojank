## Cap'n'Proto setup

Howto: https://github.com/capnproto/go-capnp/blob/main/docs/Installation.md

```
$ sudo apt install capnproto
$ go install capnproto.org/go/capnp/v3/capnpc-go@latest
$ `git clone https://github.com/capnproto/go-capnp`
$ `capnp compile -I /tmp/go-capnp/std/ -ogo proto/jobs.capnp`
```
