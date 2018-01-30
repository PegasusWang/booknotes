# 10. Batch Processing

先给不同的系统分类：

-   Services(online systems): 等待客户端请求到来尽快处理后响应给客户端。响应时间很重要
-   Batch processing systems(offline systems): 处理大量输入数据通过跑任务（job）处理，通常衡量指标是吞吐量（处理特定输入的时间）
-   Stream processing systems(near-real-time systems): 在事件发生后不久后处理，介于在线和离线之间

2004 年 google 提出Map-Reduce 批处理算法，被后续狠多系统比如 Hadoop，CouchDB，MongoDB 实现。

### Batch Processing with Unix Tools

考虑处理 nginx 请求日志

##### Simple Log Analysis

比如你想找到5个访问最多的页面： cat access.log | awk '{print $7}' | sort | uniq -c | sort -r -n | head -n 5

##### Chain of commands versus custom program

下面这段 ruby 代码做了同样的事情

```ruby
counts = Hash.new(0)
File.open('/var/log/nginx/access.log') do |file|
    file.each do |line|
        url = line.split[6]
    counts[url] += 1
    end
end
top5 = counts.map{|url, count| [count, url] }.sort.reverse[0...5]
top5.each{|count, url| puts "#{count} #{url}" }
```

##### Sorting versus in-memory aggregation

ruby 代码里保存了一个 url 的 hash 表，unix pipeline 示例没有 hash 表但是要求数据是预先排序的，那种更好呢？
如果数据量级比较小放在内存没问题， pipeline 工具能处理远大于内存的数据。数据可以被分片排序后写入到段文件，然后每个排序好的段可以被合并成一个大的排序文件。

### The Unix Philosophy

unix 几十年前的设计哲学至今一般被很多系统借鉴。does one thing well;
管道思想等。通过管道，我们能把不同人写的程序灵活组合起来实现非常强大的功能。

##### A uniform interface

如果你想让一个程序的输出成为另一个的输入，意味着这些程序需要使用相同的数据格式，换言之，兼容的接口（相同的输入/输出接口）
在 unix 中这种接口就是文件（准备说应该是文件描述符），文件只是一系列有序字节。通过这个简单的接口，
很多不同的事物可以使用相同的接口来描述:文件系统中一个文件，进程间通信管道（unix
socket，stdin，stdout），设备文件(/dev/audi) ，表示 TCP 连接的 socket等等。

##### Separation of logic and wiring

另一个 unix 工具特性是它们都是用标准输入(stdin)和标准输出(stdout)，和程序本身逻辑分离，方便重定向到任何其他地方。

但是 unix 工具有个很大的缺陷，就是只能运行在一个机器上，所以 Hadoop 横空出世。

### MapReduce and Distributed Filesystems

unix 工具使用 stdin 和 stdout 作为输入和输出，MapReduce jobs 读写文件到分布式文件系统。Hadoop 实现中使用的文件系统叫做
HDFS(Hadoop Distributed System)，作为 Google File System 的开源实现。
HDFS 基于 shared-nothing 原则，不同于使用 shared-disk 的 Net work Attached
Storage(NAS)架构。它不需要特殊的硬件，只需要设备能够连上数据中心的网络。HDFS 由每个机器上的 daemon
进程组成，暴露网络服务让其他机器上的节点访问，由叫做 NameNode 的服务器来跟踪哪个文件块落到了哪个节点上。

##### MapReduce Job Execution

MapReduce 是一种可以在分布式文件系统比如 HDFS 上处理大量数据集的编程框架。我们回到日志分析的例子，和 Mapreduce 很像：

-   读取输入文件然后分割成记录（records），处理 nginx log 的例子里每条记录就是一行 log（换行符分隔）
-   调用 mapper 函数从每个输入记录里提取出 key 和 value。之前的例子就是 `awk '{print $7}'`：提取出  url 作为 key，value 是空值。
-   根据 key 排序所有的 key 和 value 对。日志例子就是用 sort 命令
-   调用 reducer 函数迭代有序的 key-value 对，过程中可以高效处理相邻的相同 key。日志例子里就是 uniq -c 命令

上边四步 可以被一个MapReduce job 处理，步骤2(map)和步骤4(reduce)就是需要写自定义数据处理代码的地方。为了创建一个
MapRecue 任务，你需要实现两个回调函数：maper 和 reducer：

-   Mapper：mapper 对于每个输入记录都会调用，目的是从记录中提取 key 和 value，每个记录的处理都是独立的
-   Reducer：MapReduce框架会把 mapper 处理已经得到的很多 key-value 对，按照相同 key 收集起来，，通过迭代器调用
    reducer收集这些值。reducer 可以输出记录

##### Distributed execution of MapReduce

unix 命令行和 MapReduce 最大的区别就是后者可以并行地在多个机器上执行，而且不用显示写代码处理。在 Hadoop MapReduce
实现中，mapper 和 reducer 就是实现了特定接口的类，Mongodb 和 CouchDB 中是 js 函数。

