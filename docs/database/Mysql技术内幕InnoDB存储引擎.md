# 1. Mysql 体系结构和存储引擎

 mysql 是单进程多线程架构，支持插件式的表存储引擎，存储引擎基于表而不是数据库。

### 1.3 MySQL存储引擎：

##### InnoDB
支持事务，面向在线事务处理(OLTP)应用，特点是行锁设计，支持外键，非锁定读。
使用多版本并发控制(MVCC)获得高并发性。
表中数据的存储，采用聚集(clustered)方式，因此每张表的存储都是按照主键的顺序存放。

##### MyISAM
不支持事务、表锁设计，支持全文索引，主要面向OLAP数据库应用。

##### NDB
集群存储引擎，join操作在数据库层完成，不是在存储引擎层，复杂连接需要巨大网络开销

##### Memory
存放内存中，适用于临时表，默认是哈希索引

##### Archive
只支持 insert select 操作，告诉插入和压缩功能


# 2. InnoDB 存储引擎

多线程模型：

- Master Thread: 主要负责将缓冲池中的数据异步刷新到磁盘，保证数据的一致性
- IO Thread: InnoDB中大量使用了AIO(Async IO)处理IO请求
- Purge Thread: 回收已经使用并分配的 undo 页
- Page Cleaner Thread: 将之前版本中脏页的刷新操作都放入到单独的线程中来完成，减轻master工作压力

（关于linux 内存管理：https://blog.csdn.net/gatieme/article/details/52384636 ,内存被细分为多个页面帧, 页面是最基本的页面分配的单位　）

内存：

- 缓冲池：缓冲池技术提升性能（协调磁盘与cpu速度鸿沟） innodb_buffer_pool_size，新版可以有多个缓冲池
- LRU List, Free List, Flush List: 管理内存区

Checkpoint 技术：
为了避免数据丢失问题，当前事务数据库系统普遍使用了Write Ahead Log 策略，当事务提交时，先写重做日志，再修改页。
当由于发生宕机导致数据丢失时，通过重做日志完成数据恢复。 (Durability持久性)
Checkpoint 技术解决了几个问题：

- 缩短数据库恢复时间
- 缓冲池不够用时，将脏页刷新到磁盘
- 重做日志不可用时，刷新脏页

当数据库宕机，数据库不需要重做所有日志，因为checkpoint 之前的页都已经刷新回磁盘，只需要对 checkpoint
后的重做日志进行恢复，这样就大大缩减了恢复时间。

InnoDB 通过LSN(Log Sequence Number) 来标记版本。
InnoDB 中两种Checkpoint:

- Sharp Checkpoint: 发生在数据库关闭时将所有的脏页都刷新回磁盘，默认
- Fuzzy Checkpoint: 只刷新一部分

Maste Thread 实现可以参考书中的示例伪代码

InnoDB 关键特性：

- 插入缓冲(Insert Buffer)
- 两次写(Double Write)
- 自适应哈希索引(Adaptive Hash Index)
- 异步IO(Async IO)
- 刷新邻接页(Flush Neighbor Page)

延伸阅读，主要是了解B树：
B-Tree: 多叉平衡查找树

- https://zh.wikipedia.org/wiki/B%E6%A0%91
- https://blog.csdn.net/v_JULY_v/article/details/6530142
