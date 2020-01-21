#!/bin/bash
go run ./src/gameserver \
--slave=false \
--game_port=8080 \
--master_port=8070 \
--game_id=88888888-4444-2222-1111-000000000000 \
--ca_bundle_path=devsecrets/certs/devCACert.pem \
--master_cert_path=devsecrets/certs/gameserver/masterCert.pem \
--master_private_key_path=devsecrets/certs/gameserver/masterPrivateKey.pem \
--debug
echo "GameServer exited! Closing in 3s."
sleep 3

