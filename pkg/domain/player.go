package domain

type Player struct {
	Name string
}

type PlayerRepository interface {
	Add(name string) error
	Remove(name string) error
	List() ([]Player, error)
}
