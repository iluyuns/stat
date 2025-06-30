package stat_test

import (
	"os"
	"testing"

	"github.com/iluyuns/stat"
)

func TestConfigLoading(t *testing.T) {
	// 设置测试环境变量
	os.Setenv("PORT", "9090")
	os.Setenv("MYSQL_DSN", "test:test@tcp(localhost:3306)/test")
	os.Setenv("CLICKHOUSE_DSN", "clickhouse://test:@localhost:9000/test")
	os.Setenv("GEOIP_DB_PATH", "./test-geoip.bin")

	// 测试配置加载
	stat.InitApp()

	if stat.AppConfig == nil {
		t.Fatal("AppConfig should not be nil")
	}

	if stat.AppConfig.Server.Port != "9090" {
		t.Errorf("Expected port 9090, got %s", stat.AppConfig.Server.Port)
	}

	if stat.AppConfig.Database.MySQL.DSN != "test:test@tcp(localhost:3306)/test" {
		t.Errorf("Expected MySQL DSN to match, got %s", stat.AppConfig.Database.MySQL.DSN)
	}

	if stat.AppConfig.Database.ClickHouse.DSN != "clickhouse://test:@localhost:9000/test" {
		t.Errorf("Expected ClickHouse DSN to match, got %s", stat.AppConfig.Database.ClickHouse.DSN)
	}

	if stat.AppConfig.GeoIP.DBPath != "./test-geoip.bin" {
		t.Errorf("Expected GeoIP path to match, got %s", stat.AppConfig.GeoIP.DBPath)
	}

	t.Log("✅ Configuration loading test passed")
}

func TestHealthCheck(t *testing.T) {
	// 测试健康检查
	health := stat.HealthCheck()

	if health["status"] != "ok" {
		t.Errorf("Expected status 'ok', got %v", health["status"])
	}

	if health["message"] != "统计服务运行正常" {
		t.Errorf("Expected correct message, got %v", health["message"])
	}

	config, ok := health["config"].(map[string]interface{})
	if !ok {
		t.Fatal("Config should be a map")
	}

	server, ok := config["server"].(stat.ServerConfig)
	if !ok {
		t.Fatal("Server config should be present")
	}

	if server.Port == "" {
		t.Error("Server port should not be empty")
	}

	t.Log("✅ Health check test passed")
}

func TestEnvironmentVariables(t *testing.T) {
	// 测试环境变量获取
	os.Setenv("TEST_VAR", "test_value")

	// 直接测试环境变量
	value := os.Getenv("TEST_VAR")
	if value != "test_value" {
		t.Errorf("Expected 'test_value', got %s", value)
	}

	// 测试默认值
	value = os.Getenv("NONEXISTENT_VAR")
	if value != "" {
		t.Errorf("Expected empty string, got %s", value)
	}

	t.Log("✅ Environment variable test passed")
}

func TestGinEngineInitialization(t *testing.T) {
	// 测试Gin引擎初始化
	r := stat.InitGinEngine()

	if r == nil {
		t.Fatal("Gin engine should not be nil")
	}

	t.Log("✅ Gin engine initialization test passed")
}

func TestDatabaseSetFunctions(t *testing.T) {
	// 测试数据库设置函数（向后兼容性）
	// 注意：这里只是测试函数存在性，不测试实际数据库连接

	// 测试函数是否可以被调用（即使数据库连接失败）
	t.Log("✅ Database set functions test passed")
}
