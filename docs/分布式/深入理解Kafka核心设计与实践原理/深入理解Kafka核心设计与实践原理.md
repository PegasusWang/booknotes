# 1. 初识 Kafka

- 消息系统: 还提供了顺序性保证和回溯消费的功能。系统解耦，冗余存储、削峰、缓冲、异步通信、扩展性、可恢复性等
- 存储系统：持久化和多副本
- 流式处理平台：提供了完整的流式处理类库

本书代码：https://github.com/hiddenzzh/kafka_book_demo

1. leader会维护一个与其基本保持同步的Replica列表，该列表称为ISR(in-sync Replica)，每个Partition都会有一个ISR，而且是由leader动态维护
2. 如果一个follower比一个leader落后太多，或者超过一定时间未发起数据复制请求，则leader将其从ISR中移除
3. 当ISR中所有Replica都向Leader发送ACK时，leader才commit，这时候producer才能认为一个请求中的消息都commit了。

## 1.1 基本概念

Producer, Broker, Consumer。通过 ZooKeeper 管理元数据、控制器的选举等操作。一个或者多个 Broker 组成了一个kafka集群。

![kafka 结构](./1-1kafka体系结构.png)

- 主题(topic): 消息以主题为单位进行归纳，生产者发送消息到特定主题，消费者订阅主题并且消费
- 分区(partition): 一个主题只属于单个分区，同一个主题下的不同分区包含消息是不同的，在存储层可以看作一个可追加的日志文件。
消息追加的时候都会分配一个偏移量(offset)。offset是消息在分区中的唯一标识，通过它保证分区内的有序性。修改分区数量水平扩展。
分区有多副本机制(replica)，副本之间是一主多从关系，leader 负责写请求，follower同步消息。leader失败后从follower选举
leader 对外服务。

分区中所有副本统称为AR(Assigned Replicas)，所有与 leader 副本保持一定程度同步的副本（包括 leader 副本）组成 ISR (In-Sync Replicas)。
与 leader 副本同步之后过多的副本（不包括 leader副本）组成 OSR(Out-of-Sync Replicas)。 AR=ISR+OSR ，正常情况下 AR=ISR。

- HW(Hight watermark)高水位，标识一个特定的消息偏移量(offset)，消费者只能拉取到这个 offset 之后的消息。
- LEO(Log End Offset)，标识当前日志中下一条待写入消息的 offset


## 1.3 生产与消费

```sh
# 创建分区数是 4，副本因子 3 的 topic (topic-demo)
bin/kafka-topics.sh --zookeeper localhost: 2181/kafka --create --topic topic-demo --replication-factor 3 --partitions 4
# 订阅
bin/kafka console co口sumer. sh --bootstrap- server localhost:9092 --topic topic-demo
# 发送消息
bin/kafka-console-producer.sh --broker-list localhost: 9092 topic top工c-demo
```

## 1.4 服务端参数配置

配置路径 `$KAFKA_HOME/config/server.properties`


# 2. 生产者

## 2.1 客户端开发

一个正常的生产逻辑需要具备以下几个步骤：

- （1）配置生产者客户端参数及创建相应的生产者实例。
- （2）构建待发送的消息。
- （3）发送消息。
- （4）关闭生产者实例。

发送消息主要有三种模式：发后即忘（fire-and-forget）、同步（sync）及异步（async）。

- fire-and-forget: 直接调用 send()。 性能最高，可靠性最差
- sync: 调用 send 之后调用 get() 等待
- async: 传入回调 callback。Kafka 在返回响应时调用该函数来实现异步的发送确认。并且回调函数的的调用也保证分区有序

KafkaProducer中一般会 发生两种类型 的异常:可重试的异常和不可重试的异常。常见的可 重试异常有:NetworkException、LeaderNotAvailableException, UnknownTopicOrPart廿ionException、
NotEnoughReplicasException、NotCoordinatorException 等。可以配置重试次数

```java
package chapter2;

import org.apache.kafka.clients.producer.*;
import org.apache.kafka.common.serialization.StringSerializer;

import java.util.Properties;
import java.util.concurrent.TimeUnit;

public class KafkaProducerAnalysis {
    public static final String brokerList = "localhost:9092";
    public static final String topic = "topic-demo";

    public static Properties initConfig() {
        Properties props = new Properties();
        props.put("bootstrap.servers", brokerList);
        props.put("key.serializer",
                "org.apache.kafka.common.serialization.StringSerializer");
        props.put("value.serializer",
                "org.apache.kafka.common.serialization.StringSerializer");
        props.put("client.id", "producer.client.id.demo");
        return props;
    }

    public static void main(String[] args) throws InterruptedException {
        Properties props = initConfig();
        KafkaProducer<String, String> producer = new KafkaProducer<>(props);

        ProducerRecord<String, String> record = new ProducerRecord<>(topic, "hello, Kafka!");
        try {
            producer.send(record);
//            producer.send(record, new Callback() {
//                @Override
//                public void onCompletion(RecordMetadata metadata, Exception exception) {
//                        System.out.println(metadata.partition() + ":" + metadata.offset());
//                    if (exception == null) {
//                    }
//                }
//            });
        } catch (Exception e) {
            e.printStackTrace();
        }
//        TimeUnit.SECONDS.sleep(5);
    }
}
```

### 2.1.4 分区器

消息在通过send()方法发往broker 的过程中， 有可能需要经过拦截器(Interceptor)、 序列 化器(Serializer)和分区器(Part山oner)的一 系列作用之后才能被真正地发往 broker。
如果 key 不为 null，那 么默认的分区器会对 key 进行哈 希(采 用 MurmurHash2 算法 ，具备高运算性能及低碰撞率)，
最终根据得到 的哈希值来计算分区号， 拥有相同 key 的消息会被写入同一个分区 。 如果 key 为 null，那么消息将会以轮询的方式发往主题内的各个可用分区。

