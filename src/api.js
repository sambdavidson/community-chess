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

        this.id = null;

        fetch('id', {
            method: 'GET',
            accept: 'application/json'
        }).then((response)=>{
            if(response.ok) {
                return response.text();
            }
        }).then((text)=>{
            this.id = text;
        });

        //TODO: Clean this up so it doesn't happen multiple times on subscriptions.
        this.gameState =  Rx.Observable.interval(1000).startWith(0).switchMap(() => {
            const id = this.id ? this.id : '';
            return Rx.Observable.ajax({url: `gameState/${this.id}`, responseType: 'json', method: 'GET'});
        }).pluck('response').share();
    }

    castVote(pgn) {
        if(!this.id) {
            return;
        }

        const data = new FormData();
        data.append('json', JSON.stringify({pgn: pgn}));
        fetch('vote', {
            method: 'POST',
            accept: 'application/json',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                id: this.id,
                pgn: pgn
            })
        }).then((response) => {
            if(!response.ok){
                console.error(response);
            }
        });
    }



}

let apiInstance = new API();

export default apiInstance;