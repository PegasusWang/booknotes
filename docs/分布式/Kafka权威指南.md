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
