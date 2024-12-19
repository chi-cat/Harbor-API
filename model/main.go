package model

import (
	"log"
	"one-api/common"
	"os"
	"strings"
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

var LOG_DB *gorm.DB

func createRootAccountIfNeed() error {
	var user User
	//if user.Status != common.UserStatusEnabled {
	if err := DB.First(&user).Error; err != nil {
		common.SysLog("no user exists, create a root user for you: username is root, password is 123456")
		hashedPassword, err := common.Password2Hash("123456")
		if err != nil {
			return err
		}
		rootUser := User{
			Username:    "root",
			Password:    hashedPassword,
			Role:        common.RoleRootUser,
			Status:      common.UserStatusEnabled,
			DisplayName: "Root User",
			AccessToken: nil,
			Quota:       100000000,
		}
		DB.Create(&rootUser)
	}
	return nil
}

func chooseDB(envName string) (*gorm.DB, error) {
	dsn := os.Getenv(envName)
	if dsn != "" {
		if strings.HasPrefix(dsn, "postgres://") {
			// Use PostgreSQL
			common.SysLog("using PostgreSQL as database")
			common.UsingPostgreSQL = true
			return gorm.Open(postgres.New(postgres.Config{
				DSN:                  dsn,
				PreferSimpleProtocol: true, // disables implicit prepared statement usage
			}), &gorm.Config{
				PrepareStmt: true, // precompile SQL
			})
		}
		if strings.HasPrefix(dsn, "local") {
			common.SysLog("SQL_DSN not set, using SQLite as database")
			common.UsingSQLite = true
			return gorm.Open(sqlite.Open(common.SQLitePath), &gorm.Config{
				PrepareStmt: true, // precompile SQL
			})
		}
		// Use MySQL
		common.SysLog("using MySQL as database")
		// check parseTime
		if !strings.Contains(dsn, "parseTime") {
			if strings.Contains(dsn, "?") {
				dsn += "&parseTime=true"
			} else {
				dsn += "?parseTime=true"
			}
		}
		common.UsingMySQL = true
		return gorm.Open(mysql.Open(dsn), &gorm.Config{
			PrepareStmt: true, // precompile SQL
		})
	}
	// Use SQLite
	common.SysLog("SQL_DSN not set, using SQLite as database")
	common.UsingSQLite = true
	return gorm.Open(sqlite.Open(common.SQLitePath), &gorm.Config{
		PrepareStmt: true, // precompile SQL
	})
}

func InitDB() (err error) {
	db, err := chooseDB("SQL_DSN")
	if err == nil {
		if common.DebugEnabled {
			db = db.Debug()
		}
		DB = db
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		sqlDB.SetMaxIdleConns(common.GetEnvOrDefault("SQL_MAX_IDLE_CONNS", 100))
		sqlDB.SetMaxOpenConns(common.GetEnvOrDefault("SQL_MAX_OPEN_CONNS", 1000))
		sqlDB.SetConnMaxLifetime(time.Second * time.Duration(common.GetEnvOrDefault("SQL_MAX_LIFETIME", 60)))

		if !common.IsMasterNode {
			return nil
		}
		//if common.UsingMySQL {
		//	_, _ = sqlDB.Exec("DROP INDEX idx_channels_key ON channels;")             // TODO: delete this line when most users have upgraded
		//	_, _ = sqlDB.Exec("ALTER TABLE midjourneys MODIFY action VARCHAR(40);")   // TODO: delete this line when most users have upgraded
		//	_, _ = sqlDB.Exec("ALTER TABLE midjourneys MODIFY progress VARCHAR(30);") // TODO: delete this line when most users have upgraded
		//	_, _ = sqlDB.Exec("ALTER TABLE midjourneys MODIFY status VARCHAR(20);")   // TODO: delete this line when most users have upgraded
		//}
		common.SysLog("database migration started")
		err = migrateDB()
		return err
	} else {
		common.FatalLog(err)
	}
	return err
}

func InitLogDB() (err error) {
	if os.Getenv("LOG_SQL_DSN") == "" {
		LOG_DB = DB
		return
	}
	db, err := chooseDB("LOG_SQL_DSN")
	if err == nil {
		if common.DebugEnabled {
			db = db.Debug()
		}
		LOG_DB = db
		sqlDB, err := LOG_DB.DB()
		if err != nil {
			return err
		}
		sqlDB.SetMaxIdleConns(common.GetEnvOrDefault("SQL_MAX_IDLE_CONNS", 100))
		sqlDB.SetMaxOpenConns(common.GetEnvOrDefault("SQL_MAX_OPEN_CONNS", 1000))
		sqlDB.SetConnMaxLifetime(time.Second * time.Duration(common.GetEnvOrDefault("SQL_MAX_LIFETIME", 60)))

		if !common.IsMasterNode {
			return nil
		}
		//if common.UsingMySQL {
		//	_, _ = sqlDB.Exec("DROP INDEX idx_channels_key ON channels;")             // TODO: delete this line when most users have upgraded
		//	_, _ = sqlDB.Exec("ALTER TABLE midjourneys MODIFY action VARCHAR(40);")   // TODO: delete this line when most users have upgraded
		//	_, _ = sqlDB.Exec("ALTER TABLE midjourneys MODIFY progress VARCHAR(30);") // TODO: delete this line when most users have upgraded
		//	_, _ = sqlDB.Exec("ALTER TABLE midjourneys MODIFY status VARCHAR(20);")   // TODO: delete this line when most users have upgraded
		//}
		common.SysLog("database migration started")
		err = migrateLOGDB()
		return err
	} else {
		common.FatalLog(err)
	}
	return err
}

func migrateDB() error {
	err := DB.AutoMigrate(&Channel{})
	if err != nil {
		return err
	}
	err = DB.AutoMigrate(&Token{})
	if err != nil {
		return err
	}
	err = DB.AutoMigrate(&User{})
	if err != nil {
		return err
	}
	err = DB.AutoMigrate(&Option{})
	if err != nil {
		return err
	}
	err = DB.AutoMigrate(&Redemption{})
	if err != nil {
		return err
	}
	err = DB.AutoMigrate(&Ability{})
	if err != nil {
		return err
	}
	err = DB.AutoMigrate(&Log{})
	if err != nil {
		return err
	}
	err = DB.AutoMigrate(&Midjourney{})
	if err != nil {
		return err
	}
	err = DB.AutoMigrate(&TopUp{})
	if err != nil {
		return err
	}
	err = DB.AutoMigrate(&QuotaData{})
	if err != nil {
		return err
	}
	err = DB.AutoMigrate(&Task{})
	if err != nil {
		return err
	}
	common.SysLog("database migrated")
	err = createRootAccountIfNeed()
	return err
}

func migrateLOGDB() error {
	var err error
	if err = LOG_DB.AutoMigrate(&Log{}); err != nil {
		return err
	}

	// 检查PromptCacheHitTokens字段是否存在
	var exists bool

	if common.UsingMySQL {
		var result string
		err = LOG_DB.Raw("SELECT COLUMN_NAME FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'logs' AND COLUMN_NAME = 'prompt_cache_hit_tokens'").Scan(&result).Error
		exists = result != ""
	} else if common.UsingPostgreSQL {
		err = LOG_DB.Raw("SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'logs' AND column_name = 'prompt_cache_hit_tokens')").Scan(&exists).Error
	} else {
		err = LOG_DB.Raw("SELECT COUNT(*) > 0 FROM pragma_table_info('logs') WHERE name = 'prompt_cache_hit_tokens'").Scan(&exists).Error
	}

	if err != nil {
		common.SysError("检查PromptCacheHitTokens字段失败: " + err.Error())
		return err
	}

	// 如果字段不存在，添加该字段
	if !exists {
		if common.UsingPostgreSQL {
			err = LOG_DB.Exec("ALTER TABLE logs ADD COLUMN prompt_cache_hit_tokens integer DEFAULT 0").Error
		} else {
			err = LOG_DB.Exec("ALTER TABLE logs ADD COLUMN prompt_cache_hit_tokens int DEFAULT 0").Error
		}
		if err != nil {
			common.SysError("添加PromptCacheHitTokens字段失败: " + err.Error())
			return err
		}
		common.SysLog("成功添加PromptCacheHitTokens字段到logs表")
	} else {
		common.SysLog("PromptCacheHitTokens字段已存在于logs表中")
	}

	return nil
}

func closeDB(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	err = sqlDB.Close()
	return err
}

func CloseDB() error {
	if LOG_DB != DB {
		err := closeDB(LOG_DB)
		if err != nil {
			return err
		}
	}
	return closeDB(DB)
}

var (
	lastPingTime time.Time
	pingMutex    sync.Mutex
)

func PingDB() error {
	pingMutex.Lock()
	defer pingMutex.Unlock()

	if time.Since(lastPingTime) < time.Second*10 {
		return nil
	}

	sqlDB, err := DB.DB()
	if err != nil {
		log.Printf("Error getting sql.DB from GORM: %v", err)
		return err
	}

	err = sqlDB.Ping()
	if err != nil {
		log.Printf("Error pinging DB: %v", err)
		return err
	}

	lastPingTime = time.Now()
	common.SysLog("Database pinged successfully")
	return nil
}
