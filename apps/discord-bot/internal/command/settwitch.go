package command

import (
	"context"
	"fmt"
	"log"
	"regexp"

	"github.com/bwmarrin/discordgo"
	"github.com/senchabot-opensource/monorepo/apps/discord-bot/internal/command/helpers"
	"github.com/senchabot-opensource/monorepo/apps/discord-bot/internal/service"
	"github.com/senchabot-opensource/monorepo/apps/discord-bot/internal/service/streamer"
	"github.com/senchabot-opensource/monorepo/config"
	"github.com/senchabot-opensource/monorepo/helper"
)

func (c *commands) SetTwitchCommand(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate, service service.Service) {
	options := i.ApplicationCommandData().Options

	switch options[0].Name {
	case "streamer":
		options = options[0].Options
		twitchUsername := options[0].StringValue()

		commandUsername := i.Member.User.Username
		twitchUsername = helper.ParseTwitchUsernameURLParam(twitchUsername)

		response0, uInfo := streamer.GetTwitchUserInfo(twitchUsername)
		if response0 != "" {
			ephemeralRespond(s, i, response0)
			return
		}

		response1, ok := streamer.CheckIfTwitchStreamerExist(ctx, twitchUsername, uInfo, s, i, service)
		if helpers.IsChannelNameNotGiven(len(options)) && ok {
			ephemeralRespond(s, i, response1)
			return
		}

		if helpers.IsChannelNameNotGiven(len(options)) {
			channelData, err := service.GetDiscordBotConfig(ctx, i.GuildID, "stream_anno_default_channel")
			if err != nil {
				log.Println("[command.SetTwitchCommand.streamer] GetDiscordBotConfig error:", err.Error())
				ephemeralRespond(s, i, config.ErrorMessage+"#0000")
				return
			}
			if channelData == nil || channelData.Value == "" {
				ephemeralRespond(s, i, "Twitch yayıncısı eklerken daha önce `/set-twitch announcement default-channel` komutuyla varsayılan duyuru kanalı eklemiş olmalı veya isteğe bağlı kanal adını belirtmelisiniz.")
				return
			}

			ch, err := s.Channel(channelData.Value)
			if err != nil {
				ephemeralRespond(s, i, config.ErrorMessage+"#XXXY")
				return
			}

			resp := streamer.SetTwitchStreamer(ctx, uInfo, nil, ch, i.GuildID, commandUsername, service)
			ephemeralRespond(s, i, resp)
			return
		}

		channelId := options[1].ChannelValue(s).ID
		channelName := options[1].ChannelValue(s).Name

		streamerData, err := service.GetDiscordTwitchLiveAnno(ctx, uInfo.ID, i.GuildID)
		if err != nil {
			log.Println("[command.SetTwitchCommand.streamer] GetDiscordTwitchLiveAnno error:", err.Error())
			return
		}

		if streamerData != nil && channelId == streamerData.AnnoChannelID {
			ephemeralRespond(s, i, fmt.Sprintf("`%v` kullanıcı adlı Twitch yayıncısı `%v` kanalına canlı yayın duyuruları için daha önce eklenmiş.", twitchUsername, channelName))
			return
		}

		ch, err := s.Channel(channelId)
		if err != nil {
			ephemeralRespond(s, i, config.ErrorMessage+"#XXXY")
			return
		}

		resp := streamer.SetTwitchStreamer(ctx, uInfo, &channelId, ch, i.GuildID, commandUsername, service)
		ephemeralRespond(s, i, resp)

	case "event-channel":
		options = options[0].Options
		channelId := options[0].ChannelValue(s).ID
		channelName := options[0].ChannelValue(s).Name

		ok, err := service.AddAnnouncementChannel(ctx, channelId, i.GuildID, i.Member.User.Username)
		if err != nil {
			log.Println("[command.SetTwitchCommand.event-channel] AddAnnouncementChannel error:", err.Error())
			ephemeralRespond(s, i, config.ErrorMessage+"#0002")
			return
		}
		if !ok {
			ephemeralRespond(s, i, fmt.Sprintf("`%v` isimli kanal Twitch yayın duyurusu etkinlikleri için daha önce eklenmiş.", channelName))
			return
		}

		ephemeralRespond(s, i, fmt.Sprintf("`%v` isimli kanal Twitch yayın duyurusu etkinlikleri için listeye eklendi.", channelName))

	case "announcement":
		options = options[0].Options
		switch options[0].Name {
		case "default-channel":
			options = options[0].Options
			channelId := options[0].ChannelValue(s).ID
			channelName := options[0].ChannelValue(s).Name

			_, err := service.SetDiscordBotConfig(ctx, i.GuildID, "stream_anno_default_channel", channelId)
			if err != nil {
				log.Println("[command.SetTwitchCommand.announcement.default-channel] SetDiscordBotConfig error:", err.Error())
				ephemeralRespond(s, i, config.ErrorMessage+"#0001")
				return
			}

			ephemeralRespond(s, i, "`"+channelName+"` isimli kanal varsayılan duyuru kanalı olarak ayarlandı.")

		case "default-content":
			options = options[0].Options
			annoText := options[0].StringValue()

			_, err := service.SetDiscordBotConfig(ctx, i.GuildID, "stream_anno_default_content", annoText)
			if err != nil {
				log.Println("[command.SetTwitchCommand.announcement.default-content] SetDiscordBotConfig error:", err.Error())
				ephemeralRespond(s, i, config.ErrorMessage+"#0001")
				return
			}

			ephemeralRespond(s, i, "Yayın duyuru mesajı içeriği ayarlandı: `"+annoText+"`")

		case "custom-content":
			options = options[0].Options
			twitchUsername := options[0].StringValue()
			twitchUsername = helper.ParseTwitchUsernameURLParam(twitchUsername)
			annoContent := options[1].StringValue()

			response0, uInfo := streamer.GetTwitchUserInfo(twitchUsername)
			if response0 != "" {
				ephemeralRespond(s, i, response0)
				return
			}

			ok, err := service.UpdateTwitchStreamerAnnoContent(ctx, uInfo.ID, i.GuildID, &annoContent)
			if err != nil {
				log.Println("[command.SetTwitchCommand.announcement.custom-content] UpdateTwitchStreamerAnnoContent error:", err.Error())
				ephemeralRespond(s, i, config.ErrorMessage+"#001TEKNOBARBISI")
				return
			}

			if !ok {
				ephemeralRespond(s, i, "`"+twitchUsername+"` kullanıcı adlı Twitch yayıncısı veritabanında bulunamadı.")
				return
			}

			if annoContent == "" {
				cfg, err := service.GetDiscordBotConfig(ctx, i.GuildID, "stream_anno_default_content")
				if err != nil {
					log.Println("[command.SetTwitchCommand.announcement.custom-content] GetDiscordBotConfig error:", err.Error())
				}

				if cfg != nil {
					ephemeralRespond(s, i, twitchUsername+" kullanıcı adlı Twitch yayıncısı için özelleştirilmiş duyuru mesajı içeriği kaldırıldı. `/set-twitch announcement default-content komutuyla ayarladığız mesaj içeriği kullanılacak: `"+cfg.Value+"`")
					return
				}

				ephemeralRespond(s, i, twitchUsername+" kullanıcı adlı Twitch yayıncısı için özelleştirilmiş duyuru mesajı içeriği kaldırıldı. Varsayılan duyuru mesajı içeriği kullanılacak: `{stream.user}, {stream.category} yayınına başladı! {stream.url}`.")
				return
			}

			ephemeralRespond(s, i, twitchUsername+" kullanıcı adlı Twitch yayıncısı için duyuru mesajı içeriği ayarlandı: `"+annoContent+"`")

		case "category-filter":
			options = options[0].Options
			channelId := options[0].ChannelValue(s).ID
			channelName := options[0].ChannelValue(s).Name

			categoryFilterRegex := options[1].StringValue()

			conditionType := uint(options[2].UintValue())

			_, err := regexp.Compile(categoryFilterRegex)
			if err != nil {
				log.Printf("[command.SetTwitchCommand.announcement.category-filter] regexp.Compile error: %s, Expr: %s", err.Error(), categoryFilterRegex)
				ephemeralRespond(s, i, fmt.Sprintf("Error while parsing regular expression. (RegEx): `%s`", categoryFilterRegex))
				return
			}

			ok, err := service.SetDiscordChannelTwitchCategoryFilter(ctx, i.GuildID, channelId, categoryFilterRegex, conditionType, i.Member.User.ID)

			if err != nil {
				log.Println(err)
				ephemeralRespond(s, i, config.ErrorMessage+"#0001")
				return
			}

			if !ok {
				ephemeralRespond(s, i, config.ErrorMessage+"#0002")
				return
			}

			var conditionText string
			switch conditionType {
			case 0:
				conditionText = "eşleşmeyecek"
			case 1:
				conditionText = "eşleşecek"
			}

			ephemeralRespond(s, i, fmt.Sprintf("`%s` isimli duyuru kanalına atılacak Twitch yayın duyurularının kategori filtresi `%s` şekilde `%s` olarak ayarlandı.", channelName, conditionText, categoryFilterRegex))
		}
	}
}

