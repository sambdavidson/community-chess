let knownPlayers = {};

document.addEventListener('DOMContentLoaded', function(){
    /* Player Stuff */
    formSetup('player-registrar-connect-form', '/players/connect');
    formSetup('player-registrar-connection-status-form', '/players/connectionStatus');
    formSetup('create-player-form', '/players/create', playerInfoCallback);
    formSetup('get-player-form', '/players/get', playerInfoCallback);
    formSetup('login-form', 'players/login', playerCredsCallback);

    /* Game Server stuff */
    formSetup('gs-connect-form', '/games/connect');
    formSetup('gs-connection-status-form', '/games/connectionStatus');
    formSetup('gs-game-form', '/games/game');

    setVisible('players');
    console.log('JS Loaded');
});

function setVisible(divId) {
    let root = document.getElementById("sections");
    Array.from(root.children).forEach((el) => {
        el.hidden = true;
    });
    let e = document.getElementById(divId);
    e.hidden = false;
}

function formSetup(formId, url, dataCallback) {
    /** @type {HTMLFormElement} */
    let form = document.getElementById(formId);
    /** @type {HTMLPreElement} */
    let pre = form.querySelector('.output');

    if (form.attachEvent) {
        form.attachEvent('submit', processForm);
    } else {
        form.addEventListener('submit', processForm);
    }

    /**
     * 
     * @param {Event} e 
     */
    function processForm(e) {
        if (e.preventDefault) {
            e.preventDefault();
        }

        const formData = new FormData(form);

        fetch(url, {
            method: 'POST',
            body: formData,
        })
        .then((response) => response.json())
        .then((data) => {
            pre.innerText = JSON.stringify(data, null, 2);
            if (typeof dataCallback === 'function') {
                dataCallback(data);
            }
        })
        .catch((error) => {
            pre.innerText = JSON.stringify(error, null, 2);
            console.error('Error:', error);
        });

        return false;
    }
}

function playerInfoCallback(data) {
    knownPlayers[data.player.id] = data.player;
    
    /** @type {HTMLSelectElement} */
    let dropdown = document.getElementById('get-player-known-player-dropdown');
    /** @type {HTMLInputElement} */
    let input = document.getElementById('get-player-uuid');
    dropdown.onchange = () => {
        input.value = dropdown.value;
    }
    dropdown.addEventListener('click', (e) => {
        dropdown.innerHTML = '';
        for (const [key, value] of Object.entries(knownPlayers)) {
            let o = document.createElement('option');
            o.innerText = `${value.username}:${value.number_suffix}`;
            o.value = value.id;
            dropdown.appendChild(o);
        }
    })
}

function playerCredsCallback(data) {
    console.log(data);
}