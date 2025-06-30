package stat

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ip2location/ip2location-go/v9"
	"gorm.io/driver/clickhouse"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ==================== 全局变量定义 ====================

// GeoIP 数据库
var GeoIPDB *ip2location.DB
var GeoIPDBPath = "./IP2LOCATION-LITE-DB11.BIN"

// 数据库连接
var ClickHouseDB *gorm.DB // ClickHouse 统计明细库（可选）
var MysqlDB *gorm.DB      // 业务管理库（如 MySQL/PG）
var StatDB *gorm.DB       // 统计数据库（ClickHouse 或 MySQL）

// 应用配置
var AppConfig *Config

// 脚本缓存
var (
	scriptCache    = make(map[string]string)
	scriptCacheMux sync.RWMutex
	scriptTemplate *template.Template
)

// ==================== 配置结构体 ====================

// 应用主配置
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	GeoIP    GeoIPConfig
}

// 服务器配置
type ServerConfig struct {
	Port string
	Mode string // debug, release
}

// 数据库配置
type DatabaseConfig struct {
	MySQL      MySQLDBConfig
	ClickHouse ClickHouseDBConfig
}

// GeoIP 配置
type GeoIPConfig struct {
	DBPath string
}

// ClickHouse 数据库配置
type ClickHouseDBConfig struct {
	DSN string
}

// MySQL 数据库配置
type MySQLDBConfig struct {
	DSN                       string
	DefaultStringSize         uint
	DisableDatetimePrecision  bool
	DontSupportRenameIndex    bool
	DontSupportRenameColumn   bool
	SkipInitializeWithVersion bool
}

// ==================== 初始化函数 ====================

// InitApp 应用主初始化函数
func InitApp() {
	// 1. 加载配置
	loadConfig()

	// 2. 初始化数据库（同步进行）
	initDatabases()

	// 3. 初始化 GeoIP
	go InitGeoIP()

	// 4. 初始化脚本模板
	initScriptTemplate()

	log.Println("应用初始化完成")
}

// loadConfig 加载应用配置
func loadConfig() {
	AppConfig = &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
			Mode: getEnv("GIN_MODE", "debug"),
		},
		Database: DatabaseConfig{
			MySQL: MySQLDBConfig{
				DSN:                       getEnv("MYSQL_DSN", "root:password@tcp(localhost:3306)/stat?charset=utf8mb4&parseTime=True&loc=Local"),
				DefaultStringSize:         191,
				DisableDatetimePrecision:  true,
				DontSupportRenameIndex:    true,
				DontSupportRenameColumn:   true,
				SkipInitializeWithVersion: false,
			},
			ClickHouse: ClickHouseDBConfig{
				DSN: getEnv("CLICKHOUSE_DSN", "clickhouse://default:@localhost:9000/stat?dial_timeout=10s&read_timeout=20s"),
			},
		},
		GeoIP: GeoIPConfig{
			DBPath: getEnv("GEOIP_DB_PATH", "./IP2LOCATION-LITE-DB11.BIN"),
		},
	}

	// 更新全局变量
	GeoIPDBPath = AppConfig.GeoIP.DBPath

	log.Println("配置加载完成")
}

// initDatabases 初始化所有数据库连接
func initDatabases() {
	// 初始化 MySQL（必需）
	InitMySQL(AppConfig.Database.MySQL)

	// 初始化 ClickHouse（可选，优先使用）
	InitClickHouse(AppConfig.Database.ClickHouse)

	// 设置统计数据库：优先 ClickHouse，备选 MySQL
	if ClickHouseDB != nil {
		StatDB = ClickHouseDB
		log.Println("✅ 使用 ClickHouse 作为统计数据库")
	} else {
		StatDB = MysqlDB
		log.Println("⚠️  ClickHouse 不可用，使用 MySQL 作为统计数据库")
		// 在 MySQL 中创建统计表作为备选
		initMySQLStatTables()
	}
}

