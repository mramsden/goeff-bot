# Goeff Bot

A simple voice channel presence tracker for your Discord server.

## Initial setup

This guide assumes you have a server that is configured with docker and git.

### Discord

You will need to have a Discord app setup in the [developer portal](https://discord.com/developers). Under the "Bot" part of the app you will find a token. This is your "bot token" take a note of it and keep it safe for later.

Next go to the OAuth2 page and go to the OAuth2 URL Generator. Select the `bot` scope and under bot permissions select `Send Messages`. Ensure `Guild Install` is selected for the integration type. Now visit the generated URL and grant access for your bot to the server you want to monitor.

### Build and run the bot

The bot needs to be built before it can be run.

First clone the code to your server by running:

```
git clone https://github.com/mramsden/goeff-bot.git
cd goeff-bot
```

When in the `goeff-bot` directory run the following to build the docker container:

```
docker build goeff-bot:latest .
```

Now that this has been run you can run the following command to start the container:

```
docker run -e DISCORD_BOT_TOKEN=your_bot_token_here -e DISCORD_NOTIFY_CHANNEL=channel_id_to_notify --restart always -it goeff-bot:latest
```

The two environment variables are required in this command other options can be adjusted as required. You should now see the bot come online in your Discord. Join a voice channel on your server and you will see a message arrive.
