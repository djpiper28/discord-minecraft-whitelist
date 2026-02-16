package com.github.djpiper28.discord.minecraft.whitelist;

public class MinecraftUser {
    private final String id;
    private final String username;
    private final int verificationNumber;
    private final boolean banned;
    private final boolean verified;

    public MinecraftUser(final String id, final String username, final int verificationNumber, final boolean banned, final boolean verified) {
        this.id = id;
        this.username = username;
        this.verificationNumber = verificationNumber;
        this.banned = banned;
        this.verified = verified;
    }

    public String getId() {
        return this.id;
    }

    public String getUsername() {
        return username;
    }

    public int getVerificationNumber() {
        return verificationNumber;
    }

    public boolean isBanned() {
        return banned;
    }

    public boolean getVerified() {
        return verified;
    }
}
