# 1 初识Kafka

### 1.1 发布与订阅消息系统

### 1.2 Kakfa登场

Kafka 发布与订阅的消息系统，一般称为『分布式提交日志』或者『分布式流平台』

Kafka 的数据单元称为消息， 为了提高效率，消息分批次写入 Kafka，这些消息属于同一个主题和分区。

Kafka 消息通过主题(topic)分类，topic 好比数据库的表，主题可以分为若干个分区，一个分区就是一个提交日志。消息以追加方式写入分区，然后以先入先出顺序读取。因为有多个分区，所以无法保证整个 topic 范围消息顺序，但是可以保证单个分区内的顺序。 一个主题的分区可以横跨多个服务器，提供更强大的性能。

通常用『流』描述生产者移动到消费者的数据。

生产者创建消息，默认把消息均衡分布到主题的所有分区上，而不关心特定消息被写入到哪个分区。
包含同一个键的消息被写入到同一个分区。
消费者可以订阅一个或者多个主题，按照消息生成顺序读取它们，通过检查消息偏移量区分已读消息。

broker: 一个独立的 Kafka 服务器称为broker.接收来自生产者的消息，为消息设置偏移量，并提交消息到磁盘保存。每个集群都有一个 broker 同事充当了集群控制器的角色（选举出来）。集群中一个分区从属于一个 broker，该 broker 称为分区的首领。一个分区可以分配给多个 broker，发生分区复制，为分区提供消息冗余，
如果有一个 broker 失效，其他 broker 可以接管领导权。

过期：通过设置保留策略，超时或者超过特定大小字节数，删除过期旧消息。

多集群：

- 数据类型分离
- 安全需求隔离
- 多数据中心（容灾）


为什么选择 kafka？

- 多个生产者
- 多个消费者，共享一个消息流
- 基于磁盘的数据存储
- 伸缩性

### 1.4 数据生态系统

使用场景：

- 活动跟踪
- 传递消息
- 度量指标和日志记录
- 提交日志
- 流处理


# 2 安装 Kafka

依赖 Java, Zookeeper

Kafka 使用 Zookeeper  保存集群元数据库信息和消费者信息
Zookeeper 群组(Ensemble)，使用一致性协议，建议每个群组包含奇数个节点

硬件：

- 磁盘吞吐：ssd 固态硬盘胜过 传统的机械硬盘 HDD
- 磁盘容量：取决于需要保留的消息数量
- 内存：不建议 kafka 和其他程序一起部署，共享页面缓存会降低 kafka 消费者性能
- 网络：网络和磁盘存储使主要制约扩展的因素
- cpu：kafka 对计算能力要求相对较低，不是主要因素

可以修改内核参数针对 Kafka 调优。

修改 jvm 垃圾回收参数来调优。


# 3. Kafka 生产者-向 Kafka 写入数据

场景：是否允许丢失？偶尔重复是否能接受？是否有严格的延迟和吞吐量要求

```java
private Properties kafkaProps=new Properties();
kafkaProps.put("bootstrap.serves", "broker1:9092,borker2:9092");

kafkaProps.put("key.serializer", "org.apache.kafka.common.serizlization.StringSerializer");
kafkaProps.put("value.serializer", "org.apache.kafka.common.serizlization.StringSerializer");
producer = new KafkaProducer<String, String>(kafkaProps);
```

发送方式：

- 发送并忘记(fire-and-forget)
- 同步发送 (send发送返回一个 Future 对象，调用 get()方法等待，就知道是否发送成功)
- 异步发送(send()并且指定一个回调函数，服务器在返回响应时调用该函数)

```java
ProducerRecord<String, String> record=new ProducerRecord<>("CostomerCountry", "Precision Products", "France");
try {
    producer.send(record);
} catch (Exception e){
    e.printStackTrace();
}

// 同步发送
ProducerRecord<String, String> record=new ProducerRecord<>("CostomerCountry", "Precision Products", "France");
try {
    producer.send(record).get(); // 调用返回的 Future() 对象的 get() 方法等待 kafka响应
} catch (Exception e){
    e.printStackTrace();
}
// kafka 异常分为可重试和不可重试

// 异步发送
private class DemoProducerCallback implements Callback {
    @Override
    public void onCompletion(RecordMetadata recordMetadata, Exception e) {
    if (e!=null{
                e.printStackTrace();

    })

    }

}
ProducerRecord<String, String> record=new ProducerRecord<>("CostomerCountry", "Precision Products", "France");
producer.send(record, new DemoProducerCallback());
```

