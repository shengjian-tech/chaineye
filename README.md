<img src="doc/img/chaineye.png" width="240">

>chaineye是一款开源的区块链监控平台，目前已经支持百度XuperChain，基于[nightingale](https://github.com/ccfos/nightingale)二次开发,开箱即用的产品体验。


## 预览
<img src="doc/img/overview.png" width="800">

## 快速安装
- 前置:需要安装Prometheus或者其他工具作为数据源。已有正在运行的XuperChain网络。
- 克隆项目到本地 项目地址 https://github.com/shengjian-tech/chaineye
- `go mod tidy`下载依赖, `go build -ldflags "-w -s" -o chaineye ./cmd/center/main.go`编译完成。
- 执行sql文件[./docker/initsql/a-n9e.sql](./docker/initsql/a-n9e.sql) 
- 修改 [./etc/config.toml](./etc/config.toml) 配置文件。 配置Redis连接，数据库连接，Prometheus服务地址，`XuperSdkYmlPath` 配置文件,将```#UseFileAssets = true```的注释解开。
- 修改完配置文件后，在根目录执行命令即可启动`chaineye`服务。命令 `nohup ./chaineye &` , 随后可以通过查看日志输出，判断服务是否正常启动。
- 下载`chaineye`对应前端项目`front_chaineye`，仓库路径 https://github.com/shengjian-tech/front_chaineye
- 下载前端最新的release版本或者自行编译，解压后，将`pub`目录放到`chaineye`可执行文件同级目录。
- 访问`http://127.0.0.1:17000` 页面, 账号：root 密码：root.2020  
- 导入XuperChain监控大盘，XuperChain大盘文件路径 [xuper_metric.json](./xuper_metric.json) 下载后，在监控大盘中，导入即可。

## 超级链监控大盘预览
<img src="doc/img/metric.png" width="800">


## 鸣谢
[夜莺nightingale](https://github.com/ccfos/nightingale)  
[XuperChain](https://github.com/xuperchain/xuperchain)
