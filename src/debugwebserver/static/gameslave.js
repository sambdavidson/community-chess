var gameslaveModule = {};
document.addEventListener('DOMContentLoaded', function(){

    /* Game Server stuff */
    formSetup('gs-connect-form', '/games/connect');
    formSetup('gs-connection-status-form', '/games/connectionStatus');
    formSetup('gs-game-form', '/games/game');

    console.log('Gameserver Loaded');
});