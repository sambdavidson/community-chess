/**
 * Sets up Google OAuth2 endpoints on the express server.
 */
const fs = require('fs');
const path = require('path');
const {google} = require('googleapis');

const redirectURL = 'googleoauth2callback';
const permissionScopes = [
    'https://www.googleapis.com/auth/userinfo.email',
    'https://www.googleapis.com/auth/userinfo.profile'
];

/* Load secrets */
let secrets = null;
try {
    secrets = JSON.parse(fs.readFileSync(path.join(__dirname, '..', 'secrets', 'google_oauth.json')));
} catch(e) {
    console.error('Error loading Google OAuth2 secrets. Ensure they are properly configured.', e);
}

exports = module.exports;

exports.initEndpoints = function(app) {
    if (secrets === null) {
        return;
    }

    const oauth2Client = new google.auth.OAuth2(secrets.CLIENT_ID, secrets.CLIENT_SECRET, redirectURL);

    const url = oauth2Client.generateAuthUrl({
        access_type: 'offline',
        scopes: permissionScopes
    });
    console.log('init google oath');
    console.log(url);
};