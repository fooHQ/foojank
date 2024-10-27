using Go = import "/go.capnp";
@0xdcccaa5d36aa8b70;
$Go.package("capnp");
$Go.import("github.com/foojank/foojank/proto/capnp");

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
    repository @0 :Text;
    filePath @1 :Text;
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
        dummyRequest @4 :DummyRequest;
    }
    response :union {
        createWorker @5 :CreateWorkerResponse;
        destroyWorker @6 :DestroyWorkerResponse;
        getWorker @7 :GetWorkerResponse;
        execute @8 :ExecuteResponse;
    }
}
