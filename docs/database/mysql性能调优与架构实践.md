# 2. Mysql 基础

物理文件组成：
日志：
    - 错误日志
    - 二进制文件
    - 更新日志。5.0 以后废弃
    - 查询日志
    - 慢查询日志
    - InnoDB 在线 REDO 日志
数据文件：
    - ".frm" 文件：与表相关的元数据文件
    - ".MYD" 文件: 存在 MyISAM 表的数据
    - ".MYI" ： 存放 MyISAM 索引相关信息
    - ".ibd"和 ibddata: 存放 InnoDB  数据文件


Replication相关文件：
    - master.info 存放 slave的 master 端相关信息
    - relay log 和 relay log index：用于存放slave 端的I/O 线程从master读取的binary log情况
    - relay-log.info 存放slave 的IO线程写入本地的relay log相关信息

其他文件：
    - system config file
    - pid 存放进程id
    - socket文件：unix/linux 下可以不通过TCP/IP 直接用unix socket 链接mysql

Mysql Server系统架构： sql layer -> storage engine layer


# Mysql 存储引擎简介

- MyISAM 存储引擎：.frm, .MYD(表数据), .MYI(索引数据)。 支持三种索引：B-Tree, R-Tree, Full-text
- InnoDB: 事务安全；数据多版本读取；锁定机制改进（实现了行锁）；实现了外键。物理结构：数据文件；日志文件
- NDB Cluster：
    - 负责管理各个节点的 master 主机
    - SQL 层的 SQL 服务器节点（Mysql server）
    - Storage 层的NDB 数据节点，即 NDB Cluster

# 4 Mysql 安全管理
相关因素：外围网络，主机，数据库（用户管理、访问授权）
权限等级：GlobalLevel, Database Level, Table Level, COlumn Level, Routine Level

# 5 Mysql 备份与恢复

逻辑备份：
    - 生成 INSERT 语句备份: mysqldump
    - 生成特定格式的纯文本备份数据文件

物理备份：备份相应的数据文件
热物理备份：
    - MyISAM: mysqlhotcopy
    - InnoDB: ibbackupk

# 6 性能影响因素
- 商业需求对性能影响：
    - 不合理需求导致投入产出过低
    - 无用功能堆积导致系统复杂
- 系统架构及实现对系统的影响
    - 是否适合：二进制多媒体、流水队列、超大文本不适合
    - 是否应该用 cache：各种配置和规则；活跃用户基本信息；准实时的统计信息数据；访问频繁但是变更较少的数据；
    - 数据层实现
    - 过度依赖query语句功能造成操作效率低下
- query语句对系统性能影响
- Schema设计对系统性能的影响
- 硬件环境对系统性能影响
    - OLTP
    - OLAP

# 7 MySQL 数据库锁定机制
锁定机制的优劣直接影响到数据库的并发处理能力和性能
- 行级锁定(row-level)：颗粒度小，并发能力强；但是上锁和释放操作更多，消耗大，易死锁
    - 共享锁和排他锁。为了在锁定的过程中让行级锁定和表级锁定共存，InnoDB 同样使用了意向锁（表级锁定）的概念，意向共享锁和意向排他锁
    - 间隙锁：Innode的锁定是通过在指定数据记录的第一个索引键之前和最后一个索引键之后的空域空间标记锁定信息实现的。



- 表级锁定(table-level): 很好处理死锁问题；锁定资源争用概率高，并发不好
    - 分为读锁定和写锁定，主要通过4个队列来维护两种锁定：两个存放当前正在锁定的读和写锁定信息，另外两个存放等待中的读写锁定信息
        - Current read-lock queue (lock-read)
        - Pending read-lock queue (lock->read_wait)
        - Current write-lock queue (lock->write)
        - Pending write-lock queue (lock->write_wait)
    - 读锁定：一个新客户端申请获取读锁定资源的时候，需要满足两个条件：1.请求锁定的资源当前没有被写锁定；2写锁定等待队列中没有更高优先级写锁定等待
    - 写锁定：

