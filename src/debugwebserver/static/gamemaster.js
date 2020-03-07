var gameslaveMaster = {};
document.addEventListener('DOMContentLoaded', function(){

    /* Game Server stuff */
    formSetup('gm-connect-form', '/gamemaster/connect');
    formSetup('gm-connection-status-form', '/gamemaster/connectionstatus');
    formSetup('gm-initialize', '/gamemaster/initialize')

    console.log('Gamemaster Loaded');
});