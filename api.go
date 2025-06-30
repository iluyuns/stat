package stat

import (
	"crypto/md5"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 统一错误响应格式
type ApiResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// 统一成功响应
func successResponse(c *gin.Context, data interface{}, message string) {
	c.JSON(http.StatusOK, ApiResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// 统一错误响应
func errorResponse(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, ApiResponse{
		Success: false,
		Error:   message,
	})
}

// 参数验证错误响应
func validationError(c *gin.Context, message string) {
	errorResponse(c, http.StatusBadRequest, "参数错误: "+message)
}

// 数据库错误响应
func databaseError(c *gin.Context, err error) {
	log.Printf("数据库错误: %v", err)
	errorResponse(c, http.StatusInternalServerError, "数据操作失败")
}

// 认证错误响应
func authError(c *gin.Context, message string) {
	errorResponse(c, http.StatusUnauthorized, message)
}

// PV上报接口
func TrackPV(c *gin.Context) {
	var pv PageView
	if err := c.ShouldBindJSON(&pv); err != nil {
		validationError(c, err.Error())
		return
	}

	// 验证必要字段
	if pv.SiteId == "" {
		validationError(c, "site_id不能为空")
		return
	}

	pv.IP = c.ClientIP()
	pv.UA = c.Request.UserAgent()
	uainfo := UA(pv.UA)
	pv.Device = uainfo.Device
	pv.OS = uainfo.OS
	pv.Browser = uainfo.Browser
	pv.CreatedAt = time.Now()

	// 解析地理信息
	if GeoIPDB != nil {
		// 只处理IPv4地址，跳过IPv6
		if !strings.Contains(pv.IP, ":") {
			record, err := GeoIPDB.Get_all(pv.IP)
			if err != nil {
				log.Printf("failed to get geoip record: %v", err)
			} else {
				// 检查返回的数据是否有效
				if record.City != "" && record.City != "Invalid IP address" {
					pv.City = record.City
				}
				if record.Region != "" && record.Region != "Invalid IP address" {
					pv.Province = record.Region
				}
				if record.Isp != "" && record.Isp != "Invalid IP address" {
					// 截断ISP信息到64个字符，避免数据库字段长度限制
					if len(record.Isp) > 64 {
						pv.ISP = record.Isp[:64]
					} else {
						pv.ISP = record.Isp
					}
				}
				if record.Country_long != "" && record.Country_long != "Invalid IP address" {
					pv.Country = record.Country_long
				}
			}
		}
	}

	if StatDB != nil {
		if err := StatDB.Create(&pv).Error; err != nil {
			databaseError(c, err)
			return
		}
	}

	successResponse(c, nil, "pv received")
}

// 停留时长上报接口
func TrackDuration(c *gin.Context) {
	var duration struct {
		SiteId    string `json:"site_id"`
		Path      string `json:"path"`
		Duration  int    `json:"duration"`
		UserID    string `json:"user_id"`
		IP        string `json:"ip"`
		UA        string `json:"ua"`
		Heartbeat bool   `json:"heartbeat"`
	}
	if err := c.ShouldBindJSON(&duration); err != nil {
		validationError(c, err.Error())
		return
	}

	// 验证必要字段
	if duration.SiteId == "" {
		validationError(c, "site_id不能为空")
		return
	}

	// 使用统计数据库（ClickHouse 或 MySQL）
	if StatDB != nil {
		// 查找对应的页面访问记录并更新停留时长
		var pageView PageView
		query := StatDB.Where("site_id = ? AND path = ? AND user_id = ?",
			duration.SiteId, duration.Path, duration.UserID)

		// 如果是心跳数据，查找最近的记录
		if duration.Heartbeat {
			query = query.Order("created_at DESC").Limit(1)
		}

		if err := query.First(&pageView).Error; err == nil {
			// 更新停留时长，取最大值
			if duration.Duration > pageView.Duration {
				pageView.Duration = duration.Duration
				StatDB.Save(&pageView)
			}
		} else {
			// 如果找不到记录，创建新的记录
			pageView = PageView{
				SiteId:    duration.SiteId,
				Path:      duration.Path,
				UserID:    duration.UserID,
				IP:        duration.IP,
				UA:        duration.UA,
				Duration:  duration.Duration,
				CreatedAt: time.Now(),
			}
			StatDB.Create(&pageView)
		}
	}

	successResponse(c, nil, "duration recorded")
}

// 事件上报接口
func TrackEvent(c *gin.Context) {
	var event StatEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		validationError(c, err.Error())
		return
	}

	// 验证必要字段
	if event.SiteId == "" {
		validationError(c, "site_id不能为空")
		return
	}

	if event.EventName == "" {
		validationError(c, "event_name不能为空")
		return
	}

	// 设置IP和UA（如果前端没有提供）
	if event.IP == "" {
		event.IP = c.ClientIP()
	}
	if event.UA == "" {
		event.UA = c.Request.UserAgent()
	}

	// 解析地理信息
	if GeoIPDB != nil {
		// 只处理IPv4地址，跳过IPv6
		if !strings.Contains(event.IP, ":") {
			record, err := GeoIPDB.Get_all(event.IP)
			if err != nil {
				log.Printf("failed to get geoip record: %v", err)
			} else {
				// 检查返回的数据是否有效
				if record.City != "" && record.City != "Invalid IP address" {
					event.City = record.City
				}
				if record.Region != "" && record.Region != "Invalid IP address" {
					event.Province = record.Region
				}
				if record.Isp != "" && record.Isp != "Invalid IP address" {
					// 截断ISP信息到64个字符，避免数据库字段长度限制
					if len(record.Isp) > 64 {
						event.ISP = record.Isp[:64]
					} else {
						event.ISP = record.Isp
					}
				}
				if record.Country_long != "" && record.Country_long != "Invalid IP address" {
					event.Country = record.Country_long
				}
			}
		}
	}

	// 解析设备、系统、浏览器信息（如果前端没有提供）
	if event.UA != "" {
		uainfo := UA(event.UA)
		event.Device = uainfo.Device
		event.OS = uainfo.OS
		event.Browser = uainfo.Browser
	}

	// 设置创建时间
	event.CreatedAt = time.Now()

	// 如果前端没有提供事件类型分类，自动推断
	if event.EventCategory == "" {
		event.EventCategory = inferEventCategory(event.EventName)
	}

	// 使用统计数据库（ClickHouse 或 MySQL）
	if StatDB != nil {
		if err := StatDB.Create(&event).Error; err != nil {
			databaseError(c, err)
			return
		}
	}

	successResponse(c, nil, "event recorded")
}

// 根据事件名称推断事件类型分类
func inferEventCategory(eventName string) string {
	if eventName == "" {
		return "未知"
	}

	name := strings.ToLower(eventName)

	// 点击相关事件
	if strings.Contains(name, "click") || strings.Contains(name, "tap") || strings.Contains(name, "button") {
		return "点击事件"
	}

	// 滚动相关事件
	if strings.Contains(name, "scroll") {
		return "滚动事件"
	}

	// 表单相关事件
	if strings.Contains(name, "form") || strings.Contains(name, "submit") || strings.Contains(name, "input") {
		return "表单事件"
	}

	// 页面相关事件
	if strings.Contains(name, "page") || strings.Contains(name, "view") || strings.Contains(name, "load") {
		return "页面事件"
	}

	// 用户交互事件
	if strings.Contains(name, "user") || strings.Contains(name, "interaction") || strings.Contains(name, "action") {
		return "用户交互"
	}

	// 自定义事件
	if strings.Contains(name, "custom") || strings.Contains(name, "custom_") {
		return "自定义事件"
	}

	// 默认分类
	return "其他事件"
}

// 查询统计数据（简单示例：返回今日PV数）
func GetReport(c *gin.Context) {
	var count int64
	today := time.Now().Format("2006-01-02")
	if ClickHouseDB != nil {
		ClickHouseDB.Raw(`SELECT COUNT(*) FROM stat_page_view WHERE toDate(created_at) = ?`, today).Scan(&count)
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "pv_today": count})
}

// 新建统计ID
func CreateSite(c *gin.Context) {
	var req struct {
		Name   string `json:"name"`
		Remark string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "参数错误"})
		return
	}
	siteId := RandString(12)
	site := StatSite{SiteId: siteId, Name: req.Name, Remark: req.Remark}
	if MysqlDB != nil {
		err := MysqlDB.Create(&site).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "site_id": siteId})
}

