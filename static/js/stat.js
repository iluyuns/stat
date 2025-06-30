/**
 * 网站统计埋点脚本
 * 
 * 功能说明：
 * 1. 自动采集页面访问数据（PV、UV、停留时长）
 * 2. 支持自定义事件埋点上报
 * 3. 自动识别设备、浏览器、操作系统信息
 * 4. 支持事件类型自动分类
 * 
 * 使用方法：
 * 1. 引入脚本：<script src="/static/js/stat.js" site-id="your_site_id"></script>
 * 2. 事件埋点：statReportEvent('button_click', 'submit_button', {page: 'home'})
 * 
 * 作者：统计系统
 * 版本：1.0
 */
(function() {
    /**
     * 获取API地址
     * 优先级：script标签属性 > 全局变量 > 当前域名
     * @returns {string} API基础地址
     */
    function getApiUrl() {
      // 1. 尝试从当前 script 标签属性获取
      var scripts = document.getElementsByTagName('script');
      for (var i = 0; i < scripts.length; i++) {
        if (scripts[i].src && scripts[i].src.indexOf('stat.js') !== -1) {
          if (scripts[i].getAttribute('data-api-url')) {
            return scripts[i].getAttribute('data-api-url');
          }
        }
      }
      // 2. 尝试从全局变量获取
      if (window.statApiUrl) return window.statApiUrl;
      // 3. 默认使用当前域名
      return window.location.origin;
    }

    /**
     * 自动获取站点ID
     * 优先级：script标签属性 > 全局变量
     * @returns {string} 站点唯一标识
     */
    function getSiteId() {
      // 1. 尝试从当前 script 标签属性获取
      var scripts = document.getElementsByTagName('script');
      for (var i = 0; i < scripts.length; i++) {
        if (scripts[i].src && scripts[i].src.indexOf('stat.js') !== -1 && scripts[i].getAttribute('site-id')) {
          return scripts[i].getAttribute('site-id');
        }
      }
      // 2. 尝试从全局变量获取
      if (window.statSiteId) return window.statSiteId;
      return '';
    }

    /**
     * 匿名用户ID生成与获取
     * 使用Cookie存储，有效期1年
     * @returns {string} 匿名用户唯一标识
     */
    function getOrCreateAnonId() {
      var key = 'stat_anon_id';
      var match = document.cookie.match(new RegExp('(^| )' + key + '=([^;]+)'));
      var id = match ? match[2] : '';
      if (!id) {
        // 生成UUID格式的匿名ID
        id = ([1e7]+-1e3+-4e3+-8e3+-1e11).replace(/[018]/g, function(c) {
          return (c ^ crypto.getRandomValues(new Uint8Array(1))[0] & 15 >> c / 4).toString(16);
        });
        document.cookie = key + '=' + id + ';path=/;max-age=31536000';
      }
      return id;
    }

    /**
     * 自动获取用户ID
     * 优先级：全局变量 > 匿名ID
     * @returns {string} 用户唯一标识
     */
    function getUserId() {
      return window.statUserId || getOrCreateAnonId();
    }

    /**
     * 获取环境信息
     * 包含页面路径、来源、用户代理、屏幕信息等
     * @returns {object} 环境信息对象
     */
    function getEnvInfo() {
      return {
        path: location.pathname,           // 当前页面路径
        ref: document.referrer,            // 来源页面
        ua: navigator.userAgent,           // 用户代理字符串
        width: screen.width,               // 屏幕宽度
        height: screen.height,             // 屏幕高度
        colorDepth: screen.colorDepth,     // 屏幕色深
        lang: navigator.language,          // 浏览器语言
        net: navigator.connection ? navigator.connection.effectiveType : '', // 网络类型
        site_id: getSiteId(),              // 站点ID
        user_id: getUserId(),              // 用户ID
        // 可选：采集更多环境信息
      };
    }
  
    /**
     * 上报页面访问数据（PV）
     * 页面加载时自动调用
     */
    function reportPV() {
      var apiUrl = getApiUrl();
      fetch(apiUrl + '/api/track/pv', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(getEnvInfo())
      });
    }
  
    /**
     * 停留时长上报
     * 页面离开时上报用户停留时长
     */
    var enterTime = Date.now();
    window.addEventListener('beforeunload', function() {
      var duration = Math.round((Date.now() - enterTime) / 1000);
      var apiUrl = getApiUrl();
      fetch(apiUrl + '/api/track/duration', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ path: location.pathname, duration: duration })
      });
    });
  
    // 页面加载时立即上报PV
    reportPV();

    /**
     * 事件埋点上报函数
     * 全局函数，供页面调用
     * 
     * @param {string} eventName - 事件名称，如 'button_click', 'form_submit'
     * @param {string} value - 事件值，如按钮文本、表单数据等
     * @param {object} extra - 额外数据，如页面信息、用户行为等
     * 
     * 使用示例：
     * statReportEvent('button_click', 'submit_button', {page: 'home', section: 'header'})
     * statReportEvent('form_submit', 'contact_form', {form_type: 'contact'})
     * statReportEvent('scroll', 'page_scroll', {scroll_percent: 50})
     */
    window.statReportEvent = function(eventName, value, extra) {
      var data = Object.assign(getEnvInfo(), {
        event_name: eventName,                    // 事件名称
        value: value || '',                       // 事件值
        extra: extra ? JSON.stringify(extra) : '' // 额外数据（JSON字符串）
      });
      var apiUrl = getApiUrl();
      fetch(apiUrl + '/api/track/event', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data)
      });
    };
  })();