import React from 'react';
    import ReactDOM from 'react-dom';
    import ChessGame from 'chess.js';
import Countdown from './countdown';
import VotingChessboard from './votingChessboard';
import API from './api';
import MoveList from './moveList';

import './index.css';

class Game extends React.Component {

    constructor(props) {
        super(props);
        this.state = {
            userId: 'user',
            windowWidth: window.innerWidth,
            turnCount: 1,
            playerUp: 'w',
            turnInfo: {
                pgn: '',
                votePgn: null,
                endTimeMs: 0
            }
        };
        this.pendingVote = false;
        this.game = new ChessGame();
        this.resizeHandler = ()=>{this.setState({windowWidth: window.innerWidth})};
    }

    componentDidUpdate() {
    }

    componentDidMount() {
        API.gameState.subscribe(gameState => {
            this.pendingVote = false;
            this.game.reset();
            if (gameState.votePgn) {
                this.game.load_pgn(gameState.votePgn);
            } else {
                this.game.load_pgn(gameState.pgn);
            }

            let serverGame = new ChessGame();
            serverGame.load_pgn(this.state.turnInfo.pgn);
            let fenFields = serverGame.fen().split(' ');

            this.setState({
                turnInfo: gameState,
                turnCount: fenFields[5],
                playerUp: fenFields[1] === 'w' ? 'White' : 'Black'

            });
        });
        window.addEventListener("resize", this.resizeHandler);
    }

    castVoteToAPI(pgn) {
        this.pendingVote = true;
        this.setState({
            turnInfo: {
                pgn: this.state.turnInfo.pgn,
                votePgn: pgn,
                endTimeMs: this.state.turnInfo.endTimeMs
            }
        });
        API.castVote(this.state.userId, pgn);
    }

    voteMessage() {
        if (this.pendingVote) {
            return "Submitting vote."
        } else if(this.state.turnInfo.votePgn) {
            return "Voted."
        } else
        return "Vote on a move."
    }

    render() {
        return (
            <div>
                <h1>Community Chess</h1>
                <h3>Turn {this.state.turnCount} - {this.state.playerUp}</h3>
                <div id="NextMoveIn">Next move in</div>
                <Countdown endTimeMs={this.state.turnInfo.endTimeMs}/>
                <div id="GameColumns">
                    <div className="column">
                        <h2>Next Moves</h2>
                        <MoveList/>
                    </div>
                    <div className="column">
                        <h2>Board</h2>
                        <VotingChessboard
                            width={(this.state.windowWidth/3) - 100}
                            game={this.game}
                            vote={this.state.turnInfo.votePgn}
                            onVoteCast={this.castVoteToAPI.bind(this)}/>
                        <div>{this.voteMessage()}</div>
                    </div>
                    <div className="column">
                        <h2>Discussions</h2>
                        <i>Coming soon</i>
                    </div>
                </div>
            </div>
        );
    }

    componentWillUnmount() {
        window.removeEventListener("resize", this.resizeHandler);
    }
}

// ========================================

ReactDOM.render(
    <Game />,
    document.getElementById('root')
);