# 4 Kafka消费者-从 Kafka 读取数据
应用程序从 KafkaConsumer 向 Kafka 订阅主题，并从订阅的主题上接收消息。

### 4.1 KafkaConsumer 概念
如果写入消息速度远超读取速度，怎么办？对消费者进行横向伸缩

Kafka 消费者从属于消费者群组。一个群组里的消费者订阅的是同一个主题，每个消费者接收主题一部分分区的消息。但是不要让消费者数量超过分区数量，多余消费者会闲置。

消费者群组和分区再平衡：

分区所有权从一个消费者转移到另一个消费者，称之为再均衡。

消费者通过向被指派为群组协调器的 broker（不同群组可以有不同的协调器）发送心跳来维持它们和群组的
从属关系以及它们对分区的所有权关系。消费者轮询消息（为了获取消息）或者提交偏移量时发送心跳。
如果消费者停止发送心跳的时间足够长，会话就会过期，群组协调器就认为他已经死亡，出发一次再均衡。

# 4.2 创建 Kafka 消费者

```java
// 创建 KafkaConsumer 对象
Properties props = new Properties();
props.put("bootstrap.servers", "broker1:9092,borker2:9092");
props.put("group.id", "CountryCounter");
pros.put("key.deserializer", "org.apache.kafka.common.serialization.StringDeserializer");
pros.put("value.deserializer", "org.apache.kafka.common.serialization.StringDeserializer");

KafkaConsumer<String, String> consumer = new KafkaConsumer<String, String>(props);

// 订阅主题, 支持正则订阅多个主题。 consumer.subscribe("test.*")
consumer.subscribe(Collections.singletonList("customerCountries"));

// 轮询。轮询会处理所有细节：群组协调，分区再平衡、发送心跳和获取数据
// 最好一个消费者使用一个线程
try {
    while (true) { //  消费者是个死循环
        ConsumerRecords<String, String> records = consumer.poll(100); // 指定毫秒一致等待broker返回数据
        for (ConsumerRecord<String, String> record: records) {
            int updateCounr = 1;
            if (custCountryMap.containsValue(record.Value())) {
                updateCount = custCountryMap.get(record.get(record.value())) + 1;
            }
            custCountryMap.put(record.value(), updateCount);
            JSONObject json = new JSONObject(custCountryMap);
            System.out.println(json.toString(4));
        }
    }
} finally {
    consumer.close();
}
```

提交：把更新分区当前位置的操作叫做提交。消费者往一个叫做 `_consumer_offset` 的特殊主题发送消息，
消息里包含每个分区的偏移量。如果消费者一直处于运行状态，偏移量就没啥用处。
如果消费者崩溃或者有新的消费者加入群组，就会出发再均衡，完成均衡以后每个消费者可能分配到新分区，
为了能继续之前工作，消费者需要读取每个分区最后一次提交的偏移量，然后从偏移量指定地方继续处理。

如果提交的偏移量小于客户端处于的最后一个消息的偏移量，那么处于两个偏移量之间的消息会被重复处理。
如果提交的偏移量大于客户端处理的最后一个消息的偏移量，俩偏移量之间的消息将会丢失。

所以提交偏移量的方式对客户端有很大影响:

##### 4.6.1 自动提交

最简单的方式就是自动提交。enable.auto.commit 为 true，每5s 消费者就会自动把从 poll()接收到的最大偏移量提交上去。提交间隔 auto.commit.interval.ms 控制

问题：最后一次提交之后的3s 发生了再均衡，再均衡之后消费者从最后一次提交的偏移量开始读取消息，
这个时候偏移量已经落后3s，这3s 到达的消息会重复处理。没有给开发者余地避免重复处理消息.

##### 4.6.2 提交当前偏移量

大部分开发者通过控制偏移量提交时间来消除丢失消息的可能，并在发生再均衡时减少重复消息数量。
auto.commit.offset 设置为false，让程序决定何时提交。commitSync() 提交最简单可靠，
这个 api 会提交由 poll() 方法返回的最新偏移量。成功立马返回，否则抛异常。

##### 4.6.3 异步提交

consumer.commitAsync()
可以通过用一个单调递增序列号来维护异步提交顺序。每次提交偏移量之后或者在回调里提交偏移量递增。
进行重试前，先检查回调序列号和即将提交的偏移量是否相等，相等说明没有新的提交，
可以安全重试，否则说明一个新的提交已经发送出去了，应该停止重试。

### 4.11 独立消费者

有时候简单一些不想要群组，不需要订阅主题，而是自己分配分区。

```java
consumer.partitionsFor("topic");
consumer.assign(partitions);
```


