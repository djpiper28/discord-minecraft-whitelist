package main

import (
	"fmt"
	"log/slog"

	"github.com/Goscord/goscord/discord"
	"github.com/Goscord/goscord/discord/embed"
	"gorm.io/gorm"
)

type UnbanCommand struct{}

func (c *UnbanCommand) Name() string {
	return "mcunban"
}

func (c *UnbanCommand) Description() string {
	return "Unban a Discord user"
}

func (c *UnbanCommand) Category() string {
	return "admin"
}

func (c *UnbanCommand) Options() []*discord.ApplicationCommandOption {
	return []*discord.ApplicationCommandOption{
		{
			Type:        discord.ApplicationCommandOptionString,
			Name:        "user",
			Description: "The Discord user to unban",
			Required:    true,
		},
	}
}

func (c *UnbanCommand) Execute(ctx *Context) bool {
	if CheckGuild(ctx) != nil {
		return false
	}

	// Get guild settings to check for admin role
	guildid := ctx.interaction.GuildId
	var gs GuildSettings
	err := db.First(&gs, guildid).Error
	if err != nil {
		SendInternalError(err, ctx)
		return false
	}

	if !UserIsAdmin(gs, ctx.interaction.Member) && ctx.interaction.Member.Permissions&discord.BitwisePermissionFlagAdministrator == 0 {
		SendAdminPermissionsError(gs, ctx)
		return false
	}

	targetUserID := ctx.interaction.Data.Options[0].Value.(string)

	err = db.Transaction(func(tx *gorm.DB) error {
		var discordUser DiscordUser
		err := tx.Where("discord_user_id = ?", targetUserID).First(&discordUser).Error
		if err != nil {
			return err
		}

		return tx.Model(&discordUser).Updates(map[string]interface{}{
			"Banned":    false,
			"BanReason": "",
		}).Error
	})

	if err != nil {
		SendInternalError(err, ctx)
		return false
	}

	e := embed.NewEmbedBuilder()
	message := fmt.Sprintf(`**Unbanned User:** <@%s>
**Unbanned By:** <@%s>`,
		targetUserID,
		ctx.interaction.Member.User.Id)

	e.SetTitle("User Unbanned")
	e.SetDescription(message)
	ThemeEmbed(e, ctx)

	// Send response
	ctx.client.Interaction.CreateResponse(ctx.interaction.Id,
		ctx.interaction.Token,
		&discord.InteractionCallbackMessage{Embeds: []*embed.Embed{e.Embed()}})

	slog.Info("Unbanned user", "target ID", targetUserID, "by", ctx.interaction.Member.User.Id)
	return true
}
