document.addEventListener('DOMContentLoaded', function(){
    formSetup('create-player-form', '/players/create');
    formSetup('player-registrar-connect-form', '/players/connect');
    formSetup('player-registrar-connection-status-form', '/players/connectionStatus');

    console.log('JS Loaded');
});

function formSetup(formId, url) {
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
        })
        .catch((error) => {
            pre.innerText = JSON.stringify(error, null, 2);
            console.error('Error:', error);
        });

        return false;
    }
}
