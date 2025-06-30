package stat

import (
	"database/sql"
	"time"

	"gorm.io/gorm"
)

// 统计ID管理
// 用于多应用/多页面唯一标识
// 可扩展更多字段如备注、创建人等
type StatSite struct {
	ID        int64          `json:"id" gorm:"primaryKey;autoIncrement"`
	SiteId    string         `json:"site_id" gorm:"size:64;uniqueIndex"`
	Name      string         `json:"name" gorm:"size:128;index"`
	Remark    string         `json:"remark" gorm:"size:255"`
	CreatedBy string         `json:"created_by" gorm:"size:64;index"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (StatSite) TableName() string { return "stat_site" }

// 事件埋点结构体
type StatEvent struct {
	ID            int64          `json:"id" gorm:"primaryKey;autoIncrement"`
	SiteId        string         `json:"site_id" gorm:"size:64;index:idx_event_site_time,priority:1"`
	EventName     string         `json:"event_name" gorm:"size:64;index:idx_event_site_time,priority:2"`
	EventCategory string         `json:"event_category" gorm:"size:32;index"`
	UserID        string         `json:"user_id" gorm:"size:64;index"`
	Value         string         `json:"value" gorm:"size:255"`
	Extra         string         `json:"extra" gorm:"type:text"`
	Path          string         `json:"path" gorm:"size:255;index:idx_site_path,priority:2"`
	IP            string         `json:"ip" gorm:"size:64;index"`
	UA            string         `json:"ua" gorm:"size:255"`
	Referer       string         `json:"referer" gorm:"size:255"`
	City          string         `json:"city" gorm:"size:64;index"`
	Province      string         `json:"province" gorm:"size:64;index"`
	ISP           string         `json:"isp" gorm:"size:64;index"`
	Device        string         `json:"device" gorm:"size:64;index"`
	OS            string         `json:"os" gorm:"size:64;index"`
	Browser       string         `json:"browser" gorm:"size:64;index"`
	Screen        string         `json:"screen" gorm:"size:32"`
	Net           string         `json:"net" gorm:"size:32"`
	Country       string         `json:"country" gorm:"size:64;index"`
	CreatedAt     time.Time      `json:"created_at" gorm:"index:idx_event_site_time,priority:3;autoCreateTime"`
	UpdatedAt     time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt     gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (StatEvent) TableName() string { return "stat_event" }

// 页面访问结构体
type PageView struct {
	ID        int64          `json:"id" gorm:"primaryKey;autoIncrement"`
	SiteId    string         `json:"site_id" gorm:"size:64;index:idx_site_time,priority:1;index:idx_site_path,priority:1"`
	Path      string         `json:"path" gorm:"size:255;index:idx_site_path,priority:2"`
	IP        string         `json:"ip" gorm:"size:64;index"`
	UA        string         `json:"ua" gorm:"size:255"`
	Referer   string         `json:"referer" gorm:"size:255"`
	UserID    string         `json:"user_id" gorm:"size:64;index"`
	City      string         `json:"city" gorm:"size:64;index"`
	Province  string         `json:"province" gorm:"size:64;index"`
	ISP       string         `json:"isp" gorm:"size:64;index"`
	Device    string         `json:"device" gorm:"size:64;index"`
	OS        string         `json:"os" gorm:"size:64;index"`
	Browser   string         `json:"browser" gorm:"size:64;index"`
	Screen    string         `json:"screen" gorm:"size:32"`
	Net       string         `json:"net" gorm:"size:32"`
	Country   string         `json:"country" gorm:"size:64;index"`
	Duration  int            `json:"duration" gorm:"default:0"`
	CreatedAt time.Time      `json:"created_at" gorm:"index:idx_site_time,priority:2;autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (PageView) TableName() string { return "stat_page_view" }

// 用户模型
type User struct {
	ID        int64          `json:"id" gorm:"primaryKey;autoIncrement"`
	Username  string         `json:"username" gorm:"size:64;uniqueIndex;not null"`
	Email     string         `json:"email" gorm:"size:128;uniqueIndex;not null"`
	Password  string         `json:"-" gorm:"size:128;not null"` // 不返回密码
	Nickname  string         `json:"nickname" gorm:"size:64"`
	Avatar    string         `json:"avatar" gorm:"size:255"`
	Status    int            `json:"status" gorm:"default:1"` // 1:正常 0:禁用
	LastLogin sql.NullTime   `json:"last_login"`
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (User) TableName() string { return "stat_user" }

// 验证码模型
type VerifyCode struct {
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Email     string    `json:"email" gorm:"size:128;index;not null"`
	Code      string    `json:"code" gorm:"size:8;not null"`
	Type      string    `json:"type" gorm:"size:16;not null"` // register, reset, login
	Used      bool      `json:"used" gorm:"default:false"`
	ExpiredAt time.Time `json:"expired_at"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (VerifyCode) TableName() string { return "stat_verify_code" }

// 用户会话模型
type UserSession struct {
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID    int64     `json:"user_id" gorm:"index;not null"`
	Token     string    `json:"token" gorm:"size:128;uniqueIndex;not null"`
	IP        string    `json:"ip" gorm:"size:64"`
	UA        string    `json:"ua" gorm:"size:255"`
	ExpiredAt time.Time `json:"expired_at"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (UserSession) TableName() string { return "stat_user_session" }
