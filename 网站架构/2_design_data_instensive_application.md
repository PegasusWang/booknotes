# 5. Replication

Replication(副本)不同机器上保存同样的的数据。基于以下理由：

-   保证数据地理上离用户更近（降低延迟）
-   保证高可用，一个挂了切到另一个
-   增大读取吞吐量

本章讨论三种在 replication 同步数据的方式： single-leader, multi-leader, leaderless replication

### Leaders and Followers

每个存储节点都叫做一个 replica，一个不可避免的问题是：如何保证所有数据都落在所有副本上？
每次写入数据库都要同步到每个 replica 上，最常用的的方式是主从复制(leader-based replication also master-slave replication)

1.  一个 replica 被设计成 leader，client 必须像 leader 写入数据。
2.  leader 写入数据的同时发送数据变动的 log 供其他 followers 读取并更新其存储。
3.  数据读取请求可以从任意一个 replica 读取。

### Synchronous Versus Asynchronous Replication

同步的更新请求能保证和 leader 一致的数据，缺点是如果 follwer 没响应会导致无法处理写入请求，直到 follower 恢复。

一般会保证有一个是同步的，其他是异步的，如果同步的跪了就选取一个异步的为同步的，保证总有俩节点数据一致。(半同步)
通常主从复制被配置成全异步的，缺点是如果 leader 跪了所有的写入请求无法处理，好处是就算所有的 followers 跪了，依然可以处理写入请求。
主从复制不再被广泛使用了，尤其是 followers 在不同地理位置上分布的。

### Setting Up New Followers

增加新的 replica 的时候如何保证数据被精确从 leader 拷贝？

-   在某个时间点建立 leader 的快照
-   复制快照数据到新节点
-   follower 连接到 leader 请求快照时间点之后的数据
-   当处理完从快照时间点以后的数据（caught up），就可以继续处理来自 leader 的其他数据变更

### Handling Node Outages

主从复制模式下，如何保证高可用？

##### Follower failure: Catch-up recovery

如果 follower 短时间内不可用，恢复起来很容易，从 log 恢复。可以从log 中知道出错后最后一次数据处理时间点，连上 leader
后可以从该时间丢失的数据全部请求回来。

处理 leader 的失败有些麻烦，需要一个 follower 被提升为新的 leader，client 需要被配置写入新的 leader，其他的 follower
需要从新的 leader 读取数据变更。这种处理叫做 failover，可以手动或者自动处理：
1. 确认 leader 故障。crash，断电，网络问题等等。通常用响应超时来检测
2. 选取新的 leader
3. 重新配置系统使用新的 leader。写入请求到新的 leader

Failover 策略也会有一些问题：(也是分布式系统面临的问题)

-   如果是异步同步，新的 leader 有可能并未完全同步 old leader 故障前的所有数据, old leader 重新假如集群后可能导致写入冲突。比如占用了相同的主键。一种方式是丢弃这些数据，但是会影响客户端期望的持久性。
-   可能导致不同数据库之间的数据不一致。比如同时在 mysql 和 redis 使用了相同的主键
-   在特定的失败场景中，可能有两个节点都认为自己是 leader（split brain），可能会有写入冲突
-   如何正确设置超时时间认为 leader 挂了？有可能是负载高导致延迟久被误认为 leader 挂了

##### Implementation of Replication Logs

几种不同方式在实践中被使用：

-   Statement-based replication: 最简单的场景，每个 leader 的写请求都被执行然后发送 log 到它的 followers, mysql5.1 之前使用。有以下问题：

    -   未决函数。比如 NOW ， RAND 等是会返回不同的值的
    -   如果语句使用的是自增列，或者基于已有的数据，那这些语句必须在 replica 上用相同顺序执行，否则会有副作用
    -   语句可能有副作用。（触发器、存储过程、用户定义函数）可能在不同的 replica 上有不同的副作用，及时副作用是确定性的

-   Write-ahead log(WAL) shipping
    把语句变成 append-only 的字节序列当成 log 发送给 follower，当 follower 处理这些 log 的时候，就建立了和原始写入一样的数据结构
    PostgresSQL 和 Oracle 采用这种方式。缺陷：如果不同版本的 replication 协议版本有变化，升级会不兼容

-   Logic (row-based) log replication
    一种选择是对于 replication 和 数据存储引擎 使用不同的 log 格式，允许 replication log 从数据引擎log 解耦，这种叫做
    Logic log。通常是一些列记录用来描述行级粒度的写入：
    -   对于插入一行，log 包含所有列的新值
    -   对于删除一行， log 包含足够的信息来定位需要删除的行。通常是主键
    -   对于更新一行，log 包含足够的信息去更新需要更新的列数据

一个修改了多行的事务会生成多个 logic log，然后是一个记录事务被提交的记录。mysql bin-log 配置成row-based replication 使用的就是这种方式。
好处是可以跨数据库版本（和存储引擎的log解耦），也方便被其他的系统解析和使用。

-   Trigger-based replication
    前面几种方式都是数据库自己实现的，不需要应用层代码。有些场景比如你想复制特定的数据集合，可能需要把 replication
    提升到应用层。可用一些工具或者数据库提供的 触发器和存储过程 实现。d

### Problems with Replication Lag

使用 replication 的一个原因之一就是能容忍个别节点故障，但是会出现一些问题，主从复制架构中，异步复制会导致数据 followers
的数据同步落后，这就会带来数据不一致的问题，同样的查询在 leader 和 folower 上运行结果不一致，当然这只是暂时的，最终数据会同步成一致的，也就是最终一致性(eventual consistency)

##### Reading Your Own Writes

如果数据是读多写很少的，比如一条评论等，可以写入 leader，然后从 follower 读取。异步复制中有个问题，如果用户在写入之后很快就读取，这时候数据可能还没复制到
follower，对用户而言数据好像丢失了。这个时候需要 read-after-consistency(read-your-writes consistency)，保证用户在更新数据后刷新页面，能立刻看到他们的更新，但是不能保证其他用户的。
以下是一些实现方式：

-   当读取用户可能修改过的数据的时候，从 leader 读，否则从 follower 读取。总是从 leader 读取用户个人信息之类的东西，其他的从 followers 读。
-   跟踪数据变更时间，在更新的后一分钟内都从 leader 读取。
-   客户端记住最近的更新时间戳，然后系统可以从有该时间戳以后更新数据的 replica 读取，或者等待直到更新了数据。
-   如果数据中心分布在不同地理位置，请求必须路由到有改 leader 的数据中心读取。

    更麻烦的是跨设备读取，比如在移动设备上更新了数据，在其他设备上要能看到更新。

##### Monotonic Reads

第二个异步复制会出现的问题是用户可能会看到某些数据『时间倒流』，从不同的 replica 中读取数据可能遇到。
Monotonic Reads: 每个用户的读取都只从一个 relica 读。比如基于用户 id hash 来选取 replica 读取。

##### Consistent Prefix Reads

第三个异步复制会碰到的反常的地方是：如果有些分片复制速度比其他慢，一个观察者可能会先看到回答，后看到问题（因果错乱）。
避免这种反常的方式是：consistent prefix reads.  如果一系列写入操作按照确定顺序执行，其他人读取这些写入结果以相同顺序出现。

保证写入操作有关联方都写入同一个 partition

##### Solutions for Replication Lag

