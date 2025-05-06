module rate_limiter

go 1.20

require (
	github.com/go-redis/redis/v8 v8.11.5
	github.com/redis/go-redis/v9 v9.0.5
	golang.org/x/net v0.33.0
	google.golang.org/grpc v1.40.0 // 降级至与 Go 1.20 和旧代码兼容的版本
	google.golang.org/protobuf v1.25.0 // 降级至与 Go 1.20 和旧代码兼容的版本
	gopkg.in/yaml.v3 v3.0.1
	github.com/golang/protobuf v1.4.3
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/fsnotify/fsnotify v1.8.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
)

