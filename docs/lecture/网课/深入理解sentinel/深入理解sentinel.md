# 01 一次服务雪崩的排查经历

- Jedis 抛出 Read time out 的原因：由于缓存的 value 字符串太长，网络传输数据包大，导致 Jedis 执行 get 命令耗时长。
- 服务 A 出现 RPC 调用超时的原因：业务代码的缺陷、并发量的突增，以及缓存设计缺陷导致 Jedis 读操作耗时长，导致服务 B 接口执行耗时超过 3 秒，从而导致服务 A 远程 RPC 调用超时。
- 服务 A 出现服务 B 拒绝请求异常的原因：服务 A 调用超时触发 dubbo 超时重试，原本并发量就已经很高，加上耗时的接口调用，服务 B 业务线程池的线程全部处于工作状态，服务 B 已经高负荷运行，而 Dubbo 的超时重试又导致服务 B 并发量翻倍，简直雪上加霜，服务 B 处理不过来只能拒绝执行服务 A 的请求。
- 服务 A 奔溃的原因：服务 B 的不可用导致服务 A 处理不过来客户端发来的请求，而服务 A 又没有拒绝客户端的请求，客户端请求源源不断，最后服务 A 请求堆积，导致 SocketChannel 占用的文件句柄达到上限，服务 A 就奔溃了。

# 02 为什么需要服务降级以及常见的几种降级方式

常见的服务降级实现方式有：开关降级、限流降级、熔断降级。

- 限流降级: 直接拒绝;匀速排队;冷启动
- 熔断降级
  - 在每秒请求异常数超过多少时触发熔断降级
  - 在每秒请求异常错误率超过多少时触发熔断降级
  - 在每秒请求平均耗时超过多少时触发熔断降级
- 开关降级。人工开关(配置中心或者redis)。或者定时服务分时段开启

# 03 为什么选择 Sentinel，Sentinel 与 Hystrix 的对比

Sentinel	Hystrix
隔离策略	信号量隔离	线程池隔离/信号量隔离
熔断降级策略	基于响应时间或失败比率	基于失败比率
实时指标实现	滑动窗口	滑动窗口（基于 RxJava）
规则配置	支持多种数据源	支持多种数据源
扩展性	多个 SPI 扩展点	插件的形式
基于注解的支持	支持	支持
限流	基于 QPS，支持基于调用关系的限流	有限的支持
流量整形	支持慢启动、匀速器模式	不支持
系统负载保护	支持	不支持
控制台	开箱即用，可配置规则、查看秒级监控、机器发现等	不完善
常见框架的适配	Servlet、Spring Cloud、Dubbo、gRPC 等	Servlet、Spring Cloud Netflix

# 04 Sentinel 基于滑动窗口的实时指标数据统计

```java
public class MetricBucket {
    /**
     * 存储各事件的计数，比如异常总数、请求总数等
     */
    private final LongAdder[] counters;
    /**
     * 这段事件内的最小耗时
     */
    private volatile long minRt;
}
```

# 05 Sentinel 的一些概念与核心类介绍

- 资源：资源是 Sentinel 的关键概念。资源，可以是一个方法、一段代码、由应用提供的接口，或者由应用调用其它应用的接口。
- 规则：围绕资源的实时状态设定的规则，包括流量控制规则、熔断降级规则以及系统保护规则、自定义规则。
- 降级：在流量剧增的情况下，为保证系统能够正常运行，根据资源的实时状态、访问流量以及系统负载有策略的拒绝掉一部分流量。


# 16 Sentinel 动态数据源：规则动态配置

- sentinel-datasource-extension：定义动态数据源接口、提供抽象类
- sentinel-datasource-redis：基于 Redis 实现的动态数据源
- sentinel-datasource-zookeeper： 基于 ZooKeeper 实现的动态数据源


# 19 Sentinel 集群限流的实现（下）

sentinel-core 模块的 cluster 包下定义了实现集群限流功能的相关接口：

- TokenService：定义客户端向服务端申请 token 的接口，由 FlowRuleChecker 调用。
- ClusterTokenClient：集群限流客户端需要实现的接口，继承 TokenService。
- ClusterTokenServer：集群限流服务端需要实现的接口。
- EmbeddedClusterTokenServer：支持嵌入模式的集群限流服务端需要实现的接口，继承 TokenService、ClusterTokenServer。


TokenService 接口的定义如下：

```
public interface TokenService {
    TokenResult requestToken(Long ruleId, int acquireCount, boolean prioritized);
    TokenResult requestParamToken(Long ruleId, int acquireCount, Collection<Object> params);
}
```
