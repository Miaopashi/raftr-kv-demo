package service

import (
	"encoding/json"
	"errors"
	"github.com/hashicorp/raft"
	"io"
	"net/http"
	"raft-kv-demo/core"
	"strings"
)

type Service struct {
	store *core.Store
	fsm   *core.StoreFSM
}

func NewService(f *core.StoreFSM) *Service {
	return &Service{fsm: f, store: &f.Store}
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/key") {
		s.handleKeyRequest(w, r)
	} else if r.URL.Path == "/join" {
		s.handleJoin(w, r)
	} else if r.URL.Path == "/remove" {
		s.handleRemove(w, r)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (s *Service) handleKeyRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var setReq SetKVRequest
		err := json.NewDecoder(r.Body).Decode(&setReq)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err = s.store.Set(setReq.Key, setReq.Value)
		if err == raft.ErrNotLeader {
			url := s.store.FormRedirect(r, s.store.FetchLeaderHttpBind())
			http.Redirect(w, r, url, http.StatusTemporaryRedirect)
		} else if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else if r.Method == "GET" {
		k := r.URL.Query().Get("key")
		lv, err := level(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		v, err := s.store.Get(k, lv)
		if err == raft.ErrNotLeader {
			url := s.store.FormRedirect(r, s.store.FetchLeaderHttpBind())
			http.Redirect(w, r, url, http.StatusTemporaryRedirect)
		} else if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			io.WriteString(w, v)
		}
	} else if r.Method == "DELETE" {
		k := r.URL.Query().Get("key")
		err := s.store.Delete(k)
		if err == raft.ErrNotLeader {
			url := s.store.FormRedirect(r, s.store.FetchLeaderHttpBind())
			http.Redirect(w, r, url, http.StatusTemporaryRedirect)
		} else if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func level(r *http.Request) (lv core.ReadLevel, err error) {
	lv = core.ReadLevel(r.URL.Query().Get("level"))
	if lv != core.Stale && lv != core.Default && lv != core.Consistent {
		err = errors.New("[http service] level not exist")
	}
	return
}

func (s *Service) handleJoin(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	addr := r.URL.Query().Get("addr")
	err := s.store.Join(id, addr)
	if err == raft.ErrNotLeader {
		url := s.store.FormRedirect(r, s.store.FetchLeaderHttpBind())
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (s *Service) handleRemove(w http.ResponseWriter, r *http.Request) {
	id := raft.ServerID(r.URL.Query().Get("id"))
	err := s.store.Remove(id)
	if err == raft.ErrNotLeader {
		url := s.store.FormRedirect(r, s.store.FetchLeaderHttpBind())
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
