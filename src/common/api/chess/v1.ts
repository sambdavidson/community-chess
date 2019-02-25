import * as base from '../base/v1';

// Forsyth-Edwards Notation
export type FEN = string;
// Portable Game Notation
export type PGN = string;

export class Chess implements base.Models.IGame {
    gameName: string = "chess";
    gameId: base.Models.GameId;
    chatId: base.Models.ChatId;
    creationTime: Date;
    metadata: ChessMetadata;
    state: ChessState;
    players: base.Models.Player[]
}

export class ChessMetadata implements base.Models.IGameMetadata {
    gameName: string = "chess";
    version: number;
    title: string;
    publiclyVisible: boolean;
    rules: ChessRules
}

export class ChessRules implements base.Models.IGameRules {
    voteApplication: base.Models.IVoteApplication;
    balancedTeams: boolean;
}

export class ChessState implements base.Models.IGameState {
    gameName: string = "chess";
    version: number;
    turnEnd: Date;
    rounds: ChessRound[];
}

export class ChessRound {
    board: FEN;
    votes: {[playerId: string]: ChessVote};
}

export class ChessVote implements base.Models.IGameVote {
    gameName: string = "chess";
    movePGN: PGN;
    voters: base.Models.PlayerId[];
}