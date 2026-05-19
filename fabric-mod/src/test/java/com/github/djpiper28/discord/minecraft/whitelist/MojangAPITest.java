package com.github.djpiper28.discord.minecraft.whitelist;

import org.junit.jupiter.api.Test;
import static org.junit.jupiter.api.Assertions.*;

import java.io.IOException;

public class MojangAPITest {
    @Test
    public void testGetUuid() throws IOException, InterruptedException {
        // Notch's UUID is 069a79f4-44e9-4726-a5be-fca90e38aaf5
        String username = "Notch";
        String expectedUuid = "069a79f4-44e9-4726-a5be-fca90e38aaf5";
        
        String actualUuid = MojangAPI.getUuid(username);
        
        assertEquals(expectedUuid, actualUuid);
    }

    @Test
    public void testGetUuidInvalidUser() throws IOException, InterruptedException {
        String username = "thisuserdoesnotexist123456789";
        String actualUuid = MojangAPI.getUuid(username);
        
        assertNull(actualUuid);
    }
}