// waitForDB 等待数据库就绪
func waitForDB(dsn string, driver string, maxRetry int) error {
	log.Printf("🔄 等待 %s 数据库就绪...", driver)

	for i := 0; i < maxRetry; i++ {
		var db *gorm.DB
		var err error

		switch driver {
		case "mysql":
			db, err = gorm.Open(mysql.New(mysql.Config{
				DSN: dsn,
			}), &gorm.Config{})
		case "clickhouse":
			db, err = gorm.Open(clickhouse.Open(dsn), &gorm.Config{})
		}

		if err == nil {
			// 测试连接
			sqlDB, err := db.DB()
			if err == nil {
				err = sqlDB.Ping()
				if err == nil {
					log.Printf("✅ %s 数据库连接成功", driver)
					return nil
				}
			}
		}

		log.Printf("⏳ %s 数据库连接失败，重试中... (%d/%d)", driver, i+1, maxRetry)
		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("%s 数据库连接超时，已重试 %d 次", driver, maxRetry)
}

// initMySQLStatTables 在 MySQL 中初始化统计表
func initMySQLStatTables() {
	if MysqlDB == nil {
		log.Printf("❌ MySQL 不可用，无法创建统计表")
		return
	}

	log.Println("📊 在 MySQL 中创建统计表...")
	err := MysqlDB.AutoMigrate(&PageView{}, &StatEvent{})
	if err != nil {
		log.Printf("❌ MySQL 统计表迁移失败: %v", err)
	} else {
		log.Println("✅ MySQL 统计表迁移完成")
	}
}

// initScriptTemplate 初始化脚本模板
func initScriptTemplate() {
	// 读取脚本模板文件
	scriptContent, err := os.ReadFile("templates/stat.js.tmpl")
	if err != nil {
		log.Printf("警告: 无法读取脚本模板文件: %v", err)
		return
	}

	// 解析模板
	scriptTemplate, err = template.New("stat").Parse(string(scriptContent))
	if err != nil {
		log.Printf("警告: 无法解析脚本模板: %v", err)
		return
	}

	log.Println("脚本模板初始化成功")
}

// InitGeoIP 初始化 GeoIP 数据库
func InitGeoIP() {
	if _, err := os.Stat(GeoIPDBPath); os.IsNotExist(err) {
		log.Printf("警告: GeoIP 数据库文件不存在: %s", GeoIPDBPath)
		return
	}

	db, err := ip2location.OpenDB(GeoIPDBPath)
	if err != nil {
		log.Printf("警告: 无法打开 GeoIP 数据库: %v", err)
		return
	}

	GeoIPDB = db
	log.Println("GeoIP 数据库初始化成功")
}

// InitClickHouse 初始化 ClickHouse 数据库
func InitClickHouse(cfg ClickHouseDBConfig) {
	// 等待 ClickHouse 就绪
	if err := waitForDB(cfg.DSN, "clickhouse", 3); err != nil {
		log.Printf("⚠️  ClickHouse 连接失败: %v", err)
		log.Printf("💡 系统将自动使用 MySQL 作为统计数据库")
		return
	}

	db, err := gorm.Open(clickhouse.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		log.Printf("⚠️  ClickHouse 连接失败: %v", err)
		log.Printf("💡 系统将自动使用 MySQL 作为统计数据库")
		return
	}

	ClickHouseDB = db
	log.Println("ClickHouse 连接成功")

	// ClickHouse 迁移统计表
	err = db.AutoMigrate(&PageView{}, &StatEvent{})
	if err != nil {
		log.Printf("⚠️  ClickHouse 表迁移失败: %v", err)
		log.Printf("💡 系统将自动使用 MySQL 作为统计数据库")
		ClickHouseDB = nil // 重置连接
		return
	}

	log.Println("ClickHouse 统计表迁移完成")
}

// InitMySQL 初始化 MySQL 数据库
func InitMySQL(cfg MySQLDBConfig) {
	// 等待 MySQL 就绪
	if err := waitForDB(cfg.DSN, "mysql", 30); err != nil {
		log.Printf("❌ MySQL 连接失败: %v", err)
		return
	}

	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       cfg.DSN,
		DefaultStringSize:         cfg.DefaultStringSize,
		DisableDatetimePrecision:  cfg.DisableDatetimePrecision,
		DontSupportRenameIndex:    cfg.DontSupportRenameIndex,
		DontSupportRenameColumn:   cfg.DontSupportRenameColumn,
		SkipInitializeWithVersion: cfg.SkipInitializeWithVersion,
	}), &gorm.Config{})
	if err != nil {
		log.Printf("❌ MySQL 连接失败: %v", err)
		return
	}

	MysqlDB = db
	log.Println("MySQL 连接成功")

	// 迁移管理表和用户相关表
	err = db.AutoMigrate(&StatSite{}, &User{}, &VerifyCode{}, &UserSession{})
	if err != nil {
		log.Printf("❌ MySQL 迁移失败: %v", err)
	} else {
		log.Println("MySQL 管理表迁移完成")
	}
}

