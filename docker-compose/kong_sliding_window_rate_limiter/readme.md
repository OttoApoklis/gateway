# Sliding Window Rate Limiting Plugin for Kong

This plugin implements a sliding window rate limiting algorithm using Redis to limit the number of requests a client can make within a specified time window.

## Features

- Configurable sliding window size and maximum request count.
- Uses Redis for distributed rate limiting.
- Supports per-API or global configuration.

## Installation

1. Clone this repository into your Kong plugins directory (usually `/usr/local/share/lua/5.1/kong/plugins/`).
2. Add the plugin name to the `plugins` property in your Kong configuration file (`kong.conf`):

   ```plaintext
   plugins = bundled,sliding-window-rate-limiting