在一个最终一致性系统下，考虑应用在复制延迟过久（分钟到小时）的情况下如何表现是必要的。如果对用户来说很不友好，就应该考虑使用
read-after-write 等保证用户体验。

### Multi-Leader Replication

一主多从架构有个很明显的缺陷：只往一个 leader 写入。如果 leader 挂了或者无法连接了，就不能写入数据了。
multi-leader(master-master/active replication): 多主架构就是为了解决这个问题。每个 leader 对于其他 leader 表现像 follower。

#### Use Cases for Multi-Leader Replication

多主架构中一般很少只用一个数据中心，因为收益可能小于增加的复杂度。相比一主多从：

-   更好的写入性能
-   更好地容忍数据中心中断
-   更好地容忍网络问题

有些数据库支持多主配置，但是经常使用额外工具实现。比如 mysql: Tungsten Replicator
多主也有风险，比如不同数据中心写入冲突；自增 key，触发器，数据完整性限制等可能会出问题。

##### Handling Write Conflicts

多主架构最大的问题就是写入冲突，比如两个用户同时对一份数据做了不同修改，分别写入到两个数据中心
leader，当执行异步复制的时候就会检测到冲突。

-   Synchronous versus asynchronous conflict detection: 使用同步冲突检测会丧失多主可以分别写入的优点，如果想用同步冲突检测，应该使用一主多从架构
-   Conflict avoidance: 应用保证所有对于同一个记录的写入都写入到同一个 leader，冲突就可以避免。切换 leader 时会带来问题
-   Converging toward a consistent state: 达到一致状态

    -   给每次写入分配 id，最大的 id 写入视为最终写入，丢弃其他写入。或者用时间戳，最新的写入有效（last write wins,
        LWW）。这种方式很流行，但是可能造成数据丢失
    -   给每个 replica 一个独一无二的 id，高 id 的 replica 优先
    -   记录冲突并提供保留所有信息，应用层代码之后处理冲突

-   Custom conflict resolution logic: 在应用层代码决定。
    -   on write: conflict handler，在后台进程检测到 replicated 变化时处理冲突
    -   on read：读取的时候应用层代码处理冲突，然后写回数据库

自动解决并发修改数据冲突：

-   Conflict-free replicated datatypes(CRDTS), 一系列可以被并发修改和自动解决冲突的 sets，maps，orderd lists，counters 数结构
-   Mergeable persistent data structures。类似 git 版本控制，采用三路合并解决冲突
-   Operational transformation. 冲突解决算法，被协作式编辑用到（google doc）

##### Multi-Leader Replication Topologies

多主架构中常用的拓扑是 all-to-all，每个leader 把写入发送给其他 leader。mysql 默认只支持环形拓扑，一个节点只能从另外一个节点收到写入然后传播给下一个节点。 星型拓扑，一个 root
节点传播写入到其他所有节点。环形和星型会给节点一个标识，防止重复写入。但是遇到一个节点故障会停滞传播。
all-to-all 传播也会有问题，写入可能会在一些 replcias 中以错误的顺序执行，一般使用叫 version vectors 的技术解决。

### Leaderless Replication

前面俩种都是基于一个节点写入后发送到其他节点的思想，leader 决定那些写入需要被处理，followers 遵从leader 的写入顺序写入。
有些数据存储系统抛弃了这种思想，允许任意一个 replica 可写，比如 Amazon 的 Dynamo 系统， Riak，Cassandra 等开源存储的
leaderless 复制也受 Dynamo 启发（奇怪的是Amazon 的另一个 DynamoDB 用的 leader-based 架构）。一些 leaderless 架构实现中，client 直接向几个 replica
发送写入请求，其他实现方式在 client 中间使用一个协调节点实现。

##### Writing to the Database When a Node is Down

我们考虑一种三个节点有一个节点挂掉的情况，这个时候同时写入三个节点只有俩返回了 ok，我们认为写入就算成功了，忽略失败的那个。
等挂了的那个恢复了，每个读取它的 client 会漏掉它挂掉时候的写入数据。为了解决这个问题，读取的时候也同时读取三个节点，然后根据版本号确定哪个值是最新的。

###### Read repaire and anti-entropy

刚才『三缺一』的写入情况如何处理挂掉的节点恢复后跟上新数据？Dynamo-style 数据存储经常用两种机制：

-   Read repare: 读取的时候根据版本号确定最新数据，同时把新数据写入到包含只有旧数据的 replica。对经常读取的数据比较有效
-   Anti-entropy process: 一些数据存储有个后台进程一直去检测 replica 的丢失数据，然后从其他 replica 拷贝过去

###### Quorums for reading and writing

刚才讨论了三个有俩写入成功的情况，如果只有一个写入成功了呢？
为了更通用，我们假定有 n 个 replica，每次写入必须有 w 个节点被认为写入成功，每次查询必须至少有 r 个节点。
（上面的例子中 n = 3, w = 2, r = 2），只要 w + r > n 我们认为读取到了最新的值，因为至少 r 个节点了更新了值。
读取和写入遵守这个规则的 r 和 w 值成为 "quorum reads and writes"，可以认为 r 和 w 是一个合法读取和写入的最小投票数。
在 Dynamo-style 数据库中，n, w, r是可以配置的，一般的选择是 n 是个奇数（3 or 5），然后设置 w=r=(n+1)/2(上取整)

##### Limitations of Quorum Consistency

-   sloppy quorum: w 个写入可能不是 r 个节点读取的，无法保证重叠
-   无法确定并发写入顺序
-   如果读和写是并发出现的，写入只在部分 replica 有效
-   如果一个写入报告失败，后续的读不一定会返回写入的值
-   一个节点如果被其他节点旧数据恢复，存储新节点的节点数会小于 w 值，打破 quorum condition (w + r > n)

