package eweb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"time"

	// "gorm.io/driver/sqlite" // Sqlite driver based on CGO
	sqlite "github.com/glebarez/sqlite" // Pure go SQLite driver, checkout https://github.com/glebarez/sqlite for details
	"github.com/sirupsen/logrus"
	gorm "gorm.io/gorm"
)

var configModel = new(configWrap)

type configWrap struct {
	Id        int       `gorm:"primaryKey;type:INTEGER PRIMARY KEY AUTOINCREMENT"` // ID
	Data      []byte    `gorm:"type:TEXT"`                                         // 配置
	UpdatedAt time.Time `gorm:"type:DATETIME"`                                     // 更新时间
}

func (cw configWrap) TableName() string {
	return "eweb_config"
}

func (ew *EmbedWeb) initEmbedDB() {
	embedDbPath := path.Join(baseDirPath, embedDbFile)
	db, err := gorm.Open(sqlite.Open(embedDbPath), &gorm.Config{})
	if err != nil {
		ew.embedLog.Fatalf("open embed database error: %v", err)
	}
	ew.embedDB = db
}

func (ew *EmbedWeb) initConfig() {
	err := ew.embedDB.AutoMigrate(configModel)
	if err != nil {
		ew.embedLog.Fatalf("init embed config table error: %v", err)
	}
	ew.config, err = ew.getConfig(context.Background())
	if err != nil {
		ew.embedLog.Fatalf("init embed config error: %v", err)
	}
}

type config struct {
	Port     int    `json:"port,omitempty"`
	Public   bool   `json:"public,omitempty"`
	LogLevel string `json:"log_level,omitempty"`
}

func (ew *EmbedWeb) getConfig(ctx context.Context) (*config, error) {
	cfg := new(config)
	raw, _ := json.Marshal(cfg)
	var cfgWrap = &configWrap{Id: 1, Data: raw}
	err := ew.embedDB.WithContext(ctx).Where("id = 1").FirstOrCreate(cfgWrap).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	} else if err != nil {
		return nil, err
	}
	err = json.Unmarshal(cfgWrap.Data, cfg)
	if err != nil {
		err = ew.embedDB.WithContext(ctx).Where("id = 1").Save(cfgWrap).Error
		if err != nil {
			return nil, err
		}
	}
	return cfg, nil
}

func (ew *EmbedWeb) setConfigField(ctx context.Context, field string, value any) error {
	switch v := value.(type) {
	case []byte:
		str := string(v)
		expr := gorm.Expr(fmt.Sprintf("json_set(data, '$.%s', json('%s'))", field, str))
		updates := map[string]any{"data": expr, "updated_at": time.Now()}
		return ew.embedDB.WithContext(ctx).Model(configModel).Where("id = 1").Updates(updates).Error
	case json.RawMessage:
		str := string(v)
		expr := gorm.Expr(fmt.Sprintf("json_set(data, '$.%s', json('%s'))", field, str))
		updates := map[string]any{"data": expr, "updated_at": time.Now()}
		return ew.embedDB.WithContext(ctx).Model(configModel).Where("id = 1").Updates(updates).Error
	case map[string]any, []any, bool:
		raw, _ := json.Marshal(v)
		str := string(raw)
		expr := gorm.Expr(fmt.Sprintf("json_set(data, '$.%s', json('%s'))", field, str))
		updates := map[string]any{"data": expr, "updated_at": time.Now()}
		return ew.embedDB.WithContext(ctx).Model(configModel).Where("id = 1").Updates(updates).Error
	default:
		expr := gorm.Expr(fmt.Sprintf("json_set(data, '$.%s', ?)", field), v)
		updates := map[string]any{"data": expr, "updated_at": time.Now()}
		return ew.embedDB.WithContext(ctx).Model(configModel).Where("id = 1").Updates(updates).Error
	}
}

func (ew *EmbedWeb) GetWebServerPort() int {
	return ew.config.Port
}

func (ew *EmbedWeb) GetWebServerPublic() bool {
	return ew.config.Public
}

func (ew *EmbedWeb) GetLogLevel() string {
	return ew.config.LogLevel
}

func (ew *EmbedWeb) SetWebServerPort(ctx context.Context, port int) error {
	ew.embedLog.Infof("set embed web server port to %d", port)
	ew.config.Port = port
	return ew.setConfigField(ctx, "port", port)
}

func (ew *EmbedWeb) SetWebServerPublic(ctx context.Context, public bool) error {
	if public {
		ew.embedLog.Infof("set embed web server public")
	} else {
		ew.embedLog.Infof("set embed web server private")
	}
	ew.config.Public = public
	return ew.setConfigField(ctx, "public", public)
}

func (ew *EmbedWeb) SetLogLevel(ctx context.Context, level string) error {
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}
	ew.embedLog.Infof("set app log level to %s", level)
	ew.config.LogLevel = level
	ew.log.SetLevel(logLevel)
	return ew.setConfigField(ctx, "log_level", level)
}
