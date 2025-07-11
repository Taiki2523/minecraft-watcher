package interfaces

import (
	"bufio"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
	"github.com/taiki2523/minecraft-watcher/pkg/application"
)

func WatchFileLoop(
	logPath string,
	playerService *application.PlayerService,
	stopCh <-chan struct{},
) error {
	var file *os.File
	var reader *bufio.Reader
	var currentInode uint64

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	dir := filepath.Dir(logPath)
	if err := watcher.Add(dir); err != nil {
		return err
	}

	openLogFile := func() error {
		info, err := os.Stat(logPath)
		if err != nil {
			return err
		}
		sysStat, ok := info.Sys().(*syscall.Stat_t)
		if !ok {
			return errors.New("stat.Sys() の型アサーションに失敗")
		}
		inode := sysStat.Ino
		if file != nil && inode == currentInode {
			return nil
		}

		f, err := os.Open(logPath)
		if err != nil {
			return err
		}
		if _, err := f.Seek(0, io.SeekEnd); err != nil {
			f.Close()
			return err
		}
		if file != nil {
			file.Close()
		}
		file = f
		reader = bufio.NewReader(file)
		currentInode = inode
		log.Info().Str("file", logPath).Msg("ログファイルをオープンしました")
		return nil
	}

	if err := openLogFile(); err != nil {
		log.Error().Err(err).Str("file", logPath).Msg("ログファイルをオープンできません")
	}

	for {
		select {
		case <-stopCh:
			if file != nil {
				file.Close()
			}
			return nil

		case event := <-watcher.Events:
			if event.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Rename) != 0 && filepath.Base(event.Name) == filepath.Base(logPath) {
				log.Debug().Str("event", event.String()).Msg("fsnotifyイベント検出")
				if err := openLogFile(); err != nil {
					log.Error().Err(err).Str("file", logPath).Msg("ログファイルをオープンできません")
					break
				}
				if reader == nil {
					break
				}

				for {
					line, err := reader.ReadString('\n')
					if err != nil {
						if errors.Is(err, io.EOF) {
							time.Sleep(100 * time.Millisecond)
							break
						}
						return err
					}
					processLogLine(line, playerService)
				}
			}

		case err := <-watcher.Errors:
			log.Error().Err(err).Msg("watcher error")
		}
	}
}

func processLogLine(line string, playerService *application.PlayerService) {
	log.Debug().Str("line", line).Msg("Checking log line")

	if strings.Contains(line, "joined the game") {
		if name := application.ExtractPlayerName(line); name != "" {
			if err := playerService.PlayerJoined(name); err != nil {
				log.Error().Err(err).Msg("通知失敗")
			}
		}
	} else if strings.Contains(line, "left the game") {
		if name := application.ExtractPlayerName(line); name != "" {
			if err := playerService.PlayerLeft(name); err != nil {
				log.Error().Err(err).Msg("通知失敗")
			}
		}
	}
}
