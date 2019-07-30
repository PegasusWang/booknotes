>The Internet was done so well that most people think of it as a natural resource like the Pacific Ocean, rather than something that was man-made. When was the last time a tech‐ nology with a scale like that was so error-free?  —Alan Kay, in interview with Dr Dobb’s Journal (2012)

# 1. Reliable, Scalable, and Maintainable Application

大部分 web 系统我们关心三个重要概念：

### Reliability(可用性)：系统需要持续正确工作（甚至是在遇到硬件故障、软件错误、人为错误等）

-   硬件故障：磁盘 crash、内存坏掉、电源被拔掉。通常通过硬件冗余（备份）来解决
-   软件错误：比如软件 bug、资源用尽、无响应、级联错误。没有特快快速解决软件系统性错误的方法，有些方式还是有帮助的：比如做好系统之间交互的假定、全面的测试、进程隔离、崩溃自动重启、生产环境监控和分析等。
-   人为错误：人类是不可信的，经常出现操作、配置等出错误。
    -   以降低出错可能的方式设计系统。比如良好的抽象、api、admin 接口能让减少出错的可能
    -   在人们最容易出错的地方解耦。比如创建沙盒环境
    -   在所有层面测试，从单测到集成测试。自动化测试
    -   允许快速从人为失误中恢复。比如快速回滚代码和配置
    -   简历监控系统。比如性能和错误指标
    -   管理和培训

### Scalability(可扩展性)：能应付系统增长（数据集、负载、复杂度的增长等）

-   负载: 可以用一些负载参数衡量，比如 qps、缓存命中率、同时支撑用户在线数等。书中用 tweet 举例，感兴趣可以看下，feed 刘系统常见的推拉模型。
-   性能：1.负载参数增加的时候，保持系统资源（cpu、内存、带宽）不变，性能如何变化？2.负载参数增加的时候，为了保持同样性能需要增加多少资源？

在批处理系统例如 hadoop 中通常用吞吐量(throughput)衡量，即每秒处理的记录个数。在线系统中通常用响应时间来衡量。
通常使用 percentiles 而不是算术平均值(average)来衡量响应时间，percentiles
指的是把响应时间从快到慢排列后取中位数作为分界点。比如如果 p95 响应时间是 1.5秒，意味着 100 个请求有 95 个少于
1.5秒，5个多于 1.5秒。

- https://en.wikipedia.org/wiki/Percentile
- https://blogs.dnvgl.com/software/2016/12/p10-p50-and-p90/

### Maintainability(可维护性)：能够有效率地扩充功能

众所周知，软件开发的主要成本不在起初的开发阶段，而在持续维护：修复 bug、保持系统运转、调查失败、适应新平台、修改使其适用于新用例、修复技术债、添加新特性。
软件系统的三个设计原则：

-   可用性（Operability）：保持操作简单

    -   监控系统健康，快速恢复
    -   追查问题原因，比如系统故障或者性能恶化
    -   保持系统和平台更新
    -   预防不同系统之间相互影响
    -   为将来出现的问题做准备（容量升级）
    -   建立开发、配置管理的良好实践
    -   预演复杂的维护任务，比如迁移应用到其他平台
    -   在系统配置修改的时候保持安全性
    -   保持生产环境平稳
    -   保留系统的知识财富

-   简约性（Simplicity）：对新人友好，尽量减少复杂度

复杂症状：explosion of the state space，模块之间紧耦合、纠缠不清的依赖、不一致的术语和命名、为了解决性能问题各种
hack、到处都是特殊用例。
避免意外复杂度(accidental complexity)的最好工具就是抽象，良好的抽象可以隐藏细节（高级语言程序设计就是例子）。

-   可进化性（Evolvability）：容易修改适应新需求

    需求等总是在变，敏捷工作流提供了一个适应变化的框架（测试驱动开发TDD、重构）

# 2. Data Models and Query Languages

复杂应用通常有很多层，一个 api 构建在其他 api 上，但是基本思想很简单：每一层通过干净的数据模型来隐藏其下层的复杂性

### Relational Model Versus Document Model

数据库多年来一直被传统的关系型数据库占领，直到最近随着互联网各种业务的爆发，有时候传统的关系型数据库满足不了需求，Nosql(Not Only Sql)出现。
在数据交互中，JSON 格式慢慢占据主流，由于 某些格式可以用 JSON 自包含，我们可以用 MongoDB、RethinkDB、CouchDB、Espresso 等支持文档型数据的数据库来存储。
在某些数据查询中，使用 JSON 格式表示数据可以让我们避免关系型数据库中多次 join，更具效率

