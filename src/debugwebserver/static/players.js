var playersModule = {
    playersMap: {},
    simpleNamesToLoginTokens: {},
};

document.addEventListener('DOMContentLoaded', function(){
    /* Player Stuff */
    formSetup('player-registrar-connect-form', '/players/connect');
    formSetup('player-registrar-connection-status-form', '/players/connectionstatus');
    formSetup('create-player-form', '/players/create', playerInfoCallback);
    formSetup('get-player-form', '/players/get', playerInfoCallback);
    formSetup('login-form', 'players/login', playerCredsCallback);

    console.log('Players Loaded');
});

function playerInfoCallback(data) {
    if (!data || !data.player) {
        return;
    }
    playersModule.playersMap[data.player.id] = data.player;
    
    /** @type {HTMLSelectElement} */
    let dropdown = document.getElementById('get-player-known-player-dropdown');
    /** @type {HTMLInputElement} */
    let input = document.getElementById('get-player-uuid');
    dropdown.onchange = () => {
        input.value = dropdown.value;
    }
    dropdown.addEventListener('click', (e) => {
        dropdown.innerHTML = '';
        for (const [key, value] of Object.entries(playersModule.playersMap)) {
            let o = document.createElement('option');
            o.innerText = `${value.username}:${value.number_suffix}`;
            o.value = value.id;
            dropdown.appendChild(o);
        }
    });
}

/**
 * 
 * @param {any} data 
 * @param {FormData} formData 
 */
function playerCredsCallback(data, formData) {
    if (!data || !data.token) {
        return;
    }
    let simepleName = `${formData.get('login-username')}:${formData.get('login-number-suffix')}`;
    let existing = !!playersModule.simpleNamesToLoginTokens[simepleName];
    playersModule.simpleNamesToLoginTokens[simepleName] = data.token;
    if (!existing) {
        populateActiveTokenSelect();
    }
    setActivePlayerToSimpleName(simepleName);
    updatePlayerTokenHiddens();
}

function populateActiveTokenSelect() {
    /** @type {HTMLSelectElement} */
    let sel = document.getElementById('active-player-select');
    while(sel.firstChild) {
        sel.removeChild(sel.firstChild);
    }
    for (const [key, value] of Object.entries(playersModule.simpleNamesToLoginTokens)) {
        let o = document.createElement('option');
        o.innerText = key;
        o.value = value;
        sel.appendChild(o);
    }
}

function setActivePlayerToSimpleName(name) {
    /** @type {HTMLSelectElement} */
    let sel = document.getElementById('active-player-select');
    let opts = sel.options
    for (let opt, j = 0; opt = opts[j]; j++) {
        if (opt.value == name) {
            sel.selectedIndex = j;
            break;
        }
    }
}

function activePlayerToken() {
    /** @type {HTMLSelectElement} */
    let sel = document.getElementById('active-player-select');
    return sel.options[sel.selectedIndex].value;
}

function updatePlayerTokenHiddens() {
    let t = activePlayerToken();
    for (let el of document.getElementsByClassName('player-token')) {
        el.value = t;
    }
}