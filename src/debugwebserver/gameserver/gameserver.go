package gameserver

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	gs "github.com/sambdavidson/community-chess/src/proto/services/games/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

// Handler handles gameserver operation HTTP stuff.
type Handler struct {
	TLS *tls.Config
}

var (
	conn *grpc.ClientConn
	gsc  gs.GameServerClient
)

func (h *Handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	req.ParseMultipartForm(100000)
	rw.Header().Set("Content-Type", "application/json")

	switch req.URL.Path { // Handlers for following the /games/ prefix e.g. the /games/join URL
	case "connect":
		h.connect(rw, req)
	case "connectionstatus":
		h.connectionstatus(rw, req)
	case "game":
		h.game(rw, req)
	case "metadata":
		h.metadata(rw, req)
	case "state":
		h.state(rw, req)
	case "history":
		h.history(rw, req)
	case "join":
		h.join(rw, req)
	case "leave":
		h.leave(rw, req)
	case "postvote":
		h.postvote(rw, req)
	case "status":
		h.status(rw, req)
	default:
		rw.WriteHeader(http.StatusNotFound)
	}
}

func (h *Handler) connect(rw http.ResponseWriter, req *http.Request) {
	if conn != nil {
		conn.Close()
	}
	var err error
	conn, err = grpc.Dial(req.FormValue("gs-connect-address"),
		grpc.WithTransportCredentials(credentials.NewTLS(h.TLS)),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			PermitWithoutStream: true,
			Time:                time.Hour,
			Timeout:             10 * time.Second,
		}))
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rw).Encode(err)
		return
	}
	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(struct {
		Target string
		State  string
	}{conn.Target(), conn.GetState().String()})
	gsc = gs.NewGameServerClient(conn)
}

func (h *Handler) connectionstatus(rw http.ResponseWriter, req *http.Request) {
	target := ""
	state := "NOT CONNECTED"
	if conn != nil {
		target = conn.Target()
		state = conn.GetState().String()
	}
	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(struct {
		Target string
		State  string
	}{target, state})
}

func (h *Handler) game(rw http.ResponseWriter, req *http.Request) {
	if gsc == nil {
		rw.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintln(rw, "Not not connected to Game Server service.")
		return
	}
	res, err := gsc.Game(context.Background(), &gs.GameRequest{
		Detailed: req.FormValue("gs-game-detailed") == "false",
	})
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rw).Encode(err)
		return
	}
	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(res)
}

func (h *Handler) metadata(rw http.ResponseWriter, req *http.Request) {
	if gsc == nil {
		rw.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintln(rw, "Not not connected to Game Server service.")
		return
	}
	rw.WriteHeader(http.StatusNotImplemented)
}

func (h *Handler) state(rw http.ResponseWriter, req *http.Request) {
	if gsc == nil {
		rw.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintln(rw, "Not not connected to Game Server service.")
		return
	}
	rw.WriteHeader(http.StatusNotImplemented)
}
func (h *Handler) history(rw http.ResponseWriter, req *http.Request) {
	if gsc == nil {
		rw.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintln(rw, "Not not connected to Game Server service.")
		return
	}
	rw.WriteHeader(http.StatusNotImplemented)
}
func (h *Handler) join(rw http.ResponseWriter, req *http.Request) {
	if gsc == nil {
		rw.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintln(rw, "Not not connected to Game Server service.")
		return
	}
	rw.WriteHeader(http.StatusNotImplemented)
}
func (h *Handler) leave(rw http.ResponseWriter, req *http.Request) {
	if gsc == nil {
		rw.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintln(rw, "Not not connected to Game Server service.")
		return
	}
	rw.WriteHeader(http.StatusNotImplemented)
}
func (h *Handler) postvote(rw http.ResponseWriter, req *http.Request) {
	if gsc == nil {
		rw.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintln(rw, "Not not connected to Game Server service.")
		return
	}
	rw.WriteHeader(http.StatusNotImplemented)
}
func (h *Handler) status(rw http.ResponseWriter, req *http.Request) {
	if gsc == nil {
		rw.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintln(rw, "Not not connected to Game Server service.")
		return
	}
	rw.WriteHeader(http.StatusNotImplemented)
}
