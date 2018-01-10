import React from 'react';
    import ReactDOM from 'react-dom';
    import ChessGame from 'chess.js';
import Countdown from './countdown';
import VotingChessboard from './votingChessboard';
import API from './api';
import MoveList from './moveList';
import favicon from './community-chess.png';

import './index.css';

class Game extends React.Component {

    constructor(props) {
        super(props);
        this.state = {
            window: {
                width: window.innerWidth,
                height: window.innerHeight,
            },
            gameState: {
                pgn: '',
                votePgn: null,
                endTimeMs: 0
            }
        };
        this.game = new ChessGame();
        this.resizeHandler = ()=>{this.setState({window: {width: window.innerWidth, height: window.innerHeight}})};
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
        alert('Not yet implemented, sorry.\n\nIt will be okay, your vote wasn\'t that bad. :)');
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

        const chessboardSize = (Math.min(this.state.window.width, this.state.window.height)/2);

        return (
            <div>
                <div id="Header">
                    <a id="Title"
                        href="https://github.com/samdamana/community-chess"
                        rel="noopener noreferrer"
                        target="_blank"
                        title="Community Chess on GitHub"><img src={favicon} alt=""/>Community Chess</a>
                    <span id="NextMoveIn">Next Move</span>
                    <span id="MoveTimer"><Countdown endTimeMs={this.state.gameState.endTimeMs}/></span>
                </div>
                <div id="GameColumns">
                    <div className="column">
                        <h2>Next Moves</h2>
                        <MoveList/>
                    </div>
                    <div className="column">
                        <h2>Board</h2>
                        <VotingChessboard
                            width={chessboardSize}
                            game={this.game}
                            vote={this.state.gameState.votePgn}
                            onVoteCast={this.castVoteToAPI.bind(this)}/>
                        <div id="TurnInfo">{turnMessage}</div>
                        <div>{voteMessage}</div>
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