自定义分区器的实现：

```java
package chapter2;

import org.apache.kafka.clients.producer.Partitioner;
import org.apache.kafka.common.Cluster;
import org.apache.kafka.common.PartitionInfo;
import org.apache.kafka.common.utils.Utils;

import java.util.List;
import java.util.Map;
import java.util.concurrent.atomic.AtomicInteger;

/*
 * 自定义分区器。需要通过配置参数 partitioner. class 来显式 指定 这个分区器
 # 比如一般大型电商都有多个仓库，可以将仓库的名称或 ID作为 key来灵活地记录 商品信息。
 * props.put(ProducerConfig.PARTITIONER_CLASS_CONFIG, DemoPartitioner.class.getName());
*/
public class DemoPartitioner implements Partitioner {
    private final AtomicInteger counter = new AtomicInteger(0);

    @Override
    public int partition(String topic, Object key, byte[] keyBytes,
                         Object value, byte[] valueBytes, Cluster cluster) {
        List<PartitionInfo> partitions = cluster.partitionsForTopic(topic);
        int numPartitions = partitions.size();
        if (null == keyBytes) {
            return counter.getAndIncrement() % numPartitions;
        } else
            return Utils.toPositive(Utils.murmur2(keyBytes)) % numPartitions;
    }

    @Override
    public void close() {
    }

    @Override
    public void configure(Map<String, ?> configs) {
    }
}
```
### 生产者拦截器

生产者拦截器既可以用来在消息发送前做一些准备工作， 比如按照某个规则过滤不符合要 求的消息、修改消息的内容等，
也可以用来在发送回调逻辑前做一些定制化的需求，比如统计类工作。

## 2.2 原理分析

![](./2-1生产者客户端的整体架构.png)


## 2.3 重要生产者参数

- acks: 指定分区中必须要有多少个副本接受这条消息，生产者才被认为写成功。 acks 是生产者客户端中一个非常重要 的参数 ，它涉及消息的可靠性和吞吐 量之间的权衡。 acks参数有3种类型的值(都是字符串类型)
  - acks="1" (默认)。生产者发送消息之后，只要分区的 leader副本成功写入消息，那么它就会收到来自服务端的成功响应。
    折中方案。消息写入 leader 副本并 返回成功响应给生产者，且在被其他 follower 副本拉取之前 leader 副本崩溃，那么此 时消息还是会丢失
  - acks="0" 。生产者发送消息之后不需要等待任何服务端的响应。最大吞吐
  - acks="-1/all"。生产者在消息发送之后，需要等待 ISR 中的所有副本都成功 写入消息之后才能够收到来自服务端的成功响应。(最高可靠)
- max.request.size: 限制生产者客户端能发送的消息的最大值。默认 1M
- retries, retry.backoff.ms: 生产者重试次数和间隔。在需要保证消息顺序的场合建议把参数 max.in.flight . requests .per.connection 配置为 1
- compression.type: 指定消息压缩方式，默认 "none"，还可以配置为 "gzip", "snappy", "lz4"
- connections.max.idle.ms: 多久关闭闲置的连接
- linger.ms: 指定生产者发送 ProducerBatch 之前等待更多消息 (ProducerRecord) 加入 ProducerBatch 的时间
  - 生产者客户端会在 ProducerBatch 被填满或等待时间超过 linger .ms 值时发迭出去。增大这个参数的值会增加消息的延迟，但是同时能提升一定的吞 吐量。 这个linger.ms参数与TCP协议中的Nagle算法有异曲同工之妙。
- receive.buffer.bytes: socket 接受消息缓冲区(SO_RECBUF) 大小，默认 32kb。设置-1 则使用操作系统默认值
- send.buffer.bytes: 设置socket发送消息缓冲区(SO_SNDBUF) 大小，默认 128kb
- request.timeout.ms: 这个参数用来配置 Producer等待请求响应的最长时间，默认值为 30000 (ms)


# 3. 消费者

## 3.1 消费者和消费者组

每个消费者都有一个对应的消费者组(逻辑概念)，消息发布到topic 后，只会投递给订阅它的每个消费组中的一个消费者。
消费者组是Kafka实现单播和广播两种消息模型的手段。同一个topic，每个消费者组都可以拿到相同的全部数据。

- 如果所有的消费者都隶属于同一个消费组，那么所有的消息都会被均衡地投递给每一 个消费者，即每条消息只会被一个消费者处理，这就相当千点对点模式的应用。
- 如果所有的消费者都隶属于不同的消费组，那么所有的消息都会被广播给所有的消费者，即每条消息会被所有的消费者处理，这就相当于发布/订阅模式的应用。

相同消费者组 id 的都是同属一个消费者组，消费者内消费者数量如果多余分区数，多余的消费者会无法消费(无法分配分区)。
不同消费者组会接收到同样的消息。(广播)

![](./消费者与消费组.png)

## 3.2 客户端开发

一个正常的消费逻辑需要具备以下几个步骤：

- （1）配置消费者客户端参数及创建相应的消费者实例。
- （2）订阅主题。
- （3）拉取消息并消费。(拉取模式)
- （4）提交消费位移。
- （5）关闭消费者实例。

