var gameserverModule = {
    gameMap: {}
};

document.addEventListener('DOMContentLoaded', function(){

    /* Game Server stuff */
    formSetup('gs-connect-form', '/games/connect');
    formSetup('gs-connection-status-form', '/games/connectionstatus');
    formSetup('gs-game-form', '/games/game');
    formSetup('gs-join-form', '/games/join');
    formSetup('gs-leave-form', '/games/leave');

    console.log('Gameserver Loaded');
});