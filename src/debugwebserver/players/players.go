package players

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	pr "github.com/sambdavidson/community-chess/src/proto/services/players/registrar"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

// Handler handles player operation HTTP stuff.
type Handler struct {
	TLS *tls.Config
}

var (
	conn *grpc.ClientConn
	prc  pr.PlayersRegistrarClient
)

func (h *Handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	req.ParseMultipartForm(100000)
	rw.Header().Set("Content-Type", "application/json")

	switch req.URL.Path { // Handlers for following the /players/ prefix e.g. the /player/create URL
	case "connect":
		h.connect(rw, req)
	case "connectionstatus":
		h.connectionstatus(rw, req)
	case "create":
		h.create(rw, req)
	case "get":
		h.get(rw, req)
	case "login":
		h.login(rw, req)
	default:
		rw.WriteHeader(http.StatusNotFound)
	}
}

func (h *Handler) connect(rw http.ResponseWriter, req *http.Request) {
	if conn != nil {
		conn.Close()
	}
	var err error
	conn, err = grpc.Dial(req.FormValue("pr-connect-address"),
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
	prc = pr.NewPlayersRegistrarClient(conn)
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

func (h *Handler) create(rw http.ResponseWriter, req *http.Request) {
	if prc == nil {
		rw.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintln(rw, "Not not connected to playerregistrar service.")
		return
	}
	res, err := prc.RegisterPlayer(context.Background(), &pr.RegisterPlayerRequest{
		Username: req.FormValue("create-player-username"),
	})
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rw).Encode(err)
		return
	}
	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(res)
}

func (h *Handler) get(rw http.ResponseWriter, req *http.Request) {
	if prc == nil {
		rw.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintln(rw, "Not not connected to playerregistrar service.")
		return
	}
	pid := req.FormValue("get-player-uuid")
	if pid == "" {
		pid = req.FormValue("get-player-known-player-dropdown")
	}
	res, err := prc.GetPlayer(context.Background(), &pr.GetPlayerRequest{
		PlayerId: pid,
	})
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rw).Encode(err)
		return
	}
	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(res)
}

func (h *Handler) login(rw http.ResponseWriter, req *http.Request) {
	if prc == nil {
		rw.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintln(rw, "Not not connected to playerregistrar service.")
		return
	}
	suffix, err := strconv.Atoi(req.FormValue("login-number-suffix"))
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rw).Encode(err)
		return
	}
	res, err := prc.Login(context.Background(), &pr.LoginRequest{
		Username:     req.FormValue("login-username"),
		NumberSuffix: int32(suffix),
	})
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rw).Encode(err)
		return
	}
	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(res)
}
