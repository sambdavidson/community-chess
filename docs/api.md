# ![API](./img/community-chess-api.png)
Defines the API as a collection of endpoints and how they map to services.

## Dictionary
- client - Front End, Web UI, Javascript web client.

## Front Endpoints
Simplified set of all endpoints available to the client. For details on any endpoint view its' service description.
| Method | Endpoint                | Description                                           | Service                           | Request           | Returns                            |
| ------ | ----------------------- | ----------------------------------------------------- | --------------------------------- | ----------------- | ---------------------------------- |
| *      | /login                  | (TODO) Some sort of identity flow for obtaining EUCs. | TODO                              | TODO              | TODO                               |
| GET    | /players/${PlayerId}    | Get details of player ${PlayerId}                     | [Players Server](#Players-Server) |                   | [PlayerExtended](#Player-Extended) |
| GET    | /games                  | Collection of publicly available games.               | [MC Server](#mc-server)           |                   | [GamesCollection](#gamecollection) |
| POST   | /games                  | Create a new game.                                    | [MC Server](#mc-server)           | [Game](#game)     |                                    |
| GET    | /game/${GameId}         | Description of game ${GameId}.                        | [Game Server](#game-server)       |                   | [GameMetadata](#gamemetadata)      |
| POST   | /game/${GameId}/players | Join (add player) at the game ${GameId}.              | [Game Server](#game-server)       | [Player](#Player) |                                    |
| POST   | /game/${GameId}/vote    | Cast a vote to the game ${GameId}.                    | [Game Server](#game-server)       | [GameVote](#Vote) |                                    |

## Generic Structures

### Generic Aliases
```Typescript
type PlayerId = string;
type GameId = string;
type ChatId = string;
```

### Player
Referenced Types: [PlayerId](#Generic-Aliases)
```Typescript
class Player {
    playerId: PlayerId;
    nickname: string;
}
```

### PlayerExtended
Refrenced Types: [Player](#Player), [GameId](#Generic-Aliases)
```Typescript
class PlayerExtended extends Player {
    email: string;
    username: string;
    games: GameId[];
    // TODO: Figure out OAuth fields.
}
```

### GamesCollection
Referenced Types: [Date](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Date), [Game](#Game)
```Typescript
class GamesCollection {
    time: Date;
    games: Game[];
}
```

### Game
Implementations: [Chess](#chess)

Referenced Types: [GameId](#Generic-Aliases), [ChatId](#Generic-Aliases), [Date](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Date), [GameMetadata](#GameMetadata), [GameState](#GameState), [Player](#Player)
```Typescript
interface Game {
    gameName: string;
    gameId: GameId;
    chatId: ChatId;
    creationTime: Date;
    metadata: GameMetadata;
    state: GameState;
    players: Player[]
}
```

### GameMetadata
Implementations: [ChessMetadata](#chessmetadata)

Referenced Types: [GameRules](#GameRules)
```Typescript
interface GameMetadata {
    gameName: string;
    version: number;
    title: string;
    publiclyVisible: boolean;
    rules: GameRules;
}
```

### GameRules
Implementations: [ChessRules](#chessrules)

Referenced Types: [VoteApplication](#VoteApplication)
```Typescript
interface GameRules {
    voteApplication: VoteApplication;
}
```

### VoteApplication
Implementations: [VoteAppliedImmediately](#VoteAppliedImmediately), [VoteAppliedAfterTally](#VoteAppliedAfterTally)
```Typescript
interface VoteApplication {
    voteApplicationName: string;
}
```

### VoteAppliedImmediately
```Typescript
class VoteAppliedImmediately implements VoteApplication {
    voteApplicationName: string = "voteAppliedImmediately";
}
```

### VoteAppliedAfterTally

```Typescript
class VoteAppliedAfterTally implements VoteApplication {
    voteApplicationName: string = "voteAppliedAfterTally";
    voteTimeout: number;
    selectionType: SelectionType;
}
```

### SelectionType
```Typescript
enum SelectionType {
    Probability = "PROBABILITY", // Selection is made randomly, more votes means higher probability.
    MostVotes = "MOSTVOTES", // Selection is made by which has the most votes.
}
```

### GameState
Implementations: [ChessState](#chessstate)
```Typescript
interface GameState {
    gameName: string;
    version: number;
}
```

### GameVote

```Typescript
interface GameVote {
    gameName: string;
}
```

## Chess Structures

### Chess Aliases
```Typescript
type FEN = string; // Forsyth-Edwards Notation
type PGN = string; // Portable Game Notation
```

### Chess
Referenced Types: [Game](#Game), [GameId](#Generic-Aliases), [ChatId](#Generic-Aliases), [Date](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Date), [ChessMetadata](#ChessMetadata), [ChessState](#ChessState), [Player](#Player)
```Typescript
class Chess implements Game {
    gameName: string = "chess";
    gameId: GameId;
    chatId: ChatId;
    creationTime: Date;
    metadata: ChessMetadata;
    state: ChessState;
    players: Player[]
}
```

### ChessMetadata
Referenced Types: [GameMetadata](#GameMetadata), [ChessRules](#ChessRules)
```Typescript
class ChessMetadata implements GameMetadata {
    gameName: string = "chess";
    version: number;
    title: string;
    publiclyVisible: boolean;
    rules: ChessRules
}
```

### ChessRules
Referenced Types: [GameRules](#gamerules), [VoteApplication](#VoteApplication)
```Typescript
class ChessRules implements GameRules {
    voteApplication: VoteApplication;
    balancedTeams: boolean;
}
```

### ChessState
Referenced Types: [GameState](#GameState), [Date](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Date), [ChessRound](#ChessRound)
```Typescript
class ChessState implements GameState {
    gameName: string = "chess";
    version: number;
    turnEnd: Date;
    rounds: ChessRound[];
}
```

### ChessRound
Referenced Types: [FEN](#Chess-Aliases), [ChessVote](#ChessVote)
```Typescript
class ChessRound {
    board: FEN;
    votes: {[player: playerId]: ChessVote;};
}
```

### ChessVote
Referenced Types: [GameVote](#GameVote), [PGN](#Chess-Aliases), [PlayerId](#Generic-Aliases)
```Typescript
class ChessVote implements GameVote {
    gameName: string = "chess";
    movePGN: PGN;
    voters: PlayerId[];
}
```

## Services

### Player Server
Manages OAuth flow. Validates identity. Serves requests for [PlayerExtended](#PlayerExtended) objects.

### MC Server
The Master of Ceremonies Server (MC Server) is responsible for enumerating available games to the client. 

### Game Server
The Game Server run the actual game. 

