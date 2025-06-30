package stat_test

import (
	"strings"
	"testing"

	"github.com/iluyuns/stat"
)

func TestScriptGeneration(t *testing.T) {
	// 测试脚本生成
	siteID := "test_site_123"
	apiURL := "https://example.com"

	script := stat.GenerateScriptContent(siteID, apiURL)

	if script == "" {
		t.Fatal("Generated script should not be empty")
	}

	// 检查脚本是否包含必要的函数
	if !strings.Contains(script, "getApiUrl") {
		t.Error("Script should contain getApiUrl function")
	}

	if !strings.Contains(script, "getSiteId") {
		t.Error("Script should contain getSiteId function")
	}

	if !strings.Contains(script, "reportPV") {
		t.Error("Script should contain reportPV function")
	}

	if !strings.Contains(script, "statReportEvent") {
		t.Error("Script should contain statReportEvent function")
	}

	t.Log("✅ Script generation test passed")
}

func TestScriptCache(t *testing.T) {
	// 测试脚本缓存
	siteID := "cache_test_site"
	apiURL := "https://cache.example.com"

	// 第一次生成
	script1 := stat.GenerateScriptContent(siteID, apiURL)

	// 第二次生成（应该从缓存获取）
	script2 := stat.GenerateScriptContent(siteID, apiURL)

	if script1 != script2 {
		t.Error("Cached script should be identical")
	}

	// 检查缓存大小
	cacheSize := stat.GetScriptCacheSize()
	if cacheSize == 0 {
		t.Error("Cache should contain at least one entry")
	}

	t.Logf("✅ Script cache test passed, cache size: %d", cacheSize)
}

func TestScriptCacheClear(t *testing.T) {
	// 测试缓存清除

	// 生成一些脚本
	stat.GenerateScriptContent("clear_test_1", "https://test1.com")
	stat.GenerateScriptContent("clear_test_2", "https://test2.com")

	// 清除缓存
	stat.ClearScriptCache()

	finalSize := stat.GetScriptCacheSize()
	if finalSize != 0 {
		t.Errorf("Cache should be empty after clear, got %d", finalSize)
	}

	t.Log("✅ Script cache clear test passed")
}

func TestDifferentApiUrls(t *testing.T) {
	// 测试不同API地址的脚本生成
	siteID := "multi_api_test"

	apiURL1 := "https://api1.example.com"
	apiURL2 := "https://api2.example.com"

	script1 := stat.GenerateScriptContent(siteID, apiURL1)
	script2 := stat.GenerateScriptContent(siteID, apiURL2)

	// 当前实现中，脚本内容应该是相同的，因为模板没有使用API地址
	// 但缓存键应该是不同的
	if script1 != script2 {
		t.Error("Scripts should be identical for current implementation")
	}

	t.Log("✅ Different API URLs test passed")
}
