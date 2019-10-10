package chess

import (
	"context"
	"log"

	"github.com/sambdavidson/community-chess/src/proto/messages/games"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	ch "github.com/notnil/chess"
	"github.com/sambdavidson/community-chess/src/proto/messages"
	pb "github.com/sambdavidson/community-chess/src/proto/services/games/server"
)

// ChangeAcceptingVotes is called by GameServerMasters to set this GameServerSlave to no longer accept votes. Typically done at end of a voting round.
func (i *Implementation) ChangeAcceptingVotes(ctx context.Context, in *pb.ChangeAcceptingVotesRequest) (*pb.ChangeAcceptingVotesResponse, error) {
	log.Println("ChangeAcceptingVotes", in)
	i.moveMux.Lock()
	defer i.moveMux.Unlock()
	i.acceptingVotes = in.GetAcceptingVotes()
	return &pb.ChangeAcceptingVotesResponse{}, nil
}

// GetVotes is called by GameServerMasters get all votes received by this GameServerSlave for the current round.
func (i *Implementation) GetVotes(ctx context.Context, in *pb.GetVotesRequest) (*pb.GetVotesResponse, error) {
	log.Println("GetVotes", in)
	i.moveMux.Lock()
	defer i.moveMux.Unlock()

	votes := []*messages.Vote{}
	for p, m := range i.playerToMove {
		votes = append(votes, &messages.Vote{
			PlayerId: p,
			GameVote: &messages.Vote_ChessVote{
				ChessVote: &games.ChessVote{
					RoundIndex: i.roundIndex,
					Move:       m,
				},
			},
		})
	}

	return &pb.GetVotesResponse{
		Complete:   !i.acceptingVotes,
		RoundIndex: i.roundIndex,
		Votes:      votes,
	}, nil
}

// PostVote posts a vote to this game.
func (i *Implementation) PostVote(ctx context.Context, in *pb.PostVoteRequest) (*pb.PostVoteResponse, error) {
	log.Println("PostVote", in)
	if in.GetVote().GetChessVote().GetRoundIndex() != i.roundIndex {
		return nil, status.Errorf(codes.InvalidArgument, "bad round index %d; current round %d", in.GetVote().GetChessVote().GetRoundIndex(), i.roundIndex)
	}
	t, ok := i.playerToTeam[in.GetVote().GetPlayerId()]
	if !ok {
		return nil, status.Errorf(codes.PermissionDenied, "player %s has not joined this game", in.GetVote().GetPlayerId())
	}
	if t != (i.game.Position().Turn() == ch.White) {
		return nil, status.Errorf(codes.PermissionDenied, "player %s is not part of team: %s", in.GetVote().GetPlayerId(), i.game.Position().Turn())
	}
	_, err := ch.AlgebraicNotation{}.Decode(i.game.Position(), in.GetVote().GetChessVote().GetMove())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid move %s: %v", in.GetVote().GetChessVote().GetMove(), err)
	}
	i.moveMux.Lock()
	defer i.moveMux.Lock()
	if move, ok := i.playerToMove[in.GetVote().GetPlayerId()]; ok {
		i.moveToCount[move]--
	}
	i.playerToMove[in.GetVote().GetPlayerId()] = in.GetVote().GetChessVote().GetMove()
	i.moveToCount[in.GetVote().GetChessVote().GetMove()]++
	return &pb.PostVoteResponse{}, nil
}
