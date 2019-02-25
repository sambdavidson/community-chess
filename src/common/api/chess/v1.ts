import * as base from '../base/v1';

// Forsyth-Edwards Notation
export type FEN = string;
// Portable Game Notation
export type PGN = string;

export class Chess implements base.IGame {
    gameName: string = "chess";
    gameId: base.GameId;
    chatId: base.ChatId;
    creationTime: Date;
    metadata: ChessMetadata;
    state: ChessState;
    players: base.Player[]
}

export class ChessMetadata implements base.IGameMetadata {
    gameName: string = "chess";
    version: number;
    title: string;
    publiclyVisible: boolean;
    rules: ChessRules
}

export class ChessRules implements base.IGameRules {
    voteApplication: base.IVoteApplication;
    balancedTeams: boolean;
}

export class ChessState implements base.IGameState {
    gameName: string = "chess";
    version: number;
    turnEnd: Date;
    rounds: ChessRound[];
}

export class ChessRound {
    board: FEN;
    votes: {[playerId: string]: ChessVote};
}

export class ChessVote implements base.IGameVote {
    gameName: string = "chess";
    movePGN: PGN;
    voters: base.PlayerId[];
}