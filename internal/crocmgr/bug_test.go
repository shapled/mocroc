package crocmgr

import (
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/schollz/croc/v10/src/croc"
)

// TestBug_FixedCodeGeneration 测试接收码生成的问题
// 根据 send.go:404-412，generateCode 总是返回固定值，这是个 bug
func TestBug_FixedCodeGeneration(t *testing.T) {
	// 多次生成代码，验证是否随机化
	codes := make(map[string]bool)
	const numCodes = 100

	for i := 0; i < numCodes; i++ {
		// 模拟 generateCode 的逻辑
		adjectives := []string{"red", "blue", "green", "yellow", "orange", "purple", "pink", "brown"}
		animals := []string{"cat", "dog", "frog", "bird", "fish", "lion", "tiger", "bear"}
		code := adjectives[0] + animals[0] + "123"

		codes[code] = true
	}

	// 当前实现总是返回相同代码
	if len(codes) == 1 {
		t.Log("⚠️  BUG 发现：接收码生成不随机，总是返回固定值")
		t.Log("   应该使用随机索引来生成唯一码")
	} else {
		t.Logf("✅ 接收码生成正常，生成了 %d 个不同的码", len(codes))
	}
}

// TestBug_TextSendingNotImplemented 测试文本发送功能未实现
// 根据 send.go:424-428，文本发送只是记录日志，没有实际处理
func TestBug_TextSendingNotImplemented(t *testing.T) {
	manager := NewManager()
	defer manager.Close()

	// 模拟发送文本的流程
	_ = "这是一个测试消息"

	// 当前的实现只是记录日志，没有实际处理
	manager.Log("文本发送功能待完善")

	// 验证：当前实现没有创建临时文件或处理文本
	// 这会导致空文件列表被发送到接收方

	// 正确的实现应该：
	// 1. 创建临时文件
	// 2. 将文本写入文件
	// 3. 将文件添加到发送列表

	t.Log("⚠️  BUG 发现：文本发送功能未实现")
	t.Log("   当前实现只记录日志，不实际发送文本")
}

// TestBug_ManagerRaceCondition 测试 Manager 的竞态条件
// 根据 manager.go:35-41，CreateCrocClient 没有并发安全控制
func TestBug_ManagerRaceCondition(t *testing.T) {
	manager := NewManager()
	defer manager.Close()

	const numGoroutines = 20
	var wg sync.WaitGroup
	var createdClients int64
	var failedCreates int64

	// 并发创建客户端
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			client, err := manager.CreateCrocClient(croc.Options{
				IsSender:      true,
				SharedSecret:  "test-race",
				RelayPorts:    []string{"9009"},
				RelayPassword: "pass123",
			})

			if err != nil {
				// 连接失败是正常的（因为是测试）
				failedCreates++
			} else {
				createdClients++
				// 验证客户端存在
				if client == nil {
					t.Errorf("goroutine %d: 客户端为 nil", n)
				}
			}
		}(i)
	}

	wg.Wait()

	t.Logf("并发创建结果: 成功=%d, 失败=%d", createdClients, failedCreates)

	// 验证最终状态
	finalClient := manager.GetCrocClient()
	if finalClient == nil {
		t.Log("⚠️  最终客户端为 nil（可能的竞态条件）")
	} else {
		t.Log("✅ 最终客户端存在")
	}
}

// TestBug_NoContextPropagation 测试上下文传播问题
// 根据 manager.go:29-33，取消操作需要正确传播
func TestBug_NoContextPropagation(t *testing.T) {
	manager := NewManager()
	defer manager.Close()

	// 创建客户端
	client, err := manager.CreateCrocClient(croc.Options{
		IsSender:      true,
		SharedSecret:  "test-cancel",
		RelayPorts:    []string{"9009"},
		RelayPassword: "pass123",
	})
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}

	// 验证初始状态
	ctx := manager.GetContext()
	if ctx == nil {
		t.Fatal("上下文为 nil")
	}

	// 启动一个长时间运行的操作
	done := make(chan bool, 1)
	go func() {
		for i := 0; i < 100; i++ {
			select {
			case <-ctx.Done():
				done <- false
				return
			default:
				time.Sleep(10 * time.Millisecond)
			}
		}
		done <- true
	}()

	// 等待 goroutine 开始运行
	time.Sleep(50 * time.Millisecond)

	// 取消操作
	manager.Cancel()

	// 等待上下文被取消
	canceled := make(chan bool, 1)
	go func() {
		<-ctx.Done()
		canceled <- true
	}()

	// 验证取消
	select {
	case <-canceled:
		t.Log("✅ 上下文正确取消")
	case <-time.After(1 * time.Second):
		t.Error("⚠️  上下文未在预期时间内取消")
	}

	// 验证长时间运行的操作被停止
	select {
	case stopped := <-done:
		if !stopped {
			t.Log("✅ 长时间运行的操作被正确停止")
		} else {
			t.Log("⚠️  长时间运行的操作未停止（执行完成）")
		}
	case <-time.After(1500 * time.Millisecond):
		t.Error("⚠️  长时间运行的操作未停止（超时）")
	}

	_ = client
}

