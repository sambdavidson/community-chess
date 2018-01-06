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
            windowWidth: window.innerWidth,
            gameState: {
                pgn: '',
                votePgn: null,
                endTimeMs: 0
            }
        };
        this.game = new ChessGame();
        this.resizeHandler = ()=>{this.setState({windowWidth: window.innerWidth})};
    }

    componentDidMount() {
        API.gameState.subscribe(gameState => {
            let change = false;
            for(let prop in this.state.gameState) {
                if(this.state.gameState.hasOwnProperty(prop)) {
                    if(gameState[prop] !== this.state.gameState[prop]) {
                        change = true;
                        break;
                    }
                }
            }
            if (!change) {
                return;
            }
            this.game.reset();
            if (gameState.votePgn) {
                this.game.load_pgn(gameState.votePgn);
            } else {
                this.game.load_pgn(gameState.pgn);
            }

            this.setState({
                gameState: {
                    pgn: gameState.pgn,
                    votePgn: gameState.votePgn,
                    endTimeMs: gameState.endTimeMs
                },

            });
        });
        window.addEventListener("resize", this.resizeHandler);
    }

    castVoteToAPI(pgn) {
        this.setState({
            gameState: {
                pgn: this.state.gameState.pgn,
                votePgn: pgn,
                endTimeMs: this.state.gameState.endTimeMs
            }
        });
        API.castVote(pgn);
    }

    resetVote() {
        alert('TODO: Reset Vote');
    }

    render() {
        let voteMessage;
        if (this.state.gameState.votePgn) {
            voteMessage = <a onClick={this.resetVote} style={{"cursor" : "pointer"}}>(Reset Vote)</a>;
        } else {
            voteMessage = '';
        }

        let turnMessage;
        if (this.state.gameState.endTimeMs === 0) {
            turnMessage = '';
        } else {
            if(this.state.gameState.votePgn) {
                turnMessage = this.game.turn() === 'b' ? 'White\'s Turn' : 'Black\'s Turn';
            } else {
                turnMessage = this.game.turn() === 'w' ? 'White\'s Turn' : 'Black\'s Turn';
            }
        }

        return (
            <div>
                <div id="Title">Community Chess</div>
                <div id="NextMoveIn">Next move in</div>
                <Countdown endTimeMs={this.state.gameState.endTimeMs}/>
                <div id="TurnInfo">{turnMessage}</div>
                <div id="GameColumns">
                    <div className="column">
                        <h2>Next Moves</h2>
                        <MoveList/>
                    </div>
                    <div className="column">
                        <h2>Board</h2>
                        <VotingChessboard
                            width={(this.state.windowWidth/3) - 20}
                            game={this.game}
                            vote={this.state.gameState.votePgn}
                            onVoteCast={this.castVoteToAPI.bind(this)}/>
                        {voteMessage}
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
