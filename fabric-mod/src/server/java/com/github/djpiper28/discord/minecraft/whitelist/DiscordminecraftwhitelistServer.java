package com.github.djpiper28.discord.minecraft.whitelist;

import io.github.cdimascio.dotenv.Dotenv;
import net.fabricmc.api.DedicatedServerModInitializer;
import net.fabricmc.fabric.api.networking.v1.ServerPlayConnectionEvents;
import net.minecraft.network.chat.Component;

import java.sql.SQLException;

public class DiscordminecraftwhitelistServer implements DedicatedServerModInitializer {

    @Override
    public void onInitializeServer() {
        // Load the .env file and, create the database pool
        Dotenv dotenv = null;

        try {
            // Nasty hack to get dotenv to behave
            final String folder = System.getProperty("user.dir");
            dotenv = Dotenv.configure()
                    .directory(folder)
                    .ignoreIfMalformed()
                    .ignoreIfMissing()
                    .load();
        } catch (RuntimeException ex) {
            Discordminecraftwhitelist.LOGGER.error("Cannot load env vars", ex);
            throw new RuntimeException(ex);
        }

        final String db_url = dotenv.get("DB_URL"),
                username = dotenv.get("DB_USERNAME"),
                password = dotenv.get("DB_PASSWORD");

        Discordminecraftwhitelist.LOGGER.info("Connecting to the database...");
        final Database database;
        try {
            database = new Database(db_url, username, password);
        } catch (SQLException e) {
            Discordminecraftwhitelist.LOGGER.error("Cannot connect to the database");
            throw new RuntimeException(e);
        }

        ServerPlayConnectionEvents.JOIN.register((handler, sender, server) -> {
            final var player = handler.getPlayer();
            final String playerName = player.getName().getString();
            Discordminecraftwhitelist.LOGGER.info("Player {} joined the server", playerName);

            String uuid;
            try {
                uuid = MojangAPI.getUuid(playerName);
                if (uuid == null) {
                    Discordminecraftwhitelist.LOGGER.warn("Could not fetch Mojang UUID for player {}, falling back to server-provided UUID", playerName);
                    uuid = player.getUUID().toString();
                }
            } catch (Exception e) {
                Discordminecraftwhitelist.LOGGER.error("Error fetching Mojang UUID for player {}", playerName, e);
                uuid = player.getUUID().toString();
            }

            try {
                final var dbPlayer = database.getUser(playerName, uuid);

                if (!dbPlayer.getVerified()) {
                    Discordminecraftwhitelist.LOGGER.info("Player {} is not verified - kicking with code", playerName);
                    player.connection.disconnect(Component.literal(String.format("To verify please use \"/mcverify %d\" in Discord", dbPlayer.getVerificationNumber())));
                    return;
                }

                if (dbPlayer.isBanned()) {
                    Discordminecraftwhitelist.LOGGER.warn("Player {} is banned - {}", playerName, dbPlayer.banReason());
                    player.connection.disconnect(Component.literal(String.format("You are banned; %s", dbPlayer.banReason())));
                    return;
                }

                Discordminecraftwhitelist.LOGGER.info("Player {} is allowed to join the server", playerName);
            } catch (UserNotFoundException e) {
                Discordminecraftwhitelist.LOGGER.info("Player {} is not added - kicking with instructions", playerName);
                player.connection.disconnect(Component.literal("You have not registered in the Discord server, use /mcadd in the Discord server."));
                return;
            } catch (SQLException e) {
                Discordminecraftwhitelist.LOGGER.info("Cannot check player {} due to an error {}", playerName, e.toString());
                player.connection.disconnect(Component.literal(String.format("Cannot connect due to internal error %s. Please try again.", e.toString())));
                return;
            }

            try {
                String ip = player.getIpAddress();
                if (ip.contains("/")) {
                    ip = ip.substring(ip.lastIndexOf("/") + 1);
                }
                if (ip.contains(":")) {
                    ip = ip.substring(0, ip.indexOf(":"));
                }

                database.updateMinecraftUserLastAccessDetails(ip,
                        player.getX(), player.getY(), player.getZ(), player.serverLevel().dimension().location().getPath(),
                        uuid);
            } catch (SQLException e) {
                Discordminecraftwhitelist.LOGGER.warn("Could not update the last access information for the player", e);
            }
        });

        Discordminecraftwhitelist.LOGGER.info("Startup complete");
    }
}