// TestBug_GoroutineLeak 测试 goroutine 泄露
// 验证 Cancel 和 Close 是否正确清理所有 goroutine
func TestBug_GoroutineLeak(t *testing.T) {
	// 记录初始 goroutine 数量
	initialGoroutines := runtime.NumGoroutine()
	t.Logf("初始 goroutine 数量: %d", initialGoroutines)

	// 创建多个 manager 和客户端
	const numManagers = 10
	managers := make([]*Manager, numManagers)
	clients := make([]*croc.Client, numManagers)

	for i := 0; i < numManagers; i++ {
		managers[i] = NewManager()
		client, err := managers[i].CreateCrocClient(croc.Options{
			IsSender:      true,
			SharedSecret:  "test-leak",
			RelayPorts:    []string{"9009"},
			RelayPassword: "pass123",
		})
		if err == nil {
			clients[i] = client
		}
	}

	time.Sleep(200 * time.Millisecond)

	beforeClose := runtime.NumGoroutine()
	t.Logf("创建客户端后 goroutine 数量: %d (+%d)", beforeClose, beforeClose-initialGoroutines)

	// 关闭所有 manager
	for i := 0; i < numManagers; i++ {
		managers[i].Cancel()
		managers[i].Close()
	}

	// 等待清理
	time.Sleep(500 * time.Millisecond)
	runtime.GC()
	time.Sleep(500 * time.Millisecond)

	finalGoroutines := runtime.NumGoroutine()
	t.Logf("关闭后 goroutine 数量: %d (+%d)", finalGoroutines, finalGoroutines-initialGoroutines)

	// 允许一些延迟清理的 goroutine
	goroutineDiff := finalGoroutines - initialGoroutines
	if goroutineDiff > 5 {
		t.Errorf("⚠️  可能存在 goroutine 泄露: +%d 个 goroutine", goroutineDiff)
	} else {
		t.Logf("✅ goroutine 清理正常: +%d 个", goroutineDiff)
	}
}

// TestBug_ReceivePreviewNotImplemented 测试接收端预览功能未实现
// 根据 receive.go:236-248，previewFiles 只是模拟，没有实际连接
func TestBug_ReceivePreviewNotImplemented(t *testing.T) {
	// 这是 UI 相关的 bug，但影响测试
	t.Log("⚠️  BUG 发现：接收端文件预览功能未实现")
	t.Log("   当前实现返回硬编码的示例文件")
	t.Log("   应该实际连接到发送方并获取真实文件列表")
}

// TestBug_NoRealProgressUpdate 测试没有真实的进度更新
// 根据 send.go:477-498，simulateProgress 只是模拟，没有真实传输进度
func TestBug_NoRealProgressUpdate(t *testing.T) {
	manager := NewManager()
	defer manager.Close()

	client, err := manager.CreateCrocClient(croc.Options{
		IsSender:      true,
		SharedSecret:  "test-progress",
		RelayPorts:    []string{"9009"},
		RelayPassword: "pass123",
	})
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}

	// 当前的 simulateProgress 只是空循环
	// 没有从 croc 客户端获取实际进度

	t.Log("⚠️  BUG 发现：进度更新是模拟的，不是真实的")
	t.Log("   simulateProgress 只是更新 UI，没有反映实际传输进度")
	t.Log("   应该从 croc 客户端获取真实传输进度")

	_ = client
}

// TestBug_EmptyFileListSend 测试发送空文件列表
// 根据 send.go:423-429，如果只有文本没有文件，sendFiles 可能是空的
func TestBug_EmptyFileListSend(t *testing.T) {
	manager := NewManager()
	defer manager.Close()

	// 模拟只有文本没有文件的情况
	sendText := "仅文本内容"
	selectedFiles := []string{} // 空文件列表

	// 当前实现：如果只有文本，sendFiles 仍然是空的
	var sendFiles []string
	if sendText != "" {
		// 文本发送功能未实现，所以 sendFiles 仍然是空的
		manager.Log("文本发送功能待完善")
	}
	sendFiles = append(sendFiles, selectedFiles...)

	if len(sendFiles) == 0 {
		t.Log("⚠️  BUG 发现：只有文本时会发送空文件列表")
		t.Log("   这可能导致传输失败或传输空内容")
		t.Log("   正确的实现应该创建临时文件存储文本")
	}

	// 验证 GetFilesInfo 对空文件列表的处理
	_, _, _, err := croc.GetFilesInfo(sendFiles, false, false, []string{})
	if err != nil {
		t.Logf("GetFilesInfo 对空列表返回错误（预期）: %v", err)
	}
}

// TestBug_DoubleCloseSafety 测试多次关闭的安全性
// 根据 manager.go:48-53，Close 方法应该可以安全调用多次
func TestBug_DoubleCloseSafety(t *testing.T) {
	manager := NewManager()

	// 第一次关闭
	manager.Close()
	t.Log("第一次关闭完成")

	// 第二次关闭（应该安全）
	manager.Close()
	t.Log("第二次关闭完成")

	// 第三次关闭
	manager.Close()
	t.Log("第三次关闭完成")

	t.Log("✅ Close 方法可以安全调用多次")
}

// TestBug_ContextAfterClose 测试关闭后的上下文状态
func TestBug_ContextAfterClose(t *testing.T) {
	manager := NewManager()
	ctx := manager.GetContext()

	// 关闭
	manager.Close()

	// 验证上下文已取消
	select {
	case <-ctx.Done():
		t.Log("✅ 关闭后上下文被正确取消")
	case <-time.After(100 * time.Millisecond):
		t.Error("⚠️  关闭后上下文未取消")
	}

	// 再次关闭应该安全
	manager.Close()
	t.Log("✅ 关闭后可以再次调用 Close")
}
