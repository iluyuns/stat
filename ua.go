package stat

import (
	"regexp"
	"strings"
)

type UAInfo struct {
	Device  string `json:"device"`
	OS      string `json:"os"`
	Browser string `json:"browser"`
}

func UA(ua string) (pv UAInfo) {
	pv = UAInfo{}
	// 解析设备、系统、浏览器信息（完全复刻前端算法）
	if ua != "" {
		// 1. 设备类型识别（品牌+型号优先，国产品牌、Android等）
		androidModelMatch := regexp.MustCompile(`Android [\\d.]+; ([^;\\)\\[]+)[;\\)\\[]`).FindStringSubmatch(ua)
		if len(androidModelMatch) > 1 {
			model := strings.TrimSpace(androidModelMatch[1])
			if regexp.MustCompile(`HUAWEI|HONOR`).MatchString(model) {
				pv.Device = "Huawei " + model
			} else if regexp.MustCompile(`MI|Redmi|Xiaomi`).MatchString(model) {
				pv.Device = "Xiaomi " + model
			} else if regexp.MustCompile(`OPPO`).MatchString(model) {
				pv.Device = "OPPO " + model
			} else if regexp.MustCompile(`VIVO`).MatchString(model) {
				pv.Device = "Vivo " + model
			} else if regexp.MustCompile(`SM-|Samsung`).MatchString(model) {
				pv.Device = "Samsung " + model
			} else if regexp.MustCompile(`ONEPLUS`).MatchString(model) {
				pv.Device = "OnePlus " + model
			} else if regexp.MustCompile(`MEIZU`).MatchString(model) {
				pv.Device = "Meizu " + model
			} else if regexp.MustCompile(`REALME`).MatchString(model) {
				pv.Device = "Realme " + model
			} else if regexp.MustCompile(`NUBIA`).MatchString(model) {
				pv.Device = "Nubia " + model
			} else if regexp.MustCompile(`ZTE`).MatchString(model) {
				pv.Device = "ZTE " + model
			} else if regexp.MustCompile(`LENOVO`).MatchString(model) {
				pv.Device = "Lenovo " + model
			} else if regexp.MustCompile(`SONY`).MatchString(model) {
				pv.Device = "Sony " + model
			} else if regexp.MustCompile(`Pixel`).MatchString(model) {
				pv.Device = "Google Pixel " + model
			} else if regexp.MustCompile(`Lumia`).MatchString(model) {
				pv.Device = "Microsoft Lumia " + model
			} else if regexp.MustCompile(`HarmonyOS`).MatchString(ua) || regexp.MustCompile(`HarmonyOS`).MatchString(model) {
				pv.Device = "HarmonyOS " + model
			} else if regexp.MustCompile(`MiOS|Pengpai|SurgeOS`).MatchString(ua) || regexp.MustCompile(`MiOS|Pengpai|SurgeOS`).MatchString(model) {
				pv.Device = "PengpaiOS " + model
			} else if model != "" && !regexp.MustCompile(`Build`).MatchString(model) {
				pv.Device = model
			}
		}
		if pv.Device == "" {
			if regexp.MustCompile(`iPhone`).MatchString(ua) {
				pv.Device = "Apple iPhone"
			} else if regexp.MustCompile(`iPad`).MatchString(ua) {
				pv.Device = "Apple iPad"
			} else if regexp.MustCompile(`HUAWEI|HONOR`).MatchString(ua) {
				pv.Device = "Huawei"
			} else if regexp.MustCompile(`MI|Redmi|Xiaomi`).MatchString(ua) {
				pv.Device = "Xiaomi"
			} else if regexp.MustCompile(`OPPO`).MatchString(ua) {
				pv.Device = "OPPO"
			} else if regexp.MustCompile(`VIVO`).MatchString(ua) {
				pv.Device = "Vivo"
			} else if regexp.MustCompile(`SM-|Samsung`).MatchString(ua) {
				pv.Device = "Samsung"
			} else if regexp.MustCompile(`ONEPLUS`).MatchString(ua) {
				pv.Device = "OnePlus"
			} else if regexp.MustCompile(`MEIZU`).MatchString(ua) {
				pv.Device = "Meizu"
			} else if regexp.MustCompile(`REALME`).MatchString(ua) {
				pv.Device = "Realme"
			} else if regexp.MustCompile(`NUBIA`).MatchString(ua) {
				pv.Device = "Nubia"
			} else if regexp.MustCompile(`ZTE`).MatchString(ua) {
				pv.Device = "ZTE"
			} else if regexp.MustCompile(`LENOVO`).MatchString(ua) {
				pv.Device = "Lenovo"
			} else if regexp.MustCompile(`SONY`).MatchString(ua) {
				pv.Device = "Sony"
			} else if regexp.MustCompile(`Pixel`).MatchString(ua) {
				pv.Device = "Google Pixel"
			} else if regexp.MustCompile(`Lumia`).MatchString(ua) {
				pv.Device = "Microsoft Lumia"
			} else if regexp.MustCompile(`HarmonyOS`).MatchString(ua) {
				pv.Device = "HarmonyOS"
			} else if regexp.MustCompile(`MiOS|Pengpai|SurgeOS`).MatchString(ua) {
				pv.Device = "PengpaiOS"
			} else if regexp.MustCompile(`Android`).MatchString(ua) {
				pv.Device = "Android"
			} else if regexp.MustCompile(`iOS`).MatchString(ua) {
				pv.Device = "iOS"
			} else if regexp.MustCompile(`Mobile`).MatchString(ua) {
				pv.Device = "Mobile"
			} else {
				pv.Device = "Desktop"
			}
		}

		// 2. 操作系统识别
		if strings.Contains(ua, "Windows") {
			pv.OS = "Windows"
		} else if strings.Contains(ua, "Mac OS X") {
			pv.OS = "macOS"
		} else if strings.Contains(ua, "Linux") {
			pv.OS = "Linux"
		} else if strings.Contains(ua, "iOS") {
			pv.OS = "iOS"
		} else if strings.Contains(ua, "Android") {
			pv.OS = "Android"
		} else if regexp.MustCompile(`HarmonyOS`).MatchString(ua) {
			pv.OS = "HarmonyOS"
		} else if regexp.MustCompile(`MiOS|Pengpai|SurgeOS`).MatchString(ua) {
			pv.OS = "PengpaiOS"
		} else {
			pv.OS = "Unknown"
		}

		// 3. 浏览器识别（主流App/国产/主流浏览器）
		if regexp.MustCompile(`MicroMessenger`).MatchString(ua) {
			pv.Browser = "WeChat"
		} else if regexp.MustCompile(`wxwork`).MatchString(ua) {
			pv.Browser = "WeCom"
		} else if regexp.MustCompile(`DingTalk`).MatchString(ua) {
			pv.Browser = "DingTalk"
		} else if regexp.MustCompile(`Lark`).MatchString(ua) {
			pv.Browser = "Lark"
		} else if regexp.MustCompile(`QQ/|QQBrowser`).MatchString(ua) {
			pv.Browser = "QQ"
		} else if regexp.MustCompile(`Weibo`).MatchString(ua) {
			pv.Browser = "Weibo"
		} else if regexp.MustCompile(`AlipayClient`).MatchString(ua) {
			pv.Browser = "Alipay"
		} else if regexp.MustCompile(`Telegram`).MatchString(ua) {
			pv.Browser = "Telegram"
		} else if regexp.MustCompile(`baiduboxapp`).MatchString(ua) {
			pv.Browser = "BaiduApp"
		} else if regexp.MustCompile(`NewsArticle|Toutiao`).MatchString(ua) {
			pv.Browser = "Toutiao"
		} else if regexp.MustCompile(`Aweme`).MatchString(ua) {
			pv.Browser = "Douyin"
		} else if regexp.MustCompile(`XiaoHongShu`).MatchString(ua) {
			pv.Browser = "XiaoHongShu"
		} else if regexp.MustCompile(`Kwai`).MatchString(ua) {
			pv.Browser = "Kuaishou"
		} else if regexp.MustCompile(`FBAV`).MatchString(ua) {
			pv.Browser = "Facebook"
		} else if regexp.MustCompile(`Instagram`).MatchString(ua) {
			pv.Browser = "Instagram"
		} else if regexp.MustCompile(`Twitter`).MatchString(ua) {
			pv.Browser = "Twitter"
		} else if regexp.MustCompile(`HuaweiBrowser`).MatchString(ua) {
			pv.Browser = "Huawei"
		} else if regexp.MustCompile(`HarmonyOS`).MatchString(ua) {
			pv.Browser = "HarmonyOS"
		} else if regexp.MustCompile(`MiuiBrowser`).MatchString(ua) {
			pv.Browser = "Miui"
		} else if regexp.MustCompile(`HeyTapBrowser`).MatchString(ua) {
			pv.Browser = "OPPO"
		} else if regexp.MustCompile(`VivoBrowser`).MatchString(ua) {
			pv.Browser = "Vivo"
		} else if regexp.MustCompile(`SamsungBrowser`).MatchString(ua) {
			pv.Browser = "Samsung"
		} else if regexp.MustCompile(`Maxthon`).MatchString(ua) {
			pv.Browser = "Maxthon"
		} else if regexp.MustCompile(`LieBaoFast`).MatchString(ua) {
			pv.Browser = "Liebao"
		} else if regexp.MustCompile(`2345Explorer`).MatchString(ua) {
			pv.Browser = "2345"
		} else if regexp.MustCompile(`UCWEB`).MatchString(ua) {
			pv.Browser = "UCWEB"
		} else if regexp.MustCompile(`UCBrowser`).MatchString(ua) {
			pv.Browser = "UC"
		} else if regexp.MustCompile(`Opera Mini`).MatchString(ua) {
			pv.Browser = "Opera Mini"
		} else if regexp.MustCompile(`QihooBrowser|QHBrowser`).MatchString(ua) {
			pv.Browser = "360"
		} else if regexp.MustCompile(`SogouMobileBrowser`).MatchString(ua) {
			pv.Browser = "Sogou"
		} else if regexp.MustCompile(`Edg/`).MatchString(ua) {
			pv.Browser = "Edge (Chromium)"
		} else if strings.Contains(ua, "Edge") {
			pv.Browser = "Edge"
		} else if strings.Contains(ua, "OPR") || strings.Contains(ua, "Opera") {
			pv.Browser = "Opera"
		} else if strings.Contains(ua, "Chrome") {
			pv.Browser = "Chrome"
		} else if strings.Contains(ua, "Firefox") {
			pv.Browser = "Firefox"
		} else if strings.Contains(ua, "Safari") && !strings.Contains(ua, "Chrome") {
			if regexp.MustCompile(`iPhone|iPad|iPod`).MatchString(ua) {
				pv.Browser = "Safari (iOS)"
			} else if regexp.MustCompile(`Macintosh`).MatchString(ua) {
				pv.Browser = "Safari (macOS)"
			} else if regexp.MustCompile(`Vision`).MatchString(ua) {
				pv.Browser = "Safari (visionOS)"
			} else {
				pv.Browser = "Safari"
			}
		} else {
			pv.Browser = "Unknown"
		}
	}
	return
}