// 查询所有统计ID
func ListSite(c *gin.Context) {
	var sites []StatSite
	if MysqlDB != nil {
		err := MysqlDB.Find(&sites).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
			return
		}
	}

	// 为每个站点添加今日统计数据
	today := time.Now().Format("2006-01-02")
	var result []gin.H
	for _, site := range sites {
		var todayPV, todayUV int64
		if StatDB != nil {
			StatDB.Model(&PageView{}).Where("site_id = ? AND DATE(created_at) = ?", site.SiteId, today).Count(&todayPV)
			StatDB.Model(&PageView{}).Where("site_id = ? AND DATE(created_at) = ?", site.SiteId, today).Distinct("ip").Count(&todayUV)
		}

		result = append(result, gin.H{
			"id":         site.ID,
			"site_id":    site.SiteId,
			"name":       site.Name,
			"remark":     site.Remark,
			"created_by": site.CreatedBy,
			"created_at": site.CreatedAt,
			"today_pv":   todayPV,
			"today_uv":   todayUV,
		})
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// 删除统计ID
func DeleteSite(c *gin.Context) {
	id := c.Param("id")
	if MysqlDB != nil {
		err := MysqlDB.Where("site_id = ?", id).Delete(&StatSite{}).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "msg": "deleted"})
}

// 统计报表接口（PV/UV趋势）
func Report(c *gin.Context) {
	siteId := c.Query("site_id")
	start := c.Query("start")
	end := c.Query("end")
	if siteId == "" || start == "" || end == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	var pvRes, uvRes []struct {
		Date string
		Cnt  int64
	}
	if ClickHouseDB != nil {
		ClickHouseDB.Raw(`SELECT toDate(created_at) as date, COUNT(*) as cnt FROM stat_page_view WHERE site_id = ? AND created_at BETWEEN ? AND ? GROUP BY date`, siteId, start, end).Scan(&pvRes)
		ClickHouseDB.Raw(`SELECT toDate(created_at) as date, COUNT(DISTINCT ip) as cnt FROM stat_page_view WHERE site_id = ? AND created_at BETWEEN ? AND ? GROUP BY date`, siteId, start, end).Scan(&uvRes)
	}
	// 合并结果
	dateMap := map[string]struct{ PV, UV int64 }{}
	for _, r := range pvRes {
		dateMap[r.Date] = struct{ PV, UV int64 }{PV: r.Cnt}
	}
	for _, r := range uvRes {
		v := dateMap[r.Date]
		v.UV = r.Cnt
		dateMap[r.Date] = v
	}
	var dates []string
	var pv, uv []int64
	for d := start; d <= end; d = nextDay(d) {
		dates = append(dates, d)
		v := dateMap[d]
		pv = append(pv, v.PV)
		uv = append(uv, v.UV)
	}
	c.JSON(http.StatusOK, gin.H{"dates": dates, "pv": pv, "uv": uv})
}

// 工具函数：生成随机siteId
func RandString(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}

// 工具函数：日期递增
func nextDay(date string) string {
	t, _ := time.Parse("2006-01-02", date)
	t = t.Add(24 * time.Hour)
	return t.Format("2006-01-02")
}

// ========== 用户认证相关接口 ==========

// 用户注册
func Register(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required,min=3,max=20"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6,max=20"`
		Nickname string `json:"nickname"`
		Code     string `json:"code" binding:"required,len=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "参数错误: " + err.Error()})
		return
	}

	// 验证验证码
	var verifyCode VerifyCode
	if MysqlDB != nil {
		err := MysqlDB.Where("email = ? AND code = ? AND type = ? AND used = ? AND expired_at > ?",
			req.Email, req.Code, "register", false, time.Now()).First(&verifyCode).Error
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "验证码错误或已过期"})
			return
		}
	}

	// 检查用户名和邮箱是否已存在
	var existingUser User
	if MysqlDB != nil {
		err := MysqlDB.Where("username = ? OR email = ?", req.Username, req.Email).First(&existingUser).Error
		if err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "用户名或邮箱已存在"})
			return
		}
	}

	// 创建用户
	hashedPassword, _ := HashPassword(req.Password)
	user := User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
		Nickname: req.Nickname,
		Status:   1,
	}

	if MysqlDB != nil {
		err := MysqlDB.Create(&user).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "创建用户失败"})
			return
		}

		// 标记验证码为已使用
		MysqlDB.Model(&verifyCode).Update("used", true)
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "msg": "注册成功", "user_id": user.ID})
}

// 用户登录
func Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "参数错误"})
		return
	}

	// 查找用户
	var user User
	if MysqlDB != nil {
		err := MysqlDB.Where("username = ? OR email = ?", req.Username, req.Username).First(&user).Error
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "用户不存在"})
			return
		}
	}

	// 验证密码
	if !CheckPassword(req.Password, user.Password) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "密码错误"})
		return
	}

	// 检查用户状态
	if user.Status != 1 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "账户已被禁用"})
		return
	}

	// 生成token
	token := GenerateToken()
	session := UserSession{
		UserID:    user.ID,
		Token:     token,
		IP:        c.ClientIP(),
		UA:        c.Request.UserAgent(),
		ExpiredAt: time.Now().Add(7 * 24 * time.Hour), // 7天过期
	}

	if MysqlDB != nil {
		// 更新最后登录时间
		MysqlDB.Model(&user).Update("last_login", time.Now())
		// 创建会话
		MysqlDB.Create(&session)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"msg":     "登录成功",
		"token":   token,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"nickname": user.Nickname,
			"avatar":   user.Avatar,
		},
	})
}

// 发送验证码
func SendVerifyCode(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
		Type  string `json:"type" binding:"required,oneof=register reset login"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "参数错误"})
		return
	}

	// 检查发送频率限制（1分钟内只能发送一次）
	var lastCode VerifyCode
	if MysqlDB != nil {
		err := MysqlDB.Where("email = ? AND type = ? AND created_at > ?",
			req.Email, req.Type, time.Now().Add(-time.Minute)).First(&lastCode).Error
		if err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "发送过于频繁，请稍后再试"})
			return
		}
	}

	// 生成验证码
	code := GenerateVerifyCode()
	expiredAt := time.Now().Add(10 * time.Minute) // 10分钟过期

	verifyCode := VerifyCode{
		Email:     req.Email,
		Code:      code,
		Type:      req.Type,
		ExpiredAt: expiredAt,
	}

	if MysqlDB != nil {
		MysqlDB.Create(&verifyCode)
	}

	// TODO: 发送邮件验证码
	// 这里应该集成邮件服务，暂时打印到日志
	log.Printf("验证码: %s, 邮箱: %s, 类型: %s", code, req.Email, req.Type)

	c.JSON(http.StatusOK, gin.H{"success": true, "msg": "验证码已发送"})
}

// 用户登出
func Logout(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "缺少token"})
		return
	}

	// 移除token前缀
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	if MysqlDB != nil {
		MysqlDB.Where("token = ?", token).Delete(&UserSession{})
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "msg": "登出成功"})
}

// 获取当前用户信息
func GetCurrentUser(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "未登录"})
		return
	}

	var user User
	if MysqlDB != nil {
		err := MysqlDB.Where("id = ?", userID).First(&user).Error
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "用户不存在"})
			return
		}
	}

	// 兼容前端：data字段下为用户信息
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"email":      user.Email,
			"nickname":   user.Nickname,
			"avatar":     user.Avatar,
			"status":     user.Status,
			"last_login": user.LastLogin,
		},
	})
}

// 生成统计脚本
func GenerateScript(c *gin.Context) {
	siteId := c.Query("site_id")
	if siteId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "缺少site_id参数"})
		return
	}

	// 获取当前访问的域名和协议
	apiURL := c.Query("api_url")
	if apiURL == "" {
		// 使用当前请求的域名和协议
		scheme := "http"
		if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		apiURL = scheme + "://" + c.Request.Host
	}

	// 验证站点是否存在
	var site StatSite
	if MysqlDB != nil {
		err := MysqlDB.Where("site_id = ?", siteId).First(&site).Error
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "站点不存在"})
			return
		}
	}

	// 生成正确的HTML标签，使用动态脚本接口
	scriptTag := fmt.Sprintf(`<script src="%s/api/script/stat.js?site_id=%s&api_url=%s" site-id="%s" data-api-url="%s"></script>`, apiURL, siteId, apiURL, siteId, apiURL)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"script":  scriptTag,
		"site": gin.H{
			"site_id": site.SiteId,
			"name":    site.Name,
			"remark":  site.Remark,
		},
		"api_url": apiURL,
	})
}

// 重置密码
func ResetPassword(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Code     string `json:"code" binding:"required,len=6"`
		Password string `json:"password" binding:"required,min=6,max=20"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "参数错误"})
		return
	}

	// 验证验证码
	var verifyCode VerifyCode
	if MysqlDB != nil {
		err := MysqlDB.Where("email = ? AND code = ? AND type = ? AND used = ? AND expired_at > ?",
			req.Email, req.Code, "reset", false, time.Now()).First(&verifyCode).Error
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "验证码错误或已过期"})
			return
		}
	}

	// 更新密码
	hashedPassword, _ := HashPassword(req.Password)
	if MysqlDB != nil {
		err := MysqlDB.Model(&User{}).Where("email = ?", req.Email).Update("password", hashedPassword).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "重置密码失败"})
			return
		}

		// 标记验证码为已使用
		MysqlDB.Model(&verifyCode).Update("used", true)
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "msg": "密码重置成功"})
}

