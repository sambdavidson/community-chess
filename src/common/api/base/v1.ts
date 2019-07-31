import { HTTPMethods } from "../../httpMethods";

export namespace Models {
    // Unique identifier for every game created.
    export type GameId = string;
    // Unique identifier for a chat conversation thread. 
    export type ChatId = string;

    export type Player = PlayerLite | PlayerFull;
    export namespace Player {
        /**  
         * Unique identifier for a player. Global and one PlayerId per account.
         * 128-bit UUID/v4 
         **/
        export type Id = string;
    }
    export class PlayerLite {
        pType: 'playerSimple';
        playerId: Player.Id;
        nickname: string;
    }
    export class PlayerFull {
        pType: 'playerFull';
        playerId: Player.Id;
        nickname: string;
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
        voteApplication: VoteApplication;
    }

    export type VoteApplication = VoteAppliedImmediately | VoteAppliedAfterTally;

    export class VoteAppliedImmediately {
        voteApplicationName: string = 'voteAppliedImmediately';
    }

    export class VoteAppliedAfterTally {
        voteApplicationName: string = 'voteAppliedAfterTally';
        voteTimeout: number;
        selectionType: SelectionType;
    }

    export enum SelectionType {
        Probability = 'PROBABILITY', // Selection is made randomly, more votes means higher probability.
        MostVotes = 'MOSTVOTES', // Selection is made by which has the most votes.
    }

    export interface IGameState {
        gameName: string;
        version: number;
    }

    export interface IGameVote {
        gameName: string;
    }
}

export const RoutePrefix: string = '/v1';

export interface IRoute {
    Method: string;
    Path(... args: any[]): string;
}

export namespace Routes {
    export namespace Players {
        export namespace Get {
            export const Method = HTTPMethods.GET;
            export const Path = (id: Models.Player.Id) => `${RoutePrefix}/players/${id}`;
            export type BodyType = undefined;
            export type ReturnType = Models.Player; 
        }
        export namespace Patch {
            export const Method = HTTPMethods.PATCH;
            export const Path = (id: Models.Player.Id) => `${RoutePrefix}/players/${id}`;
            export type BodyType = Models.Player;
            export type ReturnType = Models.Player; 
        }
    }
    export namespace Games { 
        export namespace Create {
            export const Method = HTTPMethods.POST;
            export const Path = () => `${RoutePrefix}/games`;
            export type BodyType = Models.IGame;
            export type ReturnType = Models.IGameMetadata;
        }
        export namespace List {
            export const Method = HTTPMethods.GET;
            export const Path = () => `${RoutePrefix}/games`;
            export type BodyType = undefined;
            export type ReturnType = Models.GamesCollection;
        }
        export namespace Get {
            export const Method = HTTPMethods.GET;
            export const Path = (id: string) => `${RoutePrefix}/games/${id}`;
            export type BodyType = undefined;
            export type ReturnType = Models.IGameMetadata;
        }
        export namespace AddPlayer {
            export const Method = HTTPMethods.POST;
            export const Path = (id: string) => `${RoutePrefix}/games/${id}/players`;
            export type BodyType = Models.Player;
            export type ReturnType = undefined;
        }
        export namespace CastVote {
            export const Method = HTTPMethods.POST;
            export const Path = (id: string) => `${RoutePrefix}/games/${id}/players`;
            export type BodyType = Models.IGameVote;
            export type ReturnType = undefined;
        }
    }
}

/*
WORK IN PROGRESS 2/25/2019

I have been working on the last little while how to represent the Routes statically within the typing system. 
Using namespaces and classes and stuff to get to the route name Routes > Games > List > Method
I wanted to achieve 3 things properties:

- Method: returning the HTTP Method for this route
- Path: the URL path for this route including parameters to autopopulate it.
- ReturnType: the expected return type of this interface.

I ran into a few problems. You can't define static types in interfaces. 
I wanted interface IRoute to be reused for every end class on the namespace heirarchy. This isn't possible.
This is (sorta) solved by implementing the inteface then instantiating the end-class as a const JS object.
The problem with that is you lose the typing on Path(... args: any[]) and overloading it with Path(id: string) doesn't show up on intellisense.

The other problem is that you can't assign an interface such as IGame to an object or static type. Instead I need to have a 'Get' namespace
in parallel with the Get class or const. This works fine except now there are two parallel 'Get' objects making the file complicated.
*/