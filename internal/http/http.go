// Package http is the MVP HTTP front-end for Radica.
//
// One listener, one protocol, one port. Lets you drive Get/Set/Delete
// with curl while the rest of the system is built out.
//
// Routes:
//
//	GET    /healthz       → "ok"
//	GET    /kv/{key}      → 200 + body, or 404
//	PUT    /kv/{key}      → 204; request body is the value
//	DELETE /kv/{key}      → 204, or 404
package http

import (
	"context"
	"errors"
	"io"
	nethttp "net/http"
	"strings"

	"github.com/radicaio/radica/internal/wire"
)

// Submitter is the subset of the transport this server needs.
type Submitter interface {
	Submit(op wire.Op) <-chan wire.Reply
}

// Server is the HTTP handler.
type Server struct {
	mux *nethttp.ServeMux
	sub Submitter
}

// NewServer wires routes and returns a ready-to-serve handler.
func NewServer(sub Submitter) *Server {
	s := &Server{mux: nethttp.NewServeMux(), sub: sub}
	s.mux.HandleFunc("/healthz", s.healthz)
	s.mux.HandleFunc("/kv/", s.kv)
	return s
}

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w nethttp.ResponseWriter, r *nethttp.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) healthz(w nethttp.ResponseWriter, _ *nethttp.Request) {
	_, _ = io.WriteString(w, "ok")
}

func (s *Server) kv(w nethttp.ResponseWriter, r *nethttp.Request) {
	key := strings.TrimPrefix(r.URL.Path, "/kv/")
	if key == "" {
		nethttp.Error(w, "missing key", nethttp.StatusBadRequest)
		return
	}

	switch r.Method {
	case nethttp.MethodGet:
		s.doGet(w, r.Context(), []byte(key))
	case nethttp.MethodPut:
		body, err := io.ReadAll(r.Body)
		if err != nil {
			nethttp.Error(w, "read body: "+err.Error(), nethttp.StatusBadRequest)
			return
		}
		s.doSet(w, r.Context(), []byte(key), body)
	case nethttp.MethodDelete:
		s.doDelete(w, r.Context(), []byte(key))
	default:
		w.Header().Set("Allow", "GET, PUT, DELETE")
		nethttp.Error(w, "method not allowed", nethttp.StatusMethodNotAllowed)
	}
}

func (s *Server) doGet(w nethttp.ResponseWriter, ctx context.Context, key []byte) {
	reply, err := s.wait(ctx, wire.Op{Code: wire.OpGet, Key: key})
	if err != nil {
		nethttp.Error(w, err.Error(), nethttp.StatusGatewayTimeout)
		return
	}
	switch reply.Status {
	case wire.StatusOK:
		_, _ = w.Write(reply.Value)
	case wire.StatusMiss:
		nethttp.NotFound(w, nil)
	default:
		nethttp.Error(w, "unexpected status", nethttp.StatusInternalServerError)
	}
}

func (s *Server) doSet(w nethttp.ResponseWriter, ctx context.Context, key, value []byte) {
	reply, err := s.wait(ctx, wire.Op{Code: wire.OpSet, Key: key, Value: value})
	if err != nil {
		nethttp.Error(w, err.Error(), nethttp.StatusGatewayTimeout)
		return
	}
	if reply.Status != wire.StatusOK {
		nethttp.Error(w, "set failed", nethttp.StatusInternalServerError)
		return
	}
	w.WriteHeader(nethttp.StatusNoContent)
}

func (s *Server) doDelete(w nethttp.ResponseWriter, ctx context.Context, key []byte) {
	reply, err := s.wait(ctx, wire.Op{Code: wire.OpDelete, Key: key})
	if err != nil {
		nethttp.Error(w, err.Error(), nethttp.StatusGatewayTimeout)
		return
	}
	switch reply.Status {
	case wire.StatusOK:
		w.WriteHeader(nethttp.StatusNoContent)
	case wire.StatusMiss:
		nethttp.NotFound(w, nil)
	default:
		nethttp.Error(w, "unexpected status", nethttp.StatusInternalServerError)
	}
}

func (s *Server) wait(ctx context.Context, op wire.Op) (wire.Reply, error) {
	ch := s.sub.Submit(op)
	select {
	case reply := <-ch:
		return reply, nil
	case <-ctx.Done():
		return wire.Reply{}, errors.New("request cancelled: " + ctx.Err().Error())
	}
}
