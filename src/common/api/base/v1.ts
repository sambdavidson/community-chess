import { HTTPMethods } from "../../httpMethods";

export namespace Models {
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
        voteApplicationName: string = 'voteAppliedImmediately';
    }

    export class VoteAppliedAfterTally implements IVoteApplication {
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
        export class Get {
            static Method = HTTPMethods.GET;
            static Path
        }
    }
    export namespace Games { 
        // export class Get {
        //     static Method = HTTPMethods.GET;
        //     static Path(id: string): string {
        //         return RoutePrefix + `/games/${id}`;
        //     }
        // }
        export const Get = {
            Method: HTTPMethods.GET,
            Path: (id: string) =>  RoutePrefix + `/games/${id}`
        }
        export namespace Get {
            export type ReturnType = Models.GameId;
        }
        // export class List {
        //     static Method = HTTPMethods.GET;
        //     static Path(): string {
        //         return RoutePrefix + `/games`;
        //     }
        //     static ReturnType = Models.GamesCollection;
        // }
    }
}

interface IFoo {
    a: string | number;
}

interface IV extends IFoo {
    a: string;
}

const v: IV = {
    a: 'a',
}

const f: IFoo = v;

const j: IV = f;

const l = v;

Routes.Games.Get.Path(s);

let r: Routes.Games.Get.ReturnType = Routes.Games.Get.Path('');

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