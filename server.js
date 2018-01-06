const express = require('express');
const path = require('path');
const bodyParser = require('body-parser');
const uuidv1 = require('uuid/v1');
const Chess = require('./node_modules/chess.js/chess').Chess;

const app = express();
app.use(express.static(path.join(__dirname, 'build')));
app.use(bodyParser.json());

const game = new Chess();

let turnTimeout = null;
let turnLengthSeconds = 20;
let userVotes = {};
let votedMoves = {};
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
    for (let prop in votedMoves) {
        if (votedMoves.hasOwnProperty(prop)) {
            randomVoteSubtractor -= (votedMoves[prop].votes/totalVotes);
            if(randomVoteSubtractor < 0) {
                topMove = votedMoves[prop];
                break;
            }
        }
    }

    if(topMove) {
        game.reset();
        game.load_pgn(topMove.pgn);
        votedMoves = {};
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

app.set('port', process.env.PORT || 3001);

app.get('/gameState/:id', (req, res) => {
    const userVote = userVotes[req.params.id];

    let votes = [];
    for (let pgn in votedMoves) {
        if (votedMoves.hasOwnProperty(pgn)) {
            votes.push({pgn: pgn, votes: votedMoves[pgn].votes});
        }
    }
    res.json({
        pgn: game.pgn(),
        votePgn: userVote ? userVote : null,
        endTimeMs: turnTimeout.getTime(),
        votes: votes
    });
});

app.post('/vote', (req, res) => {

    let vote = {pgn: null};
    if(req.body && req.body.pgn && req.body.id && validateMove(req.body.pgn)) {
        const id = req.body.id;
        vote = req.body;
        if(userVotes[id]) {
            votedMoves[userVotes[id]].votes--;
            totalVotes--;
        }
        userVotes[id] = vote.pgn;
        if(votedMoves[vote.pgn]) {
            votedMoves[vote.pgn].votes++
        } else {
            // TODO: weird prop/pgn redundancy
            votedMoves[vote.pgn] = {votes: 1, pgn: vote.pgn};
        }
        totalVotes++;
    }
    res.status(200).end();
});

app.get('/', (req,res)=> {
    res.sendFile(path.join(__dirname, 'build', 'index.html'));
});

app.get('/id', (req, res)=> {
   res.send(uuidv1());
});

app.get('/reset', (req,res)=> {
    game.reset();
    res.send(`"${req.ip}" reset the game!`);
});


app.listen(app.get("port"), () => {
    // Begins the game!
    applyTurn();
    console.log(`Find the server at: http://localhost:${app.get("port")}/`); // eslint-disable-line no-console
});