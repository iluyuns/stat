# 统计模块说明

## 目录结构
- api.go：统计相关接口（如PV、事件、时长上报、报表查询、siteId管理）
- model.go：数据结构定义（事件、页面访问，含地理、设备等字段、siteId）
- service.go：业务逻辑接口、IP库接口
- middleware.go：PV/UV自动采集中间件
- stat.js：前端埋点脚本
- stat.html：统计报表可视化页面

## 功能
- PV/UV统计：自动采集页面访问数据，含地理、设备、分辨率、网络等
- 事件统计：支持手动埋点上报
- 停留时长统计：自动上报页面停留时长
- 多应用/多页面支持：每个统计对象分配唯一 siteId
- 统计报表：支持可视化展示PV/UV/事件/地域/设备等

## API接口
- POST /api/track/pv        页面访问上报（需带 site_id）
- POST /api/track/event     事件埋点上报（需带 site_id）
- POST /api/track/duration  停留时长上报（需带 site_id）
- GET  /api/stat/report     统计数据聚合查询（支持 site_id、时间段、类型等）
- POST /api/stat/site       新建统计ID（siteId）
- GET  /api/stat/site       查询所有统计ID
- DELETE /api/stat/site/:id 删除统计ID

## 字段说明
- site_id：统计唯一ID，区分不同应用/页面
- path、ref、ua、screen、net、user_id、city、province、isp、device、os、browser、duration 等

## 前端用法
- 引入 stat.js，自动采集PV、停留时长、事件等
- 支持通过 `<script src=... site-id="xxxx"></script>` 或 `window.statSiteId = 'xxxx'` 传递 siteId
- 手动事件埋点：window.statReportEvent(eventName, value, extra)
- 示例代码片段（后台生成）：
  ```html
  <script src="https://stat.minis.app/static/js/stat.js" site-id="xxxxxx"></script>
  ```

## 后端用法
- 主程序初始化时调用 stat.SetDB(db) 注入数据库
- 实现 IpResolver 接口并通过 stat.SetIpResolver 注入
- 路由注册上述API接口
- 提供 siteId 管理接口，支持多应用/多页面统计
- 统计查询接口支持按 siteId、时间段、类型等聚合

## 报表可视化（stat.html）
- 使用 ECharts/Chart.js 等开源可视化库，展示PV/UV/事件趋势、地域、设备分布等
- 支持按 siteId、时间段、类型等查询
- 示例见 stat.html

## 参考百度统计的实现思路
- 每个统计对象分配唯一 siteId，前端埋点代码需带 siteId
- 后端所有统计数据均按 siteId 归集，支持多应用/多页面分离统计
- 提供统计代码片段自动生成、管理后台、可视化报表等能力
- 支持丰富的统计维度（地理、设备、来源、行为、时长等）
- 支持事件自定义埋点，便于灵活扩展

## 扩展
- 可根据业务需求扩展统计字段和存储方式
- 可对接管理后台、报表、导出等 