# 5 深入 Kafka

### 5.1 集群成员关系
Kafka 使用 Zookeeper 维护集群成员关系，kafka 组件订阅 Zookeeper 的 /brokers/ids 路径，当有
broker 加入或者退出集群的时候，这些组件就得到通知。

### 5.2 控制器

控制器其实就是一个 broker，只不过还负责分区选举。控制器使用 epoch 避免脑裂，
脑裂指的是两个节点同时认为自己是控制器。

### 5.3 复制
- 首领副本: 每个分区都有一个首领副本，为了保持一致性，所有生产者请求和消费者请求都会经过这个副本。

- 跟随者副本: 唯一的任务就是从首领复制消息，保持和首领一致的状态。如果首领崩溃，其中一个被提升为新首领。

### 5.4 处理请求

Processor 线程 ，请求队列， 响应队列，IO 线程

请求类型：

- 元数据请求
- 生产请求
- 获取请求o

### 5.5 物理存储
Kafka 存储单元是分区。

分区分配：不同机架保证高可用

文件管理：为每个主题配置数据保留期限

文件格式：消息和偏移量保存在文件里。保存在磁盘上的格式和消息格式一样， kafka 可以通过零复制技术
给消费者发送消息，同事避免了对生产者已经压缩过的消息进行解压和再压缩。
kafka附带了一个 DumpLogSegment 的工具来查看片段内容。

索引：kafka 为每个分区维护了一个索引

清理：为每个键保留最新的值


# 6 可靠的数据传递

### 6.1 可靠性保证
保证：指的是确保系统在各种不同环境下能够发生一致的行为

kafka 可以在那些方便保证呢？

- Kafka 可以保证分区消息的顺序。同一个分区
- 只有当消息被写入分区的所有同步副本（但不一定要写入磁盘），他才被认为是已提交的。
- 只要还有一个副本是活跃的，那么已经提交的消息就不会丢失
- 消费者只能读取已经提交的信息

### 6.2 复制
kafka 复制机制和分区的多副本架构是 kafka 可靠性保证的核心。

### 6.3 broker 配置

复制系数： replication.factor。更高复制系数带来更高可用性，可靠性，更少故障。
大部分场景3够用，有些银行设置为5.

不完全的首领选举：unclean.leader.election.enalbe 只能在 broker 级别。
如果允许不同步的副本成为首领，就要承担丢失数据和不一致的风险。
如果不允许他们成为首领，就要接受较低的可用性，必须等待原先首领恢复可用。

最少同步副本: min.insync.replicas

### 在可靠系统中使用生产者

- 根据可靠性需求配置恰当的 acks 值
- 在参数配置和代码里正确处理错误

发送确认，生产者可以选择3种不同确认模式:
- acks=0. 吞吐很高，把消息发送出去就认为写入kafka。几乎一定会丢失消息
- acks=1。意味着首领在收到消息并把它写入到分区数据文件（不一定同步磁盘)时返回确认或者错误响应。
    依然可能丢失数据。（比如消息成功写入首领，但是被复制到跟随者副本之前首领崩溃）
- acks=all。意味着首领返回确认或者错误响应之前，会等待所有同步副本都收到消息。最保险也最慢

配置生产者重试参数： 可重试错误和不可重试错误。重试可能导致消息重复，做好去重或者幂等

额外的错误处理:
- 不可重试的 broker 错误，消息大小，认证错误等
- 消息发送之前的错误，比如序列化错误
- 生产者到达重试上限或者消息占用内存到达上限

# 6.5 在可靠的系统里使用消费者
已经被写入到所有同步副本的消息对消费者可用的o

消费者可靠性配置:
- group.id。
- auto.offset.reset. 指定了没有偏移量可以提交时，消费者做什么
- enable.auto.commit: 让消费者基于任务调度自动提交偏移量
- auto.commit.interval.ms : 自动提交偏移量的频率

如果想显示提交偏移量：

- 总是在处理完事件之后再提交偏移量
- 提交频率时性能和重复消息数量之间的权衡
- 确保对提交的偏移量心里有数
- 再均衡。注意处理消费者的再均衡问题
- 消费者可能要重试
- 消费者可能需要维护状态。尝试 KafkaStream
- 长时间处理。线程池
- 仅一次传递。幂等性写入

# 6.6 验证系统可靠性

配置验证: org.apache.kafka.tools包里的 VerifiableProducer VerifiableConsumer 两个类

应用程序验证:
- 客户端从服务器断开连接
- 首领选举
- 依次重启 broker
- 依次重启生产者
- 依次重启消费者

