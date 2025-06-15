package interfaces

import (
	"time"

	"github.com/taiki2523/minecraft-watcher/pkg/application"
	"github.com/taiki2523/minecraft-watcher/pkg/domain"
	"github.com/taiki2523/minecraft-watcher/pkg/internal"
)

func StartPlayerCheck(n domain.Notifier, playerService *application.PlayerService, interval time.Duration, stopCh <-chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			players := playerService.GetPlayerList()
			if len(players) == 0 {
				continue // 誰もいなければスキップ
			}

			body := internal.FormatPlayerListStatus(players)
			if body == "" {
				continue
			}
			_ = n.Send(body)

		case <-stopCh:
			return
		}
	}
}
