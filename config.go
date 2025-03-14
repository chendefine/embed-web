package embedweb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	gorm "gorm.io/gorm"
)

func (ew *EmbedWeb) initConfig() {
	err := ew.db.AutoMigrate(configModel)
	if err != nil {
		ew.log.Fatalf("init embed config table error: %v", err)
		os.Exit(1)
	}
	ew.cfg, err = ew.getConfig(context.Background())
	if err != nil {
		ew.log.Fatalf("init embed config error: %v", err)
		os.Exit(1)
	}
}

func (ew *EmbedWeb) GetPort() int {
	return ew.cfg.Port
}

func (ew *EmbedWeb) GetPublic() bool {
	return ew.cfg.Public
}

func (ew *EmbedWeb) GetLogLevel() string {
	return ew.log.level.String()
}

func (ew *EmbedWeb) SetPort(port int) {
	ew.log.Infof("set embed-web listen on port: %d", port)
	ew.cfg.Port = port
	ew.setConfigField(context.Background(), "port", port)
}

func (ew *EmbedWeb) SetPublic(public bool) {
	if public {
		ew.log.Infof("set embed-web serve public")
	} else {
		ew.log.Infof("set embed-web serve localhost")
	}
	ew.cfg.Public = public
	ew.setConfigField(context.Background(), "public", public)
}

func (ew *EmbedWeb) SetLogLevel(level string) {
	ew.log.level, _ = logrus.ParseLevel(level)
	ew.setConfigField(context.Background(), "log_level", level)
}

type config struct {
	Port     int    `json:"port,omitempty"`
	Public   bool   `json:"public,omitempty"`
	LogLevel string `json:"log_level,omitempty"`
}

var configModel = new(configWrap)

type configWrap struct {
	Id        int       `gorm:"primaryKey;type:INTEGER PRIMARY KEY AUTOINCREMENT"` // ID
	Data      []byte    `gorm:"type:TEXT"`                                         // 配置
	UpdatedAt time.Time `gorm:"type:DATETIME"`                                     // 更新时间
}

func (cw configWrap) TableName() string {
	return "eweb_config"
}

func (ew *EmbedWeb) getConfig(ctx context.Context) (*config, error) {
	cfg := new(config)
	raw, _ := json.Marshal(cfg)
	var cfgWrap = &configWrap{Id: 1, Data: raw}
	err := ew.db.WithContext(ctx).Where("id = 1").FirstOrCreate(cfgWrap).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	} else if err != nil {
		return nil, err
	}
	err = json.Unmarshal(cfgWrap.Data, cfg)
	if err != nil {
		err = ew.db.WithContext(ctx).Where("id = 1").Save(cfgWrap).Error
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
		return ew.db.WithContext(ctx).Model(configModel).Where("id = 1").Updates(updates).Error
	case json.RawMessage:
		str := string(v)
		expr := gorm.Expr(fmt.Sprintf("json_set(data, '$.%s', json('%s'))", field, str))
		updates := map[string]any{"data": expr, "updated_at": time.Now()}
		return ew.db.WithContext(ctx).Model(configModel).Where("id = 1").Updates(updates).Error
	case map[string]any, []any, bool:
		raw, _ := json.Marshal(v)
		str := string(raw)
		expr := gorm.Expr(fmt.Sprintf("json_set(data, '$.%s', json('%s'))", field, str))
		updates := map[string]any{"data": expr, "updated_at": time.Now()}
		return ew.db.WithContext(ctx).Model(configModel).Where("id = 1").Updates(updates).Error
	default:
		expr := gorm.Expr(fmt.Sprintf("json_set(data, '$.%s', ?)", field), v)
		updates := map[string]any{"data": expr, "updated_at": time.Now()}
		return ew.db.WithContext(ctx).Model(configModel).Where("id = 1").Updates(updates).Error
	}
}
