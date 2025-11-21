# 项目：百度网盘管理工具
## 项目介绍
调用百度网盘开放API，实现网盘文件管理，当前已实现功能
- 授权码登录（pan/pan.go中实现）
- 获取网盘文件列表（pan/list.go中实现）
- 下载网盘文件（pan/download.go中实现）
- 上传网盘文件（pan/upload.go中实现）
- 删除网盘文件（pan/remove.go中实现）
- 移动网盘文件（pan/move.go中实现）
- 重命名网盘文件（pan/rename.go中实现）
- 创建目录（pan/mkdir.go中实现）
- 复制文件或文件夹 （pan/copy.go中实现）
- 显示文件信息 （pan/info.go中实现）
- 显示网盘容量信息（pan/info.go中实现）
- 刷新Access Token（pan/pan.go中实现）

## 通用规则
- 编译必须使用项目目录下的build.sh脚本
- 不要创建和运行单元测试
- 只执行build.sh编译脚本测试生成的代码，不要执行编译后的可执行文件
- 实现新的网盘文件操作或管理功能时，在pan/目录下创建新的文件，并实现功能

## 百度网盘开放API文档索引
- 授权码模式-用户授权码 Code 换取 Access Token 凭证：https://pan.baidu.com/union/doc/al0rwqzzl?from=open-sdk-go
- 设备码模式-获取设备码、用户码：https://pan.baidu.com/union/doc/fl1x114ti?from=open-sdk-go
- 设备码模式-用 Device Code 轮询换取 Access Token：https://pan.baidu.com/union/doc/fl1x114ti?from=open-sdk-go
- 授权码模式-刷新 Access Token：https://pan.baidu.com/union/doc/al0rwqzzl?from=open-sdk-go
- 获取文件信息-获取文档列表：https://pan.baidu.com/union/doc/Eksg0saqp?from=open-sdk-go
- 获取文件信息-获取图片列表：https://pan.baidu.com/union/doc/bksg0sayv?from=open-sdk-go
- 获取文件信息-获取文件列表：https://pan.baidu.com/union/doc/nksg0sat9?from=open-sdk-go
- 获取文件信息-搜索文件：https://pan.baidu.com/union/doc/zksg0sb9z?from=open-sdk-go
- 管理文件-文件复制：https://pan.baidu.com/union/doc/mksg0s9l4?from=open-sdk-go
- 管理文件-文件删除：https://pan.baidu.com/union/doc/mksg0s9l4?from=open-sdk-go
- 管理文件-文件移动：https://pan.baidu.com/union/doc/mksg0s9l4?from=open-sdk-go
- 管理文件-文件重命名：https://pan.baidu.com/union/doc/mksg0s9l4?from=open-sdk-go
- 文件上传-分片上传：https://pan.baidu.com/union/doc/nksg0s9vi?from=open-sdk-go
- 文件上传-创建文件：https://pan.baidu.com/union/doc/rksg0sa17?from=open-sdk-go
- 文件上传-预上传：https://pan.baidu.com/union/doc/3ksg0s9r7?from=open-sdk-go
- 获取文件信息-递归获取文件列表：https://pan.baidu.com/union/doc/Zksg0sb73?from=open-sdk-go
- 获取文件信息-查询文件信息：https://pan.baidu.com/union/doc/Fksg0sbcm?from=open-sdk-go
- 获取网盘容量信息：https://pan.baidu.com/union/doc/Cksg0s9ic?from=open-sdk-go
- 获取用户信息：https://pan.baidu.com/union/doc/pksg0s9ns?from=open-sdk-go