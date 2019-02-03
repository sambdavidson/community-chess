const express = require('express');
const path = require('path');
const bodyParser = require('body-parser');
const uuidv1 = require('uuid/v1');

/* Server Modules */
const googleOAuth = require('./google_oath');
const game = require('./game');

const app = express();
app.use(express.static(path.join(__dirname, '..', 'build')));
app.use(bodyParser.json());

googleOAuth.initEndpoints(app);
game.initEndpoints(app);

app.set('port', process.env.PORT || 3001);

app.get('/', (req,res)=> {
    res.sendFile(path.join(__dirname, 'build', 'index.html'));
});

app.get('/id', (req, res)=> {
   res.send(uuidv1());
});


app.listen(app.get("port"), () => {
    game.begin();
    console.log(`Find the server at: http://localhost:${app.get("port")}/`); // eslint-disable-line no-console
});