生产环境监控可靠性:消息的 error-rate, retry-rate
消费者来说最重要的指标是 consumer-lag。Burrow 是一个 consumer-lag 工具


# 7 构建数据管道

kakfa可以作为数据管道各个数据段之间的大型缓冲区，有效地解耦管道数据的生产者和消费者。

### 7.1 构建数据管道需要考虑问题

- 及时性
- 可靠性。kafka 支持至少一次，结合事务模型或者唯一键特性的外部存储来实现仅一次。
- 高吞吐量和动态吞吐量
- 数据格式
- 转换。ETL(Extract-transform-load 提取-转换-加载) ELT
- 安全。kafka 支持加密传输，认证和安全
- 故障处理能力
- 耦合和灵活性

### 7.2 在 Connect API 和 客户端 API 选择

### 7.3 Kafka Connect

连接器和任务


# 8. 跨集群数据镜像
把集群间的数据复制叫做镜像。kafka内置跨集群复制工具 MirrorMaker

### 8.1 场景

- 区域集群和中心集群。多城市数据中心
- 冗余(DR)
- 云迁移

### 8.2 多集群架构

问题：

- 高延迟
- 有限带宽
- 高成本

架构原则:

- 每个数据中心至少要一个集群
- 每两个数据中心之间的数据复制要做到每个事件仅复制一次（除非错误重试）
- 如果可能，尽量从远程数据中心读取数据，而不是向远程数据中心写入数据

Hub-Spoke 架构： 一个中心集群对应多个本地集群情况

双活架构：两个或者多个数据中心需要共享数据并且每个数据库中心都可以生产和读取数据时。(active-activ场景)

主备架构(active-Standby): 失效备援
- 数据丢失和不一致性
- 失效备援之后的起始偏移量
    - 偏移量自动重置
    - 复制偏移量主题

### 8.3 Kafaka 的 MirrorMaker


监控:

- 延迟监控: 目标集群是否落后于原集群
- 度量指标监控

### 8.4 其他跨集群同步方案

- uber UReplicator
- Confluent 的 Replicator


# 9 管理 Kafka
### 9.1 主题操作

kafka-topic.sh ，创建，修改，删除和查看集群里的主题。
kakfa大部分命令行工具直接操作 Zookeeper 元数据，不会连接到 broker上。

创建主题：三个参数，主题名字，复制系数，分区数量

```
#kafka-topic.sh --zookeeper zoo1.example.com:2181/kafka-cluster --create --topic my-topic
--replication-factor 2 --partitions 8
```

增加分区；删除主题o

### 9.2 消费者群组

消费者群组信息，旧版 kafka 存储在 Zookeeper 上，新版保存在 broker 上。

kafka-consumer-groups.sh

### 9.3 动态配置变更

kakfka-configs.sh

### 9.4 分区管理

kafka-preferred-replica-election.sh

kafka-reassign-partition.sh

kafka-run-class.sh  解码日志片段

kafka-replica-verification.sh 副本验证

### 9.5 消费和生产
有时候为了验证可以手动读取和生成消息，借助 kafka-console-consumer.sh kafka-console-producer.sh工具。 注意版本要和 kafka borker 一致。

控制台消费者：

```
# 使用旧版消费者读取单个主题 my-topic
# kafka-console-consumer.sh --zookpper
zoo1.example.com:2181/kafka-cluster -- topic my-topic
sample message 1
sample message 2
```

控制台生产者，默认将命令行输入的每一行视为一个消息，消息键值通过 tab 字符分隔，没有 tab 键是 null。发送 EOF 字符来关闭客户端。

```
# kafka-console-producer.sh --broker-list
kafka1.example.com:9092,kafka2.example.com:9092 --topic my-topic
```，

### 9.6 客户端 ACL

kafka-acls.sh 处理客户端与访问控制相关的问题。

### 9.7 不安全操作


- 移动集群控制器
- 取消分区重新分配
- 移出待删除主题
- 手动删除主题，要求在线下关闭集群里的所有 broker。集群还在运行时修改 Zookeeper 上的元数据是非常危险的。


# 10 监控 Kafka

### 10.1 度量指标基础

- 度量指标在哪里？ 通过 java management Extensions(JMX) 接口来访问
- 内部或者外部度量。网络健康监控
- 应用程序健康检测。
    - 使用外部进程报告 broker 运行状态（健康检测）
    - 在 broker 停止发送度量指标的时候发出告警（也叫做stale度量指标)
- 度量指标的覆盖面

### 10.2 borker 的度量指标
确保 kafka 的监控和报警不要依赖 kafka 本身。