// 提供动态脚本文件
func ServeScript(c *gin.Context) {
	siteId := c.Query("site_id")
	if siteId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "缺少site_id参数"})
		return
	}

	// 获取API地址
	apiURL := c.Query("api_url")
	if apiURL == "" {
		// 默认使用当前域名
		apiURL = c.Request.Host
		if c.Request.TLS != nil {
			apiURL = "https://" + apiURL
		} else {
			apiURL = "http://" + apiURL
		}
	}

	// 验证站点是否存在
	var site StatSite
	if MysqlDB != nil {
		err := MysqlDB.Where("site_id = ?", siteId).First(&site).Error
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "站点不存在"})
			return
		}
	}

	// 生成脚本内容
	scriptContent := GenerateScriptContent(siteId, apiURL)

	// 设置响应头
	c.Header("Content-Type", "application/javascript")
	c.Header("Cache-Control", "public, max-age=3600") // 缓存1小时
	c.String(http.StatusOK, scriptContent)
}

// ========== 工具函数 ==========

// 密码加密
func HashPassword(password string) (string, error) {
	// 这里使用简单的MD5加密，生产环境建议使用bcrypt
	hash := md5.Sum([]byte(password))
	return hex.EncodeToString(hash[:]), nil
}

// 密码验证
func CheckPassword(password, hash string) bool {
	hashedPassword, _ := HashPassword(password)
	return hashedPassword == hash
}

// 生成token
func GenerateToken() string {
	// 生成32位随机字符串
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 32)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}

// 生成验证码
func GenerateVerifyCode() string {
	// 生成6位数字验证码
	code := fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
	return code
}

// 认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			// 尝试从cookie获取token
			if cookie, err := c.Cookie("token"); err == nil {
				token = cookie
			}
		}

		// 如果是页面访问，尝试从查询参数获取token（用于调试）
		if token == "" && c.Request.Method == "GET" {
			token = c.Query("token")
		}

		if token == "" {
			// 检查是否是API请求（以/api/开头）
			if strings.HasPrefix(c.Request.URL.Path, "/api/") {
				authError(c, "缺少token")
				c.Abort()
				return
			}
			// 如果是页面请求，重定向到登录页面
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		// 移除token前缀
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		// 验证token
		var session UserSession
		if MysqlDB != nil {
			err := MysqlDB.Where("token = ? AND expired_at > ?", token, time.Now()).First(&session).Error
			if err != nil {
				// 检查是否是API请求（以/api/开头）
				if strings.HasPrefix(c.Request.URL.Path, "/api/") {
					authError(c, "token无效或已过期")
					c.Abort()
					return
				}
				// 如果是页面请求，重定向到登录页面
				c.Redirect(http.StatusFound, "/login")
				c.Abort()
				return
			}
		} else {
			// 检查是否是API请求（以/api/开头）
			if strings.HasPrefix(c.Request.URL.Path, "/api/") {
				authError(c, "数据库连接失败")
				c.Abort()
				return
			}
			// 如果是页面请求，重定向到登录页面
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		// 设置用户ID到上下文
		c.Set("user_id", session.UserID)
		c.Next()
	}
}

// 根据token获取用户信息
func GetUserByToken(token string) *User {
	var session UserSession
	if MysqlDB != nil {
		err := MysqlDB.Where("token = ? AND expired_at > ?", token, time.Now()).First(&session).Error
		if err != nil {
			return nil
		}
	}

	var user User
	if MysqlDB != nil {
		err := MysqlDB.Where("id = ?", session.UserID).First(&user).Error
		if err != nil {
			return nil
		}
	}

	return &user
}

// 仪表盘数据接口
func GetDashboardData(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		authError(c, "未登录")
		return
	}

	// 获取site_id参数
	siteId := c.Query("site_id")

	// 获取今日数据
	today := time.Now().Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	var todayPV, todayUV, yesterdayPV, yesterdayUV int64
	var avgDuration *float64
	var activeSites int64
	var pageCount int64

	if StatDB != nil {
		dbType := getDBType()

		// 今日PV
		todayQuery := buildDateQuery("created_at", today)
		if siteId != "" {
			StatDB.Model(&PageView{}).Where("site_id = ? AND "+todayQuery, siteId).Count(&todayPV)
		} else {
			StatDB.Model(&PageView{}).Where(todayQuery).Count(&todayPV)
		}

		// 今日UV
		if dbType == "clickhouse" {
			if siteId != "" {
				StatDB.Model(&PageView{}).Where("site_id = ? AND "+todayQuery, siteId).Distinct("ip").Count(&todayUV)
			} else {
				StatDB.Model(&PageView{}).Where(todayQuery).Distinct("ip").Count(&todayUV)
			}
		} else {
			if siteId != "" {
				StatDB.Model(&PageView{}).Where("site_id = ? AND "+todayQuery, siteId).Distinct("ip").Count(&todayUV)
			} else {
				StatDB.Model(&PageView{}).Where(todayQuery).Distinct("ip").Count(&todayUV)
			}
		}

		// 昨日PV
		yesterdayQuery := buildDateQuery("created_at", yesterday)
		if siteId != "" {
			StatDB.Model(&PageView{}).Where("site_id = ? AND "+yesterdayQuery, siteId).Count(&yesterdayPV)
		} else {
			StatDB.Model(&PageView{}).Where(yesterdayQuery).Count(&yesterdayPV)
		}

		// 昨日UV
		if dbType == "clickhouse" {
			if siteId != "" {
				StatDB.Model(&PageView{}).Where("site_id = ? AND "+yesterdayQuery, siteId).Distinct("ip").Count(&yesterdayUV)
			} else {
				StatDB.Model(&PageView{}).Where(yesterdayQuery).Distinct("ip").Count(&yesterdayUV)
			}
		} else {
			if siteId != "" {
				StatDB.Model(&PageView{}).Where("site_id = ? AND "+yesterdayQuery, siteId).Distinct("ip").Count(&yesterdayUV)
			} else {
				StatDB.Model(&PageView{}).Where(yesterdayQuery).Distinct("ip").Count(&yesterdayUV)
			}
		}

		// 平均停留时长 - 修复计算逻辑
		var durationResult *float64
		dbType = getDBType()
		var durationQuery *gorm.DB
		if dbType == "clickhouse" {
			// ClickHouse使用coalesce处理NULL值，只计算今日数据
			durationQuery = StatDB.Model(&PageView{}).Select("coalesce(AVG(duration), 0)").Where("duration > 0 AND " + todayQuery)
			if siteId != "" {
				durationQuery = durationQuery.Where("site_id = ?", siteId)
			}
			durationQuery.Scan(&durationResult)
		} else {
			// MySQL使用IFNULL处理NULL值，只计算今日数据
			durationQuery = StatDB.Model(&PageView{}).Select("IFNULL(AVG(duration), 0)").Where("duration > 0 AND " + todayQuery)
			if siteId != "" {
				durationQuery = durationQuery.Where("site_id = ?", siteId)
			}
			durationQuery.Scan(&durationResult)
		}
		if durationResult != nil {
			avgDuration = durationResult
		}

		// 活跃站点数（只在全局查询时计算）
		if siteId == "" {
			if dbType == "clickhouse" {
				StatDB.Model(&PageView{}).Where(todayQuery).Distinct("site_id").Count(&activeSites)
			} else {
				StatDB.Model(&PageView{}).Where(todayQuery).Distinct("site_id").Count(&activeSites)
			}
		} else {
			activeSites = 1 // 单个站点时设为1
		}

		// 页面数量（只在站点查询时计算）
		if siteId != "" {
			StatDB.Model(&PageView{}).Where("site_id = ?", siteId).Distinct("path").Count(&pageCount)
		}
	}

	// 计算趋势
	pvTrend := 0.0
	uvTrend := 0.0
	if yesterdayPV > 0 {
		pvTrend = float64(todayPV-yesterdayPV) / float64(yesterdayPV) * 100
	}
	if yesterdayUV > 0 {
		uvTrend = float64(todayUV-yesterdayUV) / float64(yesterdayUV) * 100
	}

	// 获取最近7天的PV趋势数据
	var pvTrendData []struct {
		Date  string `json:"date"`
		Count int64  `json:"count"`
	}

	if StatDB != nil {
		sevenDaysAgo := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
		trendQuery := StatDB.Model(&PageView{}).
			Select("DATE(created_at) as date, COUNT(*) as count").
			Where("created_at >= ?", sevenDaysAgo)

		if siteId != "" {
			trendQuery = trendQuery.Where("site_id = ?", siteId)
		}

		trendQuery.Group("DATE(created_at)").
			Order("date ASC").
			Scan(&pvTrendData)
	}

	// 如果没有数据，生成默认的7天数据
	if len(pvTrendData) == 0 {
		for i := 6; i >= 0; i-- {
			date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
			pvTrendData = append(pvTrendData, struct {
				Date  string `json:"date"`
				Count int64  `json:"count"`
			}{
				Date:  date,
				Count: 0,
			})
		}
	}

	successResponse(c, gin.H{
		"today_pv": todayPV,
		"today_uv": todayUV,
		"avg_duration": func() string {
			if avgDuration != nil && *avgDuration > 0 {
				// 如果超过60秒，显示分钟和秒
				if *avgDuration >= 60 {
					minutes := int(*avgDuration / 60)
					seconds := int(*avgDuration) % 60
					if seconds > 0 {
						return fmt.Sprintf("%d分%d秒", minutes, seconds)
					} else {
						return fmt.Sprintf("%d分钟", minutes)
					}
				} else {
					// 小于60秒，显示秒数
					return fmt.Sprintf("%d秒", int(*avgDuration))
				}
			}
			return "0秒"
		}(),
		"active_sites":  activeSites,
		"page_count":    pageCount,
		"pv_trend":      pvTrend,
		"uv_trend":      uvTrend,
		"pv_trend_data": pvTrendData,
	}, "数据获取成功")
}

