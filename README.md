# Golong
### eks-sshclient Project
- 实现和ssh命令相同的功能命令行工具
### k8s-job Project
- 连接k8s的api-server 获取job信息
### log  Project
- 日志包，实现不同级别的日志记录
- 多并发实现异步写入文件
### gin/sshclient-go Project
- 使用gin框架实现远程执行SSH网页版，加载静态文件，每台主机开启一个goroutine进行处理
### aispace.com/tail_cpl Project
- Tail过滤日志文件，匹配日志后执行shell命令处理错误
### pod-gpu-metrics-exporter
- k8s的GPU-metrics-exporter，主要实现Pod级别调用了多少块GPU卡，每块GPU卡的使用率
- 最终实现以Pod的方式展现GPU使用率
