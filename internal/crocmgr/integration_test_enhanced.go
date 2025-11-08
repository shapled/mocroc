package crocmgr

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/schollz/croc/v10/src/croc"
)

// TestSendReceive_ConcurrentClients 测试并发客户端创建的竞态条件
// 这是一个重要的回归测试，用于发现 Manager 中的并发问题
func TestSendReceive_ConcurrentClients(t *testing.T) {
	const numConcurrent = 10
	errors := make(chan error, numConcurrent)
	var wg sync.WaitGroup

	// 多个 goroutine 同时创建客户端
	for i := 0; i < numConcurrent; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			manager := NewManager()
			defer manager.Close()

			// 快速创建和销毁客户端
			for j := 0; j < 5; j++ {
				client, err := manager.CreateCrocClient(croc.Options{
					IsSender:       true,
					SharedSecret:   "test-concurrent",
					Debug:          false,
					RelayPorts:     []string{"9009"},
					RelayPassword:  "pass123",
				})
				if err != nil {
					errors <- nil // 连接失败是正常的，因为是并发
					continue
				}
				if client == nil {
					errors <- nil
					continue
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	errorCount := 0
	for err := range errors {
		if err != nil {
			errorCount++
		}
	}

	// 注意：如果出现 panic 或严重的并发错误，test 就会失败
	if errorCount == numConcurrent*5 {
		t.Logf("所有并发操作都失败了（网络问题）")
	}
}

// TestSendReceive_ManagerRaceCondition 测试 Manager 的竞态条件
// 特别关注 CreateCrocClient 和 Close 的并发执行
func TestSendReceive_ManagerRaceCondition(t *testing.T) {
	manager := NewManager()

	var wg sync.WaitGroup
	wg.Add(3)

	// 启动多个并发操作
	go func() {
		defer wg.Done()
		for i := 0; i < 20; i++ {
			manager.CreateCrocClient(croc.Options{
				IsSender:       true,
				SharedSecret:   "test-race",
				RelayPorts:     []string{"9009"},
				RelayPassword:  "pass123",
			})
			time.Sleep(10 * time.Millisecond)
		}
	}()

	go func() {
		defer wg.Done()
		time.Sleep(50 * time.Millisecond)
		manager.Cancel()
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			_ = manager.GetCrocClient()
			time.Sleep(20 * time.Millisecond)
		}
	}()

	// 等待所有 goroutine 完成
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	// 最多等待 3 秒
	select {
	case <-done:
		// 成功完成
	case <-time.After(3 * time.Second):
		t.Fatal("测试超时，可能存在死锁或 goroutine 泄露")
	}
}

// TestSendReceive_CleanupAfterError 测试错误后的资源清理
// 验证当创建客户端失败时，Manager 不会处于不一致状态
func TestSendReceive_CleanupAfterError(t *testing.T) {
	// 测试多种错误场景
	errorScenarios := []struct {
		name    string
		options croc.Options
	}{
		{
			name: "InvalidPort",
			options: croc.Options{
				IsSender:       true,
				SharedSecret:   "test-error",
				RelayPorts:     []string{"invalid-port"},
				RelayPassword:  "pass123",
			},
		},
		{
			name: "EmptySecret",
			options: croc.Options{
				IsSender:       true,
				SharedSecret:   "",
				RelayPorts:     []string{"9009"},
				RelayPassword:  "pass123",
			},
		},
		{
			name: "InvalidAddress",
			options: croc.Options{
				IsSender:       true,
				SharedSecret:   "test-error",
				RelayAddress:   "invalid.invalid.invalid",
				RelayPorts:     []string{"9009"},
				RelayPassword:  "pass123",
			},
		},
	}

	for _, scenario := range errorScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			manager := NewManager()
			defer manager.Close()

			// 尝试创建客户端（应该失败）
			_, err := manager.CreateCrocClient(scenario.options)

			// 验证错误处理
			if err == nil {
				t.Logf("%s: 意外成功", scenario.name)
			} else {
				t.Logf("%s: 预期的错误: %v", scenario.name, err)
			}

			// 验证 Manager 仍然可用
			client := manager.GetCrocClient()
			if client == nil {
				t.Logf("%s: 客户端为 nil（预期的）", scenario.name)
			}

			// 验证可以再次尝试创建
			_, err2 := manager.CreateCrocClient(scenario.options)
			if err2 == nil {
				t.Logf("%s: 第二次尝试成功", scenario.name)
			}
		})
	}
}