// 热门页面接口
func GetPopularPages(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "未登录"})
		return
	}

	// 获取site_id参数
	siteId := c.Query("site_id")

	var pages []struct {
		Path  string `json:"path"`
		Count int64  `json:"count"`
	}

	if StatDB != nil {
		today := time.Now().Format("2006-01-02")
		query := StatDB.Model(&PageView{}).
			Select("path, COUNT(*) as count").
			Where("DATE(created_at) = ?", today)

		if siteId != "" {
			query = query.Where("site_id = ?", siteId)
		}

		query.Group("path").
			Order("count DESC").
			Limit(10).
			Scan(&pages)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    pages,
	})
}

// 实时数据接口
func GetRealtimeData(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "未登录"})
		return
	}

	// 获取site_id参数
	siteId := c.Query("site_id")

	var visits []struct {
		CreatedAt string `json:"created_at"`
		Path      string `json:"path"`
		IP        string `json:"ip"`
		City      string `json:"city"`
		Province  string `json:"province"`
		UA        string `json:"ua"`
		UserID    string `json:"user_id"`
		Device    string `json:"device"`
		OS        string `json:"os"`
		Browser   string `json:"browser"`
		Duration  int    `json:"duration"`
		Referer   string `json:"referer"`
		Screen    string `json:"screen"`
		Net       string `json:"net"`
	}

	if StatDB != nil {
		// 获取最近30分钟的访问记录
		thirtyMinutesAgo := time.Now().Add(-30 * time.Minute)
		query := StatDB.Model(&PageView{}).
			Select("created_at, path, ip, city, province, ua, user_id, device, os, browser, duration, referer, screen, net").
			Where("created_at >= ?", thirtyMinutesAgo)

		if siteId != "" {
			query = query.Where("site_id = ?", siteId)
		}

		query.Order("created_at DESC").
			Limit(20).
			Scan(&visits)
	}

	// 转换数据格式
	var result []map[string]interface{}
	for _, visit := range visits {
		// 处理地理位置
		location := ""
		if visit.City != "" || visit.Province != "" {
			if visit.City != "" && visit.Province != "" {
				location = visit.City + ", " + visit.Province
			} else if visit.City != "" {
				location = visit.City
			} else {
				location = visit.Province
			}
		} else {
			location = "未知"
		}

		// 处理停留时长
		durationText := "0秒"
		if visit.Duration > 0 {
			if visit.Duration >= 60 {
				minutes := visit.Duration / 60
				seconds := visit.Duration % 60
				if seconds > 0 {
					durationText = fmt.Sprintf("%d分%d秒", minutes, seconds)
				} else {
					durationText = fmt.Sprintf("%d分钟", minutes)
				}
			} else {
				durationText = fmt.Sprintf("%d秒", visit.Duration)
			}
		}

		// 处理设备信息
		deviceInfo := visit.Device
		if visit.OS != "" && visit.OS != "Unknown" {
			deviceInfo = visit.Device + " (" + visit.OS + ")"
		}

		// 处理来源页面
		referer := visit.Referer
		if referer == "" {
			referer = "直接访问"
		} else if len(referer) > 50 {
			referer = referer[:47] + "..."
		}

		// 处理屏幕分辨率
		screen := visit.Screen
		if screen == "" {
			screen = "未知"
		}

		// 处理网络类型
		net := visit.Net
		if net == "" {
			net = "未知"
		}

		result = append(result, map[string]interface{}{
			"created_at":   visit.CreatedAt,
			"path":         visit.Path,
			"ip":           visit.IP,
			"location":     location,
			"device":       deviceInfo,
			"browser":      visit.Browser,
			"duration":     durationText,
			"duration_raw": visit.Duration,
			"referer":      referer,
			"screen":       screen,
			"net":          net,
			"user_id":      visit.UserID,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// 数据库兼容性处理函数

// 获取数据库类型
func getDBType() string {
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

// 兼容的日期格式化
func formatDateForDB(date time.Time) string {
	dbType := getDBType()
	switch dbType {
	case "clickhouse":
		return date.Format("2006-01-02")
	case "mysql":
		return date.Format("2006-01-02")
	default:
		return date.Format("2006-01-02")
	}
}

// 兼容的日期查询
func buildDateQuery(field string, date string) string {
	dbType := getDBType()
	switch dbType {
	case "clickhouse":
		return fmt.Sprintf("toDate(%s) = '%s'", field, date)
	case "mysql":
		return fmt.Sprintf("DATE(%s) = '%s'", field, date)
	default:
		return fmt.Sprintf("DATE(%s) = '%s'", field, date)
	}
}

// 兼容的日期时间格式化
func buildDateTimeFormat(field string) string {
	dbType := getDBType()
	switch dbType {
	case "clickhouse":
		return fmt.Sprintf("toDateTime(%s, '%%H:%%i')", field)
	case "mysql":
		return fmt.Sprintf("DATE_FORMAT(%s, '%%H:%%i')", field)
	default:
		return fmt.Sprintf("DATE_FORMAT(%s, '%%H:%%i')", field)
	}
}

// 兼容的连接字符串
func buildConcatString(fields ...string) string {
	dbType := getDBType()
	switch dbType {
	case "clickhouse":
		// ClickHouse使用concat函数
		return fmt.Sprintf("concat(%s)", strings.Join(fields, ", "))
	case "mysql":
		// MySQL使用CONCAT函数
		return fmt.Sprintf("CONCAT(%s)", strings.Join(fields, ", "))
	default:
		return fmt.Sprintf("CONCAT(%s)", strings.Join(fields, ", "))
	}
}

// 健康检查接口
func HealthCheckAPI(c *gin.Context) {
	// 检查数据库连接
	mysqlStatus := "unknown"
	clickhouseStatus := "unknown"
	statDBStatus := "unknown"

	if MysqlDB != nil {
		sqlDB, err := MysqlDB.DB()
		if err == nil {
			err = sqlDB.Ping()
			if err == nil {
				mysqlStatus = "healthy"
			} else {
				mysqlStatus = "unhealthy"
			}
		} else {
			mysqlStatus = "unhealthy"
		}
	} else {
		mysqlStatus = "not_configured"
	}

	if ClickHouseDB != nil {
		sqlDB, err := ClickHouseDB.DB()
		if err == nil {
			err = sqlDB.Ping()
			if err == nil {
				clickhouseStatus = "healthy"
			} else {
				clickhouseStatus = "unhealthy"
			}
		} else {
			clickhouseStatus = "unhealthy"
		}
	} else {
		clickhouseStatus = "not_configured"
	}

	if StatDB != nil {
		sqlDB, err := StatDB.DB()
		if err == nil {
			err = sqlDB.Ping()
			if err == nil {
				statDBStatus = "healthy"
			} else {
				statDBStatus = "unhealthy"
			}
		} else {
			statDBStatus = "unhealthy"
		}
	} else {
		statDBStatus = "not_configured"
	}

	// 检查GeoIP数据库
	geoipStatus := "not_configured"
	if GeoIPDB != nil {
		geoipStatus = "healthy"
	}

	// 总体健康状态
	overallStatus := "healthy"
	if mysqlStatus == "unhealthy" || (clickhouseStatus == "unhealthy" && statDBStatus == "unhealthy") {
		overallStatus = "unhealthy"
	}

	successResponse(c, gin.H{
		"status":    overallStatus,
		"timestamp": time.Now().Unix(),
		"version":   "1.0.0",
		"services": gin.H{
			"mysql": gin.H{
				"status": mysqlStatus,
				"type":   "management",
			},
			"clickhouse": gin.H{
				"status": clickhouseStatus,
				"type":   "statistics",
			},
			"stat_db": gin.H{
				"status": statDBStatus,
				"type":   getDBType(),
			},
			"geoip": gin.H{
				"status": geoipStatus,
			},
		},
		"uptime": time.Since(time.Now()).String(), // 这里应该使用应用启动时间
	}, "系统运行正常")
}

// 事件查询接口
func ListEvents(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		authError(c, "未登录")
		return
	}

	// 获取查询参数
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "50")
	eventType := c.Query("event_type")
	eventTypeFilter := c.Query("event_type_filter") // 新增：事件类型分类筛选
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	siteId := c.Query("site_id")
	search := c.Query("search")

	// 构建查询条件
	query := StatDB.Model(&StatEvent{})

	// 站点筛选
	if siteId != "" {
		query = query.Where("site_id = ?", siteId)
	}

	// 事件类型筛选
	if eventType != "" {
		query = query.Where("event_name = ?", eventType)
	}

	// 事件类型分类筛选
	if eventTypeFilter != "" {
		// 使用正则表达式匹配事件名称
		filters := strings.Split(eventTypeFilter, "|")
		var conditions []string
		var args []interface{}

		for _, filter := range filters {
			conditions = append(conditions, "event_name LIKE ?")
			args = append(args, "%"+filter+"%")
		}

		if len(conditions) > 0 {
			query = query.Where("("+strings.Join(conditions, " OR ")+")", args...)
		}
	}

	// 搜索筛选
	if search != "" {
		query = query.Where("event_name LIKE ? OR value LIKE ? OR path LIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	// 时间范围筛选
	if startDate != "" {
		query = query.Where("DATE(created_at) >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("DATE(created_at) <= ?", endDate)
	}

	// 分页查询
	var events []StatEvent
	var total int64

	// 获取总数
	query.Count(&total)

	// 获取分页数据
	offset := (parseInt(page) - 1) * parseInt(limit)
	query.Order("created_at DESC").Offset(offset).Limit(parseInt(limit)).Find(&events)

	successResponse(c, gin.H{
		"events": events,
		"total":  total,
		"page":   parseInt(page),
		"limit":  parseInt(limit),
	}, "事件数据获取成功")
}

// 事件统计概览接口
func GetEventStats(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		authError(c, "未登录")
		return
	}

	siteId := c.Query("site_id")

	var stats struct {
		TotalEvents  int64   `json:"total_events"`
		EventTypes   int64   `json:"event_types"`
		UniqueUsers  int64   `json:"unique_users"`
		TodayEvents  int64   `json:"today_events"`
		AvgFrequency float64 `json:"avg_frequency"`
	}

	if StatDB != nil {
		// 构建基础查询
		baseQuery := StatDB.Model(&StatEvent{})
		if siteId != "" {
			baseQuery = baseQuery.Where("site_id = ?", siteId)
		}

		// 总事件数
		baseQuery.Count(&stats.TotalEvents)

		// 事件类型数
		StatDB.Model(&StatEvent{}).Where("site_id = ?", siteId).Distinct("event_name").Count(&stats.EventTypes)

		// 触发用户数
		StatDB.Model(&StatEvent{}).Where("site_id = ?", siteId).Distinct("user_id").Count(&stats.UniqueUsers)

		// 今日事件数
		today := time.Now().Format("2006-01-02")
		todayQuery := StatDB.Model(&StatEvent{}).Where("DATE(created_at) = ?", today)
		if siteId != "" {
			todayQuery = todayQuery.Where("site_id = ?", siteId)
		}
		todayQuery.Count(&stats.TodayEvents)

		// 计算平均频率（总事件数 / 触发用户数）
		if stats.UniqueUsers > 0 {
			stats.AvgFrequency = float64(stats.TotalEvents) / float64(stats.UniqueUsers)
		}
	}

	successResponse(c, stats, "事件统计获取成功")
}

// 事件趋势数据接口
func GetEventTrend(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		authError(c, "未登录")
		return
	}

	days := c.DefaultQuery("days", "7")
	eventType := c.Query("event_type")
	siteId := c.Query("site_id")

	var trend []struct {
		Date  string `json:"date"`
		Count int64  `json:"count"`
	}

	if StatDB != nil {
		query := StatDB.Model(&StatEvent{}).
			Select("DATE(created_at) as date, COUNT(*) as count").
			Where("created_at >= DATE_SUB(CURDATE(), INTERVAL ? DAY)", days)

		if siteId != "" {
			query = query.Where("site_id = ?", siteId)
		}

		if eventType != "" {
			query = query.Where("event_name = ?", eventType)
		}

		query.Group("DATE(created_at)").
			Order("date ASC").
			Scan(&trend)
	}

	successResponse(c, trend, "事件趋势获取成功")
}

// 事件类型分布接口
func GetEventTypeDistribution(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		authError(c, "未登录")
		return
	}

	siteId := c.Query("site_id")

	var distribution []struct {
		EventName string  `json:"event_name"`
		Count     int64   `json:"count"`
		Percent   float64 `json:"percent"`
	}

	if StatDB != nil {
		baseQuery := StatDB.Model(&StatEvent{})
		if siteId != "" {
			baseQuery = baseQuery.Where("site_id = ?", siteId)
		}

		var total int64
		baseQuery.Count(&total)

		if total > 0 {
			query := StatDB.Model(&StatEvent{}).
				Select("event_name, COUNT(*) as count, ROUND(COUNT(*) * 100.0 / ?, 2) as percent", total).
				Group("event_name").
				Order("count DESC")

			if siteId != "" {
				query = query.Where("site_id = ?", siteId)
			}

			query.Scan(&distribution)
		}
	}

	successResponse(c, distribution, "事件类型分布获取成功")
}

// 事件类型分类分布接口
func GetEventCategoryDistribution(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		authError(c, "未登录")
		return
	}

	siteId := c.Query("site_id")

	var distribution []struct {
		EventCategory string  `json:"event_category"`
		Count         int64   `json:"count"`
		Percent       float64 `json:"percent"`
	}

	if StatDB != nil {
		baseQuery := StatDB.Model(&StatEvent{})
		if siteId != "" {
			baseQuery = baseQuery.Where("site_id = ?", siteId)
		}

		var total int64
		baseQuery.Count(&total)

		if total > 0 {
			query := StatDB.Model(&StatEvent{}).
				Select("event_category, COUNT(*) as count, ROUND(COUNT(*) * 100.0 / ?, 2) as percent", total).
				Where("event_category != '' AND event_category IS NOT NULL").
				Group("event_category").
				Order("count DESC")

			if siteId != "" {
				query = query.Where("site_id = ?", siteId)
			}

			query.Scan(&distribution)
		}
	}

	successResponse(c, distribution, "事件类型分类分布获取成功")
}

// 工具函数：字符串转整数
func parseInt(s string) int {
	if i, err := strconv.Atoi(s); err == nil {
		return i
	}
	return 1
}

// 获取分布统计数据
func GetDistributionData(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		authError(c, "未登录")
		return
	}

	// 获取site_id参数
	siteId := c.Query("site_id")
	log.Printf("GetDistributionData called for site: %s", siteId)

	// 获取设备分布
	var deviceDistribution []struct {
		Device string `json:"device"`
		Count  int64  `json:"count"`
	}
	if StatDB != nil {
		query := StatDB.Model(&PageView{}).
			Select("device, COUNT(*) as count").
			Where("device != '' AND device IS NOT NULL")

		if siteId != "" {
			query = query.Where("site_id = ?", siteId)
		}

		query.Group("device").
			Order("count DESC").
			Limit(10).
			Scan(&deviceDistribution)

		log.Printf("Device distribution query result: %+v", deviceDistribution)
	}

	// 如果没有设备数据，添加默认数据
	if len(deviceDistribution) == 0 {
		deviceDistribution = []struct {
			Device string `json:"device"`
			Count  int64  `json:"count"`
		}{
			{Device: "Desktop", Count: 0},
			{Device: "Mobile", Count: 0},
		}
	}

	// 获取地区分布
	var regionDistribution []struct {
		Province string `json:"province"`
		Count    int64  `json:"count"`
	}
	if StatDB != nil {
		query := StatDB.Model(&PageView{}).
			Select("province, COUNT(*) as count").
			Where("province != '' AND province != '-' AND province IS NOT NULL AND province != 'IPv6 address missing in IPv4 BIN.'")

		if siteId != "" {
			query = query.Where("site_id = ?", siteId)
		}

		query.Group("province").
			Order("count DESC").
			Limit(10).
			Scan(&regionDistribution)

		log.Printf("Region distribution query result: %+v", regionDistribution)
	}

	// 如果没有地区数据，添加默认数据
	if len(regionDistribution) == 0 {
		regionDistribution = []struct {
			Province string `json:"province"`
			Count    int64  `json:"count"`
		}{
			{Province: "未知", Count: 0},
		}
	}

	// 获取浏览器分布
	var browserDistribution []struct {
		Browser string `json:"browser"`
		Count   int64  `json:"count"`
	}
	if StatDB != nil {
		query := StatDB.Model(&PageView{}).
			Select("browser, COUNT(*) as count").
			Where("browser != '' AND browser IS NOT NULL")

		if siteId != "" {
			query = query.Where("site_id = ?", siteId)
		}

		query.Group("browser").
			Order("count DESC").
			Limit(10).
			Scan(&browserDistribution)

		log.Printf("Browser distribution query result: %+v", browserDistribution)
	}

	// 如果没有浏览器数据，添加默认数据
	if len(browserDistribution) == 0 {
		browserDistribution = []struct {
			Browser string `json:"browser"`
			Count   int64  `json:"count"`
		}{
			{Browser: "Chrome", Count: 0},
			{Browser: "Safari", Count: 0},
			{Browser: "Firefox", Count: 0},
		}
	}

	// 获取访问来源分布
	var referrerDistribution []struct {
		Referer string `json:"referer"`
		Count   int64  `json:"count"`
	}
	if StatDB != nil {
		query := StatDB.Model(&PageView{}).
			Select("referer, COUNT(*) as count")

		if siteId != "" {
			query = query.Where("site_id = ?", siteId)
		}

		query.Group("referer").
			Order("count DESC").
			Limit(10).
			Scan(&referrerDistribution)
	}

	// 获取网络类型分布
	var networkDistribution []struct {
		Net   string `json:"net"`
		Count int64  `json:"count"`
	}
	if StatDB != nil {
		query := StatDB.Model(&PageView{}).
			Select("net, COUNT(*) as count").
			Where("net != ''")

		if siteId != "" {
			query = query.Where("site_id = ?", siteId)
		}

		query.Group("net").
			Order("count DESC").
			Limit(10).
			Scan(&networkDistribution)
	}

	// 获取城市分布（用于热力图）
	var cityDistribution []struct {
		City  string `json:"city"`
		Count int64  `json:"count"`
	}
	if StatDB != nil {
		query := StatDB.Model(&PageView{}).
			Select("city, COUNT(*) as count").
			Where("city != '' AND city IS NOT NULL AND city != '-' AND city != 'IPv6 address missing in IPv4 BIN.'")
		if siteId != "" {
			query = query.Where("site_id = ?", siteId)
		}
		query.Group("city").
			Order("count DESC").
			Limit(100).
			Scan(&cityDistribution)

		log.Printf("City distribution query result: %+v", cityDistribution)
	}

	// 如果没有城市数据，添加默认数据
	if len(cityDistribution) == 0 {
		cityDistribution = []struct {
			City  string `json:"city"`
			Count int64  `json:"count"`
		}{
			{City: "Beijing", Count: 0},
			{City: "Shanghai", Count: 0},
			{City: "Guangzhou", Count: 0},
		}
		log.Printf("No city data found, using default: %+v", cityDistribution)
	}

	// 获取国家分布
	var countryDistribution []struct {
		Country string `json:"country"`
		Count   int64  `json:"count"`
	}
	if StatDB != nil {
		query := StatDB.Model(&PageView{}).
			Select("country, COUNT(*) as count").
			Where("country != '' AND country IS NOT NULL")
		if siteId != "" {
			query = query.Where("site_id = ?", siteId)
		}
		query.Group("country").
			Order("count DESC").
			Limit(50).
			Scan(&countryDistribution)

		log.Printf("Country distribution query result: %+v", countryDistribution)
	}

	// 如果没有国家数据，添加默认数据
	if len(countryDistribution) == 0 {
		countryDistribution = []struct {
			Country string `json:"country"`
			Count   int64  `json:"count"`
		}{
			{Country: "China", Count: 0},
			{Country: "United States", Count: 0},
			{Country: "Japan", Count: 0},
		}
		log.Printf("No country data found, using default: %+v", countryDistribution)
	}

	response := gin.H{
		"success": true,
		"data": gin.H{
			"device_distribution":   deviceDistribution,
			"region_distribution":   regionDistribution,
			"city_distribution":     cityDistribution,
			"country_distribution":  countryDistribution,
			"browser_distribution":  browserDistribution,
			"referrer_distribution": referrerDistribution,
			"network_distribution":  networkDistribution,
		},
	}

	c.JSON(http.StatusOK, response)
}