// ==================== 数据库设置函数（保持向后兼容） ====================

// SetMysqlDB 设置 MySQL 数据库（向后兼容）
func SetMysqlDB(db *gorm.DB) {
	MysqlDB = db
	// 迁移管理表和用户相关表
	err := db.AutoMigrate(&StatSite{}, &User{}, &VerifyCode{}, &UserSession{})
	if err != nil {
		log.Printf("MySQL 迁移失败: %v", err)
	}
}

// SetClickHouseDB 设置 ClickHouse 数据库（向后兼容）
func SetClickHouseDB(db *gorm.DB) {
	ClickHouseDB = db
	// ClickHouse 只迁移明细表
	err := db.AutoMigrate(&PageView{}, &StatEvent{})
	if err != nil {
		log.Printf("ClickHouse 迁移失败: %v", err)
	}
}

// ==================== 脚本缓存管理 ====================

// GenerateScriptContent 生成统计脚本内容
func GenerateScriptContent(siteID string, apiURL string) string {
	cacheKey := siteID + "_" + apiURL

	// 检查缓存
	scriptCacheMux.RLock()
	if cached, exists := scriptCache[cacheKey]; exists {
		scriptCacheMux.RUnlock()
		return cached
	}
	scriptCacheMux.RUnlock()

	// 生成新脚本
	var script string
	if scriptTemplate != nil {
		// 使用模板生成
		var buf strings.Builder
		err := scriptTemplate.Execute(&buf, map[string]string{
			"SiteID": siteID,
			"ApiURL": apiURL,
		})
		if err == nil {
			script = buf.String()
		}
	} else {
		// 回退到静态文件
		script = getStaticScript()
	}

	// 缓存脚本
	scriptCacheMux.Lock()
	scriptCache[cacheKey] = script
	scriptCacheMux.Unlock()

	return script
}

// getStaticScript 获取静态脚本内容
func getStaticScript() string {
	content, err := os.ReadFile("static/js/stat.js")
	if err != nil {
		return "console.error('Failed to load stat script');"
	}
	return string(content)
}

// ClearScriptCache 清除脚本缓存
func ClearScriptCache() {
	scriptCacheMux.Lock()
	scriptCache = make(map[string]string)
	scriptCacheMux.Unlock()
	log.Println("脚本缓存已清除")
}

// GetScriptCacheSize 获取脚本缓存大小
func GetScriptCacheSize() int {
	scriptCacheMux.RLock()
	defer scriptCacheMux.RUnlock()
	return len(scriptCache)
}

// ==================== 辅助函数 ====================

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// InitGinEngine 初始化 Gin 引擎
func InitGinEngine() *gin.Engine {
	// 设置 Gin 模式
	gin.SetMode(AppConfig.Server.Mode)

	// 创建引擎
	r := gin.Default()

	// 加载HTML模板
	r.LoadHTMLGlob("templates/*.html")

	return r
}

// ==================== 健康检查 ====================

// HealthCheck 健康检查
func HealthCheck() map[string]interface{} {
	status := map[string]interface{}{
		"status":  "ok",
		"message": "统计服务运行正常",
		"config": map[string]interface{}{
			"server": AppConfig.Server,
			"database": map[string]interface{}{
				"mysql": map[string]interface{}{
					"available": IsMySQLAvailable(),
					"type":      "management",
				},
				"clickhouse": map[string]interface{}{
					"available": IsClickHouseAvailable(),
					"type":      "statistics",
				},
				"stat_db": map[string]interface{}{
					"available": StatDB != nil,
					"type":      GetStatDBType(),
				},
			},
			"geoip": GeoIPDB != nil,
		},
		"cache": map[string]interface{}{
			"script_cache_size": GetScriptCacheSize(),
		},
	}

	return status
}

// GetStatDBType 获取当前统计数据库类型
func GetStatDBType() string {
	if StatDB == nil {
		return "none"
	}
	if StatDB == ClickHouseDB {
		return "clickhouse"
	}
	if StatDB == MysqlDB {
		return "mysql"
	}
	return "unknown"
}

// IsClickHouseAvailable 检查 ClickHouse 是否可用
func IsClickHouseAvailable() bool {
	return ClickHouseDB != nil
}

// IsMySQLAvailable 检查 MySQL 是否可用
func IsMySQLAvailable() bool {
	return MysqlDB != nil
}
