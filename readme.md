# 🤖 Discord Server Nuker Bot 🤖

This project is a Discord bot that provides various destructive actions for Discord servers, such as deleting all channels, creating new channels, deleting all roles, creating new roles, and sending Direct Messages (DMs) to all members. It's built with Go and uses the DiscordGo library.

## 🚀 Features

- 🗑️ **Delete all channels**: The bot can delete all channels in a server.
- 📁 **Create new channels**: The bot can create a specified number of new channels in a server.
- ❌ **Delete all roles**: The bot can delete all roles in a server.
- 🎭 **Create new roles**: The bot can create a specified number of new roles in a server.
- 📨 **Send DMs**: The bot can send DMs to all members in a server.

## ⚙️ Configuration

The bot's behavior can be configured by modifying the `config.toml` file. Here's what each configuration option does:

- `RoleName`: The base name for the roles that the bot creates.
- `ChannelName`: The base name for the channels that the bot creates.
- `WebhookName`: The base name for the webhooks that the bot creates.
- `WebhookSpamMessage`: The message that the bot sends via webhooks.
- `TTS`: Whether the bot should use Text-To-Speech (TTS) when sending messages.
- `ProfilePictureLink`: The link to the profile picture that the bot uses.
- `ProfilePictureName`: The name of the profile picture that the bot uses.
- `ShouldMassDM`: Whether the bot should send DMs to all members.
- `MassDMMessage`: The message that the bot sends to all members.
- `NumChannels`: The number of channels that the bot creates.
- `NumWebhooksPerChannel`: The number of webhooks that the bot creates in each channel.
- `NumRoles`: The number of roles that the bot creates.

## 📚 Usage

To use the bot, you need to send a command in a Discord server where the bot is a member. The command should start with a `!` character. Here are the available commands:

- `!start`: Starts the bot. The bot will perform all actions specified in the `config.toml` file.
- `!deleteChannels`: Deletes all channels in the server.
- `!createChannel`: Creates new channels in the server.
- `!deleteRoles`: Deletes all roles in the server.
- `!createRole`: Creates new roles in the server.
- `!sendDMs`: Sends DMs to all members in the server.
- `!help`: Sends a help message to the user's DMs.

## ⚠️ Disclaimer

This bot is intended for educational purposes only. Misuse of this bot can violate Discord's Terms of Service.

## 📝 License

This project is licensed under the MIT License.