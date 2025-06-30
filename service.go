package stat

import "context"

// 存储事件埋点
type EventService interface {
	SaveEvent(ctx context.Context, event StatEvent) error
}

// 存储页面访问
type PageViewService interface {
	SavePageView(ctx context.Context, pv PageView) error
}
