package players

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
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
	case "connectionStatus":
		h.connectionStatus(rw, req)
	case "create":
		h.create(rw, req)
	default:
		rw.WriteHeader(404)
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
		rw.WriteHeader(400)
		json.NewEncoder(rw).Encode(err)
		return
	}
	rw.WriteHeader(200)
	json.NewEncoder(rw).Encode(struct {
		Target string
		State  string
	}{conn.Target(), conn.GetState().String()})
	prc = pr.NewPlayersRegistrarClient(conn)
}

func (h *Handler) connectionStatus(rw http.ResponseWriter, req *http.Request) {
	target := ""
	state := "NOT CONNECTED"
	if conn != nil {
		target = conn.Target()
		state = conn.GetState().String()
	}
	rw.WriteHeader(200)
	json.NewEncoder(rw).Encode(struct {
		Target string
		State  string
	}{target, state})
}

func (h *Handler) create(rw http.ResponseWriter, req *http.Request) {
	ctx := context.Background()
	if prc == nil {
		rw.WriteHeader(503)
		return
	}
	res, err := prc.RegisterPlayer(ctx, &pr.RegisterPlayerRequest{
		Username: req.FormValue("create-player-username"),
	})
	if err != nil {
		rw.WriteHeader(400)
		json.NewEncoder(rw).Encode(err)
		return
	}
	rw.WriteHeader(200)
	json.NewEncoder(rw).Encode(res)
}
