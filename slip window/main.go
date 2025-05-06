package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"google.golang.org/grpc"
	"rate_limiter/proto"
)

// 定义 limiterServer，并嵌入 UnimplementedRateLimiterServer
type limiterServer struct {
	proto.UnimplementedRateLimiterServer
	limiter *RateLimiter
}

func (s *limiterServer) Check(ctx context.Context, req *proto.CheckRequest) (*proto.CheckResponse, error) {
	allowed, err := s.limiter.CheckSlidingWindow(ctx, req.Api)
	if err != nil {
		return &proto.CheckResponse{Allowed: false, Message: "internal error"}, err
	}

	if !allowed {
		return &proto.CheckResponse{
			Allowed: false,
			Message: "rate limit exceeded",
		}, nil
	}

	return &proto.CheckResponse{Allowed: true, Message: "allowed"}, nil
}

func main() {
	cfg := LoadConfig("./config.yaml")
	// 连接 Redis
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port), // Redis 地址
		Password: fmt.Sprintf("%s", cfg.Redis.Password),                // 无密码
		DB:       cfg.Redis.DB,                                         // 使用默认数据库
	})

	// 初始化限流器
	rateLimiter := NewRateLimiter(client, time.Duration(cfg.RateLimiter.WindowSize)*time.Second, cfg.RateLimiter.MaxRequests, cfg.RateLimiter.KeyPrefix)

	// 启动 gRPC 服务
	server := grpc.NewServer()
	proto.RegisterRateLimiterServer(server, &limiterServer{limiter: rateLimiter})

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Println("gRPC server listening on port " + fmt.Sprintf(":%d", cfg.Server.Port))
	if err := server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

type RateLimiter struct {
	client      *redis.Client
	windowSize  time.Duration
	maxRequests int
	keyPrefix   string
}

func NewRateLimiter(client *redis.Client, windowSize time.Duration, maxRequests int, keyPrefix string) *RateLimiter {
	return &RateLimiter{
		client:      client,
		windowSize:  windowSize,
		maxRequests: maxRequests,
		keyPrefix:   keyPrefix,
	}
}

// 定义 Lua 脚本
var slidingWindowScript = redis.NewScript(`
    redis.call('ZREMRANGEBYSCORE', KEYS[1], 0, ARGV[1])
    local count = redis.call('ZCARD', KEYS[1])
    redis.call('ZADD', KEYS[1], ARGV[2], ARGV[3])
    redis.call('EXPIRE', KEYS[1], tonumber(ARGV[4]))
    return count
`)

// 滑动窗口限流检查
func (r *RateLimiter) CheckSlidingWindow(ctx context.Context, key string) (bool, error) {
	client := r.client
	windowSize := r.windowSize
	now := time.Now().UnixNano() / int64(time.Millisecond)  // 当前时间戳（毫秒）
	windowStart := now - int64(windowSize/time.Millisecond) // 窗口起始时间
	expireTimeSec := int64(windowSize/time.Second) + 1      // 过期时间（秒）
	id := GetUUID()                                         // 唯一 ID（如 Snowflake）
	member := fmt.Sprintf("%d-%s", now, id)                 // 唯一成员标识

	// 执行 Lua 脚本
	result, err := slidingWindowScript.Run(ctx, client, []string{key},
		strconv.FormatInt(windowStart, 10),   // ARGV[1] windowStart
		strconv.FormatInt(now, 10),           // ARGV[2] now
		member,                               // ARGV[3] member
		strconv.FormatInt(expireTimeSec, 10), // ARGV[4] expireTime
	).Result()

	if err != nil {
		return false, fmt.Errorf("run lua script failed: %v", err)
	}

	// 解析结果
	count, ok := result.(int64)
	if !ok {
		return false, fmt.Errorf("invalid result type: %T", result)
	}

	// 判断是否超过最大请求数
	return count < int64(r.maxRequests), nil
}

