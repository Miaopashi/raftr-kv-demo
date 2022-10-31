package global

import "time"

var (
	RKD_RaftDir             = "./"
	RKD_RetainSnapshotCount = 1
	RKD_RaftBind            = "localhost:8080"
	RKD_HttpBind            = "localhost:8079"
	RKD_RaftNodeID          = "node1"
	RKD_MaxPool             = 10
	RKD_Timeout             = time.Duration(3 * 1000 * 1000 * 1000) // wait for 3s
	RKD_RunAsFirstNode      = false
	RKD_NeedToJoin          = false
	RKD_HttpDefaultPort     = "8079"
	RKD_LeaderHttpBind      = ""
)
