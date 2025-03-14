package embedweb

import (
	"os"
	"path"

	// "gorm.io/driver/sqlite" // Sqlite driver based on CGO
	sqlite "github.com/glebarez/sqlite" // Pure go SQLite driver, checkout https://github.com/glebarez/sqlite for details
	gorm "gorm.io/gorm"
)

func (ew *EmbedWeb) initDB() {
	dbPath := path.Join(baseDirPath, embedDbFile)
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		ew.log.Fatalf("open embed database error: %v", err)
		os.Exit(1)
	}
	ew.db = db
}

func (ew *EmbedWeb) GetDB() *gorm.DB {
	return ew.db
}
