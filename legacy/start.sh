#! /bin/bash
docker build -t obtool .
# 挂载54216端口 并且将dist文件映射入容器内 当dist文件发生改变 容器内部代码同时变化 然后重启容器 目前没有公网git暂时使用映射替代
docker run --name obtool -v /home/web/obtool/dist:/app/dist -p 54216:8787 -d snk-ssr
echo "Docker映射端口----------------54216"