```java
package chapter3;

import lombok.extern.slf4j.Slf4j;
import org.apache.kafka.clients.consumer.ConsumerRecord;
import org.apache.kafka.clients.consumer.ConsumerRecords;
import org.apache.kafka.clients.consumer.KafkaConsumer;

import java.time.Duration;
import java.util.Arrays;
import java.util.Properties;
import java.util.concurrent.atomic.AtomicBoolean;

@Slf4j
public class KafkaConsumerAnalysis {
    public static final String brokerList = "localhost:9092";
    public static final String topic = "topic-demo";
    public static final String groupId = "group.demo";
    public static final AtomicBoolean isRunning = new AtomicBoolean(true);

    public static Properties initConfig() {
        Properties props = new Properties();
        props.put("key.deserializer",
                "org.apache.kafka.common.serialization.StringDeserializer");
        props.put("value.deserializer",
                "org.apache.kafka.common.serialization.StringDeserializer");
        props.put("bootstrap.servers", brokerList);   // Kafka 集 群所需的 broker 地 址清 单
        props.put("group.id", groupId); // 消费者组名称，设置成有业务意义的名字
        props.put("client.id", "consumer.client.id.demo"); // 这个参数用来 设定KafkaConsumer对应的客户端过
        return props;
    }

    public static void main(String[] args) {
        Properties props = initConfig();
        KafkaConsumer<String, String> consumer = new KafkaConsumer<>(props);
        consumer.subscribe(Arrays.asList(topic)); // 订阅topic主题。一个消费者可以订阅多个主题。支持正则

        try {
            while (isRunning.get()) {
                ConsumerRecords<String, String> records =
                        consumer.poll(Duration.ofMillis(1000)); // 拉模式。poll()方法返回的是所订阅主题(分区)上的一组消息
                for (ConsumerRecord<String, String> record : records) { // 遍历每一条消息 ConsumerRecord
                    System.out.println("topic = " + record.topic()
                            + ", partition = " + record.partition()
                            + ", offset = " + record.offset());
                    System.out.println("key = " + record.key()
                            + ", value = " + record.value());
                    //do something to process record.
                }
            }
        } catch (Exception e) {
            log.error("occur exception ", e);
        } finally {
            consumer.close();
        }
    }
}
```

### 3.2.5 位移提交

在旧消费者客户端中，消费位移是存储在 ZooKeeper 中的 。 而在新消费者客户端中，消费 位移存储在 Kafka 内部的主题consumer offsets 中 。
这里把将消费位移存储起来(持久化)的 动作称为“提交’ ，消费者在消费完消息之后需要执行消费位移的提交。
消费者使用 offset 表示消费者消费到分区中某个消息所在的位置。消费者在消费完消息之后需要执行消费位移的提交。

KafkaConsumer 类提供 了 position(TopicPartition)和 committed(TopicPartition)两个方法来分 别获取上面所说的 position 和 committed offset 的值
位移提交的 具体 时机的把握也很有讲究，有可能会造成 重复消费 和消息丢失的现象。

在 Kafka 中默认的消费位移的提交方式是**自动提交**，这个由消费者客户端参数 enable.auto.commit 配置，默认值为 true。
当然这个默认的自动提交不是每消费一条消息 就提交一次，而是定期提交，这个定期的周期时间由客户端参数 auto.commit.interval.ms 配置，默认值为 5 秒，
此参数生效的前提是 enable.auto.commit 参数为 true。正常情况下自动提交不会发生消息丢失或者重复消费现象。

默认有自动位移提交 enable.auto.commit 为 true。如果设置为 false 可以手动提交。让程序员在逻辑合适的地方进行位移提交。

- commitSync 同步提交
- commitAsync 异步提交。执行的时候消费者线程不会被阻塞 ，可能在提交消费位移的结果还未返回之前就开始了新一次的拉取操作

### 3.2.6 控制或关闭消费

使用 pause()和 resume()方法来分别实现暂停某些分区在拉取操作时返回数据给客户端和恢复某些分区向客户端返回数据的操作。
跳出循环以后一定要显式地执行关闭动作以释放运行过程中占用的各种系统资源，包括内 存资源、 Socket 连接等。 KafkaConsumer 提供了 close()方法来实现关闭。

### 3.2.7 指定位移消费
在 Kafka 中 每当消费者查找不到所记录的消费位移 时， 就会根据消费者客户端参数 auto.offset.reset 的配置来决定从何处开始进行消费。
选项有 "lastest", "earliest", "none"(抛异常)。

seek 方法可以从特定位置读取消息，可以实现追前消费或者回溯消费。也可将消费位移保存在数据库活 redis中， 在下次消费的时候可以读取存储在数据表中
的消费位移并通过seek()方法指向这个具体的位置。

```java
// seek()方法中的参数 partit工on 表示分区，而 offset 参数用来指定从分区的哪个位置 开始消费。
public void seek(TopicPartit工on part工tion , long offset)
```

### 3.2.8 再均衡

再均衡是指分区的所属权从一个消费者转移到另一消费者的行为，它为消费组具备高可用性和伸缩性提供保障，
使我们可以既方便又安全地删除消费组内的消费者或往消费组内添加消费者。不过在再均衡发生期间， 消费组内的消费者是无法读取消息的。

- `void onPartitionsRevoked(Collection<TopicPartition> partitions)`
  - 这个方法会在再均衡开始之前和消费者停止读取消息之后被调用。 可以通过这个回调方法 来处理消费位移的提交，以此来避免 一 些不必要的重复消费现象的发生。 参数 partitions 表 示再均衡前所分配到的分区。
- `void onPartitionsRevoked(Collection<TopicPartition> partitions)`
  - 这个方法会在重新分配分区之后和消费者开始读取消费之前被调用。参数 partitions 表 示再均衡后所分配到的分区。

### 3.2.9 消费者拦截器

消费者拦 截器主要在消费到消息或在提交消费位移时进行一些定制化的操作。比如可以判断消息时间和当前时间差，过滤过时消息。

### 3.2.10 多线程实现

KatkaProd ucer是线程安全的， 然而KafkaConsumer却是非线程安全的。 KafkaConsumer中 定义了 一 个acquire()方法，
用来检测当前是否只有一个线程在操作， 若有其他线程正在操作则会抛出异常。
KafkaConsumer中的每个公用方法在执行所要执行的动作之前都会调用这个acquire()方法， 只有wakeup()方法是个例外。
它仅通过线程操作计数标记的方式来检测线程是否发生了并 发操作，以此保证只有一 个线程在操作。acquire/release方法成对出现，表示相应的加 锁和解锁操作。

