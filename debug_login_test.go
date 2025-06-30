package stat_test

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"testing"
)

// TestPasswordHash 测试密码哈希计算
func TestPasswordHash(t *testing.T) {
	password := "admin123"
	expectedHash := "0192023a7bbd73250516f069df18b500"

	// 计算密码的MD5哈希
	hash := md5.Sum([]byte(password))
	calculatedHash := hex.EncodeToString(hash[:])

	t.Logf("Password: %s", password)
	t.Logf("Expected Hash: %s", expectedHash)
	t.Logf("Calculated Hash: %s", calculatedHash)

	if calculatedHash != expectedHash {
		t.Errorf("Password hash mismatch! Expected: %s, Got: %s", expectedHash, calculatedHash)
	} else {
		t.Log("✅ Password hash calculation is correct!")
	}
}

// TestPasswordVerification 测试密码验证逻辑
func TestPasswordVerification(t *testing.T) {
	testCases := []struct {
		name     string
		password string
		hash     string
		expected bool
	}{
		{
			name:     "Correct password",
			password: "admin123",
			hash:     "0192023a7bbd73250516f069df18b500",
			expected: true,
		},
		{
			name:     "Wrong password",
			password: "wrongpass",
			hash:     "0192023a7bbd73250516f069df18b500",
			expected: false,
		},
		{
			name:     "Empty password",
			password: "",
			hash:     "0192023a7bbd73250516f069df18b500",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := checkPassword(tc.password, tc.hash)
			if result != tc.expected {
				t.Errorf("Password verification failed! Expected: %v, Got: %v", tc.expected, result)
			} else {
				t.Logf("✅ %s: %v", tc.name, result)
			}
		})
	}
}

// TestLoginFlow 测试完整登录流程
func TestLoginFlow(t *testing.T) {
	// 模拟数据库中的用户数据
	userData := map[string]string{
		"admin": "0192023a7bbd73250516f069df18b500", // admin123
		"user1": "5f4dcc3b5aa765d61d8327deb882cf99", // password
	}

	testCases := []struct {
		name     string
		username string
		password string
		expected bool
	}{
		{
			name:     "Valid admin login",
			username: "admin",
			password: "admin123",
			expected: true,
		},
		{
			name:     "Invalid admin password",
			username: "admin",
			password: "wrongpass",
			expected: false,
		},
		{
			name:     "Valid user login",
			username: "user1",
			password: "password",
			expected: true,
		},
		{
			name:     "Non-existent user",
			username: "nonexistent",
			password: "anypass",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			storedHash, exists := userData[tc.username]
			if !exists {
				if tc.expected {
					t.Errorf("User %s should exist but doesn't", tc.username)
				} else {
					t.Logf("✅ User %s correctly doesn't exist", tc.username)
				}
				return
			}

			result := checkPassword(tc.password, storedHash)
			if result != tc.expected {
				t.Errorf("Login failed! Expected: %v, Got: %v", tc.expected, result)
			} else {
				t.Logf("✅ %s: %v", tc.name, result)
			}
		})
	}
}

// TestHashGeneration 测试哈希生成函数
func TestHashGeneration(t *testing.T) {
	passwords := []string{
		"admin123",
		"password",
		"123456",
		"",
		"complex_password_123!@#",
	}

	for _, password := range passwords {
		t.Run(fmt.Sprintf("Hash_%s", password), func(t *testing.T) {
			hash := generateHash(password)
			t.Logf("Password: %s -> Hash: %s", password, hash)

			// 验证哈希长度（MD5应该是32位十六进制）
			if len(hash) != 32 {
				t.Errorf("Hash length should be 32, got %d", len(hash))
			}

			// 验证哈希可以正确验证原密码
			if !checkPassword(password, hash) {
				t.Errorf("Generated hash cannot verify original password")
			}
		})
	}
}

// 辅助函数：检查密码
func checkPassword(password, hash string) bool {
	hashedPassword := generateHash(password)
	return hashedPassword == hash
}

// 辅助函数：生成哈希
func generateHash(password string) string {
	hash := md5.Sum([]byte(password))
	return hex.EncodeToString(hash[:])
}

// BenchmarkPasswordHash 性能测试
func BenchmarkPasswordHash(b *testing.B) {
	password := "admin123"
	for i := 0; i < b.N; i++ {
		generateHash(password)
	}
}

// BenchmarkPasswordCheck 性能测试
func BenchmarkPasswordCheck(b *testing.B) {
	password := "admin123"
	hash := "0192023a7bbd73250516f069df18b500"
	for i := 0; i < b.N; i++ {
		checkPassword(password, hash)
	}
}
