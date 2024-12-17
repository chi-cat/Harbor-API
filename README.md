
# Harbor API
## 

> [!IMPORTANT]
>  Harbor-API 为NewAPI 的下游作为个人玩家搭建的版本(非官方开发)，改个名字防止认错

> [!NOTE]
> 本项目为开源项目，在[New-API](https://github.com/Calcium-Ion/new-api)和[One-API](https://github.com/songquanpeng/one-api)的基础上进行二次开发

> [!IMPORTANT]
> 使用者必须在遵循 OpenAI 的[使用条款](https://openai.com/policies/terms-of-use)以及**法律法规**的情况下使用，不得用于非法用途。
> 本项目仅供个人学习使用，不保证稳定性，且不提供任何技术支持。
> 根据[《生成式人工智能服务管理暂行办法》](http://www.cac.gov.cn/2023-07/13/c_1690898327029107.htm)的要求，请勿对中国地区公众提供一切未经备案的生成式人工智能服务。

> [!TIP]
> 最新版Docker镜像：`gtxy27/harbor-api:latest`  
> 默认账号root 密码123456  
> 更新指令：
> ```
> docker run --rm -v /var/run/docker.sock:/var/run/docker.sock containrrr/watchtower -cR
> ```


## 主要修改特点
此分叉版本的主要变更如下：

1. 删除了所有个人用不到的功能
2. 更新了deepseek的价格处理函数
3. 将web访问限制默认去除，课通过系统环境`GLOBAL_WEB_RATE_LIMIT_ENABLE`修改 `false/true`
4. 首页默认为数据面板
5. 增加主页响应式布局

## 部署
### 部署要求
- 本地数据库（默认）：SQLite（Docker 部署默认使用 SQLite，必须挂载 `/data` 目录到宿主机）
- 远程数据库：MySQL 版本 >= 5.7.8，PgSQL 版本 >= 9.6


### 基于 Docker 进行部署
### 使用 Docker Compose 部署（推荐）
```shell
# 下载项目
git clone https://github.com/gtxy27/Harbor-API
cd new-api
# 按需编辑 docker-compose.yml
# 启动
docker-compose up -d
```

### 直接使用 Docker 镜像
```shell
# 使用 SQLite 的部署命令：
docker run --name new-api -d --restart always -p 3000:3000 -e TZ=Asia/Shanghai -v /home/ubuntu/data/new-api:/data gtxy27/harbor-api:latest
# 使用 MySQL 的部署命令，在上面的基础上添加 `-e SQL_DSN="root:123456@tcp(localhost:3306)/oneapi"`，请自行修改数据库连接参数。
# 例如：
docker run --name new-api -d --restart always -p 3000:3000 -e SQL_DSN="root:123456@tcp(localhost:3306)/oneapi" -e TZ=Asia/Shanghai -v /home/ubuntu/data/new-api:/data gtxy27/harbor-api:latest
```

## 相关项目
- [One API](https://github.com/songquanpeng/one-api)：原版项目
- [Midjourney-Proxy](https://github.com/novicezk/midjourney-proxy)：Midjourney接口支持
- [chatnio](https://github.com/Deeptrain-Community/chatnio)：下一代 AI 一站式 B/C 端解决方案
- [neko-api-key-tool](https://github.com/Calcium-Ion/neko-api-key-tool)：用key查询使用额度
- [New API](https://github.com/Calcium-Ion/new-api) 上游项目


