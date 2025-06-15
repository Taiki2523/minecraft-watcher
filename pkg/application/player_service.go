package application

import (
	"strings"
	"time"

	"github.com/taiki2523/minecraft-watcher/pkg/domain"
	"github.com/taiki2523/minecraft-watcher/pkg/internal"
)

type PlayerService struct {
	Repo     domain.PlayerRepository
	Notifier domain.Notifier
	Clock    func() time.Time
}

func (s *PlayerService) PlayerJoined(name string) error {
	if err := s.Repo.Add(name); err != nil {
		return err
	}
	msg := internal.FormatPlayerEvent("join", name)
	return s.Notifier.Send(msg)
}

func (s *PlayerService) PlayerLeft(name string) error {
	if err := s.Repo.Remove(name); err != nil {
		return err
	}
	msg := internal.FormatPlayerEvent("leave", name)
	return s.Notifier.Send(msg)
}

func (s *PlayerService) AnnouncePlayers() error {
	players, err := s.Repo.List()
	if err != nil || len(players) == 0 {
		return err
	}
	names := []string{}
	for _, p := range players {
		names = append(names, p.Name)
	}
	msg := internal.FormatPlayerListStatus(names)
	return s.Notifier.Send(msg)
}

// ログ行からプレイヤー名を抽出（static関数として切り出し）
func ExtractPlayerName(line string) string {
	parts := strings.Split(line, "]: ")
	if len(parts) < 2 {
		return ""
	}
	fields := strings.Fields(parts[1])
	if len(fields) > 0 {
		return fields[0]
	}
	return ""
}

// プレイヤーリストを取得
func (s *PlayerService) GetPlayerList() []string {
	players, err := s.Repo.List()
	if err != nil {
		return nil
	}
	names := []string{}
	for _, p := range players {
		names = append(names, p.Name)
	}
	return names
}
