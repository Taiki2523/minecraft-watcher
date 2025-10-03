package main

import (
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/taiki2523/minecraft-watcher/pkg/application"
	"github.com/taiki2523/minecraft-watcher/pkg/infrastructure"
	"github.com/taiki2523/minecraft-watcher/pkg/interfaces"
	"github.com/taiki2523/minecraft-watcher/pkg/internal"
)

func main() {
	internal.InitLogger()
	internal.LogStartupEnv()

	logPath := strings.Trim(os.Getenv("LOG_FILE"), `"`)
	tmpPath := "/tmp/minecraft-watcher"
	webhook := strings.Trim(os.Getenv("DISCORD_WEBHOOK_URL"), `"`)
	if logPath == "" || webhook == "" {
		log.Fatal().Msg("環境変数 LOG_FILE と DISCORD_WEBHOOK_URL を指定してください")
	}

	err := internal.LoadMessagesDictFromEnv("PLAYER_ALONE_MESSAGES")
	if err != nil {
		log.Fatal().Err(err).Msg("メッセージ辞書の読み込みに失敗しました")
	}

	//playerRepo := infrastructure.NewPlayerFileRepository(logPath)
	playerFileRepo := infrastructure.NewPlayerFileRepository(tmpPath)
	notifier := &infrastructure.DiscordNotifier{WebhookURL: webhook}

	playerService := &application.PlayerService{
		//Repo:     playerRepo,
		Repo:     playerFileRepo,
		Notifier: notifier,
		Clock:    internal.GetNow,
	}

	// 終了シグナルを受け取るためのチャネル
	stopCh := make(chan struct{})
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		close(stopCh)
	}()

	containerName := strings.Trim(os.Getenv("MC_CONTAINER_NAME"), `"`)
	monitorIntervalStr := strings.Trim(os.Getenv("MC_MONITOR_INTERVAL"), `"`)
	monitorInterval := 10 * time.Second // デフォルト
	if d, err := time.ParseDuration(monitorIntervalStr); err == nil {
		monitorInterval = d
	}
	if containerName != "" {
		go interfaces.StartStatusMonitor(notifier, containerName, monitorInterval, stopCh)
	}

	playerCheckIntervalStr := strings.Trim(os.Getenv("PLAYER_CHECK_INTERVAL"), `"`)
	playerCheckInterval := 5 * time.Minute // デフォルト
	if d, err := time.ParseDuration(playerCheckIntervalStr); err == nil {
		playerCheckInterval = d
	}

	if playerCheckInterval > 0 {
		go interfaces.StartPlayerCheck(notifier, playerService, playerCheckInterval, stopCh)
	} else {
		log.Info().Msg("PLAYER_CHECK_INTERVAL が 0 または無効な値です。プレイヤーチェックは行いません。")
	}
	
	log.Info().Msg("ログ監視を開始します")
	if err := interfaces.WatchFileLoop(logPath, playerService, stopCh); err != nil {
		log.Error().Err(err).Msg("ログ監視中にエラーが発生しました")
	}
	log.Info().Msg("終了します")
}
