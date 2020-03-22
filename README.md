# Community Chess

![CommunityChessLogo](media/images/community-chess.png)

## Description

Vote on chess moves with people from all over the world. Specifically this is Chess layered on a generic framework for casting asynchronous moves in a given game.

## Building and Running

NOTE: Make sure you have Docker (and Docker Compose) installed

1. CD to root of the git repo e.g. `github.com/sambdavidson/community-chess`

2. Build the base protos and builder image
```bash
./src/rebuild_bash.sh
```
3. Build and run all microservices with Docker Compose
```bash
docker-compose up --build
```

4. Optional, connect to the debug CLI hosted at `localhost:8080` in a web browser.


### Running individually

After Step 2. Previously.

To run the PlayerRegistrar:

```bash
docker-compose up --build playerregistrar
```

To run the GameServer Master (MAIN):

```bash
docker-compose up --build gameserver_master
```

To run the GameServer Slave (OPTIONAL):

```bash
docker-compose up --build gameserver_slave
```

To run the debugwebserver:

```bash
docker-compose up --build debugwebserver
```


## API

### [V1](docs/v1/api.md)

