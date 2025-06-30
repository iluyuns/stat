# 统计服务 API 接口文档

## 基础信息

- 基础URL: `http://localhost:8080` (开发环境) 或 `https://your-domain.com` (生产环境)
- 认证方式: Bearer Token
- 数据格式: JSON

## 响应格式

所有API接口都使用统一的响应格式：

```json
{
  "success": true,
  "message": "操作成功",
  "data": {...}
}
```

错误响应：
```json
{
  "success": false,
  "error": "错误信息"
}
```

## 用户认证接口

### 1. 发送验证码

**接口地址:** `POST /api/auth/verify-code`

**请求参数:**
```json
{
  "email": "user@example.com",
  "type": "register"  // register, reset, login
}
```

**响应示例:**
```json
{
  "msg": "验证码已发送"
}
```

### 2. 用户注册

**接口地址:** `POST /api/auth/register`

**请求参数:**
```json
{
  "username": "testuser",
  "email": "user@example.com",
  "password": "123456",
  "nickname": "测试用户",
  "code": "123456"
}
```

**响应示例:**
```json
{
  "msg": "注册成功",
  "user_id": 1
}
```

### 3. 用户登录

**接口地址:** `POST /api/auth/login`

**请求参数:**
```json
{
  "username": "testuser",  // 支持用户名或邮箱
  "password": "123456"
}
```

**响应示例:**
```json
{
  "msg": "登录成功",
  "token": "abc123def456...",
  "user": {
    "id": 1,
    "username": "testuser",
    "email": "user@example.com",
    "nickname": "测试用户",
    "avatar": ""
  }
}
```

### 4. 用户登出

**接口地址:** `POST /api/auth/logout`

**请求头:**
```
Authorization: Bearer <token>
```

**响应示例:**
```json
{
  "msg": "登出成功"
}
```

### 5. 重置密码

**接口地址:** `POST /api/auth/reset-password`

**请求参数:**
```json
{
  "email": "user@example.com",
  "code": "123456",
  "password": "newpassword"
}
```

**响应示例:**
```json
{
  "msg": "密码重置成功"
}
```

## 用户信息接口

### 6. 获取当前用户信息

**接口地址:** `GET /api/user/profile`

**请求头:**
```
Authorization: Bearer <token>
```

**响应示例:**
```json
{
  "user": {
    "id": 1,
    "username": "testuser",
    "email": "user@example.com",
    "nickname": "测试用户",
    "avatar": "",
    "status": 1,
    "last_login": "2024-01-01T12:00:00Z"
  }
}
```

## 脚本生成接口

### 生成统计脚本标签

**接口地址:** `GET /api/script/generate?site_id=<site_id>`

**功能说明:** 生成包含正确属性的统计脚本HTML标签

**请求参数:**
- `site_id` (必需): 站点ID
- `api_url` (可选): API服务器地址，默认为当前域名

**响应示例:**
```json
{
  "success": true,
  "script": "<script src=\"http://localhost:8080/api/script/stat.js?site_id=test123&api_url=http://localhost:8080\" site-id=\"test123\" data-api-url=\"http://localhost:8080\"></script>",
  "site": {
    "site_id": "test123",
    "name": "测试站点",
    "remark": "测试用"
  },
  "api_url": "http://localhost:8080"
}
```

### 动态脚本文件

**接口地址:** `GET /api/script/stat.js?site_id=<site_id>&api_url=<api_url>`

**功能说明:** 提供动态生成的JavaScript统计脚本，支持模板变量替换

**请求参数:**
- `site_id` (必需): 站点ID
- `api_url` (可选): API服务器地址

**响应:** JavaScript文件内容，模板变量会被替换为实际值

