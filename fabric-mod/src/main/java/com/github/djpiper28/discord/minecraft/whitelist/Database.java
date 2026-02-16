package com.github.djpiper28.discord.minecraft.whitelist;

import org.apache.commons.dbcp2.BasicDataSource;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import java.net.InetAddress;
import java.sql.*;
import java.util.concurrent.atomic.AtomicReference;

/**
 * Accusations that this class comes from my year 2 group project are correct. I cannot be arsed to rewrite it.
 *
 * @author Danny
 * @version 3
 */
public class Database {
    private static final Logger log = LoggerFactory.getLogger(Database.class);
    private final BasicDataSource ds = new BasicDataSource();
    private final String url;
    private final String username;
    private final String password;

    /**
     * Construct a database connector with given login details
     *
     * @param url      the url (with database name) of the database
     * @param username the username for the database
     * @param password the password for the database
     * @since 1
     */
    public Database(String url, String username, String password) throws SQLException {
        this.url = url;
        this.username = username;
        this.password = password;

        // Setup pool
        this.ds.setUrl(this.url);
        this.ds.setUsername(this.username);
        this.ds.setPassword(this.password);
        this.ds.setMinIdle(20);
        this.ds.setMaxIdle(100);
        this.ds.setMaxOpenPreparedStatements(100);
        this.ds.setMaxTotal(500);
    }

    /**
     * Runs code on the database then closes the connection, even if there is a connection
     *
     * @param runnable the code to run
     * @return whether an error occurred during execution
     * @throws SQLException an sql exception that was thrown whilst connecting
     * @since 1
     */
    public boolean runOnDatabase(DatabaseRunnable runnable) throws SQLException {
        try {
            Class.forName("org.postgresql.Driver");
        } catch (ClassNotFoundException e) {
            throw new RuntimeException(e);
        }

        Connection conn = null;
        RuntimeException ex = null;

        try {
            conn = DriverManager.getConnection(this.url, this.username, this.password);
            conn.setAutoCommit(true);
            runnable.run(conn);
        } catch (RuntimeException e) {
            log.error(e.getLocalizedMessage());
            ex = e;
        } finally {
            if (conn != null) {
                conn.close();
            }
        }

        return ex != null;
    }

    public void updateMinecraftUserLastAccessDetails(InetAddress ipaddr, double x, double y, double z, String id) throws SQLException {
        AtomicReference<SQLException> ex = new AtomicReference<>();
        this.runOnDatabase((conn) -> {
            try {
                PreparedStatement updateStatement = conn.prepareStatement("UPDATE minecraft_users " +
                    "SET last_x = ?, last_y = ?, last_z = ?, last_login_time = CURRENT_TIMESTAMP(2) " +
                    "WHERE id = ?;");
                updateStatement.setDouble(1, x);
                updateStatement.setDouble(2, y);
                updateStatement.setDouble(3, z);
                updateStatement.setString(4, id);

                updateStatement.executeUpdate();
            } catch (SQLException e) {
                ex.set(e);
            }
        });

        if (ex.get() != null) {
            throw ex.get();
        }
    }

    /**
     * Returns the user object for a given Minecraft username
     *
     * @param username Minecraft username to lookup
     * @return the user object
     * @throws SQLException thrown if any SQL errors occur when trying to get the user
     */
    public MinecraftUser getUser(String username, String id) throws SQLException, UserNotFoundException {
        AtomicReference<MinecraftUser> ret = new AtomicReference<>(null);
        AtomicReference<SQLException> ex = new AtomicReference<>(null);
        AtomicReference<UserNotFoundException> ex2 = new AtomicReference<>(null);

        this.runOnDatabase((conn -> {
            try {
                conn.setAutoCommit(false);
                PreparedStatement getMinecraftUserPs = conn.prepareStatement("SELECT * FROM minecraft_users WHERE id = ?;");
                PreparedStatement updateMinecraftUsernameCache = conn.prepareStatement("UPDATE minecraft_users SET username = ? WHERE id = ?;");
                PreparedStatement getBannedStatus = conn.prepareStatement("SELECT discord_users.banned " +
                    "FROM discord_users " +
                    "FULL OUTER JOIN discord_minecraft_users ON discord_users.discord_user_id = discord_minecraft_users.discord_user_id " +
                    "WHERE discord_minecraft_users.minecraft_user_id = ?;");

                getMinecraftUserPs.setString(1, id);
                ResultSet res = getMinecraftUserPs.executeQuery();
                if (!res.next()) {
                    throw new UserNotFoundException();
                }

                getBannedStatus.setString(1, id);
                ResultSet bannedStatus = getBannedStatus.executeQuery();
                boolean banned = false;
                if (bannedStatus.next()) {
                    banned = bannedStatus.getBoolean(1);
                }

                final MinecraftUser user = new MinecraftUser(res.getString("id"),
                    res.getString("username"),
                    res.getInt("verification_number"),
                    banned,
                    res.getBoolean("verified"));

                if (!user.getUsername().equals(username)) {
                    System.out.println("Updating the username in the cache");

                    updateMinecraftUsernameCache.setString(1, username);
                    updateMinecraftUsernameCache.setString(2, id);
                    updateMinecraftUsernameCache.executeUpdate();
                }

                ret.set(user);
                conn.commit();
            } catch (SQLException e) {
                ex.set(e);
            } catch (UserNotFoundException e) {
                ex2.set(e);
            }
        }));

        if (ex2.get() != null) {
            throw ex2.get();
        }

        if (ex.get() != null) {
            throw ex.get();
        }
        return ret.get();
    }
}
