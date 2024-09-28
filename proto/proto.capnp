using Go = import "/go.capnp";
@0xdcccaa5d36aa8b70;
$Go.package("proto");
$Go.import("github.com/foojank/foojank/proto");

struct CreateWorkerRequest {
    # Add configurable subjects: stdin, rpc
}

struct CreateWorkerResponse {
    id @0 :UInt64;
}

struct GetWorkerRequest {
    id @0 :UInt64;
}

struct GetWorkerResponse {
    serviceName @0 :Text;
    serviceId @1 :Text;
}

struct DestroyWorkerRequest {
    id @0 :UInt64;
}

struct DestroyWorkerResponse {}

struct ExecuteRequest {
    data @0 :Data;
}

struct ExecuteResponse {
    code @0 :Int64;
}

struct DummyRequest {}

struct Message {
    action :union {
        createWorker @0 :CreateWorkerRequest;
        destroyWorker @1 :DestroyWorkerRequest;
        getWorker @2 :GetWorkerRequest;
        execute @3 :ExecuteRequest;
    }
    response :union {
        createWorker @4 :CreateWorkerResponse;
        destroyWorker @5 :DestroyWorkerResponse;
        getWorker @6 :GetWorkerResponse;
        execute @7 :ExecuteResponse;
    }
}
