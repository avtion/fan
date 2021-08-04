# 饭酱: 一个单纯的干饭提醒机器人 🍢

[![Golang 1.16](https://img.shields.io/badge/Golang-1.16-blue)](https://golang.org/) [![DockerHub](https://img.shields.io/badge/Docker-avtion%2Ffan-blue)](https://hub.docker.com/r/avtion/fan) ![MIT Licence](https://img.shields.io/badge/license-MIT-brightgreen)

# 饭酱目前能做什么 🌟

- 午餐和晚餐飞书提醒
- 美餐未下单飞书提醒
- 卖萌

# 灵感来源 💞

- 自己/别人忘记点饭
- 自己/别人看见别人/自己忘记点饭
- 自认为已经点了明天的饭
- 周日过得太悠闲忘记点周一的饭
- 周四忘记取消周五的饭
- 忘记明天/中午/晚上吃什么
- 取餐时不想打开微信再打开美餐

# Usage 使用方式 🚀

## 1. Docker(推荐) 🐳

```bash
docker run --name fan -d \
--restart unless-stopped \
avtion/fan:latest -u ${你的美餐账号} -p ${你的美餐密码} -w ${你的飞书机器人Webhook}
```

## 2. Exec 直接运行 🏄

1. `git clone git@github.com:avtion/fan.git`
2. 执行`go run . -u ${你的美餐账号} -p ${你的美餐密码} -w ${你的飞书机器人Webhook}`

# Custom 自定义配置 🐍

1. 拷贝 `config.yaml.example` 文件并重命名为 `config.yaml`
2. 修改需要自定义的配置项
3. Docker运行的时候加入参数`-v ${config.yaml的路径}:/etc/config.yaml`

## 多地点过滤Flag参数

- --nohx: 排除行信
- --noxh: 排除星辉
- --nogz: 排除高志

# TODO 薛定谔的更新计划 🖋

- [x] 多地点点餐过滤
- [ ] 支持数据上报 - 收集菜品信息以了解自己吃饭情况
- [ ] 指令机器人 - 支持快速下单和修改订单
- [ ] Web支持 - 配置修改和数据查询
- [ ] 独立功能模块

# License 📡

MIT License