- 非同步分区，从 broker 的崩溃到资源的过度消耗。使用kafka-topics.sh 可以获取非同步分区清单

- broker度量指标。
    - 活跃控制器数量
    - 请求处理器空闲率
    - 主题流入字节，如果单个 broker 接收了太多流量，需要再均衡
    - 主题流出字节。由于 kafka 多客户端支持，流出速率可能是流入的6倍
    - 主题流入的消息
    - 分区数量
    - 首领数量
    - 离线分区数量
    - 请求度量指标(一个99百分位度量表示，整组取样里有99%的值小于度量指标的值)

- 主题和分区的度量指标
    - 主题实例的度量指标
    - 分区实例的度量指标

- Java 虚拟机监控
    - 垃圾回收(GC)
    - java操作系统监控

- 操作系统监控(cpu/内存/磁盘/磁盘IO/网络) 平均负载是指等待 cpu 处理的线程数

- 日志。设置级别

### 10.3 客户端监控

生产者：
- 生产者整体度量指标(MBean)
- Per-broker 和 Per-topic 度量指标

消费者：
- Fetch Manager 度量指标
- Per-broker 和 Per-topic 度量指标
- Coordinator 度量指标

配额：
kafka 可以对客户端的请求进行限流，防止客户端拖垮整个集群。对于消费者和生产者来说都是可以配置的。
当 broker 发现客户端流量超过配额时，会暂缓想客户端返回响应，等待足够长的时间，直到客户端流量降低
到配额以下。

### 10.4 延时监控

Burrow 开源工具。如果没有其他工具，消费者客户端的 redords-lag-max 提供了部分视图。

### 10.5 端到端监控

kafka monitor


# 11 流式处理

### 11.1 什么是流式处理

数据流：数据流是无边界数据集的抽象表示。无边界意味着无限和持续增长。

事件流模型还有一些其他属性；

- 事件流是有序的
- 不可变的数据记录。事件一旦发生，不能改变
- 事件流是可重播的

编程范式：

- 请求与响应。延迟最小。一般是阻塞式。 OLTP
- 批处理：高延迟和高吞吐。数据仓库(DWH) 或者商业智能(BI)系统
- 流式处理：介于上述二者之间。持续性和非阻塞，比如网络警告，实时价格调整，包裹跟踪等

### 11.2 流式处理的一些概念

##### 11.2.1 时间

- 事件时间：事件发生和记录的创建时间
- 日志追加时间：事件保存到 broker 的时间
- 处理时间：应用程序收到事件之后要对其进行处理的时间

注意时区问题。

##### 11.2.2 状态

事件与事件之间的信息被称为状态。

- 本地状态或内部状态。只能被单个应用实例访问
- 外部状态。使用外部存储来维护，一般是 nosql，比如 Cassandra。

##### 11.2.3 流和表的二元性

##### 11.2.4 时间窗口

计算移动平均数：

- 窗口大小
- 窗口移动频率（移动间隔），若移动窗口和窗口大小相等，称为滚动窗口。如果窗口随着每一条记录移动，称为滑动窗口
- 窗口的可更新时间多长

### 11.3 流式处理的设计模式

- 单个事件处理，也叫做 map 或者 filter 模式，经常用来过滤无用事件或者转换事件

- 使用本地状态。比如计算股票价格需要保存最大最小值
    - 内存使用
    - 持久化
    - 再均衡

- 多阶段处理和重分区。多个步骤聚合

- 使用外部查找：流和表的连接。比如用户点击->查询用户信息->发送新消息。
    - CDS(change data capture): 捕捉数据库的变更形成事件流

- 流与流的连接；基于时间窗口连接。

- 乱序的事件
    - 识别
    - 规定时间重排
    - 一定时间段内重排乱序事件的能力
    - 更新结果的能力

- 重新处理

### 11.4 Streams 示例

https://github.com/gwenshap/kafka-streams-wordcount


### 11.5 Kafka Streams 架构设计

### 11.6 流式处理使用场景

如果需要快速处理事件，而不是为每个批次等几个小时，但又不是真的要求毫秒级响应，
流式处理（持续处理）就可以派上用场了。

- 客户服务
- 物联网
    - 欺诈检测


### 11.7 如何选择流式处理框架

- 摄取
- 低延迟
- 异步微服务
- 几近实时的数据分析

还有一些全局考虑点：

- 系统可操作性。部署，监控和调试，伸缩，与现有系统集成，错误处理
- API 可用性和调试简单性
- 简化。细节封装
- 摄取。
