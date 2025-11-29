using Go = import "/go.capnp";
@0xdcccaa5d36aa8b70;
$Go.package("capnp");
$Go.import("github.com/foohq/foojank/proto/capnp");

const startWorkerT :Text = "FJ.API.WORKER.START.%s.%s";
const stopWorkerT :Text = "FJ.API.WORKER.STOP.%s.%s";
const writeWorkerStdinT :Text = "FJ.API.WORKER.WRITE.STDIN.%s.%s";
const writeWorkerStdoutT :Text = "FJ.API.WORKER.WRITE.STDOUT.%s.%s";
const updateWorkerStatusT :Text = "FJ.API.WORKER.UPDATE.STATUS.%s.%s";
const updateClientInfoT :Text = "FJ.API.CLIENT.UPDATE.INFO.%s";
const replyMessageT :Text = "FJ.API.MESSAGE.REPLY.%s.%s";

struct StartWorkerRequest {
    command @0 :Text;
    args @1 :List(Text);
    env @2 :List(Text);
}

struct StartWorkerResponse {
    error @0 :Text;
}

struct StopWorkerRequest {}

struct StopWorkerResponse {
    error @0 :Text;
}

struct UpdateWorkerStatus {
    status @0 :Int64;
}

struct UpdateWorkerStdio {
    data @0 :Data;
}

struct UpdateClientInfo {
    username @0 :Text;
    hostname @1 :Text;
    system @2 :Text;
    address @3 :Text;
}

struct Message {
    content :union {
        startWorkerRequest @0 :StartWorkerRequest;
        startWorkerResponse @1 :StartWorkerResponse;
        stopWorkerRequest @2 :StopWorkerRequest;
        stopWorkerResponse @3 :StopWorkerResponse;
        updateWorkerStatus @4 :UpdateWorkerStatus;
        updateWorkerStdio @5 :UpdateWorkerStdio;
        updateClientInfo @6 :UpdateClientInfo;
    }
}