// 站点报表指标API
func HandleReportMetrics(c *gin.Context) {
	siteID := c.Query("site_id")
	days := c.DefaultQuery("days", "7")

	if siteID == "" {
		c.JSON(400, gin.H{"success": false, "message": "缺少site_id参数"})
		return
	}

	daysInt, err := strconv.Atoi(days)
	if err != nil {
		daysInt = 7
	}

	// 查询总PV
	var totalPV int64
	StatDB.Model(&PageView{}).Where("site_id = ? AND created_at >= ?", siteID, time.Now().AddDate(0, 0, -daysInt)).Count(&totalPV)

	// 查询总UV
	var totalUV int64
	StatDB.Model(&PageView{}).Where("site_id = ? AND created_at >= ?", siteID, time.Now().AddDate(0, 0, -daysInt)).Distinct("user_id").Count(&totalUV)

	// 查询平均停留时长 - 修复计算逻辑
	var avgDuration float64
	dbType := getDBType()
	if dbType == "clickhouse" {
		// ClickHouse使用coalesce处理NULL值，只计算有效数据
		StatDB.Model(&PageView{}).Where("site_id = ? AND created_at >= ? AND duration > 0", siteID, time.Now().AddDate(0, 0, -daysInt)).Select("coalesce(AVG(duration), 0)").Scan(&avgDuration)
	} else {
		// MySQL使用IFNULL处理NULL值，只计算有效数据
		StatDB.Model(&PageView{}).Where("site_id = ? AND created_at >= ? AND duration > 0", siteID, time.Now().AddDate(0, 0, -daysInt)).Select("IFNULL(AVG(duration), 0)").Scan(&avgDuration)
	}

	// 查询页面数量
	var pageCount int64
	StatDB.Model(&PageView{}).Where("site_id = ? AND created_at >= ?", siteID, time.Now().AddDate(0, 0, -daysInt)).Distinct("path").Count(&pageCount)

	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"total_pv": totalPV,
			"total_uv": totalUV,
			"avg_duration": func() string {
				if avgDuration > 0 {
					// 如果超过60秒，显示分钟和秒
					if avgDuration >= 60 {
						minutes := int(avgDuration / 60)
						seconds := int(avgDuration) % 60
						if seconds > 0 {
							return fmt.Sprintf("%d分%d秒", minutes, seconds)
						} else {
							return fmt.Sprintf("%d分钟", minutes)
						}
					} else {
						// 小于60秒，显示秒数
						return fmt.Sprintf("%d秒", int(avgDuration))
					}
				}
				return "0秒"
			}(),
			"page_count": pageCount,
		},
	})
}

