const express = require('express');
const bodyParser = require('body-parser');
const Chess = require('chess.js').Chess;

const app = express();
app.use(bodyParser.json());

const chess = new Chess();

let turnTimeout = null;
let turnLengthSeconds = 10;
let userVotes = {};
let votedMoves = {};

/**
 * Applies the current votes towards a turn.
 * Only call once, automatically sets a timeout for the next turn.
 */
function applyTurn() {
    turnTimeout = new Date();
    turnTimeout.setSeconds(turnTimeout.getSeconds() + turnLengthSeconds);

    let topMove = null;
    for(let prop in votedMoves) {
        if(votedMoves.hasOwnProperty(prop)) {
            if(topMove) {
                if(topMove.votes > votedMoves[prop].votes) {
                    topMove = votedMoves[prop];
                }
            } else {
                topMove = votedMoves[prop];
            }
        }
    }

    if(topMove) {
        chess.reset();
        chess.load_pgn(topMove.pgn);
        votedMoves = {};
        userVotes = {};
        console.log('Next turn applied.')
    }

    setTimeout(applyTurn, turnLengthSeconds * 1000);
}

app.set('port', process.env.PORT || 3001);

app.get('/gameState', (req, res) => {
    const vote = userVotes[req.ip];

    let votes = [];
    for (let pgn in votedMoves) {
        if(votedMoves.hasOwnProperty(pgn)) {
            votes.push({pgn: pgn, votes: votedMoves[pgn].votes});
        }
    }
    res.json({
        pgn: chess.pgn(),
        votePgn: vote ? vote : null,
        endTimeMs: turnTimeout.getTime(),
        votes: votes
    });
});

app.post('/vote', (req, res) => {
    console.log(req.body);
    let vote = {pgn: null};
    if(req.body && req.body.pgn) {
        vote = req.body;
        //TODO Validate vote as a real move.
        console.log(vote);
        if(userVotes[req.ip]) {
            votedMoves[userVotes[req.ip]].votes--;
        }
        userVotes[req.ip] = vote.pgn;
        if(votedMoves[vote.pgn]) {
            votedMoves[vote.pgn].votes++
        } else {
            // TODO: weird prop/pgn redundancy
            votedMoves[vote.pgn] = {votes: 1, pgn: vote.pgn};
        }
    }

    res.json({
        pgn: chess.pgn(),
        votePgn: vote.pgn,
        endTimeMs: turnTimeout.getTime(),
    });
});

app.listen(app.get("port"), () => {
    // Begins the game!
    applyTurn();
    console.log(`Find the server at: http://localhost:${app.get("port")}/`); // eslint-disable-line no-console
});