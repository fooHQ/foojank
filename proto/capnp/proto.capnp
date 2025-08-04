using Go = import "/go.capnp";
@0xdcccaa5d36aa8b70;
$Go.package("capnp");
$Go.import("github.com/foohq/foojank/proto/capnp");

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
    args @0 :List(Text);
    filePath @1 :Text;
}

struct ExecuteResponse {
    code @0 :Int64;
}

struct CreateJobRequest {
    command @0 :Text;
    args @1 :List(Text);
    env @2 :List(Text);
}

struct CreateJobResponse {
   jobID @0 :Text;
   stdinSubject @1 :Text;
   stdoutSubject @2 :Text;
   error @3 :Text;
}

struct CancelJobRequest {
    jobID @0 :Text;
}

struct CancelJobResponse {
    error @0 :Text;
}

struct UpdateJob {
    jobID @0 :Text;
    exitStatus @1 :Int64;
}

struct DummyRequest {}

struct Message {
    action :union {
        createWorker @0 :CreateWorkerRequest;
        destroyWorker @1 :DestroyWorkerRequest;
        getWorker @2 :GetWorkerRequest;
        execute @3 :ExecuteRequest;
        createJob @4 :CreateJobRequest;
        cancelJob @5 :CancelJobRequest;
        dummyRequest @6 :DummyRequest;
    }
    response :union {
        createWorker @7 :CreateWorkerResponse;
        destroyWorker @8 :DestroyWorkerResponse;
        getWorker @9 :GetWorkerResponse;
        execute @10 :ExecuteResponse;
        createJob @11 :CreateJobResponse;
        cancelJob @12 :CancelJobResponse;
        updateJob @13 :UpdateJob;
    }
}
