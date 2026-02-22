package main

import (
	"fmt"
	"log/slog"

	"github.com/Goscord/goscord/discord"
	"github.com/Goscord/goscord/discord/embed"
	"gorm.io/gorm"
)

type SetupCommand struct{}

func (c *SetupCommand) Name() string {
	return "mcsetup"
}

func (c *SetupCommand) Description() string {
	return "Setup the Minecraft whitelist Discord bot."
}

func (c *SetupCommand) Category() string {
	return "general"
}

func (c *SetupCommand) Options() []*discord.ApplicationCommandOption {
	return []*discord.ApplicationCommandOption{
		{
			Type:        discord.ApplicationCommandOptionString, //Role,
			Name:        "admin",
			Description: "The role for the administrators of the minecraft server.",
			Required:    true,
		},
		{
			Type:        discord.ApplicationCommandOptionString, //Role,
			Name:        "access",
			Description: "The role for people who can use the minecraft server. Set to - to disable",
			Required:    true,
		},
		{
			Type:        discord.ApplicationCommandOptionBoolean,
			Name:        "allowregistration",
			Description: "Whether or not to allow users to register with the bot at the moment.",
			Required:    true,
		},
		{
			Type:        discord.ApplicationCommandOptionInteger,
			Name:        "maxaccounts",
			Description: "The maximum amount of accounts per user.",
			Required:    true,
			MinValue:    1,
			MaxValue:    1000,
		},
	}
}

func (c *SetupCommand) Execute(ctx *Context) bool {
	if CheckGuild(ctx) != nil {
		return false
	}

	if ctx.interaction.Member.Permissions&discord.BitwisePermissionFlagAdministrator == 0 {
		SendPermissionsError(GuildSettings{}, ctx)
		return false
	}

	adminRole := ctx.interaction.Data.Options[0].String()
	accessRole := ctx.interaction.Data.Options[1].String()
	allowRegistration := ctx.interaction.Data.Options[2].Bool()
	maxAccounts := ctx.interaction.Data.Options[3].Int()

	// Add to database
	guildid := ctx.interaction.GuildId

	err := db.Transaction(func(tx *gorm.DB) error {
		err := UpdateDisplayName(tx, ctx.interaction.User)
		if err != nil {
			return err
		}

		dest := GuildSettings{
			ID:                     guildid,
			AdminRole:              adminRole,
			AccessRole:             accessRole,
			AllowUserRegistrations: allowRegistration,
			MaxAccountsPerUser:     maxAccounts,
		}

		mdl := tx.Model(&dest)
		err = mdl.Error
		if err != nil {
			return err
		}

		err = mdl.FirstOrCreate(&dest, guildid).Error
		if err != nil {
			return err
		}

		err = mdl.Updates(map[string]interface{}{
			"AdminRole":              adminRole,
			"AccessRole":             accessRole,
			"AllowUserRegistrations": allowRegistration,
			"MaxAccountsPerUser":     maxAccounts}).Error
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
	message := fmt.Sprintf(`**Reconfigured by:** <@%s>
**Admin Role:** <@&%s>
**Access Role:** <@&%s>
**Allow Registrations:** %t
**Max Accounts Per User:** %d`,
		ctx.interaction.Member.User.Id,
		adminRole,
		accessRole,
		allowRegistration,
		maxAccounts)

	e.SetTitle("Bot Configuration Changed")
	e.SetDescription(message)
	ThemeEmbed(e, ctx)

	// Send response
	ctx.client.Interaction.CreateResponse(ctx.interaction.Id,
		ctx.interaction.Token,
		&discord.InteractionCallbackMessage{Embeds: []*embed.Embed{e.Embed()},
			Flags: discord.MessageFlagUrgent})

	slog.Info("Changed the configuration, see discord for further details", "discord ID", ctx.interaction.Member.User.Id)
	return true
}