quorums(投票数）虽然满足读取了最新的值，但实际并不容易做到。Dynamo-style 数据库通常优化满足最终一致性

###### Monitoring staleness

在 leaderless 系统中，监控节点数据更新落后更加困难。可以为数据库加入滞后检测，量化『最终一致』

##### Sloppy Quorums and Hinted Handoff

虽然满足 quorums(w + r > n) 配置的数据库能满足高可用和低延迟，但是并不能像期望一样容忍故障，但有节点故障的时候，就会无法满足 quorum 条件。
在一个大的集群中（超过 n 个节点），client 可以在网络冲突的情况下连接到一些节点，但是可能无法满足 quorum
条件，这个时候需要权衡：

-   如果达不到 quorum of w or r nodes, 请求全部返回错误？
-   或者继续接受写入请求，然后写入一些可用的节点，但这些节点不在 n 个 nodes 里？
    后者又称为 sloppy quorum, (弱法定数(不知道咋翻译))，写入和读取仍然满足 w 和 r，但是其中一些借点可能开始没被算在 n 个节点里。
    当网络恢复，每个被临时写入的节点被重新写入到合适的 "home" 节点，这个过程叫做 "hinted handoff".
    sloppy quorum 在 Dynamo 中作为可选实现（有些默认启用 Riak，有些默认不启用 Cassandr, Voldemort）

##### Multi-datacenter opration

Leaderless 复制也适合多数据中心操作，因为它就是被设计容忍并发写入冲突、网络故障、延迟尖峰

##### Detecting Concurrent Writes

出现写入冲突的根本原因是写入事件可能以不同的顺序到达不同节点（网络延迟、部分失败等）

###### Last Write Wins(discarding concurrent writes):

如果丢失数据不可接受， LWW 对于解决冲突是个不好的策略。使用 LWW 安全的前提是确保一个 key
只被写入一次然后被当做不可变，从而避免了对同一个 key 的并发写入。

###### The "happens-before" relationship and concurrency

并发有时候是时间无关的（不是同时写入就是并发），如果两个操作互相不感知，不管时间是否一致，都认为是并发的。

###### Capturing the happens-before relationship

本节讲了用 version vector 处理 leaderless 架构下并发写入冲突的问题。例子比较复杂，感兴趣的自己看书。

# 6. Partitioning

上一章讲了 replication，同样的数据存储在不同的节点上，但是对于很大的数据集或者高吞度量查询，效率就不够了。
我们需要把数据分片 (partitions)，又叫做 sharding，通过分片，我们可以把一个很大的数据集分布式存储在不同硬盘上，
查询也可以分布在不同处理器上。

### Partitioning and Replication

一般分片和复制结合起来用，每个分片存储在多个节点上。为了简单起见，本章省略 replication 内容。

### Partitioning of Key-Value Data

如果你有很多数据想要分片，你怎么决定哪些数据落在那些节点呢？我们的目标是想要分片以提供更高查询负载
如果没有公平分片，导致一些节点查询请求高，一些低，我们叫做『skewed』(偏斜)。有些极端情况查询都不行落在一个几节点上，
叫做 hot spot。
最简单的策略是每个节点随机写入记录，缺点是一旦你想查询特定项，你不知道数据落在哪个节点，只能并行查所有节点。

##### Partitioning By Key Range

一种分片方式是把一系列连续范围的 key 写到每个分片，比如我们把所有 A 开头的书放在一个书架，B 开头的放在一个，类推。
这种分片方式被 Bigtable ，等价开源实现 Hbase，RethinkDB, 和 MongoDB 2.4 版本以前 使用。
每个分片里，我们还可以保持 key 有序，这样方便实现范围扫描。这种分片方式的缺点就是可能导致 hot spot，可以用给 key
加上前缀的方式使得 key 分布更均匀。

##### Partitioning by Hash of Key

为了防止 hot spots(key 分布不均导致单个节点压力过大)，很多分布式存储使用 hash 函数决定一个 key
落在那个分片上，一个好的哈希函数应该尽量使 key 均匀分布，通常使用一致性哈希(consistent
hashing)，这里的一致性指的是重新分配 key 的方式，其实用术语 hash partitioning 好一些。
但是使用哈希分片也有个缺点，无法实现高效的 range query 查询

###### Skewed Workloads and Relieving Hot Spots

哈希分片可以帮助减少 hot spots，但是极端情况下如果都是针对同一个 key 的读写，就会把所有请求落到一个分片上。

###### Partitioning and Secondary Indexes

刚才讨论的都是 k，v 数据模型，如果涉及到二级索引会更麻烦，通常二级索引不用来识别一个唯一记录，但是提供一种寻找特定值出现的方法。
比如找到一个用户的所有行为、寻找包含 "hogwash" 的所有文章等。二级索引在关系型和文档型数据库比较常见，但是很多K/V
存储比如 Hbase 为了降低实现复杂性没有支持二级索引。主要有两种实现二级索引的方式：

-   Partitioning Secondary Indexes by Document: 每个分片都有其二级索引。读取比较麻烦
-   Partitioning Secondary Indexes by Term： 使用被分片的全局索引。读取容易写入麻烦，写入一个可能影响多个分片的索引

### Rebalancing Partitions

数据库上的东西总是在变：

-   查询负载增加，你想添加更多 cpu 处理负载
-   数据量增加，想增加更多 硬盘和内存去存储
-   一个机器挂了，其他需要接管挂掉机器的职责

    所有这些变化都会导致数据和请求从一个节点转移到另一个节点，把负载从一个集群中的节点转移到另一个叫做 "rebalancing"

-   rebalancing 后负载均衡分配到各个节点
-   rebalancing 期间数据库仍正常读取和写入
-   只有必要的数据应该从节点迁移，降低 rebalancing 网络和磁盘IO 负载

##### Strategies for Rebalancing

###### How not to do it: hash Mod N

之所以不用这种直接取模的方式是因为，如果节点数 N 变化了，大多数 key 都要被转移

###### Fixed number of partitions

有一个中比较简单的方式，创建比需要更多的分片，然后把几个分片分配到每个节点。如果一个节点增加进来，新节点可以从之前的每个节点中
『偷』一些分片直到分片被重新公平分配。这种方式被 Riak、Elasticsearch、Couchbase、Voldemort 使用。

###### Dynamic partioning

Hbase, RethinkDB 使用动态分配，当一个分片超过配置大小，数据库会分成两个近似对半分数据；相反地，当大量数据被删或者缩小到某个阈值，就会合并相邻的分片。
每个分片被分配到一个节点，每个节点可以处理多个分片。

###### Partitioning proportionally to nodes

动态分配中分片的数量正比于数据量大小，因为分割和合并的进程保证每个分片的大小在固定的最小值和最大值之间。
Cassandra 和 Ketama 使用了分片数正比于节点数的方式，也就是每个节点包含固定的分片数，保证每个分片的平稳。

##### Operations: Automatic or Manual Rebalancing

自动分片比价方便，但是可能不可预测，rebalancing 是一个昂贵的操作，因为需要重新路由请求和在节点之间迁移大量数据。如果不仔细，这个过程可能导致网络超载或者影响请求性能。
一般在 rebalancing 过程中建议人参与。

### Request Routing

我们已经知道如何把数据分布在不同节点，但是仍然有一些个问题：client 如何知道该连接哪个节点？rebalancing 后一个查一个
key连接到哪个 ip 和端口？

解决这个问题的通俗做法叫做服务发现(service discovery)，不止限定在数据库。任意联网的软件都有这个问题，尤其是为了实现高可用的。

-   允许 client 连接到任何节点。如果节点正是需要请求的节点就直接处理请求，否则路由到合适的节点，接受请求处理后并返回给 client
-   发送所有请求到先到一个路由层，决定哪个节点处理哪个请求并正确发送过去。
-   client 需要对分片的节点有感知，可以直接连接到对应节点

核心问题就是：在节点上的分片变化时，组件如何决定路由到哪个节点？
很多分布式数据系统依赖一个独立的协调服务例如 ZooKeeper 来跟踪集群中的元信息。每个节点在 ZooKeeper 中注册，ZooKeeper
维护每个分片(partitions)和节点(nodes)的映射信息，路由层可以从 ZooKeeper
中订阅信息，当分片易主或者一节点增添后，ZooKeeper 会通知路由层，路由层可以保证它的路由信息始终是最新的。

Hbase, SolrCloud, Kafka 同样使用 ZooKeeper 来追踪分片分配信息。

### Parallel Query Execution

到现在讨论的都是针对一个 key 的读写，这是大多数分布式 Nosql 提供的。然而，massively parallel processing(MPP) 关系型数据库产品，
经常用来做分析。第 10 章将会讨论

# 7. Transactions

数据处理过程中会遇到各种错误（数据库软件、硬件挂了；应用程序 crash；网络中断；多 client
并发写入覆盖；读过期数据；竞争条件等），事务被使用来简化这些问题。一个 transaction 把多个读写归成一个逻辑单元。
概念上来说就是一个事务里的操作要么全部成功（commit）或失败（abort, rollback），如果失败了，应用层可以安全重试。
事务简化了应用层的错误处理，不用担心部分失败的问题，通过事务，应用层可以忽略可能的错误和并发问题，交由数据库处理。

啥时候使用事务呢？我们需要了解事务能提供什么样的安全保证以及需要付出的关联成本。

### The Slippery Concept of a Transaction

几乎所有关系型数据库和部分非关系型数据库支持事务。

##### The Meaning of ACID

事务能提供的安全保证经常缩写为 "ACID"，代表 Atomicity, Consistency, Isolation, Durability。不同数据库的 ACID
实现是不一致的（其实成为了市场术语）。不符合 ACID 的通常叫做 BASE，Basically Available, Soft state, Eventual Consistency.

-   Atomicity: 数据库中原子性跟并发无关，而是描述了当一个 client 想做一些写入操作，但是在一些写入后出错了应该如何处理。
    比如被组织成一个原子事务的写入，一堆写入后进程挂了、数据库满了，事务可以保证丢弃这些请求。如果没有原子性，我们就无法知道
    哪些写入生效，哪些没生效。如果 client 重试就有可能造成重复写入，原子性(abortability 术语可能更好)会丢弃这些写入，client 就能安全重试。

-   Consistency: ACID 里的 一致性（这个词在最终一致性、一致性哈希、CAP里的一致性含义不同）指的是数据库处于 『good
    state』的应用层概念(区别于其他三个概念），应用层代码可能需要依赖原子性和隔离性来实现一致性，而不是数据库单独支持。
    The idea of ACID consistency is that you have certain statements about your data(invariants) that must always be true.
    比如在一个账户系统中，信誉和借贷在所有账户中都必须是平衡的，如果一个事务以一个合法的状态开始，任何这个事务中的写入都要保持合法性，

-   Isolation: 大多数情况下一个数据库都会被几个 client 同时连接，如果他们访问了同样的数据库记录，会有并发问题。隔离性表示并发执行的事务是彼此隔离的，保证同一时间只有一个在执行，就好像串行执行的

-   Durability: 持久性保证只要一个事务被提交成功，，所有的数据都保证不会丢失，即使有硬件错误后者软件crash。单个节点数据库保证写入到数据库，replica
    保证数据被成功复制到其他节点，数据库需要保证这些操作完成才会报告 commit 是成功的。

### Single-Object and Multi-Object Operations

总得来说，原子性和隔离性描述了如果一个 client 在同一个事务中的几个写入应该如何处理：

-   Atomicity: 如果期间失败，事务应该被停止(abort)，之前的写入应该被丢弃。
-   Isolation: 并发执行的事务应该互不影响。一个事务看到的另一个事务的写入应该全部执行或者都不执行。

##### Multiple-object transactions:

(修改多行记录)需要一种方式来决定哪些读写是来自同一个事务的。在关系型数据库中，通常基于客户端的 TCP 连接：对于任意连接，所有在 BEGIN TRANSACTION 和 一个 COMMIT 中的语句被认为是同一个事务的部分。

##### Single-object wirtes

Atomicity 可以通过 log 实现记录恢复；通过对每个对象加锁来实现 Isolation

##### The need for multi-object transactions

很多分布式数据库不提供 multiple-object 事务（跨分片实现复杂、影响性能）。但有些 single-object 事务实现起来效率不够：

-   关系型数据库经常有外键，多行事务允许这些外键保持合法性
-   文档型数据库缺少 join，需要多行事务保持数据同步
-   二级索引的数据库需要在数据库更新后保持更新

##### Handling erros and aborts

事务的一个关键特性是可以被打断并且安全重试。ACID 数据库基于以下哲学：如果数据库无法保证原子性、隔离性、持久性，它最好丢弃事务功能而不是部分实现它。
大多数 ORM 框架遇到错误会抛出异常到上层代码，而不是自己重试。其实事务 abort 的关键就是在于允许安全重试。

### Weak Isolation Levels

理论上隔离性应该通过保证无并发操作来使得处理更简单：串行隔离(serializable isolation)保证并发写入也能像串行执行一样。
但现实中，隔离性不幸地没那么容易实现，主要是因为性能问题。通常会使用一些弱隔离性处理并发问题，虽然更难实现并且容易出现一些难以察觉的问题。
我们来看几种弱隔离性.

##### Read Committed

最基础的是事务隔离性是 read committed:

-   no dirty reads: 从数据库读取的时候，只会看到被提交过的数据
-   no dirty wirtes: 写入的时候只会覆盖被 cimmtted 的数据

read Committed 级别的事务隔离很流行，被 Oracle, PostgresSQL, SQL Server 2012, MemSQL 和很多其他数据库使用。
通常，数据库使用行锁(row-level locks)来保证没有 dirty wirtes，一个事务想要写入一行或一个 document必须先获取行锁，知道被提交或者 aborted 后释放。
但是读取如果使用 行锁会非常影响性能，数据库采用了记录 old 和 new 值的方式，未提交之前的事务给返回旧的值，提交完后返回新的值

#### Snapshot Isolation and Repeatable Read

read committed 场景中会出现 nonrepeatable read (or read
skew)，临时不一致性。想像一个场景：如果一个用户有1000元，分成两个账户，每个500。如果在一个事务中，从一个账户转到另一个账户500元，
在事务没处理完的时候就去读，会发现一个帐号少了100，但是还没转到另一个账户，好像凭空消失了100。这种临时不一致大部分情况下是能接受的，
但是也有一些场景不能容忍：

-   backups: 大量数据复制的场景下比较耗时，在备份进程处理期间，如果数据库仍在处理写入。这样可能导致备份一部分包含旧版本数据，一部分包含新版本数据。
    从备份中恢复的时候就会导致永久不一致。

-   Analytic queries and integrity checks: 数据分析场景的查询会在不同时刻查到不同的值

Snapshot isolation(快照隔离)就是为了解决这个问题。每个事务从一致的数据库快照读取，每个事务只能看到一个时间点的旧数据。
Snapshot isolation 被 PostgresSQL, Mysql with InnoDB，Oaacle, SQL Server 等支持。
Snaashot Isolation 通常用写锁防止 dirty wirtes，核心原则是：写不 blokc 读，读也不 block 写。
通常使用 MVCC(multi version concurrency control)
维持多个时间点版本的数据来实现快照隔离。PostgresSQL实现： 当一个事务开始时，会分配一个唯一的自增的事务id
(txid)，无论何时一个事务向数据库写入数据，被写入的数据会被 txid 标记。每行都有两个字段 created_by
deleted_by，创建的时候会记录用 created_by 记录 txid，deleted_by 为空。删除的时候在 deleted_by 记录
txid，但是不会真删。过一段时间，确认没有事务在访问被标记删除的数据后，垃圾回收进程负责清理其空间。
而一个更新实际上是翻译成一个删除和新建过程

##### Visibility rules for observing a consistent snapshot

当事务从数据库读取数据的时候，利用 transtion id
决定哪个数据对其可见，通过小心定义可见规则，数据库就能实现对应用层的一致性快照:

-   每个事务开始前，数据库会列出所有其他事务（同一个进程中未提交或被 aborted ），任何这些事务的写入都被忽略
-   忽略所有被 aborted 的事务的写入
-   所有被更大的 事务 ID 写入的数据忽略，不管其有没有提交
-   所有其他写入对应用的查询可见

回到之前的例子，基于第三条规则，用户读取还是之前的 500 元，不会看到有 100 丢失。

##### Indexes and snapshot isolation

append-only B-trees,每个写入事务创建一个新的 B-tree root

##### Repeatable read and naming confusion

Snapshot isolation 在不同数据库实现用了不同的名称，不要混淆其含义，Oracle 叫做 serializable，PostgresSQL and Mysql 叫做 repeatable read

### Preventing Lost Updates

前面讨论的两种隔离都是保证在并发写下如何保证读取事务能看到的数据，忽略了两个并发写事务的会遇到的情况，最广为人知的问题
是 "lost update" 问题，经常出现在 读取-修改-写回 场景。如果两个事务并发写，就可能会导致一个丢失。比如下面场景：

-   增加计数器或者更新一个账户余额
-   更改一个复杂结构。比如 json 文档数据
-   两个用户更新同一个 wiki 页面

##### Atomic wirte operations

大部分数据库提供了更新操作来替换 读取-修改-写入 循环，比如下面的操作在大多数数据库是并发安全的：

`UPDATE counters SET value = value + 1 WHERE key='foo'`;

通常通过加互斥锁的方式来实现，当读取的时候保证没其他的事务能读取，直到更新应用到数据(cursor
stablility)，其他实现方式是把操作放在单线程中执行。

##### Explicit locking

在应用层代码里加锁显示避免竞争写入（其实感觉用队列更好一些）

##### Automatically detecting lost updates

数据库提供丢失检测功能，不需要应用层代码处理，不过 Mysql InnoDB repeatable read 没提供更新丢失检测

##### Compare-and-set

compare-and-set operation，（和数据库实现有关，不保证一定安全k）

`UPDATE wiki_page SET content='new content' where id=123 AND content='old_value'`

### Write Skew and Phantoms

考虑一个问题，有一家医院有医生值班，要求至少有两个医生值班，Alice 和 Bob 碰巧是值班的，但是他俩碰巧又都生病了，
于是几乎同事申请了调休。当系统检测是否至少有两个医生值班的时候，由于数据库使用了快照隔离，结果两个人都返回了2，
两个人都请假了，现在没有一个人值班了。你的任务就是要保证至少一个医生值班

##### Characterizing wirte skew

上边举例的反常情况叫做 write skew。两个事务同时更新两个不同的对象，算不上冲突，但是确是竞争条件。
可以用显示在 transaction 中加锁的方式解决。FOR UPDATE 会告诉数据库锁住查询返回的所有行

    begin transaction;
    select * from doctors where on_call = true and shift_id =123 for update;

以下一些场景也会出现：

-   Meeting room booking system: 限定一个会议室同一个时间只能被预订一次
-   Multiplayer game: 两个用户把相同棋子放在同一个位置
-   Claiming a username: 两个用户试图在同一个时间点创建相同用户名的账户
-   Preventing double-spending: 限定用户花费不能超过限定额度，你可能会享用余额减去花费判断余额是否为正数来实现，当有并发写入的时候就会有问题

他们有一些类似特点：

-   一个 select 语句检索出见过检查是否满足条件
-   根据上一步检查决定是否继续  (这一步实际上如果有并发写入，先验条件并不一定能成立)
-   满足条件，执行写入操作

phantom(幽灵): 当在一个事务中的写入改变了另一个事务的查询结果，叫做 phantom。为了解决这个问题，使用 Serializability 隔离

### Serializabitliy

Serializable isolation (序列化隔离)通常是最强的隔离等级，保证事务即使是并行(parallel)执行其结果就像是按照顺序来执行，
数据库组织了所有可能的竞争条件。大多数数据库使用三种方式来实现 Serializable isolation:

-   按照顺序逐个执行事务
-   两阶段锁 (Two-phase locking)，这几十年来是唯一的可行选择
-   优化的并发控制技术比如 serializable snapshot isolation

##### Actual Serial Execution

最简单的禁止并发问题的方式就是完全禁止它：单线程一次只执行一个事务。但是即使这个方案很明显，也是直到 2007
年左右才开始真正实现（过去30年为了效率都是使用多线程），让单线程成为可能的原因主要有两点：

-   RAM 更廉价了，使得把整个活动数据库存储在内存成为可能，当事务需要访问内存中的数据时，事务相比从硬盘载入数据执行更快
-   数据库设计者意识到 OLTP 事务通常很短并且只需要很少的读写操作，LTAP 分析通常只读，所以可以在一个一致快照上执行（脱离序列执行循环）

实现了序列执行事务的有 VoltDB/H-Store, Redis, Datomic. 通常由于不需要锁，单线程执行并发能力会更好，但是存储会被单核 CPU
限制，所以一般事务需要被结构化。

##### Encapsulating transactions in stored procedures

单线程事务处理中，一般不允许交互式多语句事务。 为了提高效率，单线程事务中一般会把整个事务代码同时作为一个
存储过程(stored procedure)提交，所有事务需要的数据都在内存里，存储过程的执行由于没有任何网络和磁盘IO，执行非常快。
实际上存储过程早就存在于关系型数据库中，但是一直没有好名声：

-   每个数据库都为存储过程实现了自己的语言，缺少统一的生态库
-   难以控制，难以 debug
-   容易出错

当然现在的实现使用了通常的编程语言，比如 Redis 使用 lua。通过存储过程和内存中数据，才使得单线程执行所有的事务成为可能，由于不需要等待，单线程也能实现高吞吐。

###### Partitoning

单线程序列化执行事务虽然简单，但其吞吐量受限制于CPU 核心的速度，为了扩展到多 cpu 核心，多节点，可以考虑分片。
但是夸分片事务速度会严重受限，需要结合应用层来灵活控制。

##### Two-Phase Locking (2PL)

过去30年，只有一个被广泛在数据库中序列化使用的算法：2PL (not 2PC)。之前讲过使用锁来阻止 dirty
writes，如果两个事务并发写同样的数据，通过锁来保证一个写完了再写另一个。2PL
类似，但是锁条件更严格。对一个对象，只有没有任何写入的时候才允许几个事务并发读。但是一旦有人想写入这个对象，就需要互斥访问：

-   如果事务 A 读过了对象并且事务 B 想写入它，B 必须等 A commit 或aborts了才能继续（B 不能在A背后偷偷修改它）
-   如果事务 A 写入了一个对象并且 B 事务想读取它，B 必须等待 A commit 或者 aborts 了才能继续（2PL 下不允许读旧版本的对象）

2PL 中，写入者不会阻塞其他写入者，他们只阻塞 readers，反过来也是。快照隔离中是 reders 不会阻塞 writes，writes 也不会阻塞
readers。因为 2PL 提供了序列化能力，阻止了之前所说的所有竞争条件(lost updates and write skew)。

###### Implementation of 2PL

通过锁来实现，锁可以是共享模式或者互斥模式，被这样使用：

-   如果一个事务想读取一个对象，必须现在共享模式下获取锁，不同事务允许共享模式下同时获取锁
-   如果一个事务想写一个对象，必须先在互斥模式下获取锁，其他事务不能同时获取（只能wait）
-   如果一个事务先读之后写对象，它会升级共享模式到互斥模式
-   当一个事务获取了锁，它必须持续持有锁直到事务结束(commit or abort)。这就是两阶段的含义，第一个阶段（事务执行期间）锁被获取，第二阶段（在事务结束时）等待所有的锁被释放。

使用了这么多锁，就会出现一个事务 A 一直等待事务 B 去释放锁，B 同样等 A 释放（死锁）。数据库会自动检查死锁然后 abort
掉一个，被 abort 的食物需要应用层代码重试。

###### Performance of 2PL locking

性能和延迟问题（尽量保持事务短小、或使用队列）

##### Serializable Snapshot Isolation(SSI)

SSI(2008年才提出) 提供了完整的序列执行，相比 快照隔离只有很小的性能劣势，在单节点和分布式数据库都有使用。

###### Pessimistic versus optimisitic concurrency control

2PL 使用了叫做悲观锁的控制机制，基于以下原则：如果有任何可能出错，做出任何更改之前最好等待情况安全了再去做， 用于在多线程环境中保护数据的安全性。
Serial 执行是一种极端的悲观控制：等价于每个事务都有一个互斥锁锁住整个数据库，通过让事务快速执行来弥补。
Serializable snapshot isolation
是一种乐观控制技术，与其阻塞一些危险的操作发生，继续事务的执行，寄希望于所有事情能正确执行。当一个事务想要
commit，数据库检查是否有  anything bad 发生，如果有，事务被 abort 然后需要被重试。只有序列执行的事务允许 commit。

##### Decisions based on outdated premise

premise: 事务基于 premise(在事务开始之前为真的 一个条件) 来行动。为了提供 serializable isolation，数据库必须检测一个事务是否在一个过期的 premise 上操作并在这种情况下 abort 事务
有两种方式检测一个查询结果是否过期:

-   检测一个过期的 MVCC 对象版本( 读之前有未提交的写操作)
-   检测写操作影响之前的读取（写操作发生在读之前）

# 8. The Trouble with Distributed Systems

### Faults and Partial Failures

partial failutre: 分布式系统中经常出现一部分系统以一种不可预知的方式出错。
未决的 (同样的操作产生不同的结果)的部分出错的可能造成分布式系统很难搞。

##### Unreliable Networks

如果一个发送了请求但是没有回应，无法区分到底是请求丢失、远程节点挂了还是响应丢失了。

##### Network Faults in Practice

你需要知道你的软件对于网络问题的反应，确保能恢复过来。

##### Detecting Faults

很多系统都需要自动检测失败节点，例如：

-   负载均衡检测失败节点并且停止发送请求
-   一主多从架构中，一个 leader 跪了，应该有一个 follower 提升为一个新的 leader

##### Timeouts and Unbounded Delays

如果唯一检测失败的方式就是根据超时时间，如何确定时间多长呢？这不是个简单问题，如果超时时间太短误认为节点挂了，
可能会造成重复调用。并且如果在网络负载高的情况下转移负载，可能会有级联失败，最极端的情况是所有节点都认为其他节点挂了，每个节点都不工作了。
一般可通过测试得到一个比较合理的时间，或者采用动态改变超时时间的方式(类似TCP重传）。

##### Synchronous Versus Asynchronous Networks

我们必须假定网络拥塞、排队、无限期延迟会发生，对于 超时，只能通过实验来确定。

### Unrealiable Clocks

大部分现代计算机使用两种时钟：

-   Time-of-day clocks : 时间戳，返回从 epoch 开始的秒数。需要从 NTP 服务器同步信息
-   Monotoinc clocks: 经常用来衡量时间区间（time interva），例如超时或者服务器响应时间

##### Relying on Synchronized Clocks

实际上，写入数据的时候，不同节点携带的时间戳在同步到同一个节点的时候是有可能出现问题的。比如后写入的数据却有更小的时间戳，
问题在于不同的节点时间戳并不一定能完全同步。
google 在每个数据中心都有原子钟，允许时钟同步时间控制在 7 ms

### Process Pauses

##### Response time guarantees

##### Limiting the impact of garbage collection

### Knowledge, Truth, and Lies

##### The Truth is Defined by the Majority

大部分分布式系统使用投票的方式选主

Fenching tokens: 每当一个节点从 lock server 获取到 契约的时候分配一个单调递增token，写入的时候server 会记住 token号，
并且禁止 token 小于之前写入的 node 携带的 token 写入。

### Byzantine Faults

之前讨论的 Fenching tokens 可以发现并且阻塞意外的操作（节点自己没发现契约过期），但是阻止不了故意伪造假的 token 的情况。
分布式系统难处理的一点就是节点可能撒谎，比如声称收到了一个特定消息但是实际上却没有，这种行为成为拜占庭故障，
解决这种在不信任的环境中达成一致的问题叫做『拜占庭将军问题』(<https://yeasy.gitbooks.io/blockchain_guide/content/distribute_system/bft.html>)

### System Model and Reality

-   Synchronous Model: bounded network delay,  bounded process pauses, bounded clock error （错误不会超过上界）
-   Partially Synchronous model: Synchronous most of time
-   Asynchronous Model: 不对时间做任何假设

三种最常见的节点故障：

-   Crash-stop faults: 算法假定节点只会因为一种方式失败:crashing。这意味着节点停止响应后就永远不再响应了
-   Crash-recovery faults: 节点一定时间内没响应，之后会恢复。节点应该有稳定的持久化存储保护 crash，内存中的状态可以丢失
-   Byzantine(arbitrary) faults: 节点可能做任何，比如欺骗其他节点

###### Correctness of an algorithm

在传统算法比如排序算法中，我们可以通过定义proterties（特性）来验证一个算法正确性，比如排序算法输出的每个相邻元素中，左边的都小于右边的元素。
我们就认为它是正确的。分布式算法可以类似定义，我们需要一个算法具备如下特性：

-   Uniqueness: 没有两个请求 fencing token 会得到相同的值
-   Monotonic sequence: 如果请求 x 返回 tx, y 返回 ty，并且 x 在 y 开始前完成，那么 `tx < ty`
-   Availability: 一个请求 fencing token 的节点在最终收到响应前不会crash

# 9. Consistency and Consensus

建立容错系统最好的方式是 寻找一些通用(general-purpose)的，包含有用保证的抽象，实现他们，让应用依赖这些保证。
比如，分布式系统最重要的一抽象就是 consensus(一致性)，让所有节点达成一致。

### Consistency Guarantees

大部分数据库复制的实现都至少保证了最终一致性，但是很多问题只有在系统错误或者高并发场景下才会暴露，而且很难用测试保证。 本章将讨论几种分布一致性模型。

### Linearizability(Atomic consistency)

最终一致性实现中，如果从不同副本同时请求数据，可能得到不同结果，能不能使用一种方式就好比只从一个节点读数据呢？
Linearizability (atomic consistency, strong consistency, immediate consistency, external consistency) 思想就是这样。
当一个 client 完成写入，所有 clients 读取数据必须看到写入的值。

##### What Makes a System Linearizable ?

如果一个读请求和写请求并发执行，它可能读到新值或者旧值。

Serializability: 事务的一种隔离属性，每个事务可能读写多个对象。它保证事务表现地像是按照顺序执行，这个顺序可以和实际的执行顺序不一样

Linearizability: 对一个单独对象读写的就近保证(recency guarantee on reads and writes of a register(an individua
object))。它不会把操作组合成一个事务，无法保证 write skew 问题

##### Relying on Linearizability

###### Locking and leader election

在 single-leader 架构中，需要确保只能有一个 leader，一个选举的方式是通过一个锁实现，一开始抢到锁的是 leader。
Apache ZooKeeper, etcd 经常用来实现分布式锁和 leader 选举。

##### Constraints and uniqueness grarantees

在有唯一性限制的数据库并发写入相同的数据，需要 Linearizability

##### Cross-channel timing dependencies

考虑一种场景，图片处理。客户端上传图片到 web server，web 服务器存储图片到存储服务，发送一条消息，之后图片裁剪服务会裁剪图片，如果消息发送和处理很快，图片裁剪服务可能拿不到最新的图片处理
这种问题出现是因为在 web server 和 裁剪服务之间存在两种不同的沟通渠道: 文件存储和消息队列，如果没有 linearizability
保证，两种渠道就会出现竞争条件。

##### Implementating Linearizable Systems

Linearizability 意味着：表现的数据像是只有一份拷贝。，并且所有在它上面的操作都是原子的。
通常让系统容错的方式是使用副本(replication)

- Single-leader replication (potentially, linearizable): 
一主多从架构中只有 leader 能写入，如果你从 leader 读，或者从同步更新的 follower 读，它们潜在是 linearizable

- Consensus algorithms (Linearizable)
通过一致性算法实现类似一主多从架构

- Multi-leader replication (not linearizable)
通常多主架构不是 Linearizable 的， 因为多主架构并发处理多个节点的写入需求并且异步复制到其他节点，可能造成冲突写入并且需要人工修复，无法实现 "single copy of data"

- Leaderless replication (probably not Linearizable)
无主架构中有些人声称可以通过投票读取和写入`(w+r>n)`的方式实现强一致性，但是依赖 w,r配置。

##### Linearizability and quorums
直觉上 strict quorum 读写能保证 linearizable, 但是当有网络延迟的时候，可能会出现竞争条件。


##### The Cost of Linearizability

CAP 定理：

- 如果你的应用需要 linearizability, 并且其中一些副本因为网络问题失去联系了，这些节点无法继续处理请求，他们必须等待网络恢复或者返回错误(become unavailable)
- 如果你的应用无需 linearizability，它可以用这样的方式写入：每个副本可以独立处理请求，即使跟其他副本失联。这样就能在网络问题时依然保持可用(available)，但是这种行为是 非 linearizable 的。

因此，应用无法在提供 linearizability 的情况下因为网络问题容错(CAP 理论)

CAP 经常被表示成 Consistency, Availability, Partition
tolerance，3个里只能实现俩。这种说法是误导人的，因为网络隔离是一种故障，他们始终会发生，你无法选择。
当网络正常工作，一个系统可以提供 consistency(linearizability) 和 完全的 availability，因此表述 CAP
定理的一种比较好的方式是 "either Consistent or Available when Partitioned"。CAP
在现在的分布式系统中已经慢慢更精确的概念取代。

##### Linearizability and network delays
Attiya and Welch 证明如果想实现
linearizability,读写响应时间至少和不确定的网络延迟成正比。弱一致性模型可以更快，一般在要求低延迟的系统中会折中。


### Ordering Guarantees

##### Ordering and Causality

保持有序的其中一个原因就是保持因果性。比如保证消息在接受之前是发送了的，问题是出现在回答之前的。
如果一个系统遵守因果顺序，叫做 causally consistent.

##### The causal order is not a total order
- total order(全序): 允许比较。比如自然数能比较大小
- partially ordered（偏序）: 比如 set，有些场景可以比较，比如一个集合包含另一个的所有元素，但是其他场景无法比较。

从这两个排序概念看数据一致性模型：

- Linearizability: linearizable 系统操作可以是全序的：系统表现像是只有一份数据拷贝，并且每个操作都是原子的，意味着两个操作我们可以说一个是在另一个之前发生的。
- Causality: 两个事件是有序的如果它们因果相关，但是如果它们是并发的就是无法比较的。因果性定义了偏序关系，而不是全序。(想象git分之合并)

##### Linearizability is stronger than causal consistency
Linearizability implies causality: 任意系统只要是 linearizable 将保持因果正确，但是会破坏性能和可用性。

##### Capturing causal dependencies

为了维持因果性，你需要知道哪个操作在另一个操作之前发生.通常使用 version vectors


### Sequence Number Ordering
尽管因果性是一个重要理论概念，但是跟踪因果依赖几乎不可能。我们可以用 sequence numbers or timestamps
来排序事件，时间戳使用逻辑时钟，就是一种用来唯一标识数字的算法，通常对每个操作使用递增的计数器(全序的)。

##### Noncausal sequence number generators
如果是一主多从架构，一般有这么几种方式生成序列号：
- 每个节点可以独立生成自己的序列号。比如两个节点一个生成奇数一个生成偶数
- 绑定时间戳
- 预分配序列块

但是这几种方式都不是因果一致的，不同节点无法区分顺序

##### Lamport timestamps

Leslie Lamport 1978 年提出了一个用来生成因果一致的序列号的方法：每个节点和 client
跟踪它至今看到的最大的计数器值，然后每个请求都带上它，当一个节点接收请求或者响应的时候，如果看到其计数器值比自己的大，
就把节点自己计数器增加到该值。

##### Timestamp ordering is not sufficient

尽管 Lamport timestamps 能保持因果一致，但是在分布式系统中是不够的。
为了实现唯一性限制（比如姓名），仅仅有全序的操作关系是不够的，还需要知道什么时候被最终确定

### Total Order Broadcast
单主架构中通过选一个节点为 leader 并且在一个节点的单核 cpu
上处理序列号问题，从而实现全序关系。挑战是如果一个节点处理能力不够或者leader挂了怎么办。分布式系统称这种问题叫做
全序广播(total order broadcast) 或者 原子广播(atomic broadcast)，通常需要两个安全属性：
- Reliable delivery : 没有信息丢失，如果一个消息发送给了一个节点，就被发送到所有节点
- Totally ordered delivery: 消息被以同样的顺序发送到所有节点

##### Using total order broadcast
一致性服务 ZooKeeper 和 etcd 实现了全序广播。全序广播可以用来实现 serializable transaction, lock service provides
fencing tokens

##### Implementating linearizable storage using total order broadcast

可以通过全序广播 append-only log 实现 linearizable compare-and-set 操作：
- 向 log 发送信息，临时性声明你想写入的数据（比如 username）
- 读取 log，等待消息回执
- 检查所有声明了你希望的 username 的消息，如果第一条消息是你自己的消息就算成功，可以 commit，如果是来自其他用户的就
  abort 操作。

##### Implementating total order broadcast using linearizable storage

最简单的方法是假设你有一个 linearizable 存储一个整数并且可以执行原子的 increment-and-get
操作。算法如下：对每个想通过全序广播发送的消息，递增然后获取一个整数，让消息附带上该值，然后把消息发送到所有节点，接受者将会根据
序列号投递消息。如果一个节点投递了4号消息并且收到了一个6号消息，它知道在发送6号消息钱必须等待5号消息到来。Lamport
timestamps 不需要保证，这就是全序广播(total order broadcast和时间戳排序(timestamp ordering)的区别。


### Distributed Transactions and Consensus

一致性问题看起来很简单：几个节点在某些事情上达成一致。在某些情况下一致性很重要：

- Leader election: 一主多从架构中节点需要保证只能有一个 leader（否则两个以上自认为是 leader 的就会被写入造成数据不一致）
- Atomic commit: 在跨节点或者分片的写入中，可能会出现一些事务只在其中一些节点成功了，如果需要保证ACID
  特性，需要保证所有节点不是 commit 就是 abort/roll back 了。这个问题叫做『atomic commit problem』


##### Atomic Commit and Two-Phase Commit(2PC)

###### From single-node to distributed atomic commit
在单个节点上，原子性通常由存储引擎实现。事务提交依赖于数据持久化到磁盘的顺序。
多个节点没法简单把事务发送到所有节点去执行，无法保证都能成功，就会打破原子性保证：

###### Introduction to two-phase commit
2PC: 两阶段提交是用来实现跨节点原子事务提交的一种算法，保证所有节点要么 commit，要么 abort。
2PC 使用了一个协调人(coordinator or transaction manager)，通常和应用进程实现在一起，也可以用单独的进程或者服务。包括
Narayana, JOTM, BTM, or MSDTC。
一个 2PC 事务从应用读取和写入多个数据库节点开始，在事务过程中我们称这些数据库节点叫做 参与者
participants。当应用准备好 commit，协调人准备阶段1：它向每个节点发送一个 prepare 请求，询问他们是否准备
commit。然后协调人跟踪每个节点的响应：

- 如果所有的参与者回复 yes，说明他们都准备好提交，然后协调人发送一个 commit 请求进入阶段2，最终提交行为发生
- 如果有一个参与者回复 no，协调人发送一个 abort 请求让所有节点进入阶段2.

这个过程和西方传统婚礼上，牧师分别问喜娘和新郎是否愿意嫁给彼此很像，通常都是两个人回复『我愿意』，牧师宣布两人喜结连理（commiteed），
然后好消息传给所有的出席者，否则如果有一个不愿意的，婚礼就 "abort" 了。


##### Coordinator failure
如果在参与者回复了 yes 之后 协调人挂了，该数据库节点就无法知道到底要 commit 还是 abort，只能等协调人恢复过来。

##### Three-phase commit
2PC 因为协调人可能出问题而等待协调人恢复，一般叫做 blocking commit protocol。这时候有人提出了 3PC，加上一个 perfect failure detector 
检测哪个节点 crash 了。但在一些网络可能无线延迟的场景，依然是使用 2PC


### Distributed Transactions in Practice
分布式事务尤其是实现了 2PC 的方式，一方面提供了重要的安全性保证，但是却带来运营问题、性能问题，很多云服务选择不去实现分布式事务。
两种分布式事务经常被混淆：
- Database-internal distributed transactions: VoltDB/Mysql Cluster's NDB 存储引擎支持这种事务。这种情况下所有参与事务的节点都在一个数据库软件下运行
- Heterogeneous distributed transactions: 参与者可能是多种技术，需要保证原子性提交。更难处理

##### Exactly-once message processing

原子性提交协议：

XA transactions:  X/Open XA(short for eXtended Architecture): 一种跨技术之间实现两阶段提交的标准。被很多数据库和消息中间件支持

分布式事务的限制：
- 协调者是单点的，默认会有可用性问题，或者只有一个原始的副本支持
- 很多服务端应用是无状态模型（http），如果协调人部署在应用里，会改变默认的部署方式，导致不再是无状态
- XA 只保证了很多数据系统的最基础实现，无法检测不同系统死锁，没有一种协议来解决不同系统之间的冲突问题
- 分布式事务有错误放大的问题，如果系统部分节点失败，就会导致所有事务都失败。需要容错机制

### Fault-Tolerant Consensus
一个一致性算法必须满足以下这些特性：
 - Uniform agreement: 没有两个节点决定不一致
 - Integrity: 没有节点做两次决定
 - Validity： 如果一个节点决定一个值 v，v 是被某些节点提议的
 - Termination：每个不 crash 的节点能最终决定一些值

##### Consensus algorithms and total order broadcast
最知名的容错一致性算法包括 Viewstamped Replication(VSR), Paxos, Raft, Zab，他们有相似性但是又不完全相同。
大部分这些算法没有直接用上边描述的模型（提议和决定某个值，并且满足 agreement, integrity, validity and termination
属性），而是在一系列值上做决定，使它们成为了全序广播算法(total order algorithm)。
全序广播需要消息只被发送一次，使用相同的顺序发送到所有的节点。等价于做了几轮一致性：每一轮节点提议它们接下来要发送的消息，
然后以全序关系决定下一个要发送的消息。
所以，total order broadcast 和 repeated rounds of consensus 等价，每一轮一致性决定对应一个消息发送。

- 根据 agreement 属性，所有节点决定以相同顺序发送同一个消息
- 根据 integrity 属性，消息不会重复
- 根据 validity 属性，消息不会出错或者伪造
- 根据 termination 属性，消息不会丢失

Viewstamped Replication(VSR), Raft, Zab 直接实现了 total order broadcast，因为比重复多轮一致性 更高效。
Paxos 实现了 Multi-Paxos 的优化

##### Single-leader replication and consensus
一主多从架构中只有 leader 写入，其他 followers 以相同写入顺序更新，是否就不用担心一致性问题了？
取决于 leader 如何选取的。如果是认为配置的，并且当 leader 挂了还需要人手动改配置，就不满足
Termination，因为需要人工干涉。

##### Epoch numbering and quorums
之前讨论的一致性协议都需要一个 leader，但是不保证 leader 是唯一的。而是弱保证：协议定义一个 epoch 数字(
ballot number in Paxos, view number in Viewstampted Replication, term number in Raft)，保证在每个 epoch
leader 是唯一的。
每次一个 leader 被认为挂了，其他节点就会选举 leader。每次选举都会给一个单调递增的 epoch number, 如果在两个不同的 epoch
期间有两个 leader 冲突（比如一个 leader 并没有挂）拥有更大的 epoch number 的leader 胜出。

##### Limitations of consensus
一致性算法不是应用在任何地方，收益都是有代价的：
- 一致性系统需要少数服从多数的操作，一般至少要三个节点来容忍一个失败，或者至少5个节点容忍2个失败
- 大多数一致性算法需要固定节点参与投票，意味着不能随意增删节点
- 一致性算法通常用超时探测失败节点，对于有网络延迟的系统可能造成频繁选主而不是真正做有用的工作
- 有时候一致性算法对网络问题很敏感


### Membership and Coordination Services
ZooKeeper 和 etcd 经常被形容为 "分布式k-v存储" 或者 "协调和配置服务"，可以通过 key 读写数据以及遍历 key。
应用层代码一般不会直接使用 ZooKeeper, 而是间接通过其他项目，比如 Hbase，Hadoop YARN, Openstack Nova, Kafka 都依赖
Zookeeper 在背后跑。Zookeeper 不仅实现了 order broadcast (and hence consensus)，还实现了分布式系统其他一些有用的特性：

- Linearizable atomic operations: 使用原子的 compare-and-set 操作，可以实现锁
- Total ordering of operations: ZooKeeper 通过全序所有操作（给一个单调递增的 事务 ID(zxid) 和 版本号(cversion)）实现了
  fencing token.
- Failure detection
- Change notification

##### Allocating work to nodes

##### Service discovery
ZooKeeper, etcd, Consul 经常用作服务发现：为了使用哪个服务需要连接到哪个 IP

##### Membership services
membership 服务可以用来决定集群中哪些是能工作的活跃节点。