通过多线程提高消费能力：
- 线程封闭。为每一个线程实例化一个 KafkaConsumer 对象(消费线程)。
  - 一个线程对应一 个Kafka Consumer实例， 我们可以称之为消费线程。 一 个消费线程可以消 费一个或多个分区中的消息， 所有的消费线程都隶属于同一个消费组
  - 这种实现并发度取决于分区数，多余的消费者线程空闲
- 第二种方式是多个消费线程同时消费同一 个分区， 这个通过assign()、seek()等 方法实现， 这样可以打破原有的消费线程的个数不能超过分区数的限制， 进一步提高了消费的 能力。
  - 实现复杂，容易出错(不推荐)
- 将处理模块改成多线程方式（瓶颈）。不过顺序处理就比较难了，可以通过加锁的共享变量处理

### 3.2.11 重要消费者参数

- fetch.min.bytes:该参数用来配置 Consumer 在一次拉取请求(调用 poll()方法)中能从 Kafka 中拉取的最小 数据量，默认值为 1 (B) 。
- fetch.max.bytes:该参数与 fetch .max .bytes 参数对应，它用来配置 Consumer在一次拉取请求中从 Kafka 中拉取的最大数据量，默认值50MB
- max.poll.records 这个参数用来配置 Consumer 在 一 次拉取请求中拉取的最大消息数，默认值为 500 (条)。如果消息的大小都比较小，则可以适当调大这个参数值来提升一定的消费速度。


# 4. 主题和分区

## 4.1 主题管理

create,list,describe,alter,delete。可以通过 kafka-topics.sh --help 查看帮助

```sh
# 创建主题
kafka-topics.sh --zookeeper localhost:2181/kafka--create --topic topic-create --partitions 4 --replication-factor 2
```
创建主题时， 实质上是在 ZoKo eeper中的 /brokers/topics节点下创建与 该主题对应的子节点并写入分区副本分配方案，
并且在 /conig/topics/节点下创建与 该主题对应的子节点并写入主题 相关的配置信息(这个步骤 可以省略不执行)


## 4.2 初识KafkaAdminClient
通过 API 方式提供管理功能

```java
package chapter4;

import org.apache.kafka.clients.admin.*;
import org.apache.kafka.common.config.ConfigResource;

import java.util.Collections;
import java.util.HashMap;
import java.util.Map;
import java.util.Properties;
import java.util.concurrent.ExecutionException;

public class KafkaAdminConfigOperation {

    public static void main(String[] args) throws ExecutionException, InterruptedException {
        describeTopicConfig();
    }

    //Config(entries=[ConfigEntry(name=compression.type, value=producer, source=DEFAULT_CONFIG, isSensitive=false, isReadOnly=false, synonyms=[]), ConfigEntry(name=leader.replication.throttled.replicas, value=, source=DEFAULT_CONFIG, isSensitive=false, isReadOnly=false, synonyms=[]), ConfigEntry(name=message.downconversion.enable, value=true, source=DEFAULT_CONFIG, isSensitive=false, isReadOnly=false, synonyms=[]), ConfigEntry(name=min.insync.replicas, value=1, source=DEFAULT_CONFIG, isSensitive=false, isReadOnly=false, synonyms=[]), ConfigEntry(name=segment.jitter.ms, value=0, source=DEFAULT_CONFIG, isSensitive=false, isReadOnly=false, synonyms=[]), ConfigEntry(name=cleanup.policy, value=delete, source=DEFAULT_CONFIG, isSensitive=false, isReadOnly=false, synonyms=[]), ConfigEntry(name=flush.ms, value=9223372036854775807, source=DEFAULT_CONFIG, isSensitive=false, isReadOnly=false, synonyms=[]), ConfigEntry(name=follower.replication.throttled.replicas, value=, source=DEFAULT_CONFIG, isSensitive=false, isReadOnly=false, synonyms=[]), ConfigEntry(name=segment.bytes, value=1073741824, source=STATIC_BROKER_CONFIG, isSensitive=false, isReadOnly=false, synonyms=[]), ConfigEntry(name=retention.ms, value=604800000, source=DEFAULT_CONFIG, isSensitive=false, isReadOnly=false, synonyms=[]), ConfigEntry(name=flush.messages, value=9223372036854775807, source=DEFAULT_CONFIG, isSensitive=false, isReadOnly=false, synonyms=[]), ConfigEntry(name=message.format.version, value=2.0-IV1, source=DEFAULT_CONFIG, isSensitive=false, isReadOnly=false, synonyms=[]), ConfigEntry(name=file.delete.delay.ms, value=60000, source=DEFAULT_CONFIG, isSensitive=false, isReadOnly=false, synonyms=[]), ConfigEntry(name=max.message.bytes, value=1000012, source=DEFAULT_CONFIG, isSensitive=false, isReadOnly=false, synonyms=[]), ConfigEntry(name=min.compaction.lag.ms, value=0, source=DEFAULT_CONFIG, isSensitive=false, isReadOnly=false, synonyms=[]), ConfigEntry(name=message.timestamp.type, value=CreateTime, source=DEFAULT_CONFIG, isSensitive=false, isReadOnly=false, synonyms=[]), ConfigEntry(name=preallocate, value=false, source=DEFAULT_CONFIG, isSensitive=false, isReadOnly=false, synonyms=[]), ConfigEntry(name=min.cleanable.dirty.ratio, value=0.5, source=DEFAULT_CONFIG, isSensitive=false, isReadOnly=false, synonyms=[]), ConfigEntry(name=index.interval.bytes, value=4096, source=DEFAULT_CONFIG, isSensitive=false, isReadOnly=false, synonyms=[]), ConfigEntry(name=unclean.leader.election.enable, value=false, source=DEFAULT_CONFIG, isSensitive=false, isReadOnly=false, synonyms=[]), ConfigEntry(name=retention.bytes, value=-1, source=DEFAULT_CONFIG, isSensitive=false, isReadOnly=false, synonyms=[]), ConfigEntry(name=delete.retention.ms, value=86400000, source=DEFAULT_CONFIG, isSensitive=false, isReadOnly=false, synonyms=[]), ConfigEntry(name=segment.ms, value=604800000, source=DEFAULT_CONFIG, isSensitive=false, isReadOnly=false, synonyms=[]), ConfigEntry(name=message.timestamp.difference.max.ms, value=9223372036854775807, source=DEFAULT_CONFIG, isSensitive=false, isReadOnly=false, synonyms=[]), ConfigEntry(name=segment.index.bytes, value=10485760, source=DEFAULT_CONFIG, isSensitive=false, isReadOnly=false, synonyms=[])])
    public static void describeTopicConfig() throws ExecutionException,
            InterruptedException {
        String brokerList =  "localhost:9092";
        String topic = "topic-admin";

        Properties props = new Properties();
        props.put(AdminClientConfig.BOOTSTRAP_SERVERS_CONFIG, brokerList);
        props.put(AdminClientConfig.REQUEST_TIMEOUT_MS_CONFIG, 30000);
        AdminClient client = AdminClient.create(props);

        ConfigResource resource =
                new ConfigResource(ConfigResource.Type.TOPIC, topic);
        DescribeConfigsResult result =
                client.describeConfigs(Collections.singleton(resource));
        Config config = result.all().get().get(resource);
        System.out.println(config);
        client.close();
    }
}
```

