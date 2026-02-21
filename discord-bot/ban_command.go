package main

import (
	"fmt"
	"log/slog"

	"github.com/Goscord/goscord/discord"
	"github.com/Goscord/goscord/discord/embed"
	"gorm.io/gorm"
)

type BanCommand struct{}

func (c *BanCommand) Name() string {
	return "mcban"
}

func (c *BanCommand) Description() string {
	return "Ban a Discord user and all their linked Minecraft accounts"
}

func (c *BanCommand) Category() string {
	return "admin"
}

func (c *BanCommand) Options() []*discord.ApplicationCommandOption {
	return []*discord.ApplicationCommandOption{
		{
			Type:        discord.ApplicationCommandOptionString,
			Name:        "user",
			Description: "The Discord user to ban",
			Required:    true,
		},
		{
			Type:        discord.ApplicationCommandOptionString,
			Name:        "reason",
			Description: "The reason for the ban",
			Required:    false,
		},
	}
}

func (c *BanCommand) Execute(ctx *Context) bool {
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
	reason := "No reason provided"
	if len(ctx.interaction.Data.Options) > 1 {
		reason = ctx.interaction.Data.Options[1].Value.(string)
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		var discordUser DiscordUser
		err := tx.Where("discord_user_id = ?", targetUserID).First(&discordUser).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				// Create the entry if it doesn't exist
				discordUser = DiscordUser{
					DiscordUserID: targetUserID,
					Banned:        true,
					BanReason:     reason,
				}
				return tx.Create(&discordUser).Error
			}
			return err
		}

		return tx.Model(&discordUser).Updates(map[string]interface{}{
			"Banned":    true,
			"BanReason": reason,
		}).Error
	})

	if err != nil {
		SendInternalError(err, ctx)
		return false
	}

	e := embed.NewEmbedBuilder()
	message := fmt.Sprintf(`**Banned User:** <@%s>
**Reason:** %s
**Banned By:** <@%s>`,
		targetUserID,
		reason,
		ctx.interaction.Member.User.Id)

	e.SetTitle("User Banned")
	e.SetDescription(message)
	ThemeEmbed(e, ctx)

	// Send response
	ctx.client.Interaction.CreateResponse(ctx.interaction.Id,
		ctx.interaction.Token,
		&discord.InteractionCallbackMessage{Embeds: []*embed.Embed{e.Embed()}})

	slog.Info("Banned user", "target ID", targetUserID, "reason", reason, "by", ctx.interaction.Member.User.Id)
	return true
}