- 页级锁定(page-level): 介于行锁和表锁之间。

  使用表级锁定的主要是MyISAM、Memory、CSV等非事务性存储引擎，行锁主要是 InnoDB 和 NDB Cluster 存储引擎。
  页级锁定主要是 BerkeleyDB 存储引擎。

- 合理利用锁机制：
    MyISAM:
    - 缩短锁定时间；（拆分复杂query；高效索引；控制字段信息，只存储必要信息；优化表数据文件）
    - 分离能并行的操作：Concurrent Insert (并发插入特性)
    - 合理利用读写优先级

    InnoDB:
    - 尽可能让所有的数据检索通过索引完成，避免因为无法通过索引键加锁而升级为表级锁定
    - 合理设计索引
    - 尽可能减少基于范围的数据检索过滤条件，避免间隙锁带来的负面影响锁定了不该锁定的记录
    - 控制事务大小，减少锁定的资源量和锁定时间长度。
    - 尽量用级别较低的事务隔离
    减少InnoDB死锁的建议：
    - 类似业务模块尽量用相同顺序访问
    - 同一个事务中，尽可能做到一次锁定需要的所有资源
    - 升级锁定颗粒度

# 8 Mysql数据库Query优化
- Query Optimizer:
- Query 语句优化的基本思路和原则：
    - 1.优化更需要优化的query。高并发，低消耗的query 对整个系统的影响远比地并发高消耗大
    - 2.定位优化对象的性能瓶颈，io 还是 cpu，profiling 工具
    - 3.明确的优化目标。目标范围
    - 4.从 Explain 入手。不断调试然后用 explain 验证调优结果
    - 5.用小结果集驱动大结果集。减少循环嵌套中的循环次数，以减少 IO 总量及 CPU 的运算次数。
    - 6.尽可能在索引中完成排序（利用索引进行排序操作主要是利用了索引的有序性）
    - 7.只取出自己需要的columns。带宽浪费
    - 8.仅仅使用最有效的过滤条件。过滤条件多不一定是好事
    - 9.尽可能避免复杂的join和子查询
- 充分利用 Explain 和 Profiling
- 合理设计并利用索引：mysql四种索引：
    - B-Tree索引：所有实际需要的数据都是放在Tree 的 Leaf Node，而且到任何一个 Leaf Node 的最短路径的长度都是完全相同的。
        - B+Tree: 在每一个 leaf node上除了存放索引键的相关信息外，还存储了指向与改 Leaf Node 相邻的后一个 Leaf Node
        的指针信息，主要是为了加快检索多个相邻的 Leaf Node 的效率。
    - Hash 索引：通过一定的hash算法，将需要索引的键值进行hash运算。然后将得到的hash值存入一个hash表中。
    - Full-text 索引：仅有MyISAM支持，仅支持CHAR,VARHCAR,TEXT.
    - R-Tree 索引：空间数据检索问题。 GEOMETRY 类型
- 索引利弊：
    - 利:提高检索效率,降低IO成本。降低排序成本
    - 弊：增加了更新所带来的IO和调整索引的计算量，存储空间资源消耗
- 如何判断是否需要创建索引：
    - 较频繁的作为查询条件的字段应该创见索引
    - 唯一性太差的字段*不*适合单独创建索引，即使频繁作为查询条件。比如状态字段等
    - 更新非常频繁的字段不适合创建索引。查询比更新频繁
    - 不会出现在where子句中的字段不该创见索引
- 单键索引还是组合索引：
    - 前缀索引：仅仅适用于前缀随机重复很小
    - 1.对于单键索引，尽量选择针对当前Query过滤性更好的索引
    - 2.在选择组合索引时，当前Query中过滤性最好的字段在索引字段顺序中排列越靠前越好
    - 3.在选择组合索引时，尽量选择可以包含当前Query的WHERE子句中更多字段的索引
- mysql中索引的限制：
    - BLOB 和 TEXt 类型的列只能创建前缀索引
    - 使用 != 的时候，mysql 无法使用索引
    - 过滤字段中如果使用了函数运算，mysql无法使用索引
    - 使用 like 操作的时候如果条件以通配符开始(%abc....)无法使用索引
