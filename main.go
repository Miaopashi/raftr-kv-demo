package main

import (
	"flag"
	"log"
	"net/http"
	"raft-kv-demo/core"
	"raft-kv-demo/global"
	"raft-kv-demo/service"
)

func main() {
	flag.StringVar(&global.RKD_RaftNodeID, "id", "node1", "id of node")
	flag.StringVar(&global.RKD_RaftBind, "raddr", "localhost:8080", "address of raft node")
	flag.StringVar(&global.RKD_HttpBind, "haddr", "localhost:8079", "address of http service")
	flag.BoolVar(&global.RKD_RunAsFirstNode, "bootstrap", false, "is first node to run")
	flag.StringVar(&global.RKD_LeaderHttpBind, "join", "", "leader's http address for join")
	flag.Parse()
	if len(global.RKD_LeaderHttpBind) > 0 {
		global.RKD_NeedToJoin = true
	}

	fsm, err := core.NewFSM()
	if err != nil {
		log.Fatalln("[raft booting] " + err.Error())
		return
	}

	err = fsm.Open()
	if err != nil {
		log.Fatalln(err.Error())
		return
	}

	log.Println("[raft booting] Raft booting has been completed")
	log.Println("[http service] Now let's run http service")
	http.Handle("/", service.NewService(fsm))
	http.ListenAndServe(global.RKD_HttpBind, nil)
}
