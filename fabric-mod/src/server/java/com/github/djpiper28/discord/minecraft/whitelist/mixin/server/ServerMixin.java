package com.github.djpiper28.discord.minecraft.whitelist.mixin.server;

import net.minecraft.server.dedicated.DedicatedServer;
import org.spongepowered.asm.mixin.Mixin;
import org.spongepowered.asm.mixin.injection.At;
import org.spongepowered.asm.mixin.injection.Inject;
import org.spongepowered.asm.mixin.injection.callback.CallbackInfoReturnable;

@Mixin(DedicatedServer.class)
public class ServerMixin {
    @Inject(at = @At("HEAD"), method = "initServer")
    private void init(CallbackInfoReturnable<Boolean> info) {
        // This code is injected into the start of DedicatedServer.initServer();
    }
}