package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Netflix/go-env"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"gitlab.com/fynbos/backend/wallets"
	"gitlab.com/fynbos/discordbot/cmd"
	"gitlab.com/fynbos/discordbot/ops"
	fynbos_env "gitlab.com/fynbos/env"
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

	bot, err := discordgo.New("Bot " + environment.DiscordBotToken)
	if err != nil {
		log.Fatal("Invalid bot parameters", zap.Error(err))
	}
	b := NewBackends(&environment, bot)
	defer CloseBackends(b)

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

	go watchForPaymentChanges(b)
	log.Info("Started watching payments for changes...")

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
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: "Sign up for a Fynbos wallet to pay",
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.Button{
								Label:    "Sign up",
								Style:    discordgo.LinkButton,
								Disabled: false,
								URL:      fmt.Sprintf("%s/signup", fynbos_env.GetUrl()),
							},
						},
					},
				},
			},
		})
		if err != nil {
			log.Error("Failed to respond to user", zap.Error(err))
		}

		return nil
	}

	w, err := b.Wallets().Get(ctx, identity.WalletID)
	if err != nil {
		log.Info("no wallet found", zap.String("username", discordUsername))
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags:   discordgo.MessageFlagsEphemeral,
				Content: "Your payment failed",
			},
		})
		if err != nil {
			log.Error("Failed to respond to user", zap.Error(err))
		}

		return nil
	}

	return w
}

func watchForPaymentChanges(b *Backends) {
	ticker := time.NewTicker(5 * time.Second)
	for {
		<-ticker.C

		is, err := ops.ListPaymentInteractions(context.Background(), b, 10)
		if err != nil {
			log.Error("Failed to list payment interactions", zap.Error(err))
			continue
		}

		ops.ProcessInteractions(context.Background(), b, is)
	}
}
