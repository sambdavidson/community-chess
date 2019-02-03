/**
 * Sets up and manages the game state.
 */
const Chess = require('chess.js/chess').Chess;
const game = new Chess();

let turnTimeout = null;
let turnLengthSeconds = 20;
let userVotes = {};
let votedMoves = [];
let history = [];
let totalVotes = 0;

/**
 * Applies the current votes towards a turn.
 * Only call once, automatically sets a timeout for the next turn.
 */
function applyTurn() {
    turnTimeout = new Date();
    turnTimeout.setSeconds(turnTimeout.getSeconds() + turnLengthSeconds);

    let randomVoteSubtractor = Math.random();

    let topMove = null;
    for (let i = 0; i < votedMoves.length; i++) {
        const move = votedMoves[i];
        randomVoteSubtractor -= (move.votes/totalVotes);
        if(randomVoteSubtractor < 0) {
            topMove = move;
            break;
        }
    }

    if(topMove) {
        game.reset();
        game.load_pgn(topMove.pgn);
        topMove.winner = true;
        history.push(votedMoves);
        votedMoves = [];
        userVotes = {};
        totalVotes = 0;
    } else {
    }

    setTimeout(applyTurn, turnLengthSeconds * 1000);
}

/**
 * Validates move
 * @param {string} pgn
 * @returns {boolean}
 */
function validateMove(pgn) {
    const moveChess = new Chess();
    if (!moveChess.load_pgn(pgn)) {
        return false;
    }
    const lastMove = moveChess.undo();
    if (moveChess.fen() !== game.fen()) {
        return false;
    }
    return !!(new Chess(game.fen())).move(lastMove);

}

exports = module.exports;

exports.initEndpoints = function(app) {

    app.get('/gameState/:id?', (req, res) => {
        let userVote = null;
        if(req.params.id) {
            userVote = userVotes[req.params.id];
        }
        res.json({
            pgn: game.pgn(),
            votePgn: userVote ? userVote : null,
            endTimeMs: turnTimeout.getTime(),
            votes: votedMoves
        });
    });

    app.post('/vote', (req, res) => {

        let vote = {pgn: null};
        if(req.body && req.body.pgn && req.body.id && validateMove(req.body.pgn)) {
            const id = req.body.id;
            vote = req.body;
            const move = votedMoves.find((m) => {
                return m.pgn === vote.pgn;
            });
            if(userVotes[id]) {
                if (move) {
                    move.votes--;
                }
                totalVotes--;
            }
            userVotes[id] = vote.pgn;
            if(move) {
                move.votes++
            } else {
                votedMoves.push({votes: 1, pgn: vote.pgn, winner: false});
            }
            totalVotes++;
        }
        res.status(200).end();
    });

    app.get('/history/:start?', (req, res) => {
        let start = 0;
        if(req.params.start && !isNaN(parseInt(req.params.start))) {
            start = parseInt(req.params.start);
        }
        res.json(history.slice(start));
    });

    app.get('/reset', (req,res)=> {
        game.reset();
        res.send(`"${req.ip}" reset the game!`);
    });
};

exports.begin = function() {
    // Begins the game!
    applyTurn();
};