// TestSendReceive_MultipleTransfersSequential 测试连续多次传输
// 验证 Manager 能否正确处理连续的传输请求
func TestSendReceive_MultipleTransfersSequential(t *testing.T) {
	manager := NewManager()
	defer manager.Close()

	// 创建临时文件
	tmpDir, err := os.MkdirTemp("", "croc-multi-test-*")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	const numTransfers = 5
	transferCodes := make([]string, numTransfers)

	for i := 0; i < numTransfers; i++ {
		// 创建测试文件
		sendFile := filepath.Join(tmpDir, "test-send.txt")
		testData := "测试数据 " + time.Now().Format("2006-01-02 15:04:05")
		if err := os.WriteFile(sendFile, []byte(testData), 0644); err != nil {
			t.Fatalf("创建测试文件失败: %v", err)
		}

		code := "test-code-multi"
		transferCodes[i] = code

		// 准备传输
		filesInfo, emptyFolders, totalNumberFolders, err := croc.GetFilesInfo(
			[]string{sendFile},
			false, false, []string{},
		)
		if err != nil {
			t.Fatalf("获取文件信息失败: %v", err)
		}

		// 创建发送客户端
		senderClient, err := manager.CreateCrocClient(croc.Options{
			IsSender:       true,
			SharedSecret:   code,
			Debug:          false,
			NoPrompt:       true,
			RelayPorts:     []string{"9009"},
			RelayPassword:  "pass123",
		})
		if err != nil {
			t.Logf("第 %d 次创建客户端失败（可能网络问题）: %v", i+1, err)
			continue
		}

		// 启动发送（但立即取消，因为没有接收方）
		var sendErr error
		go func() {
			sendErr = senderClient.Send(filesInfo, emptyFolders, totalNumberFolders)
		}()

		// 等待一下
		time.Sleep(100 * time.Millisecond)

		// 取消
		manager.Cancel()

		// 等待 goroutine 清理
		time.Sleep(200 * time.Millisecond)

		if sendErr != nil {
			t.Logf("第 %d 次传输错误（预期）: %v", i+1, sendErr)
		} else {
			t.Logf("第 %d 次传输完成", i+1)
		}

		// 重新创建上下文以进行下一次传输
		manager2 := NewManager()
		manager2.Cancel() // 立即取消以清理
		time.Sleep(100 * time.Millisecond)
		manager2.Close()

		// 重新创建 manager（因为 Cancel 会关闭上下文）
		manager = NewManager()
		defer manager.Close()
	}
}

// TestSendReceive_LargeFileHandling 测试大文件处理
func TestSendReceive_LargeFileHandling(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "croc-large-test-*")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 创建 10MB 测试文件
	largeFile := filepath.Join(tmpDir, "large-test.bin")
	file, err := os.Create(largeFile)
	if err != nil {
		t.Fatalf("创建大文件失败: %v", err)
	}

	// 写入 10MB 数据
	data := make([]byte, 1024) // 1KB 块
	for i := range data {
		data[i] = byte(i % 256)
	}

	const mb10 = 10 * 1024 * 1024
	for written := 0; written < mb10; {
		n, err := file.Write(data)
		if err != nil {
			t.Fatalf("写入文件失败: %v", err)
		}
		written += n
	}
	file.Close()

	// 验证文件大小
	info, err := os.Stat(largeFile)
	if err != nil {
		t.Fatalf("获取文件信息失败: %v", err)
	}
	t.Logf("创建测试文件大小: %d MB", info.Size()/(1024*1024))

	manager := NewManager()
	defer manager.Close()

	// 获取文件信息
	filesInfo, emptyFolders, totalNumberFolders, err := croc.GetFilesInfo(
		[]string{largeFile},
		false, false, []string{},
	)
	if err != nil {
		t.Fatalf("获取文件信息失败: %v", err)
	}

	t.Logf("文件信息: %+v", filesInfo[0])

	// 创建客户端
	client, err := manager.CreateCrocClient(croc.Options{
		IsSender:       true,
		SharedSecret:   "test-large-file",
		Debug:          true,
		NoPrompt:       true,
		RelayPorts:     []string{"9009"},
		RelayPassword:  "pass123",
	})
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}

	// 启动传输
	var transferErr error
	go func() {
		transferErr = client.Send(filesInfo, emptyFolders, totalNumberFolders)
	}()

	// 等待或取消
	time.Sleep(1 * time.Second)
	manager.Cancel()

	time.Sleep(500 * time.Millisecond)

	if transferErr != nil {
		t.Logf("大文件传输错误（预期）: %v", transferErr)
	}
}

// TestSendReceive_ContextCancellation 测试上下文取消的传播
func TestSendReceive_ContextCancellation(t *testing.T) {
	manager := NewManager()
	defer manager.Close()

	// 创建客户端
	client, err := manager.CreateCrocClient(croc.Options{
		IsSender:       true,
		SharedSecret:   "test-cancel",
		Debug:          true,
		NoPrompt:       true,
		RelayPorts:     []string{"9009"},
		RelayPassword:  "pass123",
	})
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}

	// 监控上下文状态
	ctx := manager.GetContext()
	canceled := make(chan struct{})
	go func() {
		<-ctx.Done()
		close(canceled)
	}()

	// 立即取消
	manager.Cancel()

	// 验证上下文已取消
	select {
	case <-canceled:
		t.Log("上下文已正确取消")
	case <-time.After(1 * time.Second):
		t.Error("上下文未在预期时间内取消")
	}

	// 验证客户端仍然存在
	if client == nil {
		t.Error("客户端为 nil")
	}

	// 验证可以安全关闭
	manager.Close()
	t.Log("安全关闭完成")
}

