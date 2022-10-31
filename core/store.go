package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/raft"
	"log"
	"net/http"
	"raft-kv-demo/global"
	"strings"
)

type ReadLevel string

const (
	Stale      ReadLevel = "stale"
	Default    ReadLevel = "default"
	Consistent ReadLevel = "consistent"
)

type Store struct {
	r *raft.Raft
	m map[string]string
}

func (s *Store) Set(k string, v string) error {
	cmd := &CMD{
		Type:  "SET",
		Key:   k,
		Value: v,
	}
	b, err := json.Marshal(cmd)
	if err != nil {
		return err
	}
	err = s.r.Apply(b, global.RKD_Timeout).Error()
	return err
}

func (s *Store) FetchLeaderHttpBind() string {
	address, _ := s.r.LeaderWithID()
	if len(address) == 0 {
		// can't find leader
		// assume self as leader
		return global.RKD_HttpBind
	}
	ip, _, _ := strings.Cut(string(address), ":")
	return ip + ":" + global.RKD_HttpDefaultPort
}

func (s *Store) Get(k string, lv ReadLevel) (v string, err error) {
	if lv == Consistent {
		err = s.r.VerifyLeader().Error()
	} else if lv == Default && s.r.State() != raft.Leader {
		err = raft.ErrNotLeader
	}
	if err != nil {
		return
	}
	v, ok := s.m[k]
	if !ok {
		err = errors.New(fmt.Sprintf("[raft kv] Cannot find value of key: %s", k))
		log.Println(err.Error())
	}
	return
}

func (s *Store) Delete(k string) error {
	cmd := &CMD{
		Type: "DELETE",
		Key:  k,
	}
	b, err := json.Marshal(cmd)
	if err != nil {
		return err
	}
	err = s.r.Apply(b, global.RKD_Timeout).Error()
	return err
}

func (s *Store) FormRedirect(r *http.Request, addr string) string {
	path, params := r.URL.Path, r.URL.Query().Encode()
	if len(params) > 0 {
		path += "?" + params
	}
	return fmt.Sprintf("http://%s%s", addr, path)
}

func (s *Store) Join(id string, addr string) error {
	return s.r.AddVoter(raft.ServerID(id), raft.ServerAddress(addr), 0, 0).Error()
}

func (s *Store) Remove(id raft.ServerID) error {
	return s.r.RemoveServer(id, 0, 0).Error()
}

func (s *Store) Open() error {
	if global.RKD_RunAsFirstNode {
		log.Println("[raft booting] Bootstrap running")
		client := raft.Configuration{
			Servers: []raft.Server{{
				ID:      raft.ServerID(global.RKD_RaftNodeID),
				Address: raft.ServerAddress(global.RKD_RaftBind),
			}},
		}
		return s.r.BootstrapCluster(client).Error()
	}
	if global.RKD_NeedToJoin {
		log.Println("[raft booting] Node booting first, need to be joined")
		return s.SendJoinToLeader(global.RKD_RaftNodeID, global.RKD_RaftBind)
	}
	log.Println("[raft booting] Welcome back~")
	return nil
}

func (s *Store) SendJoinToLeader(id string, addr string) error {
	// if a node need to be joined
	// it must show leader's http addr
	url := fmt.Sprintf("http://%s/join?id=%s&addr=%s", global.RKD_LeaderHttpBind, id, addr)
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusTemporaryRedirect {
			log.Println("[http service] Request need redirect")
		}
		return err
	}
	return nil
}
