import createAuth0Client, {Auth0Client, Auth0ClientOptions} from '@auth0/auth0-spa-js';

export class Auth0Service {
    private static fetchAuthConfig: () => Promise<Response> = () => fetch(
        "auth_config.json"
    );

    public static Auth0: Promise<Auth0Client> = new Promise((resolutionFn, rejectionFn) => {
        (async () => {
            const response = await Auth0Service.fetchAuthConfig();
            const config = <Auth0ClientOptions> await response.json();
            
            const auth0 = await createAuth0Client(config);
            if (!await auth0.isAuthenticated()) {
                const query = window.location.search;
                if (query.includes("code=") && query.includes("state=")) {
                    // Process the login state
                    await auth0.handleRedirectCallback();
                
                    // Use replaceState to redirect the user away and remove the querystring parameters
                    window.history.replaceState({}, document.title, "/");
                }
            }
            resolutionFn(auth0);
        })();
    })
}

Auth0Service.Auth0.then((a)=>{console.log("loader", a)}); // Async loader