# 1. 监控简介

指标是软件或硬件组件属性的度量。为了使指标有价值，我们会跟踪其状态，通常记录一段时间内的数据点。这些数据点称为观察点（observation），观察点通常包括值、时间戳，有时也涵盖描述观察点的一系列属性（如源或标签）。观察的集合称为时间序列。

指标的类型：

- 测量型(gauge)。上下增减的数字，比如cpu/内存/磁盘，用户访问数
- 计数型(counter):随着时间只增不减的数字。比如运行时间、登陆次数
- 直方图(histogram): 对观察点进行采样

百分位数: 度量的是占总数特定百分比的观察点的值。比如50百分位数（或p50）。对于中间数（已排好序的数据）来说，50%的值低于它，50%高于它。

监控方法论：

- USE: 使用率(Utilization), 饱和度(Saturation)，错误(Error)
- google的4 个黄金指标：延迟、流量(qps等)、错误、饱和度


# 2. Prometheus 简介

Prometheus称其可以抓取的指标来源为端点（endpoint）。端点通常对应单个进程、主机、服务或应用程序。为了抓取端点数据，Prometheus定义了名为目标（target）的配置

# 3. 安装和启动 Prometheus

`brew install prometheus`


# 4. 监控主机和容器


