package storage

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"fyne.io/fyne/v2"
)

// HistoryItem 传输历史记录项
type HistoryItem struct {
	ID         string    `json:"id"`
	Type       string    `json:"type"`       // "send" or "receive"
	FileName   string    `json:"fileName"`   // 主要文件名
	FileSize   string    `json:"fileSize"`   // 文件大小
	Code       string    `json:"code"`       // 接收码
	Status     string    `json:"status"`     // "completed", "failed", "in_progress"
	Timestamp  time.Time `json:"timestamp"`  // 创建时间
	Duration   int64     `json:"duration"`   // 传输耗时（秒）
	ClientInfo string    `json:"clientInfo"` // 客户端信息
	NumFiles   int       `json:"numFiles"`   // 文件数量
}

// HistoryStorage 历史记录存储管理器
type HistoryStorage struct {
	mu         sync.RWMutex
	app        fyne.App
	prefs      fyne.Preferences
	cache      map[string]HistoryItem // ID到记录的映射
	prefix     string                 // preferences 键前缀
	idCounter  int64                  // ID计数器
	maxRecords int                    // 最大记录数限制
	recordKeys []string               // 所有记录的key列表（按时间顺序）
}

// NewHistoryStorage 创建历史记录存储管理器
func NewHistoryStorage(a fyne.App) *HistoryStorage {
	hs := &HistoryStorage{
		app:        a,
		prefs:      a.Preferences(),
		cache:      make(map[string]HistoryItem),
		prefix:     "history_",
		maxRecords: 500, // 最多保存500条记录
		recordKeys: []string{},
	}

	// 加载现有数据
	hs.loadAll()

	return hs
}

// getKey 根据记录ID获取 preferences 键
func (hs *HistoryStorage) getKey(id string) string {
	return hs.prefix + id
}

// generateID 生成新的记录ID
func (hs *HistoryStorage) generateID() string {
	hs.idCounter++
	return fmt.Sprintf("record_%d_%d", hs.idCounter, time.Now().UnixNano())
}

// loadRecord 加载单个记录
func (hs *HistoryStorage) loadRecord(id string) (HistoryItem, error) {
	key := hs.getKey(id)
	dataStr := hs.prefs.String(key)

	if dataStr == "" {
		return HistoryItem{}, fmt.Errorf("记录不存在")
	}

	var item HistoryItem
	if err := json.Unmarshal([]byte(dataStr), &item); err != nil {
		return HistoryItem{}, fmt.Errorf("解析JSON失败: %v", err)
	}

	return item, nil
}

// saveRecord 保存单个记录
func (hs *HistoryStorage) saveRecord(item HistoryItem) error {
	key := hs.getKey(item.ID)

	data, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("编码JSON失败: %v", err)
	}

	hs.prefs.SetString(key, string(data))
	return nil
}

// saveRecordKeys 保存记录key列表
func (hs *HistoryStorage) saveRecordKeys() error {
	data, err := json.Marshal(hs.recordKeys)
	if err != nil {
		return fmt.Errorf("编码key列表失败: %v", err)
	}

	hs.prefs.SetString("history_keys", string(data))
	return nil
}

// loadRecordKeys 加载记录key列表
func (hs *HistoryStorage) loadRecordKeys() error {
	dataStr := hs.prefs.String("history_keys")
	if dataStr == "" {
		hs.recordKeys = []string{}
		return nil
	}

	var keys []string
	if err := json.Unmarshal([]byte(dataStr), &keys); err != nil {
		hs.recordKeys = []string{}
		return fmt.Errorf("解析key列表失败: %v", err)
	}

	hs.recordKeys = keys
	return nil
}

// loadAll 加载所有历史记录
func (hs *HistoryStorage) loadAll() {
	hs.mu.Lock()
	defer hs.mu.Unlock()

	hs.cache = make(map[string]HistoryItem)
	hs.idCounter = 0

	// 加载记录key列表
	if err := hs.loadRecordKeys(); err != nil {
		fmt.Printf("加载记录key列表失败: %v\n", err)
		hs.recordKeys = []string{}
	}

	// 加载ID计数器
	hs.idCounter = int64(hs.prefs.Int("history_id_counter"))

	// 加载所有记录
	loadedCount := 0
	for _, id := range hs.recordKeys {
		if item, err := hs.loadRecord(id); err == nil {
			hs.cache[id] = item
			loadedCount++
		}
	}

	fmt.Printf("加载了 %d 条历史记录\n", loadedCount)
}

// Add 添加历史记录
func (hs *HistoryStorage) Add(item HistoryItem) (string, error) {
	hs.mu.Lock()
	defer hs.mu.Unlock()

	// 生成新ID
	if item.ID == "" {
		item.ID = hs.generateID()
	}

	recordID := item.ID

	// 检查是否超出最大记录数限制
	if len(hs.recordKeys) >= hs.maxRecords {
		// 删除最旧的记录
		oldestID := hs.recordKeys[0]
		hs.recordKeys = hs.recordKeys[1:]
		delete(hs.cache, oldestID)

		// 从 preferences 中删除最旧的记录
		oldestKey := hs.getKey(oldestID)
		hs.prefs.RemoveValue(oldestKey)
	}

	// 添加到记录key列表
	hs.recordKeys = append(hs.recordKeys, recordID)

	// 保存到缓存
	hs.cache[recordID] = item

	// 保存ID计数器
	hs.prefs.SetInt("history_id_counter", int(hs.idCounter))

	// 保存记录key列表
	if err := hs.saveRecordKeys(); err != nil {
		fmt.Printf("保存记录key列表失败: %v\n", err)
		return "", err
	}

	// 同步保存记录
	if err := hs.saveRecord(item); err != nil {
		fmt.Printf("保存记录 %s 失败: %v\n", recordID, err)
		return "", err
	}

	return recordID, nil
}