##### MapReduce workflows

有时候一个 job 完成不了需求，比如我们能通过一个 job 得到每个页面的浏览，但是无法获取最受欢迎的页面，需要二轮排序。
一个通用的做法就是把几个 job 连接成一个 workflows，但是Haddop MapReduce 没有直接提供任何 workflow
支持，可以隐式通过 HDFS 设计好的目录名支持，一个 job 配置写输出到特定 HDFS 目录，然后另一个 job 配置读它作为输入。
但是必须保证第一个 job 成功之后才能执行下一个，很多不同的 workflows 调度被开发出来，比如Oozie, Azkaban, Luigi, Airflow
and Pinball.其他很多高层工具比如 Pig,Hive, Cascading, Crunch, FlumeJava 同样提供了自动串联 workflow 的工具。

### Reduce-Side Joins and Grouping

当我们在批处理的情景下讨论 join 时，一般是数据库里出现的所有关联（不同于关系数据库）
Sort-merge joins

### Map-Side Joins

-   Broadcase hash joins
-   Partitioned hash joins(bucketed map joins in Hive)
-   Map-side merge joins

### The output of Batch Workflows

##### Building serarch indexes

用来构建一些全文搜索引擎的索引

##### key-value stores as batch process output

构建机器学习系统，比如分类器（反作弊过滤、异常检测、图像识别）和推荐系统

##### Philosophy of batch process outputs

MapReduce 遵守同样的unix哲学: 把输入当做不可变的并且避免有副作用（比如写到别的数据库）

### Comparing Hadoop to Distributed Databases

### Beyond MapReduce

### Graphs and Iterative Processing

### High-Level APIS and Languages

##### The move toward declarative query Languages

应用层代码只需要指定使用哪种 join 方式，查询优化器就能以一种最优方式运行

# 11. Stream Processing

之前讨论的数据都是有界的(bounded)，对于 unbounded 数据需要用流式处理(stream processing)

### Transmitting Event Streams

批处理的输入和输出都是文件，一般处理第一步是分隔文件成记录。流式处理环境里，记录通常作为一个时间(event)，指定时间点内
包含具体细节的小块、自包含不可变对象，通常包含时间戳。流式处理中，一个 event 被一个
生产者生成，然后可能被多个消费者消费。一个文件系统中，文件名标志了相关的记录，流式系统，相关事件通常被分组为一个 topic
或者 stream，然后由消息系统发出。 

### Messaging Systems

通知消费者新事件通常使用消息系统：生产者发送一条包含事件的消息，然后 push 到消费者。一般 tcp
连接只有一个发送者和接收方，而消息系统允许多个生产者节点向同一个 topic 发送消息，也允许多个消费者节点接收同一个 topic
的消息。（发布订阅模式）

##### Direct messaging from producers to consumers

有一些消息系统直接在生产者和消费者之间使用连接而不使用中转节点：

-   UDP  多路广播. 经常使用在需要低延迟的场合比如股票市场 feeds，一般生产者需要记住哪些包已经发了方便重试
-   无中间件的消息库比如 ZeroMQ
-   StatsD, Brubeck 使用 UDP 收集指标
-   如果消费者暴露了网络服务，生产者可以直接通过 HTTP  or RPC 请求 push 消息 (webhooks就是这种思想)

##### Message brokers

一个广泛使用的选择是发送消息到一个消息代理（message broker or message queue），一般通过某种类型的数据库来优化消息流。
通过把数据中心化到代理身上，系统更好地容忍客户端失败重连、crash 等情况，消息持久化落到了 broker
身上。一般可以通过配置选择消息在内存还是持久化到数据库中。通过消息队列实现了异步化，发送者只需要等到 broker
确认消息入队就会立刻返回

##### Message brokers compared to databases

消息队列虽然有存储功能甚至有的实现了两阶段提交协议，但是和数据库还是有区别：

-   数据库存储的数据不会自动删除，消息 broker 待成功发给消费者后自动删除
-   消息 broker 假定处理集合比较小，消息会被很快处理，如果大量堆积就会影响吞吐
-   数据库经常支持二级索引，消息 broker 一般支持某种 topic 匹配的订阅
-   broker 不支持任意条件的查询

##### Multiple consumers

当多个消费者向同一个 topic 读取消息的时候，使用两种模式：

-   Load balancing: broker 可能随意找一个消费者发，你可能想增加消费者并行处理
-   Fan-out: 每条信息发送给所有的消费者

##### Acknowledgements and redelivery

一般为了防止消费者崩溃没有处理消息，会使用确认机制：消费者回执处理完消息了然后 broker
才会把消息从队列移除。但是有可能消费者处理成功了但是回执却失败了，通过原子提交协议(atomic commit protocol)处理
