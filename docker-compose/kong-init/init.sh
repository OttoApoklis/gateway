#!/bin/bash
echo "Initializing Kong Plugins and Dashboards..."
curl -s -X POST http://kong-node1:8001/plugins --data name=prometheus
curl -s -X POST http://kong-node1:8001/plugins --data name=rate-limiting --data config.second=100 --data config.policy=local

