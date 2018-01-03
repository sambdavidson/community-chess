import Rx from 'rxjs';

class API {
    /**
     * Constructor for an API service.
     * This performs the ASYNC calls to the API server for the initial state.
     *
     */
    constructor() {

        /** @private {Observer}*/
        this.gameStateEm = null;

        //TODO: Clean this up so it doesn't happen multiple times on subscriptions.
        this.gameState =  Rx.Observable.interval(1000).switchMap(() => {
            return Rx.Observable.ajax({url: 'gameState', responseType: 'json', method: 'GET'});
        }).pluck('response').share();
    }

    castVote(id, pgn) {
        // ID unused for now. IP is used as the ID.

        const data = new FormData();
        data.append('json', JSON.stringify({pgn: pgn}));
        fetch('vote', {
            method: 'POST',
            accept: 'application/json',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({pgn: pgn})
        }).then((response) => {
            if(!response.ok){
                console.error(response);
            }
        });
    }



}

let apiInstance = new API();

export default apiInstance;