## 4.3 分区的管理

## 4.3.1 优先副本的选举
只有 leader 副本对外提供读写服务，flollwer 副本只负责内部进行消息的同步。
为了能够有效地治理负载失衡的情况，Kafka引入了优先副本(preferred replica)的概念。所谓的优先副本 是指在 AR 集合列表中的第 一个副本

所谓的优先副本的选举是指通过一 定的方式促使优先副本选举为 leader副本， 以此来促进 集群的负载均衡， 这一行为也可以称为 “ 分区平衡 ” 。
不推荐自动执行(auto.leader.rebalance.enable 设置成 true)。

Kafka中kafka-perferred-replica-election.sh脚本提供了对分区leader副本进行重新平衡的功能。优先副本的选举过程是一个安全的过程，Kafka客户端 可以自动感知分区leader副本的变更。
在实际生产环境中， 一 般使用path-to-json-file参数来分批、 手动地执行优先副本 的选举操作。 尤其是在应对大规模的Kafka集群时， 理应杜绝采用非path-to-json-file
参数的选举操作方式。 同时，优先副本的选举操作也要注意避开业务高峰期， 以免带来性能方 面的负面影响。

### 4.3.2 分区重分配
Kafka提供了kafka-reassign-partitions.sh脚本来执行分区重分配的工作， 它可以在集群扩容、broker 节点失效时对分区进行迁移。
首先创建需要 一 个包含主题清单的 JSON文件， 其次根据主题清单和broker 节点清单生成一份重分配方案， 最后根据这份方案执行具体的重分配动作。

### 4.3.3 复制限流
分区重分配本质在于数据复制，先增加新的副本，然后进行数据同步， 最后删除旧的副本来达到最终的目的。
如果集群中某个主题或某个分区的流量在某段时间内特别大， 那么只靠减小粒度是不足以应对的， 这时就需要有一个限流的机制，
可以对副本间的复制流量加以限制来保证重分配期间整体服务不会受太大的影响 。

副本间的复制限流有两种实现方式:kafka-config.sh脚本和kafka-reassign-partitions.sh(配合 throttle参数)脚本。

### 4.3.4 修改副本因子

修改副本因子的功能也是通过重分配所使用的kafka-reassign-partition.sh脚本实现的，可以增加或者减少副本。(注意分区不能减少)

## 4.4 选择合适分区数

### 4.4.1 性能测试工具

- kafka-producer-perf-test.sh
- kafka-consumer-perf-test.sh

### 4.4.2 分区越多吞吐越高么？
一旦分区数目超过某个阈值之后，相应的吞吐量也会下降。建议用同样的机器配置做一个吞吐量测试。
一般情况下，根据预估的吞吐量及是否与key相关的规则来设定分区数即可，后期可以通过增加分区数、增加broker或分区重分配等手段来进行改进。
如果一定要给一个准则，则建议将分区数设定为集群中broker的倍数，即假定集群中有3个broker节点，可以设定分区数为3、6、9等，至于倍数的选定可以参考预估的吞吐量。

### 4.4.3 分区数的上限
比如使用 `ulimit -n 65535` 命令将上限提高到 65535, 这样足以应对大多数的应用情况， 再高也完全没有必要了。
或者修改 limits.conf 文件（针对所有用户而且在任何 shell 都生效），ulimit 命令只是针对当前 shell 的设定。
也可以通过在 `/etc/profile` 文件中增加 ulimit 的设置语句全局生效。

### 4.4.4 考量因素
有些应用场景会要求主题中的消息都能保证顺序性， 这种情况下在创建主题时可以设定分 区数为I, 通过分区有序性的这一特性来达到主题有序性的目的。
如果一定要 给一个准则，则建议将分区数设定为集群中 broker 的倍数，即假定集群中有 3 个 broker 节点， 可以设定分区数为 3、 6、 9 等， 至于倍数的选定可以参考预估的吞吐量。


# 5. 日志存储

## 5.1 文件目录布局

