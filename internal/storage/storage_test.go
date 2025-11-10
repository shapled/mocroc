package storage

import (
	"fmt"
	"testing"
	"time"

	"fyne.io/fyne/v2/app"
)

// setupTestStorage 创建测试用的历史存储
func setupTestStorage(t *testing.T) *HistoryStorage {
	// 为每个测试使用唯一的ID，避免数据冲突
	testApp := app.NewWithID(fmt.Sprintf("com.test.mocroc.%d", time.Now().UnixNano()))

	storage := NewHistoryStorage(testApp)
	return storage
}

// TestNewHistoryStorage 测试创建历史存储
func TestNewHistoryStorage(t *testing.T) {
	storage := setupTestStorage(t)

	if storage == nil {
		t.Fatal("创建历史存储失败")
	}

	if storage.maxRecords != 500 {
		t.Errorf("期望最大记录数为 500，实际为 %d", storage.maxRecords)
	}

	if len(storage.recordKeys) != 0 {
		t.Errorf("期望初始记录key列表为空，实际长度为 %d", len(storage.recordKeys))
	}
}

// TestAddRecord 测试添加记录
func TestAddRecord(t *testing.T) {
	storage := setupTestStorage(t)

	// 创建测试记录
	item := HistoryItem{
		Type:       "send",
		FileName:   "test.txt",
		FileSize:   "1KB",
		Code:       "test-code-123",
		Status:     "completed",
		Timestamp:  time.Now(),
		Duration:   10,
		ClientInfo: "test-client",
		NumFiles:   1,
	}

	// 添加记录
	recordID, err := storage.Add(item)
	if err != nil {
		t.Fatalf("添加记录失败: %v", err)
	}

	// 验证ID被生成
	if recordID == "" {
		t.Error("期望记录ID被自动生成")
	}

	// 验证记录被添加
	if len(storage.recordKeys) != 1 {
		t.Errorf("期望记录key列表长度为 1，实际为 %d", len(storage.recordKeys))
	}

	if len(storage.cache) != 1 {
		t.Errorf("期望缓存中有 1 条记录，实际为 %d", len(storage.cache))
	}
}

// TestGetAll 测试获取所有记录
func TestGetAll(t *testing.T) {
	storage := setupTestStorage(t)

	// 添加3条记录
	for i := 0; i < 3; i++ {
		item := HistoryItem{
			Type:      "send",
			FileName:  fmt.Sprintf("test%d.txt", i),
			Status:    "completed",
			Timestamp: time.Now().Add(time.Duration(i) * time.Hour),
			Duration:  int64(i + 1),
		}
		_, err := storage.Add(item)
		if err != nil {
			t.Fatalf("添加记录失败: %v", err)
		}
	}

	// 获取所有记录
	allItems, err := storage.GetAll()
	if err != nil {
		t.Fatalf("获取所有记录失败: %v", err)
	}

	if len(allItems) != 3 {
		t.Errorf("期望获取 3 条记录，实际为 %d", len(allItems))
	}
}

// TestClear 测试清除所有记录
func TestClear(t *testing.T) {
	storage := setupTestStorage(t)

	// 添加一些记录
	for i := 0; i < 5; i++ {
		item := HistoryItem{
			Type:      "send",
			FileName:  fmt.Sprintf("test%d.txt", i),
			Status:    "completed",
			Timestamp: time.Now(),
		}
		_, err := storage.Add(item)
		if err != nil {
			t.Fatalf("添加记录失败: %v", err)
		}
	}

	// 清除所有记录
	err := storage.Clear()
	if err != nil {
		t.Fatalf("清除记录失败: %v", err)
	}

	// 验证所有记录被清除
	if len(storage.recordKeys) != 0 {
		t.Errorf("期望记录key列表为空，实际长度为 %d", len(storage.recordKeys))
	}

	if len(storage.cache) != 0 {
		t.Errorf("期望缓存为空，实际长度为 %d", len(storage.cache))
	}
}

// TestMaxRecordsLimit 测试最大记录数限制
func TestMaxRecordsLimit(t *testing.T) {
	// 创建一个限制较小的存储用于测试
	testApp := app.NewWithID(fmt.Sprintf("com.test.mocroc.limit.%d", time.Now().UnixNano()))

	storage := &HistoryStorage{
		app:        testApp,
		prefs:      testApp.Preferences(),
		cache:      make(map[string]HistoryItem),
		prefix:     "history_",
		maxRecords: 3, // 设置最大记录数为3
		recordKeys: []string{},
	}

	// 添加5条记录
	var ids []string
	for i := 0; i < 5; i++ {
		item := HistoryItem{
			Type:      "send",
			FileName:  fmt.Sprintf("test%d.txt", i),
			Status:    "completed",
			Timestamp: time.Now().Add(time.Duration(i) * time.Hour),
		}
		recordID, err := storage.Add(item)
		if err != nil {
			t.Fatalf("添加记录失败: %v", err)
		}
		ids = append(ids, recordID)
	}

	// 验证只保留最新的3条记录
	if len(storage.recordKeys) != 3 {
		t.Errorf("期望记录key列表长度为 3，实际为 %d", len(storage.recordKeys))
	}

	if len(storage.cache) != 3 {
		t.Errorf("期望缓存中有 3 条记录，实际为 %d", len(storage.cache))
	}
}

