package com.github.djpiper28.discord.minecraft.whitelist;

import net.fabricmc.api.DedicatedServerModInitializer;
import net.fabricmc.fabric.api.networking.v1.ServerPlayConnectionEvents;

public class DiscordminecraftwhitelistServer implements DedicatedServerModInitializer {

    @Override
    public void onInitializeServer() {
        ServerPlayConnectionEvents.JOIN.register((handler, sender, server) -> {
            Discordminecraftwhitelist.LOGGER.info("Player {} joined the server", handler.getPlayer().getName().getString());
        });
    }
}