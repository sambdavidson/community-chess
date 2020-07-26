import * as React from "react";
import { Auth0Service } from "../services/Auth0Service";
import {IdToken} from '@auth0/auth0-spa-js';

// 'LoginProps' describes the shape of props.
// State is never set so we use the '{}' type.

interface LoginState {
    authenticated: boolean;
    user: IdToken;

}

export class Login extends React.Component<{}, LoginState> {

    constructor(props: {}) {
        super(props);
        this.state = {
            authenticated: false,
            user: null,
        };
        (async ()=> {
            const auth0 = await Auth0Service.Auth0;
            this.setState({
                authenticated: await auth0.isAuthenticated(),
                user: await auth0.getIdTokenClaims(),
                
            })
        })();
    }

    private async login() {
        await (await Auth0Service.Auth0).loginWithRedirect({
            redirect_uri: window.location.origin, 
        });
    }

    private async logout() {
        (await Auth0Service.Auth0).logout({
            returnTo: window.location.origin
          });
    }

    private async account() {
        console.log('todo account')
    }

    render() {
        return <div>
            <pre>{JSON.stringify(this.state.user, null, 2)}</pre>
            <pre>{}</pre>
            <p>
                <button disabled={this.state.authenticated} 
                    onClick={this.login}>Log in</button>
                <button disabled={!this.state.authenticated}
                    onClick={this.logout}>Log out</button>
                <button disabled={!this.state.authenticated}
                    onClick={this.account}>Account</button>
            </p>
        </div>;
    }
}