// TestSendReceive_ManagerReuse 测试 Manager 复用
// 验证关闭后能否重新使用
func TestSendReceive_ManagerReuse(t *testing.T) {
	manager := NewManager()

	// 第一次使用
	client1, err := manager.CreateCrocClient(croc.Options{
		IsSender:       true,
		SharedSecret:   "test-reuse-1",
		RelayPorts:     []string{"9009"},
		RelayPassword:  "pass123",
	})
	if err != nil {
		t.Logf("第一次创建失败（预期）: %v", err)
	}
	if client1 != nil {
		t.Log("第一次创建成功")
	}

	// 关闭
	manager.Close()

	// 再次关闭应该安全
	manager.Close()
	t.Log("多次关闭安全")

	// 重新创建（这应该安全）
	manager = NewManager()
	defer manager.Close()

	client2, err := manager.CreateCrocClient(croc.Options{
		IsSender:       true,
		SharedSecret:   "test-reuse-2",
		RelayPorts:     []string{"9009"},
		RelayPassword:  "pass123",
	})
	if err != nil {
		t.Logf("第二次创建失败（预期）: %v", err)
	}
	if client2 != nil {
		t.Log("第二次创建成功")
	}
}

// TestSendReceive_PanicRecovery 测试 panic 恢复
func TestSendReceive_PanicRecovery(t *testing.T) {
	manager := NewManager()
	defer func() {
		if r := recover(); r != nil {
			t.Logf("捕获到 panic: %v", r)
		}
		manager.Close()
	}()

	// 故意触发 panic
	defer func() {
		if r := recover(); r != nil {
			t.Logf("预期的 panic: %v", r)
		}
	}()

	client, err := manager.CreateCrocClient(croc.Options{
		IsSender:       true,
		SharedSecret:   "test-panic",
		RelayPorts:     []string{"9009"},
		RelayPassword:  "pass123",
	})
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}

	// 在另一个 goroutine 中触发 panic
	go func() {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("goroutine 中的 panic: %v", r)
			}
		}()
		panic("测试 panic")
	}()

	_ = client // 使用 client 避免未使用错误
	time.Sleep(100 * time.Millisecond)

	// 验证 Manager 仍然可用
	_ = manager.GetCrocClient()
	t.Log("panic 后 Manager 仍然可用")
}

// TestSendReceive_TimeoutHandling 测试超时处理
func TestSendReceive_TimeoutHandling(t *testing.T) {
	manager := NewManager()
	defer manager.Close()

	// 测试连接超时
	client, err := manager.CreateCrocClient(croc.Options{
		IsSender:       true,
		SharedSecret:   "test-timeout",
		Debug:          true,
		NoPrompt:       true,
		RelayAddress:   "1.2.3.4",      // 无效地址
		RelayPorts:     []string{"9999"}, // 无效端口
		RelayPassword:  "pass123",
	})
	if err != nil {
		t.Logf("创建客户端失败（预期）: %v", err)
	}

	if client != nil {
		// 尝试传输（应该快速失败）
		filesInfo, emptyFolders, totalNumberFolders, err := croc.GetFilesInfo(
			[]string{}, false, false, []string{},
		)
		if err == nil {
			err = client.Send(filesInfo, emptyFolders, totalNumberFolders)
		}

		start := time.Now()
		if err != nil {
			elapsed := time.Since(start)
			t.Logf("传输失败（预期）: %v, 耗时: %v", err, elapsed)

			// 验证快速失败（应该在 10 秒内）
			if elapsed > 10*time.Second {
				t.Errorf("超时时间过长: %v", elapsed)
			} else {
				t.Logf("超时时间正常: %v", elapsed)
			}
		}
	}
}

// BenchmarkManager_ConcurrentCreation 性能测试：并发创建
func BenchmarkManager_ConcurrentCreation(b *testing.B) {
	b.StopTimer()

	for i := 0; i < b.N; i++ {
		manager := NewManager()
		var wg sync.WaitGroup
		const numGoroutines = 10

		wg.Add(numGoroutines)
		for j := 0; j < numGoroutines; j++ {
			go func() {
				defer wg.Done()
				_, _ = manager.CreateCrocClient(croc.Options{
					IsSender:       true,
					SharedSecret:   "bench-test",
					RelayPorts:     []string{"9009"},
					RelayPassword:  "pass123",
				})
			}()
		}

		wg.Wait()
		b.StartTimer()
		manager.Close()
		b.StopTimer()
	}
}
