local typedefs = require "kong.db.schema.typedefs"

return {
    name = "sliding-window-rate-limiting",
    fields = {
        { consumer = typedefs.no_consumer }, -- This plugin cannot be configured per consumer
        { protocols = typedefs.protocols_http }, -- Plugin works with HTTP(S) only
        {
            config = {
                type = "record",
                fields = {
                    { redis_host = { type = "string", default = "127.0.0.1" } },
                    { redis_port = { type = "number", default = 6379 } },
                    { redis_password = { type = "string", default = "" } },
                    { window_size = { type = "number", default = 1000 } }, -- in milliseconds
                    { max_requests = { type = "number", default = 100 } },
                },
            },
        },
    },
}