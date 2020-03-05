#!/bin/bash
cd "$(dirname "$(realpath "$0")")"
GAME_ID=88888888-4444-2222-1111-000000000000
go run ./rebuild --game_id=${GAME_ID} --service_type=gameserver/master
go run ./rebuild --game_id=${GAME_ID} --service_type=gameserver/slave
go run ./rebuild --service_type=playerregistrar
go run ./rebuild --service_type=debugadmin
echo "Done!"
read -n 1 -s -r -p "Press any key to close"