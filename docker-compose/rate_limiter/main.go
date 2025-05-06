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
	allowed, err := s.limiter.IsAllowed(ctx, req.Api)
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
	// 连接 Redis
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redis 地址
		Password: "",               // 无密码
		DB:       0,                // 使用默认数据库
	})

	// 初始化限流器
	rateLimiter := NewRateLimiter(client, time.Second, 400, "rate_limit:")

	// 启动 gRPC 服务
	server := grpc.NewServer()
	proto.RegisterRateLimiterServer(server, &limiterServer{limiter: rateLimiter})

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Println("gRPC server listening on port :50051")
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

func (r *RateLimiter) IsAllowed(ctx context.Context, api string) (bool, error) {
	now := time.Now().UnixMilli()
	windowStart := now - int64(r.windowSize/time.Millisecond)
	key := r.keyPrefix + api

	pipe := r.client.TxPipeline()

	// 删除超出滑动窗口的数据，减少查询开销
	pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart, 10))

	// 统计当前窗口内的请求数量
	countCmd := pipe.ZCard(ctx, key)
	// 添加当前请求时间戳作为唯一成员
	id := GetSnowFlackID()
	member := fmt.Sprintf("%d-%d", now, id) // 避免重复
	pipe.ZAdd(ctx, key, &redis.Z{
		Score:  float64(now),
		Member: member,
	})

	// 设置过期时间（避免垃圾数据堆积）
	pipe.Expire(ctx, key, r.windowSize)

	// 执行事务
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}

	// 判断是否超过最大请求数
	return countCmd.Val() < int64(r.maxRequests), nil
}
