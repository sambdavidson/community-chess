import React from 'react';
import $ from 'jquery';
import Board from 'chessboardjs';

/* Chess pieces */
import wP from 'chessboardjs/www/img/chesspieces/alpha/wP.png';
import wR from 'chessboardjs/www/img/chesspieces/alpha/wR.png';
import wN from 'chessboardjs/www/img/chesspieces/alpha/wN.png';
import wB from 'chessboardjs/www/img/chesspieces/alpha/wB.png';
import wQ from 'chessboardjs/www/img/chesspieces/alpha/wQ.png';
import wK from 'chessboardjs/www/img/chesspieces/alpha/wK.png';
import bP from 'chessboardjs/www/img/chesspieces/alpha/bP.png';
import bR from 'chessboardjs/www/img/chesspieces/alpha/bR.png';
import bN from 'chessboardjs/www/img/chesspieces/alpha/bN.png';
import bB from 'chessboardjs/www/img/chesspieces/alpha/bB.png';
import bQ from 'chessboardjs/www/img/chesspieces/alpha/bQ.png';
import bK from 'chessboardjs/www/img/chesspieces/alpha/bK.png';

/* CSS */
import 'chessboardjs/www/css/chessboard.css';
import './votingChessboard.css';

window.$ = $;
window.jQuery = $;

class VotingChessboard extends React.Component {

    componentDidMount() {
        const cfg = {
            pieceTheme: VotingChessboard.pieceTheme,
            draggable: true,
            dropOffBoard: 'snapback', // this is the default
            onDragStart: this.onDragStart.bind(this),
            onDrop: this.onDrop.bind(this),
            onSnapEnd: this.onSnapEnd.bind(this),
            onMouseoutSquare: this.onMouseoutSquare.bind(this),
            onMouseoverSquare: this.onMouseoverSquare.bind(this),
        };
        this.board = Board('voteChessboard', cfg);
        this.highlightedPiece = null;
    }


    componentDidUpdate() {
        // TODO: Solve weird flickering when voting sometimes.
        this.board.resize();
        this.board.position(this.props.game.fen());

        // Must come after a resize
        if(this.props.vote) {

            if(this.highlightedPiece) {
                this.highlightedPiece.removeClass('highlight-vote');
            }
            const squareEl = $('#voteChessboard').find('.square-' + this.props.vote.slice(-2));
            squareEl.addClass('highlight-vote');
            this.highlightedPiece = squareEl;
        }
    }
    static pieceTheme(piece) {
        return {
            "bB": bB,
            "bK": bK,
            "bN": bN,
            "bP": bP,
            "bQ": bQ,
            "bR": bR,
            "wB": wB,
            "wK": wK,
            "wN": wN,
            "wP": wP,
            "wQ": wQ,
            "wR": wR
        }[piece];
    }

    render() {
        return (
            <div style={{"display": "inline-block"}}>
                <div id="voteChessboard" style={{"width": this.props.width + "px"}}>{null}</div>
            </div>
        );
    }

    /* Chess Rule Handlers */

    static removeSquareBackgrounds() {
        $('#voteChessboard').find('.square-55d63').css('background', '');
    };

    static colorSquare(square, wColor, bColor) {
        const squareEl = $('#voteChessboard').find('.square-' + square);


        let background = wColor;
        if (squareEl.hasClass('black-3c85d') === true) {
            background = bColor;
        }

        squareEl.css('background', background);
    };

    onMouseoverSquare(square, piece) {
        if (this.props.game.game_over() === true || !!this.props.vote) {
            return;
        }
        // get list of possible moves for this square
        const moves = this.props.game.moves({
            square: square,
            verbose: true
        });

        // exit if there are no moves available for this square
        if (moves.length === 0) return;

        // highlight the square they moused over
        VotingChessboard.colorSquare(square, '#a9a9a9', '#696969');

        // highlight the possible squares for this piece
        for (let i = 0; i < moves.length; i++) {
            VotingChessboard.colorSquare(moves[i].to, '#a9a9a9', '#696969');
        }
    };

    onMouseoutSquare(square, piece) {
        VotingChessboard.removeSquareBackgrounds();
    };

    onDragStart(source, piece, position, orientation) {
        if (this.props.game.game_over() === true ||
            !!this.props.vote ||
            (this.props.game.turn() === 'w' && piece.search(/^b/) !== -1) ||
            (this.props.game.turn() === 'b' && piece.search(/^w/) !== -1)) {
            return false;
        }
    };

    onDrop(source, target) {
        // see if the move is legal
        const move = this.props.game.move({
            from: source,
            to: target,
            promotion: 'q' // NOTE: always promote to a queen for example simplicity
        });

        // illegal move
        if (move === null) return 'snapback';

        this.props.onVoteCast(this.props.game.pgn());
    };

    // update the board position after the piece snap
    // for castling, en passant, pawn promotion
    onSnapEnd() {
        this.board.position(this.props.game.fen());
    };
}

export default VotingChessboard;