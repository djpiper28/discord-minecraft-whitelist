package main

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/Goscord/goscord/discord"
	"github.com/Goscord/goscord/discord/embed"
	"gorm.io/gorm"
)

type VerifyCommand struct{}

func (c *VerifyCommand) Name() string {
	return "mcverify"
}

func (c *VerifyCommand) Description() string {
	return "Verify a Minecraft account as yours."
}

func (c *VerifyCommand) Category() string {
	return "general"
}

func (c *VerifyCommand) Options() []*discord.ApplicationCommandOption {
	return []*discord.ApplicationCommandOption{
		{
			Type:        discord.ApplicationCommandOptionString,
			Name:        "minecraftname",
			Description: "Minecraft Account Name i.e: Herobrine",
			Required:    true,
		},
		{
			Type:        discord.ApplicationCommandOptionInteger,
			Name:        "verificationcode",
			Description: "Verification code as from the Minecraft server",
			Required:    true,
		},
	}
}

func (c *VerifyCommand) Execute(ctx *Context) bool {
	if CheckGuild(ctx) != nil {
		return false
	}

	accountName := ctx.interaction.Data.Options[0].String()
	verificationCode := ctx.interaction.Data.Options[1].Int()

	// Add to database
	guildid := ctx.interaction.GuildId

	err := db.Transaction(func(tx *gorm.DB) error {
		err := UpdateDisplayName(tx, ctx.interaction.User)
		if err != nil {
			return err
		}

		// Get guild settings
		var gs GuildSettings
		minecraftUserModel := tx.Model(&gs)
		err = minecraftUserModel.Error
		if err != nil {
			return err
		}

		err = minecraftUserModel.First(&gs, guildid).Error
		if err != nil {
			return err
		}

		// Check for access role
		if gs.AccessRole != "-" {
			if !Contains(ctx.interaction.Member.Roles, gs.AccessRole) {
				return errors.New(fmt.Sprintf("Invalid Permissions - you require the role <@&%s> to use this command", gs.AccessRole))
			}
		}

		discordUser := DiscordUser{
			HasAdminRole:  UserIsAdmin(gs, ctx.interaction.Member),
			Banned:        false,
			DiscordUserID: ctx.interaction.Member.User.Id,
		}
		minecraftUserModel = tx.Model(&discordUser)
		err = minecraftUserModel.Error
		if err != nil {
			return err
		}

		userLookup, err := GetMinecraftUser(accountName)
		if err != nil {
			return err
		}

		// Get the discord minecraft user
		dcmcUser := DiscordMinecraftUser{}
		discordUserModel := tx.Model(&dcmcUser)
		err = discordUserModel.Error
		if err != nil {
			return err
		}

		err = discordUserModel.First(&dcmcUser, "discord_user_id = ? AND minecraft_user_id = ?", ctx.interaction.Member.User.Id, userLookup.Id).Error
		if err != nil {
			return err
		}

		// Get the minecraft user with the accountName
		mcUser := MinecraftUser{}
		minecraftUserModel = tx.Model(&mcUser)
		err = minecraftUserModel.Error
		if err != nil {
			return err
		}

		err = minecraftUserModel.Select("username", "verification_number").
			First(&mcUser, "id = ?", userLookup.Id).Error
		if err != nil {
			return err
		}

		if verificationCode != mcUser.VerificationNumber {
			return errors.New("Incorrect verification code")
		}

		// Mark the user as verified
		minecraftUserModel = tx.Model(&mcUser)
		err = minecraftUserModel.Where("id = ?", userLookup.Id).
			Update("verified", true).Error
		if err != nil {
			return err
		}

		err = tx.Delete(&DiscordMinecraftUser{}, "minecraft_user_id = ? AND discord_user_id <> ?",
			userLookup.Id,
			ctx.interaction.User.Id).Error
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
**Minecraft User:** %s`,
		ctx.interaction.Member.User.Id,
		accountName)

	e.SetTitle("Verified a Minecraft User")
	e.SetDescription(message)
	ThemeEmbed(e, ctx)

	// Send response
	ctx.client.Interaction.CreateResponse(ctx.interaction.Id,
		ctx.interaction.Token,
		&discord.InteractionCallbackMessage{Embeds: []*embed.Embed{e.Embed()},
			Flags: discord.MessageFlagUrgent})

	slog.Info("Verifed Minecraft account", "discord ID", ctx.interaction.Member.User.Id, "name", accountName)

	return true
}