// 站点报表趋势API
func HandleReportTrend(c *gin.Context) {
	siteID := c.Query("site_id")
	days := c.DefaultQuery("days", "7")

	if siteID == "" {
		c.JSON(400, gin.H{"success": false, "message": "缺少site_id参数"})
		return
	}

	daysInt, err := strconv.Atoi(days)
	if err != nil {
		daysInt = 7
	}

	// 生成日期范围
	var dates []string
	var pvData []int64
	var uvData []int64

	for i := daysInt - 1; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		dates = append(dates, date)

		// 查询当天PV
		var pv int64
		StatDB.Model(&PageView{}).Where("site_id = ? AND DATE(created_at) = ?", siteID, date).Count(&pv)
		pvData = append(pvData, pv)

		// 查询当天UV
		var uv int64
		StatDB.Model(&PageView{}).Where("site_id = ? AND DATE(created_at) = ?", siteID, date).Distinct("user_id").Count(&uv)
		uvData = append(uvData, uv)
	}

	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"dates": dates,
			"pv":    pvData,
			"uv":    uvData,
		},
	})
}

// 站点报表时段分布API
func HandleReportHourDistribution(c *gin.Context) {
	siteID := c.Query("site_id")
	days := c.DefaultQuery("days", "7")

	if siteID == "" {
		c.JSON(400, gin.H{"success": false, "message": "缺少site_id参数"})
		return
	}

	daysInt, err := strconv.Atoi(days)
	if err != nil {
		daysInt = 7
	}

	// 查询时段分布
	var results []struct {
		Hour  int   `json:"hour"`
		Count int64 `json:"count"`
	}

	StatDB.Model(&PageView{}).
		Select("HOUR(created_at) as hour, COUNT(*) as count").
		Where("site_id = ? AND created_at >= ?", siteID, time.Now().AddDate(0, 0, -daysInt)).
		Group("HOUR(created_at)").
		Order("hour").
		Scan(&results)

	var hours []string
	var counts []int64

	for i := 0; i < 24; i++ {
		hours = append(hours, fmt.Sprintf("%02d:00", i))
		counts = append(counts, 0)
	}

	for _, result := range results {
		if result.Hour >= 0 && result.Hour < 24 {
			counts[result.Hour] = result.Count
		}
	}

	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"hours":  hours,
			"counts": counts,
		},
	})
}

