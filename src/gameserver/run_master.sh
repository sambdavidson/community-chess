#!/bin/bash
go run ./src/gameserver \
--slave=false \
--game_port=8080 \
--master_port=8090 \
--game_id=88888888-4444-2222-1111-000000000000 \
--debug
echo "GameServer exited! Closing in 3s."
sleep 3

