package domain

type Notifier interface {
	Send(message string) error
}