// TestGetStats 测试获取统计信息
func TestGetStats(t *testing.T) {
	storage := setupTestStorage(t)

	// 添加不同状态的记录
	records := []struct {
		status string
	}{
		{"completed"},
		{"completed"},
		{"failed"},
		{"in_progress"},
		{"completed"},
	}

	for i, record := range records {
		item := HistoryItem{
			Type:      "send",
			FileName:  fmt.Sprintf("test%d.txt", i),
			Status:    record.status,
			Timestamp: time.Now(),
		}
		_, err := storage.Add(item)
		if err != nil {
			t.Fatalf("添加记录失败: %v", err)
		}
	}

	// 获取统计信息
	total, completed, failed, inProgress, err := storage.GetStats()
	if err != nil {
		t.Fatalf("获取统计信息失败: %v", err)
	}

	if total != 5 {
		t.Errorf("期望总记录数为 5，实际为 %d", total)
	}

	if completed != 3 {
		t.Errorf("期望完成记录数为 3，实际为 %d", completed)
	}

	if failed != 1 {
		t.Errorf("期望失败记录数为 1，实际为 %d", failed)
	}

	if inProgress != 1 {
		t.Errorf("期望进行中记录数为 1，实际为 %d", inProgress)
	}
}

// TestExportImport 测试导出导入功能
func TestExportImport(t *testing.T) {
	storage1 := setupTestStorage(t)

	// 添加一些记录
	for i := 0; i < 3; i++ {
		item := HistoryItem{
			Type:       "send",
			FileName:   fmt.Sprintf("test%d.txt", i),
			FileSize:   fmt.Sprintf("%dKB", i+1),
			Code:       fmt.Sprintf("code-%d", i),
			Status:     "completed",
			Timestamp:  time.Now().Add(time.Duration(i) * time.Hour),
			Duration:   int64(i + 5),
			ClientInfo: "test-client",
			NumFiles:   i + 1,
		}
		_, err := storage1.Add(item)
		if err != nil {
			t.Fatalf("添加记录失败: %v", err)
		}
	}

	// 导出数据
	exportedData, err := storage1.Export()
	if err != nil {
		t.Fatalf("导出数据失败: %v", err)
	}

	if exportedData == "" {
		t.Fatal("导出的数据为空")
	}

	// 创建新的存储并导入数据
	storage2 := setupTestStorage(t)
	err = storage2.Import(exportedData)
	if err != nil {
		t.Fatalf("导入数据失败: %v", err)
	}

	// 验证导入的数据
	allItems, err := storage2.GetAll()
	if err != nil {
		t.Fatalf("获取导入后的记录失败: %v", err)
	}

	if len(allItems) != 3 {
		t.Errorf("期望导入后有 3 条记录，实际为 %d", len(allItems))
	}
}

// TestPersistenceWrite 测试数据确实被写入Preferences
func TestPersistenceWrite(t *testing.T) {
	// 使用固定的应用ID，确保数据可以持久化
	testApp := app.NewWithID("com.test.mocroc.persistence.test")
	storage := NewHistoryStorage(testApp)

	// 清除可能存在的旧数据
	storage.Clear()

	// 添加一些记录
	expectedRecords := 3
	for i := 0; i < expectedRecords; i++ {
		item := HistoryItem{
			Type:       "send",
			FileName:   fmt.Sprintf("persistent%d.txt", i),
			Status:     "completed",
			Timestamp:  time.Now(),
			Duration:   int64(i + 1),
			ClientInfo: "persistent-test",
		}
		recordID, err := storage.Add(item)
		if err != nil {
			t.Fatalf("添加记录失败: %v", err)
		}

		// 验证记录ID被生成
		if recordID == "" {
			t.Errorf("记录 %d 的ID为空", i)
		}
	}

	// 验证内存中的数据
	if len(storage.cache) != expectedRecords {
		t.Errorf("期望内存中有 %d 条记录，实际为 %d", expectedRecords, len(storage.cache))
	}

	// 创建新的存储实例，模拟应用重启
	storage2 := NewHistoryStorage(testApp)

	// 验证数据被重新加载
	allItems, err := storage2.GetAll()
	if err != nil {
		t.Fatalf("获取持久化记录失败: %v", err)
	}

	if len(allItems) != expectedRecords {
		t.Errorf("期望重新加载 %d 条记录，实际为 %d", expectedRecords, len(allItems))
	}

	// 验证数据内容
	for i, item := range allItems {
		if item.Status != "completed" {
			t.Errorf("记录 %d 的状态不正确", i)
		}
		if item.ClientInfo != "persistent-test" {
			t.Errorf("记录 %d 的客户端信息不正确", i)
		}
	}

	// 清理测试数据
	storage2.Clear()
}

// BenchmarkAddRecord 性能测试：添加记录
func BenchmarkAddRecord(b *testing.B) {
	storage := setupTestStorage(&testing.T{})

	item := HistoryItem{
		Type:       "send",
		FileName:   "benchmark.txt",
		Status:     "completed",
		Timestamp:  time.Now(),
		Duration:   10,
		ClientInfo: "benchmark-client",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := storage.Add(item)
		if err != nil {
			b.Fatalf("添加记录失败: %v", err)
		}
	}
}

// BenchmarkGetAll 性能测试：获取所有记录
func BenchmarkGetAll(b *testing.B) {
	storage := setupTestStorage(&testing.T{})

	// 预先添加一些记录
	for i := 0; i < 100; i++ {
		item := HistoryItem{
			Type:      "send",
			FileName:  fmt.Sprintf("benchmark%d.txt", i),
			Status:    "completed",
			Timestamp: time.Now().Add(time.Duration(i) * time.Hour),
		}
		storage.Add(item)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := storage.GetAll()
		if err != nil {
			b.Fatalf("获取所有记录失败: %v", err)
		}
	}
}
