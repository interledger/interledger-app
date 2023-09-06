package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Netflix/go-env"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"gitlab.com/fynbos/backend/wallets"
	"gitlab.com/fynbos/discordbot/cmd"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

func main() {
	envFile := os.Getenv("ENV_FILE")
	if envFile != "" {
		err := godotenv.Load(envFile)
		if err != nil {
			log.Fatal("Error loading .env file", zap.Error(err))
		}
	}

	var environment Environment
	_, err := env.UnmarshalFromEnviron(&environment)
	if err != nil {
		log.Fatal("Failed to parse environment variables.", zap.Error(err))
	}

	b := NewBackends(&environment)
	defer CloseBackends(b)

	bot, err := discordgo.New("Bot " + environment.DiscordBotToken)
	if err != nil {
		log.Fatal("Invalid bot parameters", zap.Error(err))
	}
	bot.Identify.Intents = discordgo.IntentsGuildMessages
	bot.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) { log.Info("Bot is up!") })
	bot.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(30*time.Second))
		defer cancel()

		w := lookupWallet(ctx, b, s, i)
		walletContext := context.WithValue(ctx, wallets.CtxKey, w)

		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			// TODO: route interaction
			cmd.PaySlashCommandHandler(walletContext, b, s, i)
		}
	})
	err = bot.Open()
	if err != nil {
		log.Fatal("Cannot open the session", zap.Error(err))
	}
	defer bot.Close()

	commands := []*discordgo.ApplicationCommand{
		&cmd.PaySlashCommandSchema,
	}
	_, err = bot.ApplicationCommandBulkOverwrite(bot.State.User.ID, "", commands)
	if err != nil {
		log.Fatal("Cannot register commands", zap.Error(err))
	}
	log.Info("Registered commands...")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
}

func lookupWallet(ctx context.Context, b *Backends, s *discordgo.Session, i *discordgo.InteractionCreate) *wallets.Wallet {
	var discordUsername string
	if i.User != nil {
		discordUsername = i.User.Username
	} else if i.Member != nil {
		discordUsername = i.Member.User.Username
	} else {
		return nil
	}

	identity, err := b.Identities().GetByIdentifier(ctx, discordUsername)
	if err != nil {
		log.Info("no identity found", zap.String("username", discordUsername))
		return nil
	}

	w, err := b.Wallets().Get(ctx, identity.WalletID)
	if err != nil {
		log.Info("no wallet found", zap.String("username", discordUsername))
		return nil
	}

	return w
}
