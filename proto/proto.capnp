using Go = import "/go.capnp";
@0xdcccaa5d36aa8b70;
$Go.package("proto");
$Go.import("github.com/foojank/vessel/proto");

struct CreateWorkerRequest {
    // Add configurable subjects: stdin, rpc
}

struct CreateWorkerResponse {
    id @0 :UInt64;
}

struct GetWorkerRequest {
    id @0 :UInt64;
}

struct GetWorkerResponse {
    serviceId @0 :Text;
}

struct DestroyWorkerRequest {
    id @0 :UInt64;
}

struct DestroyWorkerResponse {}

struct Message {
    action :union {
        createWorker @0 :CreateWorkerRequest;
        destroyWorker @1 :DestroyWorkerRequest;
        getWorker @2 :GetWorkerRequest;
    }
    response :union {
        createWorker @3 :CreateWorkerResponse;
        destroyWorker @4 :DestroyWorkerResponse;
        getWorker @5 :GetWorkerResponse;
    }
}