- Join 的实现原理和优化思路:
    - mysql中只有一种join算法，就是Nested Loop Join。通过驱动表的结果集作为循环基础数据，然后将该结果集中的数据作为过滤条件一条条地到下一个表中查询数据，最后合并结果
- Join 语句优化：
    - 1.尽可能减少Join语句中的 Nested Loop 的循环总次数。小结果集驱动大结果集，减少循环次数
    - 2.优先优化 Nexted Loop 的内层循环
    - 3.保证join语句中被驱动表的join条件字段已经被索引
    - 4.当无法保证被驱动表的Join条件字段被索引且内存充足，可以适当增大Join Buffer的设置

- Order by 语句优化:
    - 两种类型的 order by。一种是通过索引直接取得有序数据；mysql排序算法将返回的数据进行排序
    - 1.加大max_length_for_sort_data 何止
    - 2.去掉不必要的返回字段
    - 3.增大 sort_buffer_size 设置

- Groupby 的实现和优化：
    - 三种实现方式：
        - 使用松散(Loose)索引扫描实现 groupby
        - 使用紧凑(Tight)索引扫描实现 groupby
        - 使用临时表实现 groupby

- Distinct 的实现与优化
    - 与groupby相似，只是在groupby 之后的每组中只取一条记录而已
    - 尽量利用索引，如果无法利用索引，确保尽量不要在大结果集上用distinct操作。

# 9 Mysql数据库 schema 设计性能优化
- 高效的模型设计：
    - 1.适度冗余，让Query尽量减少 Join
    - 2.大字段垂直拆分。将自己的字段拆分到单独的表里.访问频率低
    - 3.大表水平拆分，基于类型的分拆优化
- 合适的数据类型：
    - 更小的数据类型和合适的类型
- 规范的对象命名
    - 数据库和表名尽可能和业务模块名一致
    - 同一子模块采用子模块为前缀或者后缀
    - 索引名包含字段缩写

# 10 Mysql server 性能优化
- 安装优化
- 日志设置优化
- Query cache 优化
- 其他优化：网络链接、线程管理、table管理
    - mysql 每个线程独立打开自己的文件描述符。mysql Table Cache
    -
# 11 常用存储引擎优化
- MyISAM 引擎优化（读请求为主的非事务系统）
    - 索引缓存优化
    - 多 key cache 的使用
    - key cache 的 mutex问题
    - key cache 预加载
    - NULL 值的处理方式 myisam_stats_method 收集统计信息的时候认为 NULL值是相同的还是不同的
    - 表读取缓存优化
    - 并发优化（表级锁定，读写互斥）

- InnoDB 引擎优化(4个主要区别 缓存机制、事务支持、锁定实现、存储方式的差异)
    - 缓存相关优化：innodb_buffer_pool_size
    - 事务优化: 合适的事务隔离级别
        - READ UNCOMMITTED: dirty reads（脏读），最低隔离级别。非一致性读(non consistent reads)
        - READ COMMITTED: 语句级别的隔离。不会出现dirty read，可能有non-repeatable-reads(不可重复读)和phantom reads(幻读)
        - REAPEATABLE READ: innodb默认事务隔离级。仍存在 Phantom Reads 可能性
        - SERIALIZABLE: 最高级。事务中的任何时候看到的数据都是事务启动时的状态
    - 数据存储优化：数据和索引放在相同文件中。以 page 为最小物理单位存放，page默认 16KB。extent 是多个连续的 page
      组成的一个物理存储单位，一般为 64 个 page。segment（实际上代表files），由一个或者多个 extent
      组成，每个segment都存放同一种数据。tablespace最大物理结构单位，多个 segment 组成。
    - 分散 IO 提升磁盘响应:数据文件和事务日志文件存放在不同的磁盘

# 12 Mysql 可扩展设计的基本原则

- 可扩展性
    - scale out : 横向扩展，通过增加处理节点的方式
    - scale up：纵向扩展。增加当前节点的处理能力
- 事务相关性最小化原则
    - 分布式事务：
        - 合理设计切分规则
        - 大事务切分成很多小事务
        - 两种结合
