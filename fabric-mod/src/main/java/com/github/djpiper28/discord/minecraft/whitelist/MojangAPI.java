package com.github.djpiper28.discord.minecraft.whitelist;

import com.google.gson.JsonObject;
import com.google.gson.JsonParser;

import java.io.IOException;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;

public class MojangAPI {
    private static final HttpClient client = HttpClient.newBuilder()
            .followRedirects(HttpClient.Redirect.NORMAL)
            .build();

    public static String getUuid(String username) throws IOException, InterruptedException {
        HttpRequest request = HttpRequest.newBuilder()
                .uri(URI.create("https://api.mojang.com/users/profiles/minecraft/" + username))
                .GET()
                .build();

        HttpResponse<String> response = client.send(request, HttpResponse.BodyHandlers.ofString());

        if (response.statusCode() != 200) {
            return null;
        }

        JsonObject json = JsonParser.parseString(response.body()).getAsJsonObject();
        String id = json.get("id").getAsString();

        return formatUuid(id);
    }

    private static String formatUuid(String undashed) {
        if (undashed.length() != 32) return undashed;
        return undashed.substring(0, 8) + "-" +
                undashed.substring(8, 12) + "-" +
                undashed.substring(12, 16) + "-" +
                undashed.substring(16, 20) + "-" +
                undashed.substring(20);
    }
}
