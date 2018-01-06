import React from 'react';
import Board from 'chessboardjs';
import VotingChessboard from './votingChessboard';

/* CSS */
import 'chessboardjs/www/css/chessboard.css';
import './moveChessboard.css';
import $ from "jquery";

const FEN_PIECE_NAMES = {
    "p": "♟",
    "n": "♞",
    "b": "♝",
    "r": "♜",
    "q": "♛",
    "k": "♚",
    "P": "♙",
    "N": "♘",
    "B": "♗",
    "R": "♖",
    "Q": "♕",
    "K": "♔",
};

class MoveChessboard extends React.Component {

    constructor(props) {
        super(props);
        this.state = {
            move: props.game.history({verbose: true}).reverse()[0],
            hovered: false,
            mx: 0,
            my: 0,
        }
    }

    componentDidMount() {
        const cfg = {
            showNotation: false,
            pieceTheme: VotingChessboard.pieceTheme,
            position: this.props.game.fen(),
            draggable: false
        };
        this.board = Board('moveChessboard_'+this.state.move.san, cfg);
        const squareEl = $('#moveChessboard_'+this.state.move.san).find('.square-' + this.state.move.to);
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
        return (
            <div>
                <span
                    className={"moveName"}
                    onMouseEnter={(e)=>{this.setHovered(true,e)}}
                    onMouseLeave={(e)=>{this.setHovered(false,e)}}>
                    {FEN_PIECE_NAMES[this.state.move.piece.toLowerCase()]} to {this.state.move.to.toUpperCase()}
                </span>
                <div className={this.state.hovered ? "moveChessboard" : "hidden"}
                     style={{
                         "left": (this.state.mx + offset) + "px",
                         "top": (this.state.my - (chessboardSize + offset)) + "px"}}>
                    <div id={"moveChessboard_"+this.state.move.san} style={{"width": chessboardSize + "px"}}>{null}</div>
                </div>
            </div>
        );
    }
}

export default MoveChessboard;