- 事务一致性原则：保证最终一致性
    - BASE模型：基本可用、柔性状态、基本一致、最终一致
- 高可用及数据安全原则

# 13 。可扩展性设计之 Mysql Replication
 Replication 将一个server 的instance 中的数据完整复制到另一个 server 中。异步复制，延迟很小
- Relplcation 实现原理：
    - Replaction 线程
    - 复制实现级别：
        - Row Level: mysql的复制可以是基于一条语句（statement level）or 一条记录(row level)
        - Statement level:
        - Mixed level:
- Replication 常用架构
    - Master-Salves(常规复制架构): 一主多从
    - Dual Master（master-master）：两个master server互相将对方作为自己的master，自己作为对方的slave 来复制 (避免循环复制问题)
    - 级联复制架构（master-slaves-slaves...)：通过二级复制减少对master 的复制压力。延时风险
    - Dual master 与级联复制结合架构(master-master-slaves)
- Replication 搭建实现：
    - 1.准备工作：保证binary log打开；创建一个具有 replication salve 权限的账户
    - 2.获取master端的备份快照
        - (通过数据库全库冷备份；用LVM或者ZFS具有snapshot快照功能的软件进行热备份(停止写入))
        - 通过mysqldump客户端程序导出逻辑数据序
        - 通过现有某一个slave端进行『热备份』
    - 3.slave 端恢复备份『快照』
    - 4.配置并启动slave

# 14.可扩展性设计之数据切分

sharding：通过某种特定的条件，将存放在同一个数据库中的数据分散到多个数据库上面，分散单台设备负载。
两种切分模式：
    - 不同的表切分到不同主机上，成为垂直（纵向）切分.适合业务逻辑清晰的系统，耦合低。
    - 根据逻辑关系，将同一个表中的数据按照某种条件拆分到多台数据库上，成为水平拆分。
    - 组合拆分：两种方式一起使用。先垂直、再水平拆分。

垂直拆分:
    - 按照功能划分。防止跨数据库的 join 存在
    - 根据实际应用场景平衡成本和收益

    缺点：
    - 部分表关联无法在数据库级别完成，要在程序中完成
    - 事务处理相对复杂
    - 切分达到一定程度后，扩展性受限制
    - 过度切分造成系统过于复杂难以维护

水平切分:
    - 将某个访问极其频繁的表按照某个字段的某种规则分散到多个表中，每个表包含一部分数据

    优点:
    - 事务处理相对简单
    - 切分规则定义好，基本上较难遇到扩展性限制
    缺点：
    - 切分规则复杂， 很难抽象出一个满足整个数据库的切分规则
    - 后期维护难度增加，人工定位数据更难
    - 应用系统耦合度较高

- 垂直和水平切分联合使用
    - 一开始会选择垂直切分，成本小。后期加上水平拆分


- 自行开发中间件代理层
- MySQL Proxy 实现数据切分和整合 + lua 脚本
- 利用 Amoeba 实现数据切分和整合 Amoeba For Mysql
- 利用 HiveDB 实现数据切分和整合


数据库切分可能存在的问题：（尽量在程序层解决）
    - 引入分布式事务: 各个数据库解决自身的事务，然后通过应用程序来控制多个数据库上的事务。
    - 跨节点 join：通过应用程序处理，先在驱动表所在的Mysql server中取出驱动结果集，然后再去被驱动的server中取
    - 跨节点合并排序分页:程序并行获取数据


# 15章：可扩展性设计之Cache 和 Search 的利用
- 分布式缓存 memcachedcache。Lucene
- 利用 Search 实现高效全文检索


# 16章：Mysql Cluster
基于 NDB Cluster 存储引擎的的完整的分布式数据库系统

# 17章：高可用设计思路和方案
- 利用Replication 实现高可用架构
  - 常规 master-slave 架构实现基本的主备设计
  - 通过 dual master 解决master 故障
  - dual master 与 级联复制结合解决异常故障下的高可用
- 利用 mysql cluster 实现整体高可用
- 利用 DRBD 保证数据的高安全可靠

# 18章：高可用设计之 Mysql 监控
- 系统设计
  - 信息采集；分析；存储；处理
