import React from 'react';
import API from './api';
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
            return (
                <tr key={move.pgn}>
                    <td><MoveChessboard movePgn={move.pgn} /></td>
                    <td>{move.votes}</td>
                    <td>{Math.floor((move.votes / this.state.totalVotes) * 100)}%</td>
                </tr>
            )
        });

        return (
            <table id={"moveList"}>
                <tbody>
                    <tr>
                        <th>Move</th>
                        <th>Votes</th>
                        <th>Chance</th>
                    </tr>
                    {movesElements}
                </tbody>
            </table>
        );
    }
}
export default MoveList;