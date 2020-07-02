FROM node:12
# 创建目录
RUN mkdir -p /app
# 设为工作目录
WORKDIR /app
# 优先拷贝Package.json
COPY package*.json ./
# 包安装
RUN npm install
# 拷贝项目文件
COPY . ./
# 应用程序绑定8787
EXPOSE 8787

CMD [ "node", "dist/app.js" ]