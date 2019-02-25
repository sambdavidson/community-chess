// Unique identifier for a player. Global and one PlayerId per account.
export type PlayerId = string;
// Unique identifier for every game created.
export type GameId = string;
// Unique identifier for a chat conversation thread. 
export type ChatId = string;

export class Player {
    playerId: PlayerId;
    nickname: string;
}

export class PlayerExtended extends Player {
    email: string;
    username: string;
    games: GameId[];
    // TODO: Figure out OAuth fields.
}

export class GamesCollection {
    time: Date;
    games: IGame[];
}

export interface IGame {
    gameName: string;
    gameId: GameId;
    chatId: ChatId;
    creationTime: Date;
    metadata: IGameMetadata;
    state: IGameState;
    players: Player[]
}

export interface IGameMetadata {
    gameName: string;
    version: number;
    title: string;
    publiclyVisible: boolean;
    rules: IGameRules;
}

export interface IGameRules {
    voteApplication: IVoteApplication;
}

export interface IVoteApplication {
    voteApplicationName: string;
}

export class VoteAppliedImmediately implements IVoteApplication {
    voteApplicationName: string = "voteAppliedImmediately";
}

export class VoteAppliedAfterTally implements IVoteApplication {
    voteApplicationName: string = "voteAppliedAfterTally";
    voteTimeout: number;
    selectionType: SelectionType;
}

export enum SelectionType {
    Probability = "PROBABILITY", // Selection is made randomly, more votes means higher probability.
    MostVotes = "MOSTVOTES", // Selection is made by which has the most votes.
}

export interface IGameState {
    gameName: string;
    version: number;
}

export interface IGameVote {
    gameName: string;
}