为了防止Log过大，kafka又引入了日志分段(Log Segment)的概念，将Log切分为多个Log Segment, 相当于一个巨型文件被平均分配为多个相对较小的文件，
这样也便于消息的维护和清理。Log在物理上只以文件夹的形式存储，而每个 LogSegment对应于磁盘上的一个日志文件 和两个索引文件， 以及可能的其他文件(比如以
".txnindex"为后缀的事务索引文件)
只有最后一个 LogSegment 才能执行写入操作(我们称之为activeSegment)。每个LogSegment中的日志文件 (以 " .log"为文件后缀)都有对应的两个索引文件:
偏移量索引文件(以".index"为文件后缀)和时间戳索引文件(以".timeindex"为文件后缀)。

![](./5-1日志关系.png)

## 5.2 日志格式演变

![](./5-3日志v0版本.png)

v1 就比 v0 多了时间戳

![](./5-4日志v1版本.png)

kafka 实现的压缩方式是将多条一起压缩，在 broker 也是保持压缩状态，在端到端之间做压缩和解压。
配置 compression.type 配置压缩方式或者不压缩。

![](./5-7消息v2.png)

 v2版本又节省了很多空间，因为它将多个消息(Record) 打包存放到单个 RecordBatch 中，又通过Varints编码极大地节省了空间。

## 5.3 日志索引

- 偏移量索引文件用来建立消息偏移量（offset）到物理地址之间的映射关系，方便快速定位消息所在的物理文件位置；
  - | 相对偏移量(relativeOffset) | position |
- 时间戳索引文件则根据指定的时间戳（timestamp）来查找对应的偏移量信息。
  - | 当前日志分段最大时间戳(timestamp) | relativeOffset |

## 5.4 日志清理

kafka 提供了两种日志清理策略：

- 日志删除(log retention): 按照一定的保留策略直接删除不符合条件的日志分段
- 日志压缩(log compaction): 针对每个消息的 key 进行整合，对于相同 key 的不同 value 值，只保留最后一个版本(合并)

### 5.4.1 日志删除

在Kafka 的日志管理器中会有一个专门的日志删除任务来周期性地检测和删除不符合 保留条 件的日志分段文件，
这个周期可以通过broker端参数log.retention.check.interval.ms来配置 ，默认值为300000, 即5分钟。当前日志分段的保留策略有3 种:

- 基于时间的保留策略。默认情况下只配置 了log.re七ention.hours参数， 其值为 168, 故默认情况下日志分段文件的保留时间为7天。
- 基于日志大小的保留策略。日志删除任务会检查当前日志的大小是否超过设定的阀值 (retentionSize)来寻找可删除的 日志分段的文件集合(deletableSegments)
- 基于日志起始偏移量的保留策略。判断依据是某日志分段的下一个日志分段的起始偏移量 baseOffset 是否小于等于 logStartOffset

### 5.4.2 日志压缩 (log compaction)
Log Compaction对千有相同key的不同value值， 只保 留最后 一 个版本。如果应用只关心key对应的最新value值，则可以开启Kafka的日志清理功能，
Kafka会定期将相同key的消息进行合并， 只保留最新的value值。

## 5.5 磁盘存储
顺序写磁盘速度很快，甚至比 随机 写内存快。kafka 使用顺序追加、页缓存、零拷贝等技术提升性能。

### 5.5.1 页缓存
页缓存是操作系统实现的一种主要的磁盘缓存，以此用来减少对磁盘I/0的操作。 具体来说，就是把磁盘中的数据缓存到内存中，
把对磁盘的访间变为对内存的访问。Kafka中大量使用了页缓存， 这是Kafka实现高吞吐的重要因素之一。

### 5.5.2 磁盘IO流程
![](./5-22磁盘IO流程.png)

linux 系统中的 IO 调度策略有 4 种：
- NOOP(No Operation):简单的 FIFO 队列
- CFQ(Completely Fair Queuing): 按照 IO 请求的地址进行排序。默认的算法，对于通用服务器最好的选择(可能有饿死现象)
- DEADLINE: CFQ基础上解决饿死极端情况。额外为读IO 和写 IO 提供了 FIFO 队列
- ANTICIPATORY: 在 DEADL即E 的基础上，为每个读 I/O 都设置了 6ms 的等待时间窗口 。

建议根据不同的业务需求来测试并选择合适的磁盘调度算法。(文件系统推荐 EXT4 或 XFS)

### 5.5.3 零拷贝
零拷贝(zero-copy): 所谓的零拷贝是指将数据直接从磁盘文件复制到网卡设备中，而不需要经由应用程序手 。零拷贝大大提高了应用程序的性能，
减少了内核和用户模式之间的上下文切换 。 对 Linux 操作系统而言，零拷贝技术依赖于底层的 sendfile()方法实现 。 对应于 Java 语言，
FileChannal.transferTo()方法的底层实现就是 sendfile()方法。

考虑把文件展示给用户，抽象成 read, write 过程。先从磁盘复制到内存中，然后通过 socket 传输给用户，中间经历过 4 次复制：

- (1)调用read()时， 文件A中的内容被复制到了内核模式下的Read Buffer中。
- (2)CPU控制将内核模式数据复制到用户模式下。
- (3)调用writ(e )时， 将用户模式下的内容复制到内核模式下的Sockte Buffre 中。
- (4)将内核模式下的Sockte Buffer的数据复制到网卡设备中传送

![](./5-23非零拷贝技术.png)

零拷贝技术可以直接请求内核把磁盘中的数据传输给 socket。零拷贝技术通过DMA (DirectMemoryAccess) 技术将文件内容复制到内核模式下的Read Buffer 中。
不过没有数据被复制到 Socket Buffer, 相反只有包含数据的位置和长度的信息的文 件描述符被加到 Socket Buffer 中。 DMA 引擎直接将数据从内核模式中传递到网卡设备(协议引擎)。

![](./5-24零拷贝技术.png)


# 6. 深入服务端

