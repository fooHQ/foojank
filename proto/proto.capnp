using Go = import "/go.capnp";
@0xdcccaa5d36aa8b70;
$Go.package("proto");
$Go.import("github.com/foojank/vessel/proto");

struct CreateWorkerRequest {
}

struct CreateWorkerResponse {
    id @0 :Text;
}

struct DestroyWorkerRequest {
    id @0 :Text;
}

struct Message {
    action :union {
        createWorker @0 :CreateWorkerRequest;
        destroyWorker @1 :DestroyWorkerRequest;
    }
}
