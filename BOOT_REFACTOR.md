# Boot.go 重构说明

## 重构概述

本次重构将所有涉及初始化或全局变量的代码整合到 `boot.go` 文件中，实现了统一的配置管理和初始化流程。

## 主要变更

### 1. 文件整合
- ✅ 删除了 `var.go` 文件
- ✅ 将所有全局变量定义移至 `boot.go`
- ✅ 将所有配置结构体移至 `boot.go`
- ✅ 将所有初始化函数移至 `boot.go`
- ✅ 保留了 `debug_login_test.go` 和 `boot_test.go` 测试文件

### 2. 测试文件说明
- `debug_login_test.go` - 密码验证和登录流程测试
- `boot_test.go` - 配置加载和初始化功能测试
- 所有测试文件都保留用于功能验证

### 2. 新增功能

#### 配置管理
```go
type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    GeoIP    GeoIPConfig
}
```

#### 统一初始化
```go
func InitApp() {
    loadConfig()      // 加载配置
    initDatabases()   // 初始化数据库
    InitGeoIP()       // 初始化GeoIP
}
```

#### 环境变量支持
- `PORT` - 服务器端口（默认：8080）
- `GIN_MODE` - Gin模式（默认：debug）
- `MYSQL_DSN` - MySQL连接字符串
- `CLICKHOUSE_DSN` - ClickHouse连接字符串
- `GEOIP_DB_PATH` - GeoIP数据库路径

### 3. 简化的主程序

#### 重构前（cmd/main.go）
```go
func main() {
    initDatabases()           // 手动初始化数据库
    stat.InitGeoIP()          // 手动初始化GeoIP
    r := gin.Default()        // 手动创建引擎
    r.LoadHTMLGlob("templates/*.html")
    // ... 其他代码
}
```

#### 重构后（cmd/main.go）
```go
func main() {
    stat.InitApp()            // 一键初始化
    r := stat.InitGinEngine() // 获取预配置的引擎
    // ... 其他代码
}
```

## 使用方式

### 1. 基本使用
```go
package main

import "github.com/iluyuns/stat"

func main() {
    // 一键初始化所有组件
    stat.InitApp()
    
    // 获取预配置的Gin引擎
    r := stat.InitGinEngine()
    
    // 启动服务器
    r.Run(":" + stat.AppConfig.Server.Port)
}
```

### 2. 环境变量配置
```bash
# 服务器配置
export PORT=8080
export GIN_MODE=release

# 数据库配置
export MYSQL_DSN="user:pass@tcp(localhost:3306)/stat?charset=utf8mb4&parseTime=True&loc=Local"
export CLICKHOUSE_DSN="clickhouse://default:@localhost:9000/stat?dial_timeout=10s&read_timeout=20s"

# GeoIP配置
export GEOIP_DB_PATH="./IP2LOCATION-LITE-DB11.BIN"
```

### 3. 健康检查
```go
// 获取系统健康状态
health := stat.HealthCheck()
// 返回格式：
// {
//   "status": "ok",
//   "message": "统计服务运行正常",
//   "config": {
//     "server": {...},
//     "database": {"mysql": true, "clickhouse": true},
//     "geoip": true
//   }
// }
```

## 优势

### 1. 统一管理
- 所有配置和初始化逻辑集中在一个文件中
- 便于维护和调试

### 2. 环境变量支持
- 支持通过环境变量灵活配置
- 便于容器化部署

### 3. 错误处理改进
- 数据库连接失败不会导致程序崩溃
- 提供详细的错误日志

### 4. 向后兼容
- 保留了原有的 `SetMysqlDB` 和 `SetClickHouseDB` 函数
- 现有代码无需大幅修改

### 5. 测试友好
- 配置加载逻辑独立，便于单元测试
- 健康检查功能便于集成测试

## 注意事项

1. **数据库连接**：如果数据库连接失败，程序会继续运行但相关功能不可用
2. **GeoIP数据库**：如果GeoIP数据库文件不存在，程序会记录警告但继续运行
3. **环境变量**：所有配置都支持环境变量覆盖，便于不同环境部署

## 迁移指南

如果您的代码中直接使用了全局变量，无需修改，因为所有变量都保持了原有的名称和类型。

如果您的代码中调用了初始化函数，建议更新为使用 `stat.InitApp()` 进行统一初始化。 