import React from 'react';
import Board from 'chessboardjs';
import ChessGame from 'chess.js';
import VotingChessboard from './votingChessboard';

/* CSS */
import 'chessboardjs/www/css/chessboard.css';
import './moveChessboard.css';
import $ from "jquery";

class MoveChessboard extends React.Component {

    constructor(props) {
        super(props);

        this.state = {
            hovered: false,
            mx: 0,
            my: 0,
        }
    }

    componentDidMount() {
        let game = new ChessGame();
        game.load_pgn(this.props.movePgn);
        const cfg = {
            showNotation: false,
            pieceTheme: VotingChessboard.pieceTheme,
            position: game.fen(),
            draggable: false
        };
        let moveCode = this.props.movePgn.split(' ').reverse()[0];
        this.board = Board('moveChessboard_'+moveCode, cfg);
        const squareEl = $('#moveChessboard_'+moveCode).find('.square-' + moveCode.slice(-2));
        squareEl.addClass('highlight-vote');
        this.highlightedPiece = squareEl;
        this.highlightedPiece = null;
    }

    setHovered(val, event) {
        if(val) {
            this.setState({
               hovered: true,
               mx: event.clientX,
               my: event.clientY
            });
        } else {
            this.setState({
                hovered: false,
            })
        }
    }

    render() {
        let offset = 20;
        let chessboardSize = 150;
        let moveCode = this.props.movePgn.split(' ').reverse()[0];
        return (
            <div>
                <span
                    className={"moveName"}
                    onMouseEnter={(e)=>{this.setHovered(true,e)}}
                    onMouseLeave={(e)=>{this.setHovered(false,e)}}>
                    {moveCode}
                </span>
                <div className={this.state.hovered ? "moveChessboard" : "hidden"}
                     style={{
                         "left": (this.state.mx + offset) + "px",
                         "top": (this.state.my - (chessboardSize + offset)) + "px"}}>
                    <div id={"moveChessboard_"+moveCode} style={{"width": chessboardSize + "px"}}>{null}</div>
                </div>
            </div>
        );
    }
}

export default MoveChessboard;