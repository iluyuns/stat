package stat_test

import (
	"testing"

	"github.com/iluyuns/stat"
)

func TestDatabaseCompatibility(t *testing.T) {
	// 初始化应用
	stat.InitApp()

	// 测试数据库类型检测
	dbType := stat.GetStatDBType()
	t.Logf("当前统计数据库类型: %s", dbType)

	// 测试数据库可用性
	mysqlAvailable := stat.IsMySQLAvailable()
	clickhouseAvailable := stat.IsClickHouseAvailable()

	t.Logf("MySQL 可用: %v", mysqlAvailable)
	t.Logf("ClickHouse 可用: %v", clickhouseAvailable)

	// 验证逻辑：至少应该有一个数据库可用
	if !mysqlAvailable && !clickhouseAvailable {
		t.Error("至少应该有一个数据库可用")
	}

	t.Log("✅ 数据库兼容性测试通过")
}

func TestHealthCheckDatabaseInfo(t *testing.T) {
	// 初始化应用
	stat.InitApp()

	// 测试健康检查中的数据库信息
	health := stat.HealthCheck()

	// 检查数据库信息结构
	database, ok := health["config"].(map[string]interface{})["database"].(map[string]interface{})
	if !ok {
		t.Fatal("健康检查中应该包含数据库信息")
	}

	// 检查 MySQL 信息
	mysql, ok := database["mysql"].(map[string]interface{})
	if !ok {
		t.Fatal("应该包含 MySQL 信息")
	}
	if mysql["type"] != "management" {
		t.Error("MySQL 类型应该是 management")
	}

	// 检查 ClickHouse 信息
	clickhouse, ok := database["clickhouse"].(map[string]interface{})
	if !ok {
		t.Fatal("应该包含 ClickHouse 信息")
	}
	if clickhouse["type"] != "statistics" {
		t.Error("ClickHouse 类型应该是 statistics")
	}

	// 检查统计数据库信息
	statDB, ok := database["stat_db"].(map[string]interface{})
	if !ok {
		t.Fatal("应该包含统计数据库信息")
	}

	dbType, ok := statDB["type"].(string)
	if !ok {
		t.Fatal("统计数据库类型应该是字符串")
	}

	t.Logf("统计数据库类型: %s", dbType)

	// 验证统计数据库类型应该是 clickhouse 或 mysql
	if dbType != "clickhouse" && dbType != "mysql" && dbType != "none" {
		t.Errorf("统计数据库类型应该是 clickhouse、mysql 或 none，实际是: %s", dbType)
	}

	t.Log("✅ 健康检查数据库信息测试通过")
}
