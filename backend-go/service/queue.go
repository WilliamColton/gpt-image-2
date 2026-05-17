package service

import (
	"log/slog"
	"sync"
	"time"

	"gpt-image-playground/backend/config"
)

type endpointLimiter struct {
	sem chan struct{}
}

var (
	limiters   sync.Map // baseURL -> *endpointLimiter
	slotNotify = make(chan struct{}, 100)
)

func getLimiter(ep config.ApiEndpoint) *endpointLimiter {
	if ep.MaxConcurrency <= 0 {
		return nil
	}
	if v, ok := limiters.Load(ep.BaseURL); ok {
		return v.(*endpointLimiter)
	}
	lim := &endpointLimiter{sem: make(chan struct{}, ep.MaxConcurrency)}
	actual, loaded := limiters.LoadOrStore(ep.BaseURL, lim)
	if loaded {
		return actual.(*endpointLimiter)
	}
	return lim
}

func tryAcquire(ep config.ApiEndpoint) (release func(), ok bool) {
	lim := getLimiter(ep)
	if lim == nil {
		return func() {}, true
	}
	select {
	case lim.sem <- struct{}{}:
		once := &sync.Once{}
		return func() {
			once.Do(func() {
				<-lim.sem
				select {
				case slotNotify <- struct{}{}:
				default:
				}
			})
		}, true
	default:
		return nil, false
	}
}

// TryAcquireSlot 非阻塞尝试获取任意端点的槽位
func TryAcquireSlot(endpoints []config.ApiEndpoint) (int, func(), bool) {
	return TryAcquireSlotFrom(endpoints, 0)
}

func TryAcquireSlotFrom(endpoints []config.ApiEndpoint, start int) (int, func(), bool) {
	for i := start; i < len(endpoints); i++ {
		if release, ok := tryAcquire(endpoints[i]); ok {
			return i, release, true
		}
	}
	return 0, nil, false
}

// AcquireSlot 阻塞等待，按端点顺序尝试获取槽位
// onAcquired 在成功获取槽位时调用一次（用于将任务状态从 queued 更新为 running）
func AcquireSlot(endpoints []config.ApiEndpoint, onAcquired func()) (int, func()) {
	return AcquireSlotFrom(endpoints, 0, onAcquired)
}

func AcquireSlotFrom(endpoints []config.ApiEndpoint, start int, onAcquired func()) (int, func()) {
	acquired := false
	for {
		if i, release, ok := TryAcquireSlotFrom(endpoints, start); ok {
			if !acquired && onAcquired != nil {
				acquired = true
				onAcquired()
			}
			return i, release
		}
		slog.Debug("所有端点已满，等待槽位释放")
		select {
		case <-slotNotify:
		case <-time.After(time.Second):
		}
	}
}

// RefreshLimiters 清空所有限流器（端点配置变更后调用）
func RefreshLimiters() {
	limiters.Range(func(key, _ any) bool {
		limiters.Delete(key)
		return true
	})
	slog.Info("端点限流器已重置")
}
