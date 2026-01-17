package app

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// InitRedis 初始化 Redis 連線並執行 Ping 測試
func InitRedis(addr string) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	// 使用超時 Context 進行連線測試
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := rdb.Ping(ctx).Result(); err != nil {
		return nil, err
	}

	return rdb, nil
}
