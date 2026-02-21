package main

import (
	"fmt"
	"log/slog"

	"github.com/Goscord/goscord/discord"
	"github.com/Goscord/goscord/discord/embed"
)

type PlayerInfoCommand struct{}

func (c *PlayerInfoCommand) Name() string {
	return "mcplayer"
}

func (c *PlayerInfoCommand) Description() string {
	return "Get information about a player via their Discord account"
}

func (c *PlayerInfoCommand) Category() string {
	return "general"
}

func (c *PlayerInfoCommand) Options() []*discord.ApplicationCommandOption {
	return []*discord.ApplicationCommandOption{
		{
			Type:        discord.ApplicationCommandOptionString,
			Name:        "user",
			Description: "The Discord user to get information about",
			Required:    true,
		},
	}
}

func (c *PlayerInfoCommand) Execute(ctx *Context) bool {
	if CheckGuild(ctx) != nil {
		return false
	}

	discordId := ctx.interaction.Data.Options[0].Value.(string)

	var discordUser DiscordUser
	err := db.Model(&discordUser).
		Where("discord_user_id = ?", discordId).First(&discordUser).Error
	if err != nil {
		SendInternalError(err, ctx)
		return false
	}

	err = UpdateDisplayName(db, ctx.interaction.User)
	if err != nil {
		return false
	}

	var minecraftUsers []MinecraftUser
	err = db.Model(&minecraftUsers).
		InnerJoins("RIGHT JOIN discord_minecraft_users ON discord_minecraft_users.minecraft_user_id = minecraft_users.id").
		Where("discord_minecraft_users.discord_user_id = ?", discordId).
		Scan(&minecraftUsers).Error
	if err != nil {
		SendInternalError(err, ctx)
		return false
	}

	e := embed.NewEmbedBuilder()
	message := fmt.Sprintf("<@%s>\nBanned: %t; %s\nAdmin: %t\n",
		discordId,
		discordUser.Banned,
		discordUser.BanReason,
		discordUser.HasAdminRole)

	for _, user := range minecraftUsers {
		verificationStatus := "✅"

		if !user.Verified {
			verificationStatus = "❌"
		}
		message += fmt.Sprintf("\n%s: `%s`\tlast slogged in at <t:%d:f> at (%02f, %02f, %02f) in %s\n",
			verificationStatus,
			user.Username,
			user.LastLoginTime.Unix(),
			user.LastX,
			user.LastY,
			user.LastZ,
			user.LastDimension)
	}

	e.SetTitle("Information About Player")
	e.SetDescription(message)
	ThemeEmbed(e, ctx)

	// Send response
	ctx.client.Interaction.CreateResponse(ctx.interaction.Id,
		ctx.interaction.Token,
		&discord.InteractionCallbackMessage{Embeds: []*embed.Embed{e.Embed()},
			Flags: discord.MessageFlagUrgent})
	slog.Info("Data lookup for complete", "discord ID", discordId)

	return true
}
