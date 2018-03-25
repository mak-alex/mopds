package models

type DBConfig struct {
	DBType   string `arg:"-t" help:"type db: sqlite/postgres/mysql"`
	DBLog    bool   `json:"db-log"`
	DBParams string `json:"db-params"`
}