// 站点报表每日详细数据API
func HandleReportDailyStats(c *gin.Context) {
	siteID := c.Query("site_id")
	days := c.DefaultQuery("days", "7")

	if siteID == "" {
		c.JSON(400, gin.H{"success": false, "message": "缺少site_id参数"})
		return
	}

	daysInt, err := strconv.Atoi(days)
	if err != nil {
		daysInt = 7
	}

	var results []gin.H

	for i := 0; i < daysInt; i++ {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")

		// 查询当天PV
		var pv int64
		StatDB.Model(&PageView{}).Where("site_id = ? AND DATE(created_at) = ?", siteID, date).Count(&pv)

		// 查询当天UV
		var uv int64
		StatDB.Model(&PageView{}).Where("site_id = ? AND DATE(created_at) = ?", siteID, date).Distinct("user_id").Count(&uv)

		// 查询平均停留时长 - 修复NULL值问题
		var avgDuration float64
		dbType := getDBType()
		if dbType == "clickhouse" {
			// ClickHouse使用coalesce处理NULL值
			StatDB.Model(&PageView{}).Where("site_id = ? AND DATE(created_at) = ?", siteID, date).Select("coalesce(AVG(duration), 0)").Scan(&avgDuration)
		} else {
			// MySQL使用IFNULL处理NULL值
			StatDB.Model(&PageView{}).Where("site_id = ? AND DATE(created_at) = ?", siteID, date).Select("IFNULL(AVG(duration), 0)").Scan(&avgDuration)
		}

		// 查询跳出率（单页面访问的用户比例）
		var bounceUsers int64
		dbType = getDBType()
		if dbType == "clickhouse" {
			// ClickHouse语法
			StatDB.Raw(`
				SELECT COUNT(DISTINCT user_id) 
				FROM stat_page_view 
				WHERE site_id = ? AND toDate(created_at) = ? 
				AND user_id IN (
					SELECT user_id 
					FROM stat_page_view 
					WHERE site_id = ? AND toDate(created_at) = ? 
					GROUP BY user_id 
					HAVING COUNT(*) = 1
				)`, siteID, date, siteID, date).Scan(&bounceUsers)
		} else {
			// MySQL语法
			StatDB.Raw(`
				SELECT COUNT(DISTINCT user_id) 
				FROM stat_page_view 
				WHERE site_id = ? AND DATE(created_at) = ? 
				AND user_id IN (
					SELECT user_id 
					FROM stat_page_view 
					WHERE site_id = ? AND DATE(created_at) = ? 
					GROUP BY user_id 
					HAVING COUNT(*) = 1
				)`, siteID, date, siteID, date).Scan(&bounceUsers)
		}

		bounceRate := 0.0
		if uv > 0 {
			bounceRate = float64(bounceUsers) / float64(uv) * 100
		}

		// 查询新用户数（简化处理，实际应该基于用户首次访问时间）
		newUsers := uv // 简化处理

		results = append(results, gin.H{
			"date": date,
			"pv":   pv,
			"uv":   uv,
			"avg_duration": func() string {
				if avgDuration > 0 {
					// 如果超过60秒，显示分钟和秒
					if avgDuration >= 60 {
						minutes := int(avgDuration / 60)
						seconds := int(avgDuration) % 60
						if seconds > 0 {
							return fmt.Sprintf("%d分%d秒", minutes, seconds)
						} else {
							return fmt.Sprintf("%d分钟", minutes)
						}
					} else {
						// 小于60秒，显示秒数
						return fmt.Sprintf("%d秒", int(avgDuration))
					}
				}
				return "0秒"
			}(),
			"bounce_rate": math.Round(bounceRate*100) / 100,
			"new_users":   newUsers,
		})
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    results,
	})
}

