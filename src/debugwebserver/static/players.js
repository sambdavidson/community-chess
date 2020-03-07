var playersModule = {
    playersMap: {},
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
    })
}

function playerCredsCallback(data) {
    console.log(data);
}