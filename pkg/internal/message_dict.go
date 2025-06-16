package internal

import (
	"bufio"
	"log"
	"math/rand"
	"os"
	"strings"
)

var messages []string

func LoadMessagesDict(path string) error {
	log.Printf("[DEBUG] LoadMessagesDict: loading from file: %s", path)
	file, err := os.Open(path)
	if err != nil {
		log.Printf("[ERROR] LoadMessagesDict: failed to open file: %v", err)
		return err
	}
	defer file.Close()

	var result []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			result = append(result, line)
			log.Printf("[DEBUG] LoadMessagesDict: loaded message: %s", line)
		}
	}
	messages = result
	log.Printf("[DEBUG] LoadMessagesDict: total %d messages loaded", len(messages))
	return scanner.Err()
}

func LoadMessagesDictFromEnv(envKey string) error {
	raw := os.Getenv(envKey)
	log.Printf("[DEBUG] LoadMessagesDictFromEnv: loading from env: %s", envKey)
	if raw == "" {
		log.Printf("[WARN] LoadMessagesDictFromEnv: env var %s is empty", envKey)
		return nil // 空でもエラーにしない
	}
	parts := strings.Split(raw, "|")
	messages = nil
	for _, msg := range parts {
		msg = strings.TrimSpace(msg)
		if msg != "" {
			messages = append(messages, msg)
			log.Printf("[DEBUG] LoadMessagesDictFromEnv: loaded message: %s", msg)
		}
	}
	log.Printf("[DEBUG] LoadMessagesDictFromEnv: total %d messages loaded", len(messages))
	return nil
}

func GetRandomMessage() string {
	if len(messages) == 0 {
		log.Printf("[WARN] GetRandomMessage: no messages loaded")
		return ""
	}
	idx := rand.Intn(len(messages))
	log.Printf("[DEBUG] GetRandomMessage: selected index %d, message: %s", idx, messages[idx])
	return messages[idx]
}