func SetTwitchCommandMetadata() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        "set-twitch",
		Description: "Discord bot configuration",
		DescriptionLocalizations: &map[discordgo.Locale]string{
			discordgo.Turkish: "Discord botunu yapılandırma ayarları",
		},
		DefaultMemberPermissions: &setdeletePermissions,
		Options: []*discordgo.ApplicationCommandOption{
			// set-twitch streamer
			{
				Name:        "streamer",
				Description: "Add a streamer for live stream announcements. You can specify custom channel.",
				DescriptionLocalizations: map[discordgo.Locale]string{
					discordgo.Turkish: "Yayın duyuru mesajı atılacak yayıncıyı ekle. Özel kanal atayabilirsin.",
				},
				Type: discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "twitch-username-or-url",
						Description: "Twitch profile url or username",
						DescriptionLocalizations: map[discordgo.Locale]string{
							discordgo.Turkish: "Twitch kullanıcı profil linki veya kullanıcı adı",
						},
						Required: true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionChannel,
						Name:        "channel",
						Description: "Text channel for live stream announcements",
						DescriptionLocalizations: map[discordgo.Locale]string{
							discordgo.Turkish: "Duyuruların yapılacağı yazı kanalı adı",
						},
						ChannelTypes: []discordgo.ChannelType{
							discordgo.ChannelTypeGuildNews,
							discordgo.ChannelTypeGuildText,
						},
						Required: false,
					},
				},
			},
			// set-twitch announcement
			{
				Name:        "announcement",
				Description: "Announcement group",
				Options: []*discordgo.ApplicationCommandOption{
					// set-twitch announcement default-channel
					{
						Name:        "default-channel",
						Description: "Set the default channel for live stream announcements.",
						DescriptionLocalizations: map[discordgo.Locale]string{
							discordgo.Turkish: "Yayın duyuru mesajlarının atılacağı varsayılan kanalı ayarla.",
						},
						Type: discordgo.ApplicationCommandOptionSubCommand,
						Options: []*discordgo.ApplicationCommandOption{
							{
								Type:        discordgo.ApplicationCommandOptionChannel,
								Name:        "channel",
								Description: "Text channel",
								DescriptionLocalizations: map[discordgo.Locale]string{
									discordgo.Turkish: "Yazı kanalı",
								},
								ChannelTypes: []discordgo.ChannelType{
									discordgo.ChannelTypeGuildNews,
									discordgo.ChannelTypeGuildText,
								},
								Required: true,
							},
						},
					},
					// set-twitch announcement default-content
					{
						Name:        "default-content",
						Description: "Default message content for live stream announcements.",
						DescriptionLocalizations: map[discordgo.Locale]string{
							discordgo.Turkish: "Varsayılan yayın duyuru mesajı ayarla.",
						},
						Type: discordgo.ApplicationCommandOptionSubCommand,
						Options: []*discordgo.ApplicationCommandOption{
							{
								Type:        discordgo.ApplicationCommandOptionString,
								Name:        "announcement-content",
								Description: "Stream announcement content ({twitch.username} {twitch.url} {stream.category} {stream.title})",
								DescriptionLocalizations: map[discordgo.Locale]string{
									discordgo.Turkish: "Yayın mesaj duyuru içeriği ({twitch.username} {twitch.url} {stream.category} {stream.title})",
								},
								Required: true,
							},
						},
					},
					// set-twitch announcement custom-content
					{
						Name:        "custom-content",
						Description: "Streamer specific custom live stream announcement message content.",
						DescriptionLocalizations: map[discordgo.Locale]string{
							discordgo.Turkish: "Yayıncıya özgü yayın duyuru mesajı ayarla.",
						},
						Type: discordgo.ApplicationCommandOptionSubCommand,
						Options: []*discordgo.ApplicationCommandOption{
							{
								Type:        discordgo.ApplicationCommandOptionString,
								Name:        "twitch-username-or-url",
								Description: "Twitch profile url or username",
								DescriptionLocalizations: map[discordgo.Locale]string{
									discordgo.Turkish: "Twitch kullanıcı profil linki veya kullanıcı adı",
								},
								Required: true,
							},
							{
								Type:        discordgo.ApplicationCommandOptionString,
								Name:        "announcement-content",
								Description: "Stream announcement content ({twitch.username} {twitch.url} {stream.category} {stream.title})",
								DescriptionLocalizations: map[discordgo.Locale]string{
									discordgo.Turkish: "Yayın mesaj duyuru içeriği ({twitch.username} {twitch.url} {stream.category} {stream.title})",
								},
								Required: true,
							},
						},
					},
					// set-twitch announcement category-filter
					{
						Name:        "category-filter",
						Description: "Filtering Discord channel-specific Twitch stream category for announcement. (?i)Just Chatting",
						DescriptionLocalizations: map[discordgo.Locale]string{
							discordgo.Turkish: "Discord kanalına özgü yayın duyurularının filtrelenmesi. (?i)Just Chatting",
						},
						Type: discordgo.ApplicationCommandOptionSubCommand,
						Options: []*discordgo.ApplicationCommandOption{
							{
								Type:        discordgo.ApplicationCommandOptionChannel,
								Name:        "channel",
								Description: "Text channel",
								DescriptionLocalizations: map[discordgo.Locale]string{
									discordgo.Turkish: "Yazı kanalı",
								},
								ChannelTypes: []discordgo.ChannelType{
									discordgo.ChannelTypeGuildNews,
									discordgo.ChannelTypeGuildText,
								},
								Required: true,
							},
							{
								Type:        discordgo.ApplicationCommandOptionString,
								Name:        "regex",
								Description: "RegEx string. For example: (?i)Just Chatting",
								DescriptionLocalizations: map[discordgo.Locale]string{
									discordgo.Turkish: "RegEx dizesi. Örneğin: (?i)Just Chatting",
								},
								Required: true,
							},
							{
								Type:        discordgo.ApplicationCommandOptionInteger,
								Name:        "condition",
								Description: "Condition",
								DescriptionLocalizations: map[discordgo.Locale]string{
									discordgo.Turkish: "Koşul",
								},
								Choices: []*discordgo.ApplicationCommandOptionChoice{
									{Name: "matches",
										NameLocalizations: map[discordgo.Locale]string{
											discordgo.Turkish: "eşleşir",
										},
										Value: 1},
									{Name: "does not match",
										NameLocalizations: map[discordgo.Locale]string{
											discordgo.Turkish: "eşleşmez",
										},
										Value: 0},
								},
								Required: true,
							},
						},
					},
				},
				Type: discordgo.ApplicationCommandOptionSubCommandGroup,
			},
			// set-twitch event-channel
			{
				Name:        "event-channel",
				Description: "Select the live stream announcements channel to create Discord events for live streams.",
				DescriptionLocalizations: map[discordgo.Locale]string{
					discordgo.Turkish: "Canlı yayınların Discord etkinliklerini oluşturmak için canlı yayın duyuruları kanalını seç.",
				},
				Type: discordgo.ApplicationCommandOptionSubCommand,
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionChannel,
						Name:        "channel",
						Description: "Text channel to follow Twitch live stream announcements",
						DescriptionLocalizations: map[discordgo.Locale]string{
							discordgo.Turkish: "Twitch yayın duyurularının takip edileceği yazı kanalı",
						},
						ChannelTypes: []discordgo.ChannelType{
							discordgo.ChannelTypeGuildNews,
							discordgo.ChannelTypeGuildText,
						},
						Required: true,
					},
				},
			},
		},
	}
}
