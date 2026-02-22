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
            Discordminecraftwhitelist.LOGGER.error(".env file is missing, see ../README.md");
            throw new RuntimeException(e);
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
            Discordminecraftwhitelist.LOGGER.info("Player {} joined the server", player.getName().getString());

            try {
                final var dbPlayer = database.getUser(player.getName().getString(), player.getId() + "");

                if (!dbPlayer.getVerified()) {
                    Discordminecraftwhitelist.LOGGER.info("Player {} is not verified - kicking with code", player.getName().getString());
                    player.connection.disconnect(Component.literal(String.format("To verify please use \"/mcverify %d\" in Discord", dbPlayer.getVerificationNumber())));
                    return;
                }

                if (dbPlayer.isBanned()) {
                    Discordminecraftwhitelist.LOGGER.warn("Player {} is banned - {}", player.getName().getString(), dbPlayer.banReason());
                    player.connection.disconnect(Component.literal(String.format("You are banned; %s", dbPlayer.banReason())));
                    return;
                }

                Discordminecraftwhitelist.LOGGER.info("Player {} is allowed to join the server", player.getName().getString());
            } catch (UserNotFoundException e) {
                Discordminecraftwhitelist.LOGGER.info("Player {} is not added - kicking with instructions", player.getName().getString());
                player.connection.disconnect(Component.literal("You have not registered in the Discord server, us /mcadd in the Discord server."));
            } catch (SQLException e) {
                Discordminecraftwhitelist.LOGGER.info("Cannot check player {} due to an error {}", player.getName().getString(), e.toString());
                player.connection.disconnect(Component.literal(String.format("Cannot connect due to internal error %s. Please try again.", e.toString())));
            }
        });

        Discordminecraftwhitelist.LOGGER.info("Startup complete");
    }
}