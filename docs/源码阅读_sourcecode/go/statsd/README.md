# statsd client 源码

https://github.com/alexcesaro/statsd/tree/v2.0.0

# 简介 statsd

https://github.com/longtian/introduction-to-statsd

StatsD 的协议其实非常简单，每一行就是一条数据。

```
<metric_name>:<metric_value>|<metric_type>
```

以监控系统负载为例，假如某一时刻系统 1分钟的负载 是 0.5，通过以下命令就可以写入 StatsD。

```
echo "system.load.1:0.5|g" | nc 127.0.0.1 8251
```

结合其中的 system.load.1:0.5|g，各个部分分别为：

指标名称 metric_name = system.load.1
指标的值 metric_value = 0.5
指标类型 metric_type = g(gauge)

指标名称:指标命名没有特别的限制，但是一般的约定是使用点号分割的命名空间。
指标的值:指标的值是大于等于 0 的浮点数。
指标类型:

- gauge : 一维变量，可增可减。比如 cpu 使用率
- counter: 累加变量。
- ms: 记录执行时间。比如请求响应时间是 12s。 `echo "request.time:12|ms" | nc 127.0.0.1 8251`
- set: 统计变量不重复的个数。比如在线用户。`echo "online.user:9587|s" | nc 127.0.0.1 8251`
