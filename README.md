# 统计系统

一个轻量级的网站访问统计系统，支持PV统计、事件埋点、用户行为分析等功能。

## 功能特性

- 📊 **实时统计**: PV、UV、停留时长等基础统计
- 🎯 **事件埋点**: 自定义事件追踪和分析
- 👥 **用户管理**: 完整的用户注册、登录、权限管理
- 🏢 **站点管理**: 多站点支持，独立统计
- 📱 **响应式设计**: 现代化的管理界面
- 🔒 **安全认证**: Token认证，数据安全

## 快速开始

### 1. 环境要求

- Go 1.19+
- MySQL 5.7+
- ClickHouse 21.8+
- IP2Location数据库文件

### 2. 安装部署

```bash
# 克隆项目
git clone <repository-url>
cd stat

# 安装依赖
go mod tidy

# 配置环境变量
cp env.example .env
# 编辑 .env 文件，配置数据库连接信息

# 初始化数据库
mysql -u root -p < init.sql

# 启动服务
go run cmd/main.go
```

### 3. 访问系统

- 登录页面: http://localhost:8080/login
- 管理后台: http://localhost:8080/admin
- 统计报表: http://localhost:8080/report

### 4. 默认账号

- 用户名: `admin`
- 密码: `admin123`

## 使用指南

### 管理后台功能

#### 1. 用户认证
- **登录**: 支持用户名或邮箱登录
- **注册**: 邮箱验证码注册
- **密码重置**: 邮箱验证码重置密码

#### 2. 站点管理
- **创建站点**: 创建新的统计站点
- **生成脚本**: 自动生成统计代码
- **删除站点**: 删除不需要的站点

#### 3. 数据统计
- **实时数据**: 查看今日PV、UV等数据
- **趋势分析**: 查看历史数据趋势
- **事件分析**: 分析用户行为事件

### 统计脚本使用

#### 1. 基础使用

```html
<!-- 在网站页面中引入统计脚本 -->
<script src="http://localhost:8080/static/js/stat.js" site-id="your-site-id"></script>
```

#### 2. 事件埋点

```javascript
// 上报自定义事件
window.statReportEvent('button_click', 'submit', {
  button_type: 'primary',
  page: 'home'
});
```

#### 3. 高级配置

```javascript
// 设置用户ID
window.statUserId = 'user-123';

// 设置站点ID
window.statSiteId = 'your-site-id';
```

## API 接口

### 认证接口

- `POST /api/auth/register` - 用户注册
- `POST /api/auth/login` - 用户登录
- `POST /api/auth/logout` - 用户登出
- `POST /api/auth/verify-code` - 发送验证码
- `POST /api/auth/reset-password` - 重置密码

### 数据上报接口

- `POST /api/track/pv` - PV上报
- `POST /api/track/event` - 事件上报
- `POST /api/track/duration` - 停留时长上报

### 管理接口

- `GET /api/site` - 查询站点列表
- `POST /api/site` - 创建站点
- `DELETE /api/site/:id` - 删除站点
- `GET /api/script/generate` - 生成统计脚本

### 统计查询接口

- `GET /api/stat/report` - 简单统计查询
- `GET /api/stat/trend` - PV/UV趋势报表

详细API文档请参考 [API.md](./API.md)

## 数据库结构

### MySQL 表结构

- `stat_user` - 用户表
- `stat_verify_code` - 验证码表
- `stat_user_session` - 用户会话表
- `stat_site` - 站点管理表

### ClickHouse 表结构

- `stat_page_view` - 页面访问记录
- `stat_event` - 事件记录

## 配置说明

### 环境变量

```bash
# 数据库配置
MYSQL_DSN=root:password@tcp(localhost:3306)/stat?charset=utf8mb4&parseTime=True&loc=Local
CLICKHOUSE_DSN=clickhouse://default:@localhost:9000/stat?dial_timeout=10s&read_timeout=20s

# 服务配置
PORT=8080
GEOIP_DB_PATH=./IP2LOCATION-LITE-DB11.BIN
```

