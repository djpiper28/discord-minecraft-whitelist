package main

import (
	"fmt"
	"log/slog"

	"github.com/Goscord/goscord/discord"
	"github.com/Goscord/goscord/discord/embed"
	"gorm.io/gorm"
)

type WhoIsCommand struct{}

func (c *WhoIsCommand) Name() string {
	return "mcwhois"
}

func (c *WhoIsCommand) Description() string {
	return "Lookup a Minecraft user name and find out who they are"
}

func (c *WhoIsCommand) Category() string {
	return "general"
}

func (c *WhoIsCommand) Options() []*discord.ApplicationCommandOption {
	return []*discord.ApplicationCommandOption{
		{
			Type:        discord.ApplicationCommandOptionString,
			Name:        "minecraft_user",
			Description: "The Minecraft user to get information about",
			Required:    true,
		},
	}
}

func (c *WhoIsCommand) Execute(ctx *Context) bool {
	if CheckGuild(ctx) != nil {
		return false
	}

	minecraftUserName := ctx.interaction.Data.Options[0].Value.(string)
	minecraftAccount, err := GetMinecraftUser(minecraftUserName)
	if err != nil {
		SendInternalError(err, ctx)
		return false
	}

	var discordMinecraftUsers []DiscordMinecraftUser
	var minecraftUser MinecraftUser
	err = db.Transaction(func(tx *gorm.DB) error {
		err := UpdateDisplayName(tx, ctx.interaction.User)
		if err != nil {
			return err
		}

		err = tx.Model(&discordMinecraftUsers).
			Where("minecraft_user_id = ?", minecraftAccount.Id).
			Find(&discordMinecraftUsers).Error

		if err != nil {
			return err
		}

		err = tx.Model(&minecraftUser).
			Where("id = ?", minecraftAccount.Id).
			First(&minecraftUser).Error

		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		SendInternalError(err, ctx)
		return false
	}

	e := embed.NewEmbedBuilder()
	message := ""

	if minecraftUser.Verified {
		message = fmt.Sprintf("Minecraft User: `%s`\nIs Owned By: <@%s> (verified owner)\nRun: `/%s %s` for more information.",
			minecraftAccount.Name,
			discordMinecraftUsers[0].DiscordUserID,
			new(PlayerInfoCommand).Name(),
			discordMinecraftUsers[0].DiscordUserID)
	} else {
		message = fmt.Sprintf("Minecraft User: `%s`\n**No verified owner**\nPotential Owners:",
			minecraftAccount.Name)

		for _, user := range discordMinecraftUsers {
			message += fmt.Sprintf("\n<@%s>", user.DiscordUserID)
		}
	}

	e.SetTitle("Minecraft Player Lookup")
	e.SetDescription(message)
	ThemeEmbed(e, ctx)

	// Send response
	ctx.client.Interaction.CreateResponse(ctx.interaction.Id,
		ctx.interaction.Token,
		&discord.InteractionCallbackMessage{Embeds: []*embed.Embed{e.Embed()},
			Flags: discord.MessageFlagUrgent})
	slog.Info("Data lookup for complete", "name", minecraftUserName)

	return true
}
