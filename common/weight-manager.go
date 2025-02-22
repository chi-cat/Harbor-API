package common

import (
	"fmt"
	"sync"
	"time"
)

// ChannelFailureRecord 记录渠道失败信息
type ChannelFailureRecord struct {
	// 失败次数
	FailureCount int
	// 最后一次失败时间
	LastFailureTime time.Time
	// 降权值 (动态计算，避免永久降权)
	PenaltyWeight int
}

// ChannelWeightManager 管理渠道权重的结构
type ChannelWeightManager struct {
	failureRecords map[int]*ChannelFailureRecord
	mutex          sync.RWMutex
	// 恢复时间 - 多久后恢复一部分权重
	recoveryDuration time.Duration
	// 每次降权的基础值
	basePenalty int
}

// NewChannelWeightManager 创建一个新的渠道权重管理器
func NewChannelWeightManager(recoveryDuration time.Duration, basePenalty int) *ChannelWeightManager {
	return &ChannelWeightManager{
		failureRecords:   make(map[int]*ChannelFailureRecord),
		recoveryDuration: recoveryDuration,
		basePenalty:      basePenalty,
	}
}

// RecordFailure 记录渠道失败
func (m *ChannelWeightManager) RecordFailure(channelID int) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	now := time.Now()

	record, exists := m.failureRecords[channelID]
	if !exists {
		record = &ChannelFailureRecord{
			FailureCount:    0,
			LastFailureTime: now,
			PenaltyWeight:   0,
		}
		m.failureRecords[channelID] = record
	}

	record.FailureCount++
	record.LastFailureTime = now
	// 每次失败增加基础降权值，连续失败可以增加更多
	record.PenaltyWeight += m.basePenalty * record.FailureCount
	SysLog(fmt.Sprintf("RecordFailure: channelId=%d, failureCount=%d, lastFailureTime=%v", channelID, record.FailureCount, record.LastFailureTime))

}

// GetPenaltyWeight 获取渠道的当前降权值，考虑时间衰减
func (m *ChannelWeightManager) GetPenaltyWeight(channelID, maxPenaltyWeight int) int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	record, exists := m.failureRecords[channelID]
	if !exists {
		return 0
	}

	// 计算距离上次失败过去了多久
	now := time.Now()
	elapsedDuration := now.Sub(record.LastFailureTime)

	// 根据时间衰减降权值
	recoveryFactor := float64(elapsedDuration) / float64(m.recoveryDuration)
	if recoveryFactor >= 1.0 {
		// 超过恢复时间，完全恢复
		record.PenaltyWeight = 0
		record.FailureCount = 0
		return 0
	}

	// 部分恢复
	reducedPenalty := int(float64(record.PenaltyWeight) * (1.0 - recoveryFactor))
	if reducedPenalty > maxPenaltyWeight {
		return maxPenaltyWeight
	}
	return reducedPenalty
}

// 清理长期未使用的记录
func (m *ChannelWeightManager) CleanupOldRecords(maxAge time.Duration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	now := time.Now()
	for channelID, record := range m.failureRecords {
		if now.Sub(record.LastFailureTime) > maxAge {
			delete(m.failureRecords, channelID)
		}
	}
}

var ChannelWeights *ChannelWeightManager

// 定期清理长期未使用的记录
func setupCleanupTask() {
	ticker := time.NewTicker(10 * time.Minute)
	go func() {
		for range ticker.C {
			ChannelWeights.CleanupOldRecords(30 * time.Minute) // 清理半个小时未使用的记录
		}
	}()
}

func init() {
	// 初始化渠道权重管理器
	// 设置恢复时间为10分钟，基础降权值为1
	ChannelWeights = NewChannelWeightManager(10*time.Minute, 2)
	setupCleanupTask()
}
