#!/bin/bash
cd "$(dirname "$(realpath "$0")")"
GAME_ID=88888888-4444-2222-1111-000000000000
go run ./rebuild --game_id=${GAME_ID} --service_type=gameserver/master
go run ./rebuild --game_id=${GAME_ID} --service_type=gameserver/slave
go run ./rebuild --service_type=playerregistrar
echo "Done! Closing in 5s."
sleep 5