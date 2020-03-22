package gamemaster

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sambdavidson/community-chess/src/proto/messages"
	"github.com/sambdavidson/community-chess/src/proto/messages/games"
	gs "github.com/sambdavidson/community-chess/src/proto/services/games/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

// Handler handles game master operation HTTP stuff.
type Handler struct {
	TLS *tls.Config
}

var (
	conn              *grpc.ClientConn
	gmc               gs.GameServerMasterClient
	errorNotConnected = fmt.Errorf("not not connected to Game Master service")
)

func (h *Handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	req.ParseMultipartForm(100000)
	rw.Header().Set("Content-Type", "application/json")

	switch req.URL.Path { // Handlers for following the /gamemaster/ prefix e.g. the /gamemaster/initialize URL
	case "connect":
		h.connect(rw, req)
	case "connectionstatus":
		h.connectionstatus(rw, req)
	case "initialize":
		h.initialize(rw, req)
	case "addslave":
		h.addslave(rw, req)
	case "addplayers":
		h.addplayers(rw, req)
	case "removeplayers":
		h.removeplayers(rw, req)
	case "stopgame":
		h.stopgame(rw, req)
	default:
		rw.WriteHeader(http.StatusNotFound)
	}
}

func (h *Handler) connect(rw http.ResponseWriter, req *http.Request) {
	if conn != nil {
		conn.Close()
	}
	var err error
	conn, err = grpc.Dial(req.FormValue("gm-connect-address"),
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
	gmc = gs.NewGameServerMasterClient(conn)
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

func (h *Handler) initialize(rw http.ResponseWriter, req *http.Request) {
	if gmc == nil {
		rw.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintln(rw, errorNotConnected)
		return
	}
	res, err := gmc.Initialize(context.Background(), &gs.InitializeRequest{
		Game: &messages.Game{
			Id:   req.FormValue("gm-init-id"),
			Type: messages.Game_CHESS,
			Metadata: &messages.Game_Metadata{
				Title: req.FormValue("gm-init-title"),
				Rules: &messages.Game_Metadata_Rules{
					VoteApplication: &messages.Game_Metadata_Rules_VoteAppliedImmediately_{
						VoteAppliedImmediately: &messages.Game_Metadata_Rules_VoteAppliedImmediately{},
					},
					GameSpecific: &messages.Game_Metadata_Rules_ChessRules{
						ChessRules: &games.ChessRules{
							BalanceEnforcement: &games.ChessRules_TolerateDifference{
								TolerateDifference: 10,
							},
						},
					},
				},
			},
			State: &messages.Game_State{
				Game: &messages.Game_State_ChessState{
					ChessState: &games.ChessState{
						BoardFen:       req.FormValue("gm-init-chessboard-fen"),
						RoundIndex:     1,
						RoundStartTime: time.Now().UnixNano(),
						RoundEndTime:   time.Now().Add(time.Minute * 10).UnixNano(),
						Details: &games.ChessState_Details{
							PlayerIdToTeam: map[string]bool{},
							PlayerToMove:   map[string]string{},
						},
					},
				},
			},
			History: &messages.Game_History{
				Game: &messages.Game_History_ChessHistory{
					ChessHistory: &games.ChessHistory{
						StateHistory: []*games.ChessState{},
					},
				},
			},
		},
	})
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rw).Encode(err)
		return
	}
	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(res)
}

func (h *Handler) addslave(rw http.ResponseWriter, req *http.Request) {
	if gmc == nil {
		rw.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintln(rw, errorNotConnected)
		return
	}
	rw.WriteHeader(http.StatusNotImplemented)
}

func (h *Handler) addplayers(rw http.ResponseWriter, req *http.Request) {
	if gmc == nil {
		rw.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintln(rw, errorNotConnected)
		return
	}
	rw.WriteHeader(http.StatusNotImplemented)
}

func (h *Handler) removeplayers(rw http.ResponseWriter, req *http.Request) {
	if gmc == nil {
		rw.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintln(rw, errorNotConnected)
		return
	}
	rw.WriteHeader(http.StatusNotImplemented)
}

func (h *Handler) stopgame(rw http.ResponseWriter, req *http.Request) {
	if gmc == nil {
		rw.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintln(rw, errorNotConnected)
		return
	}
	rw.WriteHeader(http.StatusNotImplemented)
}
