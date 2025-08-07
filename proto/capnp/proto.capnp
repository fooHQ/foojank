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

struct UpdateStdioLine {
    text @0 :Text;
}

struct DummyRequest {}

struct Message {
    content :union {
        createJobRequest @0 :CreateJobRequest;
        cancelJobRequest @1 :CancelJobRequest;
        createJobResponse @2 :CreateJobResponse;
        cancelJobResponse @3 :CancelJobResponse;
        updateJob @4 :UpdateJob;
        updateStdioLine @5 :UpdateStdioLine;
    }
    action :union {
        createWorker @6 :CreateWorkerRequest;
        destroyWorker @7 :DestroyWorkerRequest;
        getWorker @8 :GetWorkerRequest;
        execute @9 :ExecuteRequest;
        createJob @10 :CreateJobRequest;
        cancelJob @11 :CancelJobRequest;
        dummyRequest @12 :DummyRequest;
    }
    response :union {
        createWorker @13 :CreateWorkerResponse;
        destroyWorker @14 :DestroyWorkerResponse;
        getWorker @15 :GetWorkerResponse;
        execute @16 :ExecuteResponse;
        createJob @17 :CreateJobResponse;
        cancelJob @18 :CancelJobResponse;
        updateJob @19 :UpdateJob;
    }
}
