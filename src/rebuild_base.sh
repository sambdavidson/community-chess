#!/bin/bash
cd "$(dirname "$(realpath "$0")")" # change PWD to here.
REPOSRC=//github.com/sambdavidson/community-chess/src # Opening double slash fix for Windows Git Bash bug.
docker build ./proto/ -t proto:latest --build-arg REPOSRC=${REPOSRC}
docker build . -f ./gosrcbase.Dockerfile -t gosrcbase:latest --build-arg REPOSRC=${REPOSRC}
echo "Done!"
read -n 1 -s -r -p "Press any key to close"