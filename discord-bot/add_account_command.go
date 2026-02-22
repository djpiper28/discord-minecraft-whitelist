package main

import (
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"os"

	"github.com/Goscord/goscord/discord"
	"github.com/Goscord/goscord/discord/embed"
	"gorm.io/gorm"
)

const MAX_VERIFCATION_NUMBER = 999999

type AddAccountCommand struct{}

func (c *AddAccountCommand) Name() string {
	return "mcadd"
}

func (c *AddAccountCommand) Description() string {
	return "Add an account to the Minecraft server's whitelist"
}

func (c *AddAccountCommand) Category() string {
	return "general"
}

func (c *AddAccountCommand) Options() []*discord.ApplicationCommandOption {
	return []*discord.ApplicationCommandOption{
		{
			Type:        discord.ApplicationCommandOptionString,
			Name:        "minecraftname",
			Description: "Minecraft Account Name i.e: Herobrine",
			Required:    true,
		},
	}
}

func (c *AddAccountCommand) Execute(ctx *Context) bool {
	if CheckGuild(ctx) != nil {
		return false
	}

	accountName := ctx.interaction.Data.Options[0].String()

	// Add to database
	guildid := ctx.interaction.GuildId

	err := db.Transaction(func(tx *gorm.DB) error {
		err := UpdateDisplayName(tx, ctx.interaction.User)
		if err != nil {
			return err
		}

		// Get guild settings
		var gs GuildSettings
		mdl := tx.Model(&gs)
		err = mdl.Error
		if err != nil {
			return err
		}

		err = mdl.First(&gs, guildid).Error
		if err != nil {
			return err
		}

		// Check for access role
		if gs.AccessRole != "" {
			if !Contains(ctx.interaction.Member.Roles, gs.AccessRole) {
				return errors.New(fmt.Sprintf("Invalid Permissions - you require the role <@&%s> to use this command.", gs.AccessRole))
			}
		}

		// Create the entries
		discordUser := DiscordUser{
			HasAdminRole:  UserIsAdmin(gs, ctx.interaction.Member),
			Banned:        false,
			DiscordUserID: ctx.interaction.Member.User.Id,
		}
		mdl = tx.Model(&discordUser)
		err = mdl.Error
		if err != nil {
			return err
		}

		err = mdl.Where("discord_user_id = ?", ctx.interaction.Member.User.Id).
			FirstOrCreate(&discordUser, ctx.interaction.Member.User.Id).Error
		if err != nil {
			return err
		}

		if discordUser.Banned {
			return errors.New(fmt.Sprintf(`You have been banned from registering accounts.
If this is in fault please contact <@&%s>`, gs.AdminRole))
		}

		// Check that nobody else has verified this account
		mdl = tx.Model(&MinecraftUser{})
		err = mdl.Error
		if err != nil {
			return err
		}

		// Add user
		mcUser, err := GetMinecraftUser(accountName)
		if err != nil {
			return err
		}

		var count int64
		err = mdl.Where("verified = true AND id = ?", mcUser.Id).Count(&count).Error
		if err != nil {
			return err
		}

		if count != 0 {
			return errors.New("This account has already been verified by somebody.")
		}

		// Check that the user has under the maximum accounts
		mdl = tx.Model(&DiscordMinecraftUser{})
		err = mdl.Error
		if err != nil {
			return err
		}

		err = mdl.Where("discord_user_id = ?", ctx.interaction.Member.User.Id).Count(&count).Error
		if err != nil {
			return err
		}

		if count >= gs.MaxAccountsPerUser {
			return errors.New("Maximum user count has been reached for your account.")
		}

		// Create a new unverified user
		minecraftUser := MinecraftUser{
			Id:            mcUser.Id,
			Username:      mcUser.Name,
			LastIpAddress: SetInet("0.0.0.0"),
		}
		mdl = tx.Model(&minecraftUser)
		err = mdl.Error
		if err != nil {
			return err
		}

		err = mdl.FirstOrCreate(&minecraftUser, "Id = ?", mcUser.Id).Error
		if err != nil {
			return err
		}

		// Update verification number
		mdl.Updates(map[string]interface{}{"VerificationNumber": rand.Intn(MAX_VERIFCATION_NUMBER)})

		// Add discord minecraft user
		discordMinecraftUser := DiscordMinecraftUser{
			DiscordUserID:   ctx.interaction.Member.User.Id,
			MinecraftUserID: mcUser.Id,
		}
		mdl = tx.Model(&discordMinecraftUser)
		err = mdl.Error
		if err != nil {
			return err
		}

		err = mdl.FirstOrCreate(&discordMinecraftUser, ctx.interaction.Member.User.Id).Error
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
	message := fmt.Sprintf(`**User:** <@%s>
**Minecraft User:** %s

Please join **%s** with this user to verify the account and, **use /mcverify with the code that you get.**`,
		ctx.interaction.Member.User.Id,
		accountName,
		os.Getenv("MINECRAFT_IP"))

	e.SetTitle("Added A New Minecraft User To The Whitelist")
	e.SetDescription(message)
	ThemeEmbed(e, ctx)

	// Send response
	ctx.client.Interaction.CreateResponse(ctx.interaction.Id,
		ctx.interaction.Token,
		&discord.InteractionCallbackMessage{Embeds: []*embed.Embed{e.Embed()},
			Flags: discord.MessageFlagUrgent})

	slog.Info("added user to the whitelist, it is not verified yet", "discord ID", ctx.interaction.Member.User.Id, "name", accountName)
	return true
}
