# minecraft-server

mc server bot and, plugin to allow for a compsoc whitelist

## Minecraft Server

(The same for Fabric and the deprecated Paper plugin).

Create a file called `.env` in your server directory with the following environment variables defined:

| Env Var     | Description                              |
| ----------- | ---------------------------------------- |
| DB_URL      | Database URL in for of JDBC Postgres URL |
| DB_USERNAME | Database username                        |
| DB_PASSWORD | Database password                        |

### Example `.env` file

```bash
DB_URL=jdbc:postgresql://localhost/database
DB_USERNAME=user
DB_PASSWORD=pass
```

> Note that it will not work with environment variables.

## Discord Bot

Env vars:

| Env Var          | Description                                                                                                                 |
| ---------------- | --------------------------------------------------------------------------------------------------------------------------- |
| DISCORD_TOKEN    | Discord bot token, go to https://discord.com/developers/applications                                                        |
| DISCORD_GUILD_ID | Discord guild id, go to https://support.discord.com/hc/en-us/articles/206346498-Where-can-I-find-my-User-Server-Message-ID- |
| DATABASE_URL     | Gorm.io database url, go to https://gorm.io/docs/connecting_to_the_database.html                                            |
| MINECRAFT_IP     | Public IP for your server i.e: 1.2.3.4:25565                                                                                |

Health check server is on `:8080`

## User flow

### 1 Register User

The user will use the mcadd Discord command: `/mcadd <minecraft username>`.

### 2 Verify User

The user joins the Miencraft Server, they will be kicked with a message showing a code.

Example Message:

```
You need to verify your account, use /mcverify 123456
```

### 3 User Verifies Themself

The user uses `/mcverify 123456` on Discord to verify themselves.

### 4 The User Can Now Use The Server

They should have no more issues using the server.

## Banning A User

On Minecraft, or Discord a user can be `/mcban`ed, the user is then kicked and not allowed to join
the Discord user and all of their alt accounts will be tagged as banned.

## Unbanning A User

Users can be unbanned by `/mcunban`ing someone on the Discord Bot, they will have all alt accounts freed up.

## Users Leaving Discord Server TODO

When a user leaves the discord server, they will be kicked from the Minecraft server, and have all their records deleted (GDPR or something like that I guess).

## Logging

Users will have the following data logged:

- Last join IP
- Last coordinates (x, y, z)
- Last skin image
- Last chunk (top down image)

## All "Administrators" are given OP TODO

People with the configurable admin role, are given OP powers in minecraft.

## Only people with the "Access" role can join the Server TODO

People with the configurable access role, are allowed to join the server.

## Check Current State On Startup (Discord Bot) TODO

Checks the OP list, and players list.