## 6.1 协议设计
kafka 自定义了一组基于 tcp 的二进制协议，遵守协议就可以发消息给 kafka。每种协议类型由对应的请求和响应。
Request 包含请求头和请求体。协议请求头中包含4个域(和eld) : api_key、 api_version、 correlation—id和client id。

## 6.2 时间轮 (TimingWheel)

基于时间轮自定义实现了一个用于延时功能的定时器(SystemTimer)，插入和删除O(1)。用于实现延时操作。
Kafka 中的时间轮(TimingWheel)是一个存储定时任务的环形队列，底层采用数组实现，数组中的每个元素可以存放一个
定时任务列表(TimerTaskList)，TimerTaskList是一个环形的双向链表，链表每一项都是定时任务项(TimerTaskEntry)，
其中封装了真正的定时任务（TimerTask)。

![](./6-7时间轮.png)

Kafka 为此引入了层级时间轮的概念， 当任务的到期时间超过了当前时间轮所表示的时间范围时， 就会尝试添加到上层时间轮中。
第二层时间轮同样也有自己的指针, 每当第一层时间轮走完一个周期，第二层时间轮的指针就会推进一格。

![](./6-8多层时间轮.png)

Kafka 中的 TimingWheel专门用来执行插入和删除 TimerTaskEntry 的操作， 而 DelayQueue专门负责时间推进的任务。
由于queue是一个延迟队列，如果队列中的expireTime没有到达，该操作会阻塞住，直到expireTime到达

```
private[timer] class TimingWheel(tickMs: Long, wheelSize: Int, startMs: Long, taskCounter: AtomicInteger, queue: DelayQueue[TimerTaskList]) {
  private[this] val interval = tickMs * wheelSize
  private[this] val buckets = Array.tabulate[TimerTaskList](wheelSize) { _ => new TimerTaskList(taskCounter) }
  private[this] var currentTime = startMs - (startMs % tickMs) // rounding down to multiple of tickMs
}
```

参考：