### 邮件配置

系统支持邮件验证码功能，需要配置SMTP服务：

```bash
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASS=your-password
```

## 部署说明

### Docker 部署

```bash
# 构建镜像
docker build -t stat-system .

# 运行容器
docker-compose up -d
```

### 生产环境部署

1. **数据库优化**
   - 配置MySQL主从复制
   - 优化ClickHouse集群配置
   - 设置合适的索引

2. **服务优化**
   - 使用Nginx反向代理
   - 配置SSL证书
   - 设置CDN加速

3. **监控告警**
   - 配置服务监控
   - 设置错误告警
   - 监控数据库性能

## 开发说明

### 项目结构

```
stat/
├── cmd/
│   └── main.go              # 主程序入口
├── templates/               # HTML模板和JS模板
│   ├── layout.html          # 布局模板
│   ├── login.html           # 登录页面
│   ├── dashboard.html       # 仪表盘页面
│   ├── sites.html           # 站点管理页面
│   └── stat.js.tmpl         # 统计脚本模板（Go模板）
├── static/                  # 静态资源
│   └── js/
│       └── stat.js          # 静态统计脚本（备用）
├── api.go                   # API接口实现
├── boot.go                  # 应用初始化和配置
├── model.go                 # 数据模型定义
├── service.go               # 业务逻辑服务
├── config.yaml              # 配置文件
├── go.mod                   # Go模块文件
└── README.md                # 项目说明
```

### 开发环境

```bash
# 启动开发服务器
go run cmd/main.go

# 运行测试
go test ./...

# 代码格式化
go fmt ./...
```

## 贡献指南

1. Fork 项目
2. 创建功能分支
3. 提交代码
4. 创建 Pull Request

## 许可证

MIT License

## 联系方式

如有问题或建议，请提交 Issue 或联系开发团队。

## 项目结构

```
stat/
├── cmd/
│   └── main.go              # 主程序入口
├── templates/               # HTML模板和JS模板
│   ├── layout.html          # 布局模板
│   ├── login.html           # 登录页面
│   ├── dashboard.html       # 仪表盘页面
│   ├── sites.html           # 站点管理页面
│   └── stat.js.tmpl         # 统计脚本模板（Go模板）
├── static/                  # 静态资源
│   └── js/
│       └── stat.js          # 静态统计脚本（备用）
├── api.go                   # API接口实现
├── boot.go                  # 应用初始化和配置
├── model.go                 # 数据模型定义
├── service.go               # 业务逻辑服务
├── config.yaml              # 配置文件
├── go.mod                   # Go模块文件
└── README.md                # 项目说明
```

## 脚本文件说明

### 动态脚本生成
- **模板文件**: `templates/stat.js.tmpl` - 包含Go模板语法的JavaScript模板
- **动态接口**: `/api/script/stat.js` - 根据参数生成个性化的JavaScript代码
- **生成接口**: `/api/script/generate` - 生成包含正确属性的script标签

### 静态脚本文件
- **静态文件**: `static/js/stat.js` - 纯JavaScript文件，不包含模板语法
- **访问路径**: `/static/js/stat.js` - 直接访问静态文件

### 脚本使用方式

1. **推荐方式**（动态生成）：
```html
<script src="http://your-domain.com/api/script/stat.js?site_id=YOUR_SITE_ID&api_url=http://your-domain.com" 
        site-id="YOUR_SITE_ID" 
        data-api-url="http://your-domain.com"></script>
```

2. **备用方式**（静态文件）：
```html
<script src="http://your-domain.com/static/js/stat.js" 
        site-id="YOUR_SITE_ID" 
        data-api-url="http://your-domain.com"></script>
```

3. **通过API生成**：
```bash
curl "http://your-domain.com/api/script/generate?site_id=YOUR_SITE_ID"
``` 