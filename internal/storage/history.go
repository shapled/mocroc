package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// HistoryItem 传输历史记录项
type HistoryItem struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`        // "send" or "receive"
	FileName    string    `json:"fileName"`    // 主要文件名
	FileSize    string    `json:"fileSize"`    // 文件大小
	Code        string    `json:"code"`        // 接收码
	Status      string    `json:"status"`      // "completed", "failed", "in_progress"
	Timestamp   time.Time `json:"timestamp"`   // 创建时间
	Duration    int64     `json:"duration"`    // 传输耗时（秒）
	ClientInfo  string    `json:"clientInfo"`  // 客户端信息
	NumFiles    int       `json:"numFiles"`    // 文件数量
}

// HistoryManager 历史记录管理器
type HistoryManager struct {
	items     []HistoryItem
	mu        sync.RWMutex
	dataFile  string
}

// NewHistoryManager 创建历史记录管理器
func NewHistoryManager() *HistoryManager {
	// 获取用户配置目录
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		userConfigDir = os.TempDir()
	}

	// 创建应用目录
	appDir := filepath.Join(userConfigDir, "mocroc")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		fmt.Printf("创建配置目录失败: %v\n", err)
	}

	dataFile := filepath.Join(appDir, "history.json")

	manager := &HistoryManager{
		dataFile: dataFile,
	}

	// 加载历史记录
	manager.Load()

	return manager
}

// Add 添加历史记录
func (h *HistoryManager) Add(item HistoryItem) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.items = append(h.items, item)
	h.Save()
}

// Update 更新历史记录
func (h *HistoryManager) Update(id string, updater func(*HistoryItem)) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for i := range h.items {
		if h.items[i].ID == id {
			updater(&h.items[i])
			break
		}
	}
	h.Save()
}

// GetAll 获取所有历史记录
func (h *HistoryManager) GetAll() []HistoryItem {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// 返回副本
	result := make([]HistoryItem, len(h.items))
	copy(result, h.items)
	return result
}

// Clear 清除所有历史记录
func (h *HistoryManager) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.items = []HistoryItem{}
	h.Save()
}

// Save 保存历史记录到文件
func (h *HistoryManager) Save() {
	data, err := json.MarshalIndent(h.items, "", "  ")
	if err != nil {
		fmt.Printf("序列化历史记录失败: %v\n", err)
		return
	}

	if err := os.WriteFile(h.dataFile, data, 0644); err != nil {
		fmt.Printf("保存历史记录失败: %v\n", err)
	}
}

// Load 从文件加载历史记录
func (h *HistoryManager) Load() {
	data, err := os.ReadFile(h.dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			// 文件不存在，使用空数据
			h.items = []HistoryItem{}
			return
		}
		fmt.Printf("读取历史记录失败: %v\n", err)
		return
	}

	var items []HistoryItem
	if err := json.Unmarshal(data, &items); err != nil {
		fmt.Printf("反序列化历史记录失败: %v\n", err)
		h.items = []HistoryItem{}
		return
	}

	h.items = items
}

// GetStats 获取统计信息
func (h *HistoryManager) GetStats() (total, completed, failed, inProgress int) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	total = len(h.items)
	for _, item := range h.items {
		switch item.Status {
		case "completed":
			completed++
		case "failed":
			failed++
		case "in_progress":
			inProgress++
		}
	}

	return
}
