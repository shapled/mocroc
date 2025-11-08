package crocmgr

import (
	"fmt"
	"net"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"github.com/schollz/croc/v10/src/croc"
)

// TestConnectionDiagnostics 用于诊断连接问题
func TestConnectionDiagnostics(t *testing.T) {
	t.Log("=== 连接诊断测试开始 ===")

	// 1. 检查网络连接
	t.Run("NetworkConnectivity", func(t *testing.T) {
		testNetworkConnectivity(t)
	})

	// 2. 测试本地 relay
	t.Run("LocalRelayTest", func(t *testing.T) {
		testLocalRelayConnection(t)
	})

	// 3. 测试远程 relay（可能失败，但我们想看到具体的错误）
	t.Run("RemoteRelayTest", func(t *testing.T) {
		testRemoteRelayConnection(t)
	})

	// 4. 测试中继服务器列表
	t.Run("RelayServerList", func(t *testing.T) {
		testRelayServerList(t)
	})

	t.Log("=== 连接诊断测试结束 ===")
}

func testNetworkConnectivity(t *testing.T) {
	// 测试 DNS 解析
	t.Log("测试 DNS 解析...")
	addresses, err := net.LookupHost("croc.schollz.com")
	if err != nil {
		t.Errorf("DNS 解析失败: %v", err)
	} else {
		t.Logf("DNS 解析成功: %v", addresses)
	}

	// 测试到默认端口的连接
	t.Log("测试到 croc.schollz.com:9009 的 TCP 连接...")
	conn, err := net.DialTimeout("tcp", "croc.schollz.com:9009", 5*time.Second)
	if err != nil {
		t.Errorf("TCP 连接失败: %v", err)
	} else {
		t.Logf("TCP 连接成功: %s -> %s",
			conn.LocalAddr().String(),
			conn.RemoteAddr().String())
		conn.Close()
	}
}

func testLocalRelayConnection(t *testing.T) {
	manager := NewManager()
	defer manager.Close()

	// 创建本地 relay 配置
	options := croc.Options{
		IsSender:       true,
		SharedSecret:   "test-code-diagnostics",
		Debug:          true,
		NoPrompt:       true,
		DisableLocal:   false,
		OnlyLocal:      false,
		RelayPorts:     []string{"9009", "9010", "9011", "9012", "9013"},
		RelayPassword:  "pass123",
		Stdout:         false,
		NoMultiplexing: false,
	}

	t.Log("创建 Croc 客户端...")
	_, err := manager.CreateCrocClient(options)
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}

	// 检查 goroutine 数量（用于检测 goroutine 泄露）
	initialGoroutines := runtime.NumGoroutine()
	t.Logf("当前 goroutine 数量: %d", initialGoroutines)

	// 注意：实际发送会尝试连接远程服务器
	// 我们这里只测试客户端创建，不进行实际传输
}

func testRemoteRelayConnection(t *testing.T) {
	manager := NewManager()
	defer manager.Close()

	// 使用官方服务器
	options := croc.Options{
		IsSender:       true,
		SharedSecret:   "test-code-remote",
		Debug:          true,
		NoPrompt:       true,
		DisableLocal:   false,
		OnlyLocal:      false,
		RelayAddress:   "croc.schollz.com",
		RelayPorts:     []string{"9009", "9010", "9011", "9012", "9013"},
		RelayPassword:  "pass123",
		Stdout:         false,
		NoMultiplexing: false,
	}

	t.Log("尝试创建连接到远程服务器...")
	client, err := manager.CreateCrocClient(options)
	if err != nil {
		t.Errorf("创建客户端失败: %v", err)
		return
	}

	t.Logf("客户端创建成功: %T", client)

	// 检查是否有新的 goroutine
	runtime.GC()
	time.Sleep(100 * time.Millisecond)
	currentGoroutines := runtime.NumGoroutine()
	t.Logf("创建客户端后的 goroutine 数量: %d", currentGoroutines)
}

func testRelayServerList(t *testing.T) {
	// 测试多个备用的中继服务器
	relayServers := []string{
		"croc.schollz.com:9009",
		"croc6.schollz.com:9009",
	}

	for _, server := range relayServers {
		t.Logf("测试连接: %s", server)
		conn, err := net.DialTimeout("tcp", server, 3*time.Second)
		if err != nil {
			t.Logf("  ❌ 连接失败: %v", err)
		} else {
			t.Logf("  ✅ 连接成功: %s", conn.RemoteAddr().String())
			conn.Close()
		}
	}
}

