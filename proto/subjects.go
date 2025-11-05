package proto

import "github.com/foohq/foojank/proto/capnp"

func StartWorkerSubject(agentID, workerID string) string {
	return replaceStringPlaceholders(capnp.StartWorkerT, agentID, workerID)
}

func StopWorkerSubject(agentID, workerID string) string {
	return replaceStringPlaceholders(capnp.StopWorkerT, agentID, workerID)
}

func WriteWorkerStdinSubject(agentID, workerID string) string {
	return replaceStringPlaceholders(capnp.WriteWorkerStdinT, agentID, workerID)
}

func WriteWorkerStdoutSubject(agentID, workerID string) string {
	return replaceStringPlaceholders(capnp.WriteWorkerStdoutT, agentID, workerID)
}

func UpdateWorkerStatusSubject(agentID, workerID string) string {
	return replaceStringPlaceholders(capnp.UpdateWorkerStatusT, agentID, workerID)
}

func UpdateClientInfoSubject(agentID string) string {
	return replaceStringPlaceholders(capnp.UpdateClientInfoT, agentID)
}

func ReplyMessageSubject(agentID, msgID string) string {
	return replaceStringPlaceholders(capnp.ReplyMessageT, agentID, msgID)
}

func replaceStringPlaceholders(s string, values ...string) string {
	result := s
	valIndex := 0

	for valIndex < len(values) {
		found := false
		for i := 0; i < len(result)-1; i++ {
			if result[i] == '%' && result[i+1] == 's' {
				result = result[:i] + values[valIndex] + result[i+2:]
				valIndex++
				found = true
				break
			}
		}
		if !found {
			break
		}
	}
	return result
}
