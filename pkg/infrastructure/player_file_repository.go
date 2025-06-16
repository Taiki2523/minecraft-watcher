package infrastructure

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/taiki2523/minecraft-watcher/pkg/domain"
)

type PlayerFileRepository struct {
	tmpPath string
	mu      sync.Mutex
}

func NewPlayerFileRepository(tmpPath string) *PlayerFileRepository {
	return &PlayerFileRepository{tmpPath: tmpPath}
}

func (r *PlayerFileRepository) filePath() string {
	_ = os.MkdirAll(r.tmpPath, 0755)
	return filepath.Join(r.tmpPath, "active_players.txt")
}

func (r *PlayerFileRepository) Add(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	players, _ := r.List()
	playerSet := map[string]struct{}{}
	for _, p := range players {
		playerSet[p.Name] = struct{}{}
	}
	playerSet[name] = struct{}{}
	file, err := os.Create(r.filePath())
	if err != nil {
		return err
	}
	defer file.Close()
	for n := range playerSet {
		fmt.Fprintln(file, n)
	}
	return nil
}

func (r *PlayerFileRepository) Remove(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	players, _ := r.List()
	playerSet := map[string]struct{}{}
	for _, p := range players {
		playerSet[p.Name] = struct{}{}
	}
	delete(playerSet, name)
	file, err := os.Create(r.filePath())
	if err != nil {
		return err
	}
	defer file.Close()
	for n := range playerSet {
		fmt.Fprintln(file, n)
	}
	return nil
}

func (r *PlayerFileRepository) List() ([]domain.Player, error) {
	file, err := os.Open(r.filePath())
	if err != nil {
		return nil, nil
	}
	defer file.Close()
	var players []domain.Player
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		players = append(players, domain.Player{Name: scanner.Text()})
	}
	return players, nil
}
