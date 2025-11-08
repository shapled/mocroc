package crocmgr

import (
	"testing"
	"time"

	"github.com/schollz/croc/v10/src/croc"
)

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatal("NewManager returned nil")
	}
	if m.ctx == nil {
		t.Fatal("context is nil")
	}
	if m.cancel == nil {
		t.Fatal("cancel function is nil")
	}
}

func TestManager_Lifecycle(t *testing.T) {
	m := NewManager()
	defer m.Close()

	// 测试上下文是否有效
	select {
	case <-m.ctx.Done():
		t.Fatal("context done prematurely")
	default:
	}

	// 取消上下文
	m.Cancel()
	select {
	case <-m.ctx.Done():
		// 预期行为
	case <-time.After(100 * time.Millisecond):
		t.Fatal("context not done after cancel")
	}
}

func TestCreateCrocClient_DefaultOptions(t *testing.T) {
	m := NewManager()
	defer m.Close()

	// 使用默认配置
	options := croc.Options{
		IsSender:     true,
		SharedSecret: "test-code-1234",
		Debug:        true,
		NoPrompt:     true,
	}

	client, err := m.CreateCrocClient(options)
	if err != nil {
		t.Fatalf("CreateCrocClient failed: %v", err)
	}
	if client == nil {
		t.Fatal("client is nil")
	}

	if m.GetCrocClient() != client {
		t.Fatal("client not stored in manager")
	}
}

func TestCreateCrocClient_InvalidOptions(t *testing.T) {
	m := NewManager()
	defer m.Close()

	// 测试无效配置
	options := croc.Options{}

	_, err := m.CreateCrocClient(options)
	// 可能成功也可能失败，取决于 croc 的实现
	// 我们只是确保不会 panic
	_ = err
}

func TestCreateCrocClient_MultipleClients(t *testing.T) {
	m := NewManager()
	defer m.Close()

	options1 := croc.Options{
		IsSender:     true,
		SharedSecret: "test-code-1111",
		Debug:        true,
	}

	_, err := m.CreateCrocClient(options1)
	if err != nil {
		t.Fatalf("First CreateCrocClient failed: %v", err)
	}

	// 创建第二个客户端
	options2 := croc.Options{
		IsSender:     true,
		SharedSecret: "test-code-2222",
		Debug:        true,
	}

	client2, err := m.CreateCrocClient(options2)
	if err != nil {
		t.Fatalf("Second CreateCrocClient failed: %v", err)
	}

	// 第二个客户端应该覆盖第一个
	if m.GetCrocClient() != client2 {
		t.Fatal("client not updated in manager")
	}
}

func TestManager_Close(t *testing.T) {
	m := NewManager()

	// 创建客户端
	options := croc.Options{
		IsSender:     true,
		SharedSecret: "test-code-1234",
		Debug:        true,
	}
	_, err := m.CreateCrocClient(options)
	if err != nil {
		t.Fatalf("CreateCrocClient failed: %v", err)
	}

	// 关闭 manager
	m.Close()

	// 验证状态
	if m.crocClient != nil {
		t.Fatal("client not nil after Close")
	}

	// 验证上下文已取消
	select {
	case <-m.ctx.Done():
		// 预期行为
	case <-time.After(100 * time.Millisecond):
		t.Fatal("context not done after Close")
	}

	// 再次关闭应该安全
	m.Close()
}

func TestManager_GetContext(t *testing.T) {
	m := NewManager()
	defer m.Close()

	ctx := m.GetContext()
	if ctx == nil {
		t.Fatal("GetContext returned nil")
	}

	// 验证是同一个上下文
	if ctx != m.ctx {
		t.Fatal("GetContext returned different context")
	}
}

func TestManager_GetCrocClient_BeforeCreation(t *testing.T) {
	m := NewManager()
	defer m.Close()

	client := m.GetCrocClient()
	if client != nil {
		t.Fatal("GetCrocClient should return nil before creation")
	}
}

func TestManager_ConcurrentAccess(t *testing.T) {
	m := NewManager()
	defer m.Close()

	// 模拟并发创建客户端
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(n int) {
			options := croc.Options{
				IsSender:     true,
				SharedSecret: "test-code-" + string(rune('0'+n)),
				Debug:        true,
			}
			_, err := m.CreateCrocClient(options)
			if err != nil {
				t.Logf("goroutine %d failed: %v", n, err)
			}
			done <- true
		}(i)
	}

	// 等待所有 goroutine 完成
	timeout := time.After(5 * time.Second)
	for i := 0; i < 10; i++ {
		select {
		case <-done:
			// 成功
		case <-timeout:
			t.Fatal("timeout waiting for goroutines")
		}
	}
}

// 测试日志功能
func TestManager_Log(t *testing.T) {
	m := NewManager()
	defer m.Close()

	// 这应该不会 panic
	m.Log("test log message")

	// 多次调用
	m.Log("test log 1")
	m.Log("test log 2")
}

// 测试取消和上下文传播
func TestManager_Cancel_Propagates(t *testing.T) {
	m := NewManager()
	defer m.Close()

	// 验证初始状态
	ctx := m.GetContext()
	if ctx == nil {
		t.Fatal("context is nil")
	}

	// 取消
	m.Cancel()

	// 验证上下文已取消
	select {
	case <-ctx.Done():
		// 预期行为
	case <-time.After(100 * time.Millisecond):
		t.Fatal("context not cancelled")
	}
}
