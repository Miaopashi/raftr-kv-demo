package core

import (
	"fmt"
	"github.com/hashicorp/go-msgpack/codec"
	"github.com/hashicorp/raft"
	"github.com/hashicorp/raft-boltdb"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"raft-kv-demo/global"
	"sync"
)

type StoreFSM struct {
	Store
	sync.Mutex
	logs           [][]byte
	configurations []raft.Configuration
}

func NewFSM() (fsm *StoreFSM, err error) {
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(global.RKD_RaftNodeID)
	fsm = &StoreFSM{}
	fsm.m = make(map[string]string)
	logStore, err := raftboltdb.NewBoltStore(filepath.Join(global.RKD_RaftDir, "raft-log.db"))
	if err != nil {
		return
	}
	stableStore, err := raftboltdb.NewBoltStore(filepath.Join(global.RKD_RaftDir, "raft-stable.db"))
	if err != nil {
		return
	}
	snapshots, err := raft.NewFileSnapshotStore(global.RKD_RaftDir, global.RKD_RetainSnapshotCount, os.Stderr)
	if err != nil {
		return
	}
	addr, err := net.ResolveTCPAddr("tcp", global.RKD_RaftBind)
	if err != nil {
		return
	}
	transport, err := raft.NewTCPTransport(global.RKD_RaftBind, addr, global.RKD_MaxPool, global.RKD_Timeout, os.Stderr)
	if err != nil {
		return
	}
	fsm.r, err = raft.NewRaft(config, fsm, logStore, stableStore, snapshots, transport)
	if err != nil {
		return
	}
	return
}

func (f *StoreFSM) Apply(log *raft.Log) interface{} {
	f.Lock()
	defer f.Unlock()
	f.applyCMD(log.Data)
	f.logs = append(f.logs, log.Data)
	return len(f.logs)
}

func (f *StoreFSM) applyCMD(cmd []byte) {
	c := NewCMD(cmd)
	if c.Type == CMD_Set {
		f.applySet(c.Key, c.Value)
	} else if c.Type == CMD_Delete {
		f.applyDelete(c.Key)
	}
}

func (f *StoreFSM) applySet(k string, v string) {
	f.m[k] = v
	log.Println(fmt.Sprintf("[raft kv] set %s = %s success", k, v))
}

func (f *StoreFSM) applyDelete(k string) {
	delete(f.m, k)
	log.Println(fmt.Sprintf("[raft kv] successfully delete key: %s.", k))
}

func (f *StoreFSM) Snapshot() (raft.FSMSnapshot, error) {
	f.Lock()
	defer f.Unlock()
	return &MySnapshot{f.logs, len(f.logs)}, nil
}

func (f *StoreFSM) Restore(inp io.ReadCloser) error {
	f.Lock()
	defer f.Unlock()
	defer inp.Close()
	hd := codec.MsgpackHandle{}
	dec := codec.NewDecoder(inp, &hd)

	f.logs = nil
	err := dec.Decode(&f.logs)
	if err != nil {
		return err
	}

	for _, cmd := range f.logs {
		f.applyCMD(cmd)
	}
	return nil
}

type MySnapshot struct {
	logs     [][]byte
	maxIndex int
}

func (m *MySnapshot) Persist(sink raft.SnapshotSink) error {
	hd := codec.MsgpackHandle{}
	enc := codec.NewEncoder(sink, &hd)
	if err := enc.Encode(m.logs[:m.maxIndex]); err != nil {
		sink.Cancel()
		return err
	}
	sink.Close()
	return nil
}

func (m *MySnapshot) Release() {
}
