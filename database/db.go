package database

import (
	"bytes"
	"io"
	"io/fs"
	"log"
	"os"
	"path"
	"slices"

	"github.com/agassiz/3x-ui/v2/config"
	"github.com/agassiz/3x-ui/v2/database/model"
	"github.com/agassiz/3x-ui/v2/util/crypto"
	"github.com/agassiz/3x-ui/v2/util/random"
	"github.com/agassiz/3x-ui/v2/xray"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

const (
	defaultUsername = "admin"
	defaultPassword = "admin"
)

func initModels() error {
	models := []any{
		&model.User{},
		&model.Inbound{},
		&model.OutboundTraffics{},
		&model.Setting{},
		&model.InboundClientIps{},
		&xray.ClientTraffic{},
		&model.HistoryOfSeeders{},
		&model.ClashSubscription{},
	}
	for _, model := range models {
		if err := db.AutoMigrate(model); err != nil {
			log.Printf("Error auto migrating model: %v", err)
			return err
		}
	}
	return nil
}

// initUser creates a default admin user if the users table is empty.
func initUser() error {
	empty, err := isTableEmpty("users")
	if err != nil {
		log.Printf("Error checking if users table is empty: %v", err)
		return err
	}
	if empty {
		hashedPassword, err := crypto.HashPasswordAsBcrypt(defaultPassword)

		if err != nil {
			log.Printf("Error hashing default password: %v", err)
			return err
		}

		user := &model.User{
			Username: defaultUsername,
			Password: hashedPassword,
		}
		return db.Create(user).Error
	}
	return nil
}

// runSeeders migrates user passwords to bcrypt and records seeder execution to prevent re-running.
func runSeeders(isUsersEmpty bool) error {
	empty, err := isTableEmpty("history_of_seeders")
	if err != nil {
		log.Printf("Error checking if users table is empty: %v", err)
		return err
	}

	if empty && isUsersEmpty {
		hashSeeder := &model.HistoryOfSeeders{
			SeederName: "UserPasswordHash",
		}
		return db.Create(hashSeeder).Error
	} else {
		var seedersHistory []string
		db.Model(&model.HistoryOfSeeders{}).Pluck("seeder_name", &seedersHistory)

		if !slices.Contains(seedersHistory, "UserPasswordHash") && !isUsersEmpty {
			var users []model.User
			db.Find(&users)

			for _, user := range users {
				hashedPassword, err := crypto.HashPasswordAsBcrypt(user.Password)
				if err != nil {
					log.Printf("Error hashing password for user '%s': %v", user.Username, err)
					return err
				}
				db.Model(&user).Update("password", hashedPassword)
			}

			hashSeeder := &model.HistoryOfSeeders{
				SeederName: "UserPasswordHash",
			}
			return db.Create(hashSeeder).Error
		}
	}

	return nil
}

func initDefaultSettings() error {
	empty, err := isTableEmpty("settings")
	if err != nil {
		log.Printf("Error checking if settings table is empty: %v", err)
		return err
	}

	// 只在settings表为空时才初始化默认配置
	if empty {
		log.Println("Initializing default settings...")

		// 导入service包来获取默认配置
		// 注意：这里我们需要避免循环导入，所以直接定义默认配置
		defaultSettings := map[string]string{
			"secret":                      random.Seq(32),
			"webListen":                   "",
			"webDomain":                   "",
			"webPort":                     "2053",
			"webCertFile":                 "",
			"webKeyFile":                  "",
			"webBasePath":                 "/",
			"sessionMaxAge":               "60",
			"pageSize":                    "50",
			"expireDiff":                  "0",
			"trafficDiff":                 "0",
			"remarkModel":                 "-ieo",
			"timeLocation":                "Local",
			"tgBotEnable":                 "false",
			"tgBotToken":                  "",
			"tgBotProxy":                  "",
			"tgBotAPIServer":              "",
			"tgBotChatId":                 "",
			"tgRunTime":                   "@daily",
			"tgBotBackup":                 "false",
			"tgBotLoginNotify":            "true",
			"tgCpu":                       "80",
			"tgLang":                      "en-US",
			"twoFactorEnable":             "false",
			"twoFactorToken":              "",
			"subEnable":                   "false",
			"subTitle":                    "",
			"subListen":                   "",
			"subPort":                     "2096",
			"subPath":                     "/sub/",
			"subJsonPath":                 "/json/",
			"subDomain":                   "",
			"subCertFile":                 "",
			"subKeyFile":                  "",
			"subUpdates":                  "12",
			"subEncrypt":                  "true",
			"subShowInfo":                 "true",
			"subURI":                      "",
			"subJsonURI":                  "",
			"subJsonFragment":             "",
			"subJsonNoises":               "",
			"subJsonMux":                  "",
			"subJsonRules":                "",
			"datepicker":                  "gregorian",
			"warp":                        "",
			"externalTrafficInformEnable": "false",
			"externalTrafficInformURI":    "",
		}

		// 批量插入默认配置
		for key, value := range defaultSettings {
			setting := &model.Setting{
				Key:   key,
				Value: value,
			}
			if err := db.Create(setting).Error; err != nil {
				log.Printf("Error creating default setting %s: %v", key, err)
				return err
			}
		}

		// 注意：不再在数据库初始化时设置xrayTemplateConfig
		// xrayTemplateConfig应该从实际的Xray配置中动态生成和更新

		log.Printf("Successfully initialized %d default settings + xrayTemplateConfig", len(defaultSettings))
	}

	return nil
}

func isTableEmpty(tableName string) (bool, error) {
	var count int64
	err := db.Table(tableName).Count(&count).Error
	return count == 0, err
}

// InitDB sets up the database connection, migrates models, and runs seeders.
func InitDB(dbPath string) error {
	dir := path.Dir(dbPath)
	err := os.MkdirAll(dir, fs.ModePerm)
	if err != nil {
		return err
	}

	var gormLogger logger.Interface

	if config.IsDebug() {
		gormLogger = logger.Default
	} else {
		gormLogger = logger.Discard
	}

	c := &gorm.Config{
		Logger: gormLogger,
	}
	db, err = gorm.Open(sqlite.Open(dbPath), c)
	if err != nil {
		return err
	}

	if err := initModels(); err != nil {
		return err
	}

	isUsersEmpty, err := isTableEmpty("users")
	if err != nil {
		log.Printf("Error checking if users table is empty: %v, assuming table is empty", err)
		isUsersEmpty = true // 假设表为空，继续初始化流程
	}

	if err := initUser(); err != nil {
		return err
	}

	if err := initDefaultSettings(); err != nil {
		return err
	}

	return runSeeders(isUsersEmpty)
}

// CloseDB closes the database connection if it exists.
func CloseDB() error {
	if db != nil {
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// GetDB returns the global GORM database instance.
func GetDB() *gorm.DB {
	return db
}

// IsNotFound checks if the given error is a GORM record not found error.
func IsNotFound(err error) bool {
	return err == gorm.ErrRecordNotFound
}

// IsSQLiteDB checks if the given file is a valid SQLite database by reading its signature.
func IsSQLiteDB(file io.ReaderAt) (bool, error) {
	signature := []byte("SQLite format 3\x00")
	buf := make([]byte, len(signature))
	_, err := file.ReadAt(buf, 0)
	if err != nil {
		return false, err
	}
	return bytes.Equal(buf, signature), nil
}

// Checkpoint performs a WAL checkpoint on the SQLite database to ensure data consistency.
func Checkpoint() error {
	// Update WAL
	err := db.Exec("PRAGMA wal_checkpoint;").Error
	if err != nil {
		return err
	}
	return nil
}