// Update 更新历史记录
func (hs *HistoryStorage) Update(id string, updater func(*HistoryItem)) error {
	hs.mu.Lock()
	defer hs.mu.Unlock()

	item, exists := hs.cache[id]
	if !exists {
		return fmt.Errorf("未找到ID为 %s 的记录", id)
	}

	// 更新记录
	updater(&item)
	hs.cache[id] = item

	// 同步保存
	if err := hs.saveRecord(item); err != nil {
		fmt.Printf("保存记录 %s 失败: %v\n", id, err)
		return err
	}

	return nil
}

// GetAll 获取所有历史记录
func (hs *HistoryStorage) GetAll() ([]HistoryItem, error) {
	hs.mu.RLock()
	defer hs.mu.RUnlock()

	// 按照 recordKeys 的顺序获取记录（已经按时间顺序排列）
	allItems := make([]HistoryItem, 0, len(hs.recordKeys))
	for _, id := range hs.recordKeys {
		if item, exists := hs.cache[id]; exists {
			allItems = append(allItems, item)
		}
	}

	// 反转数组，让最新的记录在前
	for i, j := 0, len(allItems)-1; i < j; i, j = i+1, j-1 {
		allItems[i], allItems[j] = allItems[j], allItems[i]
	}

	return allItems, nil
}

// GetRecent 获取最近的N条记录
func (hs *HistoryStorage) GetRecent(limit int) ([]HistoryItem, error) {
	allItems, err := hs.GetAll()
	if err != nil {
		return nil, err
	}

	if len(allItems) <= limit {
		return allItems, nil
	}

	return allItems[:limit], nil
}

// Clear 清除所有历史记录
func (hs *HistoryStorage) Clear() error {
	hs.mu.Lock()
	defer hs.mu.Unlock()

	// 删除所有记录
	for id := range hs.cache {
		key := hs.getKey(id)
		hs.prefs.RemoveValue(key)
	}

	// 清除记录key列表
	hs.recordKeys = []string{}
	hs.prefs.RemoveValue("history_keys")

	// 重置计数器
	hs.prefs.SetInt("history_id_counter", 0)

	hs.cache = make(map[string]HistoryItem)
	hs.idCounter = 0

	return nil
}

// GetStats 获取统计信息
func (hs *HistoryStorage) GetStats() (total, completed, failed, inProgress int, err error) {
	hs.mu.RLock()
	defer hs.mu.RUnlock()

	total = len(hs.cache)
	for _, item := range hs.cache {
		switch item.Status {
		case "completed":
			completed++
		case "failed":
			failed++
		case "in_progress":
			inProgress++
		}
	}

	return total, completed, failed, inProgress, nil
}

// GetStorageInfo 获取存储信息
func (hs *HistoryStorage) GetStorageInfo() (recordCount int, totalSize int64, err error) {
	hs.mu.RLock()
	defer hs.mu.RUnlock()

	recordCount = len(hs.cache)

	// 计算总大小
	for id, item := range hs.cache {
		data, jsonErr := json.Marshal(item)
		if jsonErr != nil {
			continue
		}
		totalSize += int64(len(data))
		totalSize += int64(len(hs.getKey(id))) // 键的大小
	}

	return recordCount, totalSize, nil
}

// Delete 删除指定记录
func (hs *HistoryStorage) Delete(id string) error {
	hs.mu.Lock()
	defer hs.mu.Unlock()

	if _, exists := hs.cache[id]; !exists {
		return fmt.Errorf("记录不存在")
	}

	// 从缓存中删除
	delete(hs.cache, id)

	// 从记录key列表中删除
	for i, key := range hs.recordKeys {
		if key == id {
			hs.recordKeys = append(hs.recordKeys[:i], hs.recordKeys[i+1:]...)
			break
		}
	}

	// 保存更新后的记录key列表
	if err := hs.saveRecordKeys(); err != nil {
		fmt.Printf("保存记录key列表失败: %v\n", err)
		return err
	}

	// 从 preferences 中删除
	key := hs.getKey(id)
	hs.prefs.RemoveValue(key)

	return nil
}

// Export 导出历史记录到 JSON 字符串
func (hs *HistoryStorage) Export() (string, error) {
	allItems, err := hs.GetAll()
	if err != nil {
		return "", err
	}

	data, err := json.MarshalIndent(allItems, "", "  ")
	if err != nil {
		return "", fmt.Errorf("导出失败: %v", err)
	}

	return string(data), nil
}

// Import 从 JSON 字符串导入历史记录
func (hs *HistoryStorage) Import(jsonData string) error {
	var items []HistoryItem
	if err := json.Unmarshal([]byte(jsonData), &items); err != nil {
		return fmt.Errorf("导入失败: %v", err)
	}

	// 清除现有数据
	if err := hs.Clear(); err != nil {
		return fmt.Errorf("清除现有数据失败: %v", err)
	}

	// 添加导入的数据
	for _, item := range items {
		if _, err := hs.Add(item); err != nil {
			return fmt.Errorf("添加记录失败: %v", err)
		}
	}

	return nil
}
