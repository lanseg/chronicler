package storage

type Settings struct {
	Root string `json:"root" env:"STORAGE_ROOT"`
}
