local BasePlugin = require "kong.plugins.base_plugin"
local redis = require "resty.redis"
local cjson = require "cjson"

local SlidingWindowRateLimitingHandler = BasePlugin:extend()

function SlidingWindowRateLimitingHandler:new()
    SlidingWindowRateLimitingHandler.super.new(self, "sliding-window-rate-limiting")
end

-- Helper function to connect to Redis
local function connect_redis(conf)
    local red = redis:new()
    red:set_timeout(1000)  -- 1 second timeout

    local ok, err = red:connect(conf.redis_host, conf.redis_port)
    if not ok then
        return nil, "Failed to connect to Redis: " .. err
    end

    -- Optional: Authenticate if needed
    if conf.redis_password and conf.redis_password ~= "" then
        local ok, err = red:auth(conf.redis_password)
        if not ok then
            return nil, "Failed to authenticate with Redis: " .. err
        end
    end

    return red
end

-- Helper function to execute the Lua script
local function execute_lua_script(red, key, windowStart, now, member, expireTime, windowSize, maxRequests)
    local script = [[
        redis.call('ZREMRANGEBYSCORE', KEYS[1], 0, ARGV[1])
        local count = redis.call('ZCARD', KEYS[1])
        redis.call('ZADD', KEYS[1], ARGV[2], ARGV[3])
        redis.call('EXPIRE', KEYS[1], tonumber(ARGV[4]))
        return count
    ]]

    local res, err = red:eval(script, 1, key, windowStart, now, member, expireTime)
    if err then
        return nil, "Redis Lua script error: " .. err
    end

    -- The script returns the count, but we need to ensure it's a number
    if not tonumber(res) then
        return nil, "Unexpected response from Redis Lua script"
    end

    return tonumber(res), nil
end

function SlidingWindowRateLimitingHandler:access(conf)
    SlidingWindowRateLimitingHandler.super.access(self)

    local red, err = connect_redis(conf)
    if not red then
        return kong.response.error(500, err)
    end
    -- Ensure the connection is closed in case of an error later
    local ok, close_err = pcall(red.close, red)
    if not ok then
        kong.log.err("Failed to close Redis connection: ", close_err)
    end

    -- Fetch parameters
    local key = kong.request.get_path()  -- Using request path as key
    local windowSize = conf.window_size or 1000  -- Default to 1000 ms if not set
    local maxRequests = conf.max_requests or 100  -- Default to 100 requests if not set

    local now = ngx.now() * 1000  -- Convert to milliseconds
    local windowStart = now - windowSize
    local expireTime = (windowSize / 1000) + 1  -- in seconds
    local member = kong.client.get_ip()  -- Use client IP as member

    -- Ensure expireTime is an integer
    expireTime = math.floor(expireTime)

    -- Execute Lua script
    local res, err = execute_lua_script(red, key, windowStart, now, member, expireTime, windowSize, maxRequests)
    if not res then
        return kong.response.error(500, err)
    end

    -- Check if the request count exceeds maxRequests
    if res >= maxRequests then
        return kong.response.error(429, "Rate limit exceeded")
    end

    -- Optional: Set a header to indicate remaining requests
    -- You would need to query Redis again for the current count, which might not be efficient.
    -- Alternatively, adjust the Lua script to return the remaining requests.
end

return SlidingWindowRateLimitingHandler