FROM kong:3.4.1
USER root
# 将自定义插件复制到容器中
COPY ./kong_sliding_window_rate_limiter /usr/local/kong/plugins/kong_sliding_window_rate_limiter

# 安装 LuaRocks 依赖项

RUN luarocks install kong_sliding_window_rate_limiter


# 设置环境变量启用插件
ENV KONG_PLUGINS=bundled,kong_sliding_window_rate_limiter
