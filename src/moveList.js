import React from 'react';
import API from './api';
import ChessGame from 'chess.js';
import MoveChessboard from './moveChessboard';

import './moveList.css';

class MoveList extends React.Component {

    constructor(props) {
        super(props);
        this.state = {
            moveList: [],
            totalVotes: 0
        };
    }

    componentDidMount() {
        API.gameState.subscribe((gameState) => {
            let moveList = gameState.votes;
            moveList.sort((a,b) => { // Sort in reverse
               return b.votes- a.votes;
            });
            let totalVotes = 0;
            moveList.forEach((m)=>{totalVotes += m.votes;});
            this.setState({
                moveList: moveList,
                totalVotes: totalVotes
            });
        });
    }

    render() {
        const movesElements = this.state.moveList.map((move) => {
            const game = new ChessGame();
            game.load_pgn(move.pgn);
            return (
                <tr key={move.pgn} className={(move.winner ? 'winning-move' : '')}>
                    <td>{move.votes}</td>
                    <td><MoveChessboard game={game}/></td>
                    <td>{Math.floor((move.votes / this.state.totalVotes) * 100)}%</td>
                </tr>
            )
        });

        return (
            <div id="MoveList">
                <table cellPadding={10} cellSpacing={0}>
                    <tbody>
                    <tr>
                        <th>Votes</th>
                        <th>Move</th>
                        <th>Chance</th>
                    </tr>
                    {movesElements}
                    </tbody>
                </table>
                {/*<div id="MoveHistoryButtons">1 2 3 4</div>*/}
            </div>
        );
    }
}
export default MoveList;