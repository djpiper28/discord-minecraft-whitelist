# These can be changed in the .env file
DISCORD_TOKEN?=CHANGEME
DISCORD_GUILD_ID?=CHANGEME
include .env

MINECRAFT_IP=localhost:25525

CONTAINER_NAME=postgres
POSTGRES_PASSWORD=test
POSTGRES_DB=postgres
POSTGRES_USER=postgres
DATABASE_URL="host=localhost user=$(POSTGRES_USER) password=$(POSTGRES_PASSWORD) dbname=$(POSTGRES_DB) port=5432 sslmode=disable TimeZone=Europe/London"

reset_postgres:
	docker rm $(CONTAINER_NAME)
	$(MAKE) postgres

postgres:
	docker run --name $(CONTAINER_NAME) -e POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) -p 5432:5432 -d postgres

bot_build:
	cd discord-bot && go build

bot: bot_build
	DISCORD_GUILD_ID=$(DISCORD_GUILD_ID) DISCORD_TOKEN=$(DISCORD_TOKEN) DATABASE_URL=$(DATABASE_URL) ./discord-bot/minecraft-discord-bot

default:
	$(MAKE) reset_postgres || true
	$(MAKE) postgres
	$(MAKE) bot