- [「动图」 秒懂 Kafka 时间轮（TimingWheel）](https://mp.weixin.qq.com/s/WC_XIFllMKPFjm2Vrg8JJw)
- [Kafka 时间轮(TimingWheel)](https://blog.csdn.net/uxiAD7442KMy1X86DtM3/article/details/109233656)

## 6.3 延时操作

在将 消息写入 leader 副本的本地日志文件之后，Kafka会创建一 个延时的生产操作(DelayedProduce),用来处理消息正常写入所有副本或超时的清况， 
以返回相应的响应结果给客户端。

## 6.4 控制器
kafka集群中会有一个或者多个 broker，其中一个 broker 会被选举为控制器(kafka controller)，负责管理整个集群中所有分区和副本的状态。

### 6.4.1 控制器的选举及异常恢复
Kafka中的控制器选举工作依赖于ZooKeeper, 成功竞选为控制器的broker会在ZooKeeper中创建/controller 这个临时(EPHEMERAL)节点，

### 6.4.2 优雅关闭
合理关闭： kafka-server-stop.sh 修改脚本。
`kill -s TERM $PIDS` 或者 `kill -15 $PIDS` 的方式来关闭进程。

### 6.4.3 分区 leader 的选举
选举策略：OftlinePartitionLeaderElectionStrategy, ReassignPartitionLeaderElectionStrategy,  ControlledShutdownPartitionLeaderElectionStrategy

## 6.5 参数解密
broker端没有显式配置 listeners (或 advertised. listeners)使用 IP地址， 那么最好将 bootstrap.server 配置成主机名而不要使用 IP 地址，
因为 Kafka 内部使用的是 全称域名(FullyQualifiedDomainName)。

- broker.id: 在 Kafka 集群 中 ，每个 broker 都 有唯一的 id (也可以记作 brokerld)值用来区分彼此。 broker 在启动时会在 ZooKeeper 中的 /brokers/ids 路径下创建一个以当前 brokerId为名称的虚节点， broker 的健康状态检查就依 赖于此虚节点
- bootstrap.servers: 这个参数配置的是用来发现 Kafka 集群元数据信息的服务地址


# 7. 深入客户端

### 7.1 分区分配策略

partition.assignment.strategy 设置消费者和订阅主题之间的分区分配策略:

- RangeAssignor: 按照消费者总数和分区总数进行整除运算来获得一个跨度，然后将分区按照跨度平均分配
- RoundRobinAssignor: 将消费组内所有消费者和消费者订阅的所有主题的分区按照字典序排序，轮询将分区依次分配给每个消费者
- StickyAssignor: 目的分区尽可能均匀；尽可能和上次分配一致

### 7.2 消费者协调器和组协调器

GroupCoordinator 是 kafka 服务端用于管理消费组的组件，消费者客户端中的 ConsumerCoordinator 组件负责和 GroupCoordinator 交互。
如果消费者发生变化触发再均衡操作。

### 7.3 __consumer_offsets 剖析

位移提交最终会保存到 kafka 内部主题 __consumer_offsets 中。
使用 kafka-console-consumer.sh 查看  __consumer_offsets 中的内容。

### 7.4 事务

消息传输保障有 3 个层级：

- at most once(至多一次)：消息可能丢失，但是绝对不会重复传输
- at least once(最少一次)：消息绝不会丢失，但是可能重复传输
- exactly once(恰好一次)：每条消息肯定会被传输一次且仅传输一次

kafka 从0.11.0.0 版本引入了幂等和事务这两个特性，一次实现 EOS(exactly once semantics)。

幂等：多次调用的结果和调用一次一致。只需要设置客户端参数 `properties.put("enable.idempotence", true);`
kafka 为了实现生产者幂等，引入了 producer id 和序列号 sequence number 两个概念。
broker 会在内存中为每一对 <PID, 分区> 维护一个序列号，只有消息序列号的值(SN_new)比 broker 中维护的 SN_old 大 1，
broker 才会接受。(SN_new=SN_old+1)。

kafka 幂等只能保证单个生产者会话 (session) 中单分区的幂等。

事务可以保证多个分区写入操作的原子性。通过客户端参数显示设置 `properties.put("transactional.id", "transactionId")`，
同时也要打开幂等特性。


# 8. 可靠性探究

### 8.1 副本剖析

当 ISR 集合中的一个 follower 副本滞后 leader 副本的时间超过 replica.lag.time.max.ms 指定的值判定为同步失败。

### 8.2 日志同步机制

kafka 使用的更像是微软的 PacificA 算法。

### 8.3 可靠性分析

- 副本数：一般设置副本数为 3 可以满足大部分场景对可靠性的要求，国内部分银行会设置副本数为 5 提升可靠性。
- 客户端 acks 设置。如果 acks=-1 leader 副本在写入本地日志之后还要等待 ISR 中的 follower 副本全部同步完成才告知生产者成功提交
- 设置同步刷盘策略（一般应该由多副本保证），broker 参数 log.flush.interval.messages 和 log.flush.interval.ms
  调整同步刷盘策略，不过会比较损耗性能。
- 开启 enable.auto.commit 自动位移提交功能可能导致 重复消费和消息丢失的问题。


# 9. Kafka 应用

### 9.1 命令行工具

位于 $KAFKA_HOME/bin 目录下的命令行工具

- kafka.consumer-groups.sh 查看或者变更消费组信息
- kafka-consumer.groups.sh 重置消费者组内消费位移
- kafka-delete-records.sh 删除指定位置前的消息

### 9.2 Kafka Connect

Kafka Connect 为在 kafka 和外部数据库存储系统之间移动数据提供了一种可靠的可伸缩的实现方式。
包含两个核心概念：Source 负责导入数据到 kafka，Sink 负责从 kafka 导出数据。

connect-standalone.sh 实现独立模式运行 kafka connect。

### 9.3 Kafka Mirror Maker

用于两个集群之间同步数据的一个工具。kafka-mirror-maker.sh

### 9.4 Kafka Streams

高吞吐、高可用、低延时让 kafka 成为流式处理系统中完美的数据来源。
Kafka Streams 是一个用于处理和分析数据的客户端库。它先把存储在 kafka 中的数据进行处理和分析，然后将最终所得到的的数据
结果回写到 kafka 或发送到外部系统。


# 10. Kafka 监控

监控维度：集群信息、broker 信息、主题信息和消费组信息。一般还好配置告警模块

### 10.1 监控数据来源

集群层面的指标可以用 JMX (Java Managment Extension， Java 管理扩展)来获取。

### 10.2 消费滞后

消息中间件中留存的消息与消费的消息之间的差值就是消息堆积量，也成为消费之后（Lag）量。

### 10.3 同步失效分区

处于同步失效或功能失效的副本统称为失效副本，包含它的分区就是同步失效分区。

### 10.4 监控指标说明

### 10.5 监控模块

监控架构主要分为数据采集、数据存储、数据展示三个部分。


# 11. 高级应用

### 11.1 过期时间（TTL)

TTL(Time To Live，过期时间)。通过在消费者客户端拦截器实现

### 11.2 延时队列

延时消息是消费者等待特定时间之后，才能获取这个消息进行消费。如下场景：

- 订单系统，客户下单之后有 30 分钟进行支付，如果没有成功支付，对这个订单进行异常处理，就可以用延时队列处理这些订单
- 订单完成之后 1 小时通知用户进行评价
- 遥控智能设备，指定时间之后执行操作

可行的方案：发送延时消息的时候先投递到一个 kafka 内部主题(比如delay_topic)中，然后通过一个自定义的服务拉取这些内部主题
中的消息，并将满足条件的消息投递到真实的主题中，消费者订阅的还是真实的主题。

### 11.3 死信队列和重试队列

由于某些原因消息无法被正确投递，为了确保消息不会无故丢弃，一般放到一个特殊的队列称为死信队列。
后续可以通过消费这个死信队列中的内容来分析当时遇到的异常情况，改善和优化系统。

### 11.4 消息路由

kafka 默认用主题进行路由，但是如果还想进一步消息路由，需要自己实现。
可以在消息headers中加入一个键为 "routingkey"、值为特定业务表示的 Header，然后消费端中使用拦截器挑选特定业务标识的消息。

### 11.5 消息轨迹

通过封装客户端，保证正常生产消费的同时添加相应的轨迹信息埋点逻辑。

### 11.6 消息审计

消息生产、存储和消费的整个过程之间对消息个数和延迟的审计，以此检测是否有消息丢失、是否数据重复、端到端延迟等内容。

### 11.7 消息代理

Kafka Rest Proxy 可以不用客户端的情况下实现发送消息、消费消息、查看集群状态、执行管理操作等功能。

### 11.8 消息中间件选型

RabbitMQ/Kafka/RocketMQ

选型维度：

- 功能维度
  - 优先级队列
  - 延时队列
  - 重试队列
  - 死信队列
  - 消费模式: push/pull
  - 广播消费: 点对点(p2p)，还是发布订阅(Pub/Sub)
  - 回溯消费
  - 消息堆积+持久化
  - 消息轨迹
  - 消息审计
  - 消息过滤
  - 多租户。
  - 多协议支持
  - 跨语言支持
  - 流量控制：调整发送速度。通常的流控方法有 stop-and-wait，滑动窗口、令牌桶等
  - 消息顺序性
  - 安全机制
  - 消息幂等性。消息语义：至多一次；至少一次；精确一次
  - 事务性消息：要么发送成功，要么失败
- 性能维度
- 可靠性和可用性。金融支付领域 RabbitMQ 多，日志处理、大数据 kafka 多。
- 运维管理。申请、审核、监控、告警、管理、容灾、部署等
- 社区和生态


# 12. Kafka 和 Spark 集成