文档模型的优点在于存储模式灵活(schema-on-read)、性能更好，对于某些应用来说更符合应用使用的数据结构(比如自包含结构，一对多树形结构）。

关系型数据库则更好地支持 joins，多对一和多对多关系。

### Query Languages for Data

讲了声明式查询 (Declarative query lagunage) 和 命令式查询(Imperative query)

Mapreduce query: Mongodb 支持 mapreduce 查询，但是限制 map 和 reduce 是纯函数 (pure function: 无副作用且无法执行额外数据库查询)

### Graph-Like Data Models

图由交点（节点）和边（关系）组成。很多关系可用图来建模：比如社交网络、web 网页、路线图。
很多知名算法用来操作图关系，比如最短路径算法、page rank

Property graph model: Neo4j(Use Cypher Query Language), Titan, InfiniteGraph.

# 3. Storage and Retrieval

### Data Structures That Power Your Database

-   Hash Indexex: 类似于大多数编程语言中实现的dict，采用 hash map 实现，k,v 结构
-   SSTables(Sorted String Table) and LSM-Trees(Log-Structured Merge-Tree): basic idea of LSM-trees: Keeping a cascade of SSTables that merged in the background.
-   Btrees:几乎是所有关系型数据库实现索引的方式。

### Transaction Processing or Analytics?

-   OLTP (online transcation processing): 交互式应用中访问数据的模式，要求低延迟，通常只访问小部分数据
-   OLAP (online analytics processing): 用来数据分析，通常需要扫描很大的数据集，每个记录通常只访问很少的列，通常用来聚集分析而不是直接返回给用户，结果用来作为商业分析和决策。

##### Data Warehousing(数据仓库)

ETL (Extract-Transform-Load): 数据仓库是一个独立的用来分析的数据库，防止影响 OLTP 操作。大部分是从 OLTP 数据库同步过来的只读数据，转换成比较容易分析的格式，然后导入数据仓库。
常用的开源的有 Apaceh Hive, Spark SQL, Cloudera Impala Facebook Presto, Apache Tajo, Apache Drill.

##### Stars and Snowflakes: Schemas for Analytics

star scehma(dimensional modeling): the fact table in the middle, surrounded by its dimension
snowflake schema: dimensions are further brokern down into subdimensions.

### Column-Oriented Storage

OLAP
应用中大部分时间我们只需要一行中的某几列。传统的行式关系型数据库会把每行所有数据载入，效率比较低(即使过滤了行也是先加载所有数据之后过滤)。
于是，列式存储出现了。列式存储的思想很简单：不要把所有值存储在一行，而是把所有值存储在每列。这样存储就变成了从

| date_key | product_sk |
| -------- | ---------- |
| 140102   | 69         |
| 140102   | 69         |
| 140103   | 70         |

到：

date_key file contents: 140102, 140102, 140103
product_sk file contens: 69, 69, 70

列压缩：一列中有很多元素是重复的，我们可以把一列分成不同的 n 个值，然后转化成 n 个不同的 bitmap，每个 bitmap
代表一个值，每行用一个 bit 表示。bit 如果是 1 代表改行有相应值，0 代表没有。

# 4. Encoding and Evolution

需求总是在变，这时候需要对原有的数据库模式和代码做出变更，同时保证兼容：

-   Backward compatibility: 新代码可以读取旧代码写入的数据
-   Forwared compatibility: 旧代码可以读取新代码写入的数据

### Formats for Encoding Data

程序经常操作两种类型的数据表现方式：

-   内存中：数据被保存成对象、结构、列表、数组、哈希表、树等。这些数据结构可以被高效率读取和被 cpu 操作
-   需要写入文件或者网络的自包含的 bytes 序列：比如 JSON

把内存里的数据 编码成比特序列的过程叫做 Encoding（serialization or marshalling），反过来叫做 Decoding(parsing, deserialization, unmarshalling)

##### JSON, XML and Binary Variants

JSON，XML，csv 经常用作数据交换，它们都是基于 文本的，可读的。但是有以下问题：

-   在数字编码上有歧义。 XML 和 csv 无法区分包含数字的字符串和数字。 JSON 虽然区分但是无法精确处理大数字（js）
-   JSON 和 XML 不支持二进制 string（可以 base64编码，但是会增加体积
-   There is optional schema support for both XML and JSON
-   csv 不支持任何模式

##### Binary encoding

内部使用的一些交换格式可以使用一些更高效率的编解码方式。比如 二进制 encoding for JSON(MessagePack, BSON, BJSON, UBJSON,
BISON, and Simle)。这些方式各有使用场景，但是依然没有 json 和 xml 广泛。而且，它们没有预定义一个
schema，所以需要把所有对象的字段名称包含到 encoded 的数据里。

### Thrift and Protocol Buffers

Apache Thrift and Protocol Buffers(protobuf) 都是二进制编码库。Protocol Buffers 最开始 google 开发，Thrift facebook
开发。

Thrift interface definition Language(IDL):

    struct Person {
       1: required string username,
       2: optional ii64 favoriteNumber,
       3: optional list<string> interests
    }

在 Protocol Buffers 等价定义：

    message Person {
       required string user_name = 1;
       optional int64 favorite_number = 2;
       repeated string interests = 3;
    }

两者都是实现了代码生成工具来生成对应语言的代码。

Thrift 有两种不同的二进制编码格式：BinaryProtocol and CompactProtocol

##### Field tags and schema evolution

schema 变了以后 Thrift and Protocol Buffers 是如何处理 schema 变化的来保持前向和后向兼容的？
你可以改变字段名，但是不能修改它的 tag 名称。可以通过增加 tag 数字号来增加新的字段，client 端可以简单忽略新增的字段。
由于每个 字段都被一个 tag 数字标识，所以新代码可以直接读旧代码的数据，因为每个 tag
代表的字段含义不变。唯一的要求就是如果想添加一个新字段，不能是
required，因为老代码无法写入新增的这个字段。所以新加的字段必须是 optional 或者有一个默认值。

##### Datatypes and schema evolution

如果改变一个字段的类型呢，可以但是会有损失精度或者被截断的风险。

### Avro

Apache Avro 是另一种二进制编码格式，Hadoop 子项目。包含两种格式定义语言(Avro IDL)，用于人类编辑，还有一种机器易读的基于 JSON 的定义。

### Modes of Dataflow

How data flows between processes:

-   via database

数据可能会在新、老代码交换数据之间丢失。

-   via service call
-   via asynchronous message passing

### Dataflow Through Services: Rest and RPC

不同进程之间通过网络交换数据，通常采用 client-server 模型。
SOA(service oriented architecture): 最近慢慢演变成 microservices architecture。既能自己提供服务，也能调用其他服务获取数据。
这种架构使得修改和维护服务更容易。每个服务都可以由一个单独的团队来维护。

##### web service

-   REST: rest 不是协议，更像是一种基于 http 原则的设计哲学，经常用在微服务架构。遵守REST原则设计的 api 成为restful
-   SOAP: xml-based protocol, 大多构建在 http 上，但是目标更倾向于独立于 http 并且避免使用 大部分 http 特性。jo

##### The problems with remote procedure calls(RPCs)

RPC：远程过程调用，目标是使得远程服务调用就像是在本地进程里。这个目标有些瑕疵，网络调用和本地函数调用是有很大区别的：

-   本地调用是可预测的，成功或者失败，取决于传参控制。但是远程网络调用是不可预测的，比如网络环境、远程服务不可用
-   本地调用返回结果、抛出异常或者因为程序死循环等用不返回，远程调用可能会因为超时返回一个没值的结果
-   如果你重试一个失败的网络请求，可能会发生请求发出去了只是没有返回结果的情况，这个时候重试可能导致多次请求， 除非建立一个机制防止重复。
-   本地调用基本都是执行一样的时间，网络调用延迟可能会因为负载高而延迟
-   本地调用可以传递指针，远程调用必须都传递过去
-   rpc 跨语言，可能在转换的时候出问题，比如某些语言不支持大数

### Asynchronous Message-Passing Dataflow

和 rpc 类似的是通常一个客户端请求(usually called a message)
被发送到另一个进程，消息不是直接通过网络请求发送，而是通过一个叫做 message broker(also called a message queue or message-oriented middleware)，用来临时存储消息。
相比 rpc，有几个好处：

-   如果接受者负载高，可以作为一个缓存 buffer，提高系统的可用性
-   可以自动重发消息如果进程 crash，保护消息避免丢失
-   发送方不必知道接收方的 ip、端口号（使用云部署平台的时候比较有用）
-   允许一个消息发送个多个接受者
-   逻辑上解耦了发送者和接收者

发送发一般不用关心消息的返回结果，从而实现异步

##### Message brokers

历史上消息 broker 曾被商业公司垄断，最近开源的 RabbitMQ, ActiveMQ, HornetQ, NATS, Apache Kafka 慢慢流行。
message broker 通常如下使用：一个进程发送消息到命名队列或者topic，broker保证消息被分发到一个或者多个消费者(订阅了queue或者 topic的)
同一个topic上可以有多个生产者和消费者

##### Distributed actor frameworks

actor 模型是一种在单个进程上实现并发的编程模型。与使用线程实现不同，逻辑被封装成 actors。每个 actor 代表一个实体或者
client（可以有自己的状态），通过异步消息和其他 actor 通信，但是消息分发不能被保证。一个actor
每次只处理一个消息。不用担心i安城，每个 actor 可以被框架来调度。在分布式 actor
框架中，一个应用被扩展在多个节点，消息传递是透明的。
几个流行的 分布式 actor 框架：

-   Akka
-   Orleans
-   Erlang OTP
