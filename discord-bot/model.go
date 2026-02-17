package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Goscord/goscord/discord"
	"github.com/Goscord/goscord/discord/embed"
	"github.com/Goscord/goscord/gateway"
	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"gorm.io/gorm"
)

// Bot database model - a "singleton" for the server
type GuildSettings struct {
	gorm.Model
	// Guild Id
	ID string `gorm:"primaryKey"`
	// This is the admin role (@Committe 2022/23)
	AdminRole string
	// Role to get access to the system (@Student)
	AccessRole string
	// A master toggle to turn off user registrations
	AllowUserRegistrations bool
	MaxAccountsPerUser     int64
}

type DiscordUser struct {
	// This is a cache of whether or not the user has the admin role
	HasAdminRole bool `gorm:"index"`
	// This is whether the user and, all their accounts are banned
	Banned                bool                   `gorm:"index"`
	DiscordUserID         string                 `gorm:"primaryKey"`
	DiscordMinecraftUsers []DiscordMinecraftUser `gorm:"foreignKey:DiscordUserID"`
	DisplayName           string
	// i.e: #ffffff
	ColourHex string `gorm:"default:#ffffff"`
}

type DiscordMinecraftUser struct {
	DiscordUserID   string    `gorm:"foriegnKey:discord_user.discord_user_id,index,unique,composite:discord_minecraft_user"`
	MinecraftUserID uuid.UUID `gorm:"foreignKey:minecraft_user.id,index,unique,composite:discord_minecraft_user"`
}

type MinecraftUser struct {
	Id uuid.UUID `gorm:"primaryKey"`
	// Cached username, updated every time the user logs in to the server
	Username           string
	LastLoginTime      time.Time
	LastX              float32
	LastY              float32
	LastZ              float32
	LastIpAddress      pgtype.Inet `gorm:"type:inet"`
	LastChunkImage     []byte
	LastSkinImage      []byte
	VerificationNumber int64
	Verified           bool
}

func AutoMigrateModel() {
	err := db.AutoMigrate(
		&DiscordUser{},
		&MinecraftUser{},
		&DiscordMinecraftUser{},
		&GuildSettings{},
	)

	if err != nil {
		log.Fatalf("Cannot migrate database: %w", err)
	}
}

// Helper function to set IP addresses, probably won't be used lmao
func SetInet(ip string) pgtype.Inet {
	var inet pgtype.Inet
	inet.Set(ip)
	return inet
}

type Context struct {
	client      *gateway.Session
	interaction *discord.Interaction
}

type Command interface {
	Name() string
	Description() string
	Category() string
	Options() []*discord.ApplicationCommandOption
	Execute(ctx *Context) bool
}

func Register(cmd Command, client *gateway.Session, commands map[string]Command) {
	appCmd := &discord.ApplicationCommand{
		Name:        cmd.Name(),
		Type:        discord.ApplicationCommandChat,
		Description: cmd.Description(),
		Options:     cmd.Options(),
	}

	_, err := client.Application.RegisterCommand(client.Me().Id, "", appCmd)
	if err != nil {
		log.Printf("Error registering command '%s' - %s", cmd.Name(), err)
	} else {
		log.Printf("Registered command '%s'", cmd.Name())
	}
	commands[cmd.Name()] = cmd
}

func ThemeEmbed(e *embed.Builder, ctx *Context) {
	e.SetFooter(ctx.client.Me().Username, ctx.client.Me().AvatarURL())
	e.SetColor(embed.Green)
	e.SetThumbnail(ctx.interaction.Member.User.AvatarURL())
}

func SendError(message string, ctx *Context) {
	e := embed.NewEmbedBuilder()

	e.SetTitle("An Error Occurred During Your Command")
	e.SetDescription(message)
	ThemeEmbed(e, ctx)

	// Send response
	ctx.client.Interaction.CreateResponse(ctx.interaction.Id,
		ctx.interaction.Token,
		&discord.InteractionCallbackMessage{Embeds: []*embed.Embed{e.Embed()},
			Flags: discord.MessageFlagEphemeral})
}

func SendAdminPermissionsError(gs GuildSettings, ctx *Context) {
	SendError(fmt.Sprintf("You require the <@&%s> role to perform this command.", gs.AdminRole), ctx)
}

func SendPermissionsError(gs GuildSettings, ctx *Context) {
	SendError(fmt.Sprintf("You require the <@&%s> role to perform this command.", gs.AccessRole), ctx)
}

func SendBannedError(ctx *Context) {
	SendError("You cannot use this command as you have been banned from using the server", ctx)
}

func SendWrongGuildError(ctx *Context) {
	SendError("You cannot use this bot from outside of the owner's server.", ctx)
}

func SendInternalError(err error, ctx *Context) {
	log.Print(err)
	SendError(fmt.Sprintf("An internal error has occurred:\n```\n%s\n```", err), ctx)
}

const DISCORD_GUILD_ID = "DISCORD_GUILD_ID"

func CheckGuild(ctx *Context) error {
	requiredGuild := os.Getenv(DISCORD_GUILD_ID)
	guildid := ctx.interaction.GuildId
	if guildid != requiredGuild {
		SendWrongGuildError(ctx)
		log.Printf("Guild %s is not the guild (%s).", guildid, requiredGuild)
		return errors.New("Wrong guild.")
	}
	return nil
}

// Vile linear search (grim)
func Contains(arr []string, key string) bool {
	for _, item := range arr {
		if item == key {
			return true
		}
	}
	return false
}

func UserIsAdmin(gs GuildSettings, user *discord.GuildMember) bool {
	return Contains(user.Roles, gs.AdminRole)
}

func UpdateDisplayName(tx *gorm.DB, user *discord.User) error {
	if user == nil {
		log.Print("User is nil")
		// This is non-fatal as the bot may be misconfigured
		return nil
	}

	return tx.Model(&DiscordUser{}).
		Where("discord_user_id = ?", user.Id).
		Update("display_name", user.Username).Error
}

func UpdateThread(client *gateway.Session) {
	for {
		guild, err := client.State().Guild(os.Getenv(DISCORD_GUILD_ID))
		if err != nil {
			log.Printf("Cannot find the DISCORD_GUILD_ID guild, %s.", err)
			time.Sleep(time.Minute * 5)
			continue
		}

		members := guild.Members
		log.Printf("Updating member list with %d members", len(members))

		for _, member := range members {
			log.Printf("Updating member %s", member.User.Username)

			err = UpdateDisplayName(db, member.User)
			if err != nil {
				log.Printf("Cannot update user %s in the database, %s", member.User.Username, err)
			}
		}

		time.Sleep(time.Hour)
	}
}
