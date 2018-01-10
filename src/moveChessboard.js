import React from 'react';
import Board from 'chessboardjs';
import VotingChessboard from './votingChessboard';

/* CSS */
import 'chessboardjs/www/css/chessboard.css';
import './moveChessboard.css';
import $ from "jquery";

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
        //let offset = 20;
        let chessboardSize = 150;
        return (
            <div id={"moveChessboard_"+this.state.move.san} style={{"width": chessboardSize + "px"}}>{null}</div>
        );
    }
}

export default MoveChessboard;