**使用示例:**
```html
<!-- 动态脚本（推荐） -->
<script src="http://your-domain.com/api/script/stat.js?site_id=YOUR_SITE_ID&api_url=http://your-domain.com" 
        site-id="YOUR_SITE_ID" 
        data-api-url="http://your-domain.com"></script>

<!-- 静态脚本（备用） -->
<script src="http://your-domain.com/static/js/stat.js" 
        site-id="YOUR_SITE_ID" 
        data-api-url="http://your-domain.com"></script>
```

## 脚本文件结构

### 模板文件
- **位置**: `templates/stat.js.tmpl`
- **类型**: Go模板文件
- **用途**: 动态生成个性化的JavaScript代码
- **特点**: 支持 `{{.ApiURL}}` 和 `{{.SiteID}}` 模板变量

### 静态文件
- **位置**: `static/js/stat.js`
- **类型**: 纯JavaScript文件
- **用途**: 备用静态脚本文件
- **特点**: 不包含模板语法，通过属性获取配置

## 数据上报接口

### 8. PV 上报

**接口地址:** `POST /api/track/pv`

**请求参数:**
```json
{
  "site_id": "demo123456",
  "path": "/home",
  "user_id": "user123",
  "width": 1920,
  "height": 1080,
  "lang": "zh-CN",
  "net": "4g"
}
```

### 9. 事件上报

**接口地址:** `POST /api/track/event`

**请求参数:**
```json
{
  "site_id": "demo123456",
  "event_name": "button_click",
  "user_id": "user123",
  "value": "submit",
  "extra": "{\"button_type\":\"primary\"}",
  "path": "/home"
}
```

### 10. 停留时长上报

**接口地址:** `POST /api/track/duration`

**请求参数:**
```json
{
  "path": "/home",
  "duration": 120
}
```

## 站点管理接口

### 11. 创建站点

**接口地址:** `POST /api/site`

**请求参数:**
```json
{
  "name": "我的网站",
  "remark": "网站描述"
}
```

**响应示例:**
```json
{
  "site_id": "abc123def456"
}
```

### 12. 查询所有站点

**接口地址:** `GET /api/site`

**响应示例:**
```json
[
  {
    "id": 1,
    "site_id": "demo123456",
    "name": "示例站点",
    "remark": "这是一个示例统计站点",
    "created_by": "admin",
    "created_at": "2024-01-01T12:00:00Z"
  }
]
```

### 13. 删除站点

**接口地址:** `DELETE /api/site/:id`

**响应示例:**
```json
{
  "msg": "deleted"
}
```

## 统计查询接口

### 14. 简单统计查询

**接口地址:** `GET /api/stat/report`

**响应示例:**
```json
{
  "pv_today": 1000
}
```

### 15. PV/UV 趋势报表

**接口地址:** `GET /api/stat/trend?site_id=<site_id>&start=<start_date>&end=<end_date>`

**响应示例:**
```json
{
  "dates": ["2024-01-01", "2024-01-02"],
  "pv": [100, 200],
  "uv": [50, 80]
}
```

## 错误码说明

| 状态码 | 说明 |
|--------|------|
| 200 | 成功 |
| 400 | 参数错误 |
| 401 | 未认证 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

## 使用示例

### 完整注册流程

1. 发送验证码
```bash
curl -X POST http://localhost:8080/api/auth/verify-code \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","type":"register"}'
```

2. 用户注册
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","email":"user@example.com","password":"123456","code":"123456"}'
```

3. 用户登录
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"123456"}'
```

4. 创建站点
```bash
curl -X POST http://localhost:8080/api/site \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"name":"我的网站","remark":"网站描述"}'
```

5. 生成统计脚本
```bash
curl -X GET "http://localhost:8080/api/script/generate?site_id=<site_id>"
```

## 注意事项

1. 所有需要认证的接口都需要在请求头中携带 `Authorization: Bearer <token>`
2. 验证码有效期为10分钟
3. Token 有效期为7天
4. 密码长度要求6-20位
5. 用户名长度要求3-20位
6. 邮箱格式必须正确 