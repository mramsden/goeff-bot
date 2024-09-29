# Goeff Bot

A simple voice channel presence tracker for your Discord server.

## Initial setup

This guide assumes you have a server that is configured with docker and git.

### Discord

You will need to have a Discord app setup in the [developer portal](https://discord.com/developers). Under the "Bot" part of the app you will find a token. This is your "bot token" take a note of it and keep it safe for later.

Next go to the OAuth2 page and go to the OAuth2 URL Generator. Select the `bot` scope and under bot permissions select `Send Messages`. Ensure `Guild Install` is selected for the integration type. Now visit the generated URL and grant access for your bot to the server you want to monitor.

### Run the bot

Run the following command to start the bot:

```
docker run -e DISCORD_BOT_TOKEN=your_bot_token_here -e DISCORD_NOTIFY_CHANNEL=channel_id_to_notify --restart always -it --name goeff-bot ghcr.io/mramsden/goeff-bot:latest
```

The two environment variables are required in this command other options can be adjusted as required. You should now see the bot come online in your Discord. Join a voice channel on your server and you will see a message arrive.

## Updates

To update to the latest version of the bot run:

```
docker pull ghcr.io/mramsden/goeff-bot:latest
```

Next stop the running bot and remove it:

```
docker stop goeff-bot && docker rm goeff-bot
```

Start it again using the usual command:

```
docker run -e DISCORD_BOT_TOKEN=your_bot_token_here -e DISCORD_NOTIFY_CHANNEL=channel_id_to_notify --restart always -it --name goeff-bot ghcr.io/mramsden/goeff-bot:latest
```
