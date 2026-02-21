package main

import (
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"slices"
	"time"

	"github.com/Goscord/goscord"
	"github.com/Goscord/goscord/discord"
	"github.com/Goscord/goscord/gateway"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

func main() {
	fmt.Printf(" -> Environment information: \"%s\"\n", runtime.Version())
	fmt.Println("Please send above data in any bug reports or support queries.")

	err := godotenv.Load()
	if err != nil {
		slog.Info("Cannot load a .env file, using normal env vars instead", err)
	}

	// Setup database
	databaseUrl := os.Getenv("DATABASE_URL")
	if databaseUrl == "" {
		slog.Error("The DATABASE_URL has not been set.")
		os.Exit(1)
	}

	db, err = gorm.Open(postgres.Open(databaseUrl), &gorm.Config{}) // *gorm.DB
	if err != nil {
		slog.Error("Cannot open database", "err", err)
		os.Exit(1)
	}

	sqlDB, err := db.DB()
	if err != nil {
		slog.Error("Cannot use database", "err", err)
		os.Exit(1)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetConnMaxLifetime(time.Hour)

	slog.Info("Migrating Database")
	AutoMigrateModel()

	// Setup commands map
	commands := make(map[string]Command)

	// Add all commands here:
	commandsList := []Command{
		new(PlayerInfoCommand),
		new(WhoIsCommand),
		new(SetupCommand),
		new(AddAccountCommand),
		new(VerifyCommand),
		new(BanCommand),
		new(UnbanCommand),
	}

	// Create client instance
	client := goscord.New(&gateway.Options{
		Token:   os.Getenv("DISCORD_TOKEN"),
		Intents: gateway.IntentGuilds | gateway.IntentGuildMembers,
	})

	// Setup events
	err = client.On("ready", func() {
		go func() {
			cmdNames := make([]string, 0)
			slog.Info("Finding slash commands")
			cmds, err := client.Application.GetCommands(client.Me().Id, os.Getenv(DISCORD_GUILD_ID))
			if err != nil {
				slog.Error("Cannot find slash commands", "err", err)
			} else {
				for _, cmd := range cmds {
					cmdNames = append(cmdNames, cmd.Name)
				}
			}

			slog.Info("Registering slash commands")
			for _, command := range commandsList {
				if !slices.Contains(cmdNames, command.Name()) {
					Register(command, client, commands)
					time.Sleep(time.Second * 2)
				}
			}
			slog.Info("All commands registered")
		}()

		slog.Info("Setting activity")
		err = client.SetActivity(&discord.Activity{Name: os.Getenv("MINECRAFT_IP"), Type: discord.ActivityListening})
		if err != nil {
			slog.Warn("Cannot set activity", "err", err)
		}

		// err = db.Transaction(func(tx *gorm.DB) error {
		// 	slog.Info("Updating all players in current guilds")
		// 	var discordUsers []DiscordUser
		// 	err = tx.Model(&discordUsers).Find(&discordUsers).Error
		// 	if err != nil {
		// 		slog.Warn("Cannot get all users", "err", err)
		// 		return err
		// 	}
		//
		// 	members, found := client.State().Members()[os.Getenv(DISCORD_GUILD_ID)]
		// 	if !found {
		// 		slog.Warn("Cannot find guild")
		// 	}
		//
		// 	for id := range members {
		// 		for i, existingUser := range discordUsers {
		// 			if existingUser.DiscordUserID == id {
		// 				discordUsers = slices.Delete(discordUsers, i, i)
		// 				break
		// 			}
		// 		}
		// 	}
		//
		// 	for _, deletedUser := range discordUsers {
		// 		err := tx.Model(&deletedUser).UpdateColumns(map[string]any{
		// 			"banned":     true,
		// 			"ban_reason": "Not in server",
		// 		}).Error
		// 		if err != nil {
		// 			slog.Warn("Could not ban user for not being in the guild", "discord ID", deletedUser.DiscordUserID, "name", deletedUser.DisplayName, "err", err)
		// 		} else {
		// 			slog.Info("Banned user for not being in the guild", "discord ID", deletedUser.DiscordUserID, "name", deletedUser.DisplayName)
		// 		}
		// 	}
		//
		// 	return nil
		// })
		// if err != nil {
		// 	slog.Warn("Could not check members on startup", "err", err)
		// }
	})
	if err != nil {
		slog.Error("Cannot execute onReady hook", "err", err)
		os.Exit(1)
	}

	err = client.On("interactionCreate", func(interaction *discord.Interaction) {
		if interaction.Member == nil {
			return
		}

		if interaction.Member.User.Bot {
			return
		}

		cmd := commands[interaction.Data.Name]

		if cmd != nil {
			success := cmd.Execute(&Context{client: client, interaction: interaction})
			if !success {
				slog.Info("Failed to run command", "name", cmd.Name())
			}
		}
	})
	if err != nil {
		slog.Error("Cannot run interactionCreat hook", "err", err)
		os.Exit(1)
	}

	// Login client
	if err := client.Login(); err != nil {
		slog.Error("Cannot login client", "err", err)
		os.Exit(1)
	}

	// Keep bot running
	slog.Info("Bot started")

	go HealthCheckServer()
	go UpdateThread(client)
	select {}
}
