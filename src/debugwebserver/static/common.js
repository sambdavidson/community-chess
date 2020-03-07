let knownPlayers = {};

document.addEventListener('DOMContentLoaded', function(){
    setVisible('players');
    console.log('Common Loaded');
});

/**
 * 
 * @param {string} divId is the id of the div to set visible
 */
function setVisible(divId) {
    let root = document.getElementById("sections");
    Array.from(root.children).forEach((el) => {
        el.hidden = true;
    });
    let e = document.getElementById(divId);
    e.hidden = false;
}

/**
 * 
 * @param {string} formId ID of the form element
 * @param {string} url URL to call on server
 * @param {function(any)} dataCallback callback function that will be called on 2XX response code with JSON of body.
 */
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