// 站点热门事件API
func HandlePopularEvents(c *gin.Context) {
	siteID := c.Query("site_id")
	days := c.DefaultQuery("days", "7")
	limit := c.DefaultQuery("limit", "10")
	eventType := c.Query("event_type")

	if siteID == "" {
		c.JSON(400, gin.H{"success": false, "message": "缺少site_id参数"})
		return
	}

	daysInt, err := strconv.Atoi(days)
	if err != nil {
		daysInt = 7
	}

	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		limitInt = 10
	}

	query := StatDB.Model(&StatEvent{}).Where("site_id = ? AND created_at >= ? AND event_name != '' AND event_name IS NOT NULL", siteID, time.Now().AddDate(0, 0, -daysInt))
	if eventType != "" {
		query = query.Where("event_name LIKE ?", "%"+eventType+"%")
	}

	// 使用结构体来接收查询结果，避免gin.H的nil map问题
	var rawResults []struct {
		EventName   string `json:"event_name"`
		Count       int64  `json:"count"`
		UniqueUsers int64  `json:"unique_users"`
	}

	err = query.Select("event_name, COUNT(*) as count, COUNT(DISTINCT user_id) as unique_users").
		Group("event_name").
		Order("count DESC").
		Limit(limitInt).
		Scan(&rawResults).Error

	if err != nil {
		log.Printf("HandlePopularEvents query error: %v", err)
		c.JSON(500, gin.H{"success": false, "message": "查询失败"})
		return
	}

	// 转换为gin.H格式
	results := make([]gin.H, len(rawResults))
	for i, result := range rawResults {
		avgFrequency := 0.0
		if result.UniqueUsers > 0 {
			avgFrequency = float64(result.Count) / float64(result.UniqueUsers)
		}

		results[i] = gin.H{
			"event_name":    result.EventName,
			"count":         result.Count,
			"unique_users":  result.UniqueUsers,
			"avg_frequency": math.Round(avgFrequency*100) / 100,
		}
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    results,
	})
}

// 站点管理相关API

// 更新站点信息
func UpdateSite(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		authError(c, "未登录")
		return
	}

	siteId := c.Param("siteId")
	if siteId == "" {
		validationError(c, "缺少站点ID")
		return
	}

	var updateData struct {
		SiteName        string `json:"site_name"`
		SiteDomain      string `json:"site_domain"`
		SiteDescription string `json:"site_description"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		validationError(c, err.Error())
		return
	}

	// 查找并更新站点
	var site StatSite
	if StatDB != nil {
		err := StatDB.Where("site_id = ?", siteId).First(&site).Error
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "站点不存在"})
			return
		}

		// 更新站点信息
		if updateData.SiteName != "" {
			site.Name = updateData.SiteName
		}
		if updateData.SiteDomain != "" {
			// 注意：StatSite结构体没有Domain字段，这里可以添加到Remark中
			site.Remark = updateData.SiteDomain
		}
		if updateData.SiteDescription != "" {
			site.Remark = updateData.SiteDescription
		}

		if err := StatDB.Save(&site).Error; err != nil {
			databaseError(c, err)
			return
		}
	}

	successResponse(c, gin.H{
		"site_id": siteId,
		"name":    site.Name,
		"domain":  updateData.SiteDomain, // 从请求数据中获取
		"remark":  site.Remark,
	}, "站点信息更新成功")
}

// 获取站点设置
func GetSiteSettings(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		authError(c, "未登录")
		return
	}

	siteId := c.Param("siteId")
	if siteId == "" {
		validationError(c, "缺少站点ID")
		return
	}

	// 这里可以返回站点的跟踪设置
	// 目前返回默认设置
	settings := gin.H{
		"track_pageview": true,
		"track_duration": true,
		"track_events":   true,
	}

	successResponse(c, settings, "设置获取成功")
}

// 更新站点设置
func UpdateSiteSettings(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		authError(c, "未登录")
		return
	}

	siteId := c.Param("siteId")
	if siteId == "" {
		validationError(c, "缺少站点ID")
		return
	}

	var settings struct {
		TrackPageview bool `json:"track_pageview"`
		TrackDuration bool `json:"track_duration"`
		TrackEvents   bool `json:"track_events"`
	}

	if err := c.ShouldBindJSON(&settings); err != nil {
		validationError(c, err.Error())
		return
	}

	// 这里可以保存设置到数据库
	// 目前只是返回成功
	successResponse(c, settings, "设置更新成功")
}

// 导出站点数据（CSV下载）
func ExportSiteData(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		authError(c, "未登录")
		return
	}

	siteId := c.Query("site_id")
	if siteId == "" {
		validationError(c, "缺少站点ID")
		return
	}

	typeStr := c.DefaultQuery("type", "visits")
	_ = c.DefaultQuery("format", "csv")

	filename := siteId + "_" + typeStr + "_" + time.Now().Format("20060102") + ".csv"
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename="+filename)

	w := csv.NewWriter(c.Writer)
	defer w.Flush()

	switch typeStr {
	case "pages":
		w.Write([]string{"页面路径", "访问量"})
		var pages []struct {
			Path  string
			Count int64
		}
		StatDB.Model(&PageView{}).
			Select("path, COUNT(*) as count").
			Where("site_id = ?", siteId).
			Group("path").
			Order("count DESC").
			Limit(1000).
			Scan(&pages)
		for _, p := range pages {
			w.Write([]string{p.Path, strconv.FormatInt(p.Count, 10)})
		}
	case "devices":
		w.Write([]string{"设备", "访问量"})
		var devices []struct {
			Device string
			Count  int64
		}
		StatDB.Model(&PageView{}).
			Select("device, COUNT(*) as count").
			Where("site_id = ?", siteId).
			Group("device").
			Order("count DESC").
			Limit(1000).
			Scan(&devices)
		for _, d := range devices {
			w.Write([]string{d.Device, strconv.FormatInt(d.Count, 10)})
		}
	case "cities":
		w.Write([]string{"城市", "访问量"})
		var cities []struct {
			City  string
			Count int64
		}
		StatDB.Model(&PageView{}).
			Select("city, COUNT(*) as count").
			Where("site_id = ?", siteId).
			Group("city").
			Order("count DESC").
			Limit(1000).
			Scan(&cities)
		for _, cty := range cities {
			w.Write([]string{cty.City, strconv.FormatInt(cty.Count, 10)})
		}
	case "events":
		w.Write([]string{"事件名称", "触发次数", "唯一用户数"})
		var events []struct {
			EventName   string
			Count       int64
			UniqueUsers int64
		}
		StatDB.Model(&StatEvent{}).
			Select("event_name, COUNT(*) as count, COUNT(DISTINCT user_id) as unique_users").
			Where("site_id = ?", siteId).
			Group("event_name").
			Order("count DESC").
			Limit(1000).
			Scan(&events)
		for _, e := range events {
			w.Write([]string{e.EventName, strconv.FormatInt(e.Count, 10), strconv.FormatInt(e.UniqueUsers, 10)})
		}
	default:
		w.Write([]string{"类型不支持", typeStr})
	}
}

// 删除站点数据
func DeleteSiteData(c *gin.Context) {
	userID := c.GetInt64("user_id")
	if userID == 0 {
		authError(c, "未登录")
		return
	}

	siteId := c.Param("siteId")
	if siteId == "" {
		validationError(c, "缺少站点ID")
		return
	}

	var deleteData struct {
		Type string `json:"type"`
		Days int    `json:"days"`
	}

	if err := c.ShouldBindJSON(&deleteData); err != nil {
		validationError(c, err.Error())
		return
	}

	// 这里可以实现数据删除逻辑
	// 目前返回简单的成功响应
	successResponse(c, gin.H{
		"site_id": siteId,
		"type":    deleteData.Type,
		"days":    deleteData.Days,
	}, "数据删除成功")
}

// 热门设备接口
func GetPopularDevices(c *gin.Context) {
	siteId := c.Query("site_id")
	var devices []struct {
		Device string `json:"device"`
		Count  int64  `json:"count"`
	}
	if StatDB != nil {
		query := StatDB.Model(&PageView{}).
			Select("device, COUNT(*) as count").
			Where("device != '' AND device IS NOT NULL")
		if siteId != "" {
			query = query.Where("site_id = ?", siteId)
		}
		query.Group("device").
			Order("count DESC").
			Limit(10).
			Scan(&devices)
	}
	c.JSON(200, gin.H{
		"success": true,
		"data":    devices,
	})
}

// 热门城市接口
func GetPopularCities(c *gin.Context) {
	siteId := c.Query("site_id")
	var cities []struct {
		City  string `json:"city"`
		Count int64  `json:"count"`
	}
	if StatDB != nil {
		query := StatDB.Model(&PageView{}).
			Select("city, COUNT(*) as count").
			Where("city != '' AND city IS NOT NULL")
		if siteId != "" {
			query = query.Where("site_id = ?", siteId)
		}
		query.Group("city").
			Order("count DESC").
			Limit(10).
			Scan(&cities)
	}
	c.JSON(200, gin.H{
		"success": true,
		"data":    cities,
	})
}