// TestConcurrentClientCreation 测试并发创建客户端的行为
func TestConcurrentClientCreation(t *testing.T) {
	manager := NewManager()
	defer manager.Close()

	const numClients = 5
	errors := make([]error, numClients)

	// 创建计数器跟踪成功/失败
	var successCount int64
	var errorCount int64

	// 并发创建客户端
	for i := 0; i < numClients; i++ {
		go func(n int) {
			options := croc.Options{
				IsSender:       true,
				SharedSecret:   fmt.Sprintf("test-code-%d", n),
				Debug:          false, // 关闭 debug 减少日志
				NoPrompt:       true,
				RelayPorts:     []string{"9009"},
				RelayPassword:  "pass123",
			}

			_, err := manager.CreateCrocClient(options)
			errors[n] = err

			if err == nil {
				atomic.AddInt64(&successCount, 1)
			} else {
				atomic.AddInt64(&errorCount, 1)
				t.Logf("客户端 %d 创建失败: %v", n, err)
			}
		}(i)
	}

	// 等待所有客户端创建完成
	time.Sleep(2 * time.Second)

	t.Logf("并发创建结果: 成功=%d, 失败=%d", successCount, errorCount)

	if errorCount > 0 {
		t.Log("部分客户端创建失败，这可能是网络问题")
	}
}

// TestGracefulDisconnect 测试优雅关闭
func TestGracefulDisconnect(t *testing.T) {
	manager := NewManager()
	defer manager.Close()

	// 创建客户端
	options := croc.Options{
		IsSender:       true,
		SharedSecret:   "test-graceful",
		Debug:          true,
		NoPrompt:       true,
		RelayPorts:     []string{"9009"},
		RelayPassword:  "pass123",
	}

	_, err := manager.CreateCrocClient(options)
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}

	// 获取初始 goroutine 数量
	initialGoroutines := runtime.NumGoroutine()
	t.Logf("初始 goroutine 数量: %d", initialGoroutines)

	// 取消上下文
	manager.Cancel()

	// 等待 goroutine 清理
	time.Sleep(500 * time.Millisecond)
	runtime.GC()
	time.Sleep(500 * time.Millisecond)

	finalGoroutines := runtime.NumGoroutine()
	t.Logf("关闭后 goroutine 数量: %d", finalGoroutines)

	// 差异
	diff := finalGoroutines - initialGoroutines
	if diff > 5 {
		t.Logf("WARNING: 可能存在 goroutine 泄露: 差异 %d", diff)
	}
}

// TestConnectionTimeout 测试连接超时
func TestConnectionTimeout(t *testing.T) {
	// 测试连接到无效地址的超时行为
	invalidAddresses := []string{
		"1.2.3.4:9999",  // 无效 IP
		"nonexistent.server:9009", // 不存在的域名
	}

	for _, addr := range invalidAddresses {
		t.Logf("测试超时连接: %s", addr)
		start := time.Now()

		conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
		elapsed := time.Since(start)

		if err != nil {
			t.Logf("  ✅ 正确处理超时: %v (耗时: %v)", err, elapsed)
		} else {
			t.Logf("  ⚠️ 意外连接成功: %s", conn.RemoteAddr().String())
			conn.Close()
		}

		// 验证超时时间大致正确（允许一些误差）
		if elapsed < 1800*time.Millisecond || elapsed > 2500*time.Millisecond {
			t.Logf("WARNING: 超时时间异常: %v", elapsed)
		}
	}
}

// BenchmarkCreateCrocClient 性能测试
func BenchmarkCreateCrocClient(b *testing.B) {
	manager := NewManager()
	defer manager.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		options := croc.Options{
			IsSender:       true,
			SharedSecret:   "bench-test",
			Debug:          false,
			NoPrompt:       true,
			RelayPorts:     []string{"9009"},
			RelayPassword:  "pass123",
		}

		_, err := manager.CreateCrocClient(options)
		if err != nil {
			b.Fatalf("创建客户端失败: %v", err)
		}
	}
}
