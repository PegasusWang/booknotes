# 2. 简单动态字符串

SDS: simple dynamic string，用作 redis 默认字符串表示，还被用作buffer，AOF 缓冲区和客户端状态中的输入缓冲区。
sds能减少缓冲区溢出漏洞。

SDS是二进制安全的，不仅可以保存文本，还可以是任意二进制数据

```c
struct sdshdr {
  int len; //保存字符串的长度，获取长度变成了 O(1)
  int free; //未使用的字节数量，实现了空间预分配和惰性空间释放两种优化策略
  char buf[]; //保存字符串
}
```

# 3. 链表

redis 实现的实际上是 无环的双端链表

```c
// adlist.h/listNode
typedef struct listNode {
  struct listNode *prev;
  struct listNode *next;
  void *value;
}listNode;

//adlist.h/list
typedef stuct list {
  listNode *head;
  listNode *tail;
  unsigned long len;
  void *(*dup)(void *ptr);
  void (*free)(void *ptr);
  int (*match)(void *ptr, void *key);
}list;
```

# 4. 字典
保存键值对，通过哈希表作为底层实现

哈希表定义：

```c
/* dict.h/dictht */
typedef struct dictht {
  dictEntry **table; //哈希表数组，表其实是个数组
  unsigned long size;
  unsigned long sizemark; //哈希表大小掩码，用于计算索引值总是等于 size-1
  unsigned long used;
}dictht;
```

哈希表节点 dictEntry，保存一个键值对

```c
typedef struct dictEntry {
  void *key;
  union {
    void *val;
    uint64_tu64;
    int64_ts64;
  }v;
  struct dictEntry *next;  // 解决key 冲突问题
}dictEntry;
```

字典的定义：

```c
typedef struct dict {
  dictType *type; // dictType 保存了用于操作特定类型键值对的函数,redis为不用的字典设置不同的类型特定函数
  void *privdata; //保存需要传给特定类型特定函数的可选参数
  dictht ht[2]; //ht属性是一个包含两个项的数组，数组中的每个项都是一个dictht哈希表，一般情况下，字典只使用ht[0]哈希表，ht[1]哈希表只会在对ht[0]哈希表进行rehash时使用。
  in trehashidx; //它记录了rehash目前的进度，如果目前没有在进行rehash，那么它的值为-1。
}
```


```c
typedef struct dictType {
  // 计算哈希值的函数
  unsigned int (*hashFunction)(const void *key);
  // 复制键的函数
  void *(*keyDup)(void *privdata, const void *key);
  // 复制值的函数
  void *(*valDup)(void *privdata, const void *obj);
  // 对比键的函数
  int (*keyCompare)(void *privdata, const void *key1, const void *key2);
  // 销毁键的函数
  void (*keyDestructor)(void *privdata, void *key);
  // 销毁值的函数
  void (*valDestructor)(void *privdata, void *obj);
} dictType;
```

当字典被用作数据库的底层实现，或者哈希键的底层实现时，Redis使用MurmurHash2算法来计算键的哈希值。
MurmurHash算法最初由Austin Appleby于2008年发明，这种算法的优点在于，即使输入的键是有规律的，算法仍能给出一个很好的随机分布性，并且算法的计算速度也非常快。

Redis的哈希表使用链地址法（separate chaining）来解决键冲突，每个哈希表节点都有一个next指针，多个哈希表节点可以用next指针构成一个单向链表，被分配到同一个索引上的多个节点可以用这个单向链表连接起来，这就解决了键冲突的问题。

rehash: 随着操作的不断执行，哈希表保存的键值对会逐渐地增多或者减少，为了让哈希表的负载因子（load
factor）维持在一个合理的范围之内，当哈希表保存的键值对数量太多或者太少时，程序需要对哈希表的大小进行相应的扩展或者收缩。渐进式的
负载因子= 哈希表已保存节点数量/ 哈希表大小
load_factor = ht[0].used / ht[0].size

# 5. 跳跃表
skiplist: 一种有序数据结构，它通过在每个节点中维持多个指向其他节点的指针，从而达到快速访问节点的目的。
Redis使用跳跃表作为有序集合键的底层实现之一，如果一个有序集合包含的元素数量比较多，又或者有序集合中元素的成员（member）是比较长的字符串时，Redis就会使用跳跃表来作为有序集合键的底层实现。
Redis只在两个地方用到了跳跃表，一个是实现有序集合键，另一个是在集群节点中用作内部数据结构，除此之外，跳跃表在Redis里面没有其他用途。

```c
// redis.h/zskiplistNode
typedef struct zskiplistNode {
  // 层
  struct zskiplistLevel {
    // 前进指针
    struct zskiplistNode *forward;
    // 跨度
    unsigned int span;
  } level[]; //每次创建一个新跳跃表节点的时候，程序都根据幂次定律（power law，越大的数出现的概率越小）随机生成一个介于1和32之间的值作为level数组的大小，这个大小就是层的“高度”。
  // 后退指针
  struct zskiplistNode *backward;
  // 分值,跳跃表中的所有节点都按照分值从小到大排序，相同分值按照成员对象字典序排序
  double score;
  // 成员对象
  robj *obj;
} zskiplistNode;

typedef struct zskiplist {
  // 表头节点和表尾节点
  structz skiplistNode *header, *tail;
  // 表中节点的数量
  unsigned long length;
  // 表中层数最大的节点的层数
  int level;
} zskiplist;
```

# 6.整数集合
整数集合（intset）是集合键的底层实现之一，当一个集合只包含整数值元素，并且这个集合的元素数量不多时，Redis就会使用整数集合作为集合键的底层实现。

```c
// intset.h/intset
typedef struct intset {
  // 编码方式
  uint32_t encoding;
  // 集合包含的元素数量
  uint32_t length;
  // 保存元素的数组，虽然声明是 int8_t，但是实际类型由 encoding 决定
  int8_t contents[];  //元素有序
}intset;
```

升级操作：
每当我们要将一个新元素添加到整数集合里面，并且新元素的类型比整数集合现有所有元素的类型都要长时，整数集合需要先进行升级（upgrade），然后才能将新元素添加到整数集合里面。

升级操作的好处：
- 提升灵活性: 避免类型错误
- 节省内存: 需要的时候才升级

intset 不支持降级操作
insert 操作是 O(N)，find 操作是 O(logN)

# 7. 压缩列表

压缩列表（ziplist）是列表键和哈希键的底层实现之一。当一个列表键只包含少量列表项，并且每个列表项要么就是小整数值，要么就是长度比较短的字符串，那么Redis就会使用压缩列表来做列表键的底层实现。
另外，当一个哈希键只包含少量键值对，比且每个键值对的键和值要么就是小整数值，要么就是长度比较短的字符串，那么Redis就会使用压缩列表来做哈希键的底层实现。

```sh
redis> RPUSH lst 1 3 5 10086 "hello" "world"
(integer)6
redis> OBJECT ENCODING lst
"ziplist"
redis> HMSET profile "name" "Jack" "age" 28 "job" "Programmer"
OK
redis> OBJECT ENCODING profile
"ziplist"
```

压缩列表是Redis为了节约内存而开发的，是由一系列特殊编码的连续内存块组成的顺序型（sequential）数据结构。一个压缩列表可以包含任意多个节点（entry），每个节点可以保存一个字节数组或者一个整数值。

|zlbytes|zltail|zllen|entry1|entry2|...|entryN|zlend|

每个压缩列表节点由 三个部分组成，通过当前指针的值减去 previous_entry_length 可以计算得到上一个节点的开始位置

|previous_entry_length|encoding|content|

连锁更新(cascade update)：特殊情况下产生的连续多次空间扩展操作，增加或者删除节点都可能会引发连锁更新。
因为连锁更新在最坏情况下需要对压缩列表执行N次空间重分配操作，而每次空间重分配的最坏复杂度为O（N），所以连锁更新的最坏复杂度为O（N
2）。 要注意的是，尽管连锁更新的复杂度较高，但它真正造成性能问题的几率是很低的:

- 首先，压缩列表里要恰好有多个连续的、长度介于250字节至253字节之间的节点，连锁更新才有可能被引发，在实际中，这种情况并不多见；
- 其次，即使出现连锁更新，但只要被更新的节点数量不多，就不会对性能造成任何影响：比如说，对三五个节点进行连锁更新是绝对不会影响性能的；

| 函数          | 作用                       | 时间复杂度              |
|---------------|----------------------------|-------------------------|
| ziplistPush   | 创建新节点并添加到表头或尾 | avg: O(N) worst: O(N^2) |
| ziplistInsert |                            | avg: O(N) worst: O(N^2) |
| ziplistFind   |                            | avg: O(N)               |
| ziplistDelete |                            | avg: O(N) worst: O(N^2) |

# 8. 对象

### Redis 对象
```c
typedef struct redisObject {
  // 类型, REDIS_STRING, REDIS_LIST, REDIS_HASH, REDIS_SET, REDIS_ZSET
  unsigned type:4;
  // 编码，记录这个对象底层使用了什么数据结构来实现
  unsigned encoding:4;
  // 指向底层实现数据结构的指针
  void *ptr;
  //...
} robj;
```

### 字符串对象
字符串对象的编码可以是 int , raw 或者 embstr。
如果一个字符串对象保存的是整数值，并且这个整数值可以用long类型来表示，那么字符串对象会将整数值保存在字符串对象结构的ptr属性里面（将void*转换成long），并将字符串对象的编码设置为int。
如果字符串对象保存的是一个字符串值，并且这个字符串值的长度小于等于32字节，那么字符串对象将使用embstr编码的方式来保存这个字符串值。
如果字符串对象保存的是一个字符串值，并且这个字符串值的长度大于32字节，那么字符串对象将使用一个简单动态字符串（SDS）来保存这个字符串值，并将对象的编码设置为raw。
可以用long double类型表示的浮点数在Redis中也是作为字符串值来保存的。如果我们要保存一个浮点数到字符串对象里面，那么程序会先将这个浮点数转换成字符串值，然后再保存转换所得的字符串值。

### 列表对象
列表对象的编码可以是ziplist或者linkedlist。
ziplist编码的列表对象使用压缩列表作为底层实现，每个压缩列表节点（entry）保存了一个列表元素。

当列表对象可以同时满足以下两个条件时，列表对象使用ziplist编码：

- 列表对象保存的所有字符串元素的长度都小于64字节；
- 列表对象保存的元素数量小于512个；不能满足这两个条件的列表对象需要使用linkedlist编码。

以上两个条件的上限值是可以修改的，具体请看配置文件中关于list-max-ziplist-value选项和list-max-ziplist-entries选项的说明。

#### 哈希对象
哈希对象的编码可以是ziplist或者hashtable。

当哈希对象可以同时满足以下两个条件时，哈希对象使用ziplist编码：

- 哈希对象保存的所有键值对的键和值的字符串长度都小于64字节；
- 哈希对象保存的键值对数量小于512个；不能满足这两个条件的哈希对象需要使用hashtable编码。

这两个条件的上限值是可以修改的，具体请看配置文件中关于hash-max-ziplist-value选项和hash-max-ziplist-entries选项的说明。

### 集合对象
集合对象的编码可以是intset或者hashtable。
intset编码的集合对象使用整数集合作为底层实现，集合对象包含的所有元素都被保存在整数集合里面。
hashtable编码的集合对象使用字典作为底层实现，字典的每个键都是一个字符串对象，每个字符串对象包含了一个集合元素，而字典的值则全部被设置为NULL。

当集合对象可以同时满足以下两个条件时，对象使用intset编码：

- 集合对象保存的所有元素都是整数值；
- 集合对象保存的元素数量不超过512个。

不能满足这两个条件的集合对象需要使用hashtable编码。

#### 有序集合对象
有序集合的编码可以是ziplist或者skiplist。

当有序集合对象可以同时满足以下两个条件时，对象使用ziplist编码：

- 有序集合保存的元素数量小于128个；
- 有序集合保存的所有元素成员的长度都小于64字节；

不能满足以上两个条件的有序集合对象将使用skiplist编码。

注意

以上两个条件的上限值是可以修改的，具体请看配置文件中关于zset-max-ziplist-entries选项和zset-max-ziplist-value选项的说明。

### 类型检查与命令多态

其中一种命令可以对任何类型的键执行，比如说DEL命令、EXPIRE命令、RENAME命令、TYPE命令、OBJECT命令等。
而另一种命令只能对特定类型的键执行，比如说：

- SET、GET、APPEND、STRLEN等命令只能对字符串键执行；
- HDEL、HSET、HGET、HLEN等命令只能对哈希键执行；
- RPUSH、LPOP、LINSERT、LLEN等命令只能对列表键执行；
- SADD、SPOP、SINTER、SCARD等命令只能对集合键执行；
- ZADD、ZCARD、ZRANK、ZSCORE等命令只能对有序集合键执行；

在执行一个类型特定的命令之前，Redis会先检查输入键的类型是否正确，然后再决定是否执行给定的命令。
类型特定命令所进行的类型检查是通过redisObject结构的type属性来实现的

### 内存回收
因为C语言并不具备自动内存回收功能，所以Redis在自己的对象系统中构建了一个引用计数（reference counting）技术实现的内存回收机制，通过这一机制，程序可以通过跟踪对象的引用计数信息，在适当的时候自动释放对象并进行内存回收。
对象的引用计数信息会随着对象的使用状态而不断变化：

- 在创建一个新对象时，引用计数的值会被初始化为1；
- 当对象被一个新程序使用时，它的引用计数值会被增一；
- 当对象不再被一个程序使用时，它的引用计数值会被减一；
- 当对象的引用计数值变为0时，对象所占用的内存会被释放。

### 对象共享
目前来说，Redis会在初始化服务器时，创建一万个字符串对象，这些对象包含了从0到9999的所有整数值，当服务器需要用到值为0到9999的字符串对象时，服务器就会使用这些共享对象，而不是新创建对象。
创建共享字符串对象的数量可以通过修改redis.h/REDIS_SHARED_INTEGERS常量来修改。
尽管共享更复杂的对象可以节约更多的内存，但受到CPU时间的限制，Redis只对包含整数值的字符串对象进行共享。

### 对象的空转时长
除了前面介绍过的type、encoding、ptr和refcount四个属性之外，redisObject结构包含的最后一个属性为lru属性，该属性记录了对象最后一次被命令程序访问的时间：

```c
typedef struct redisObject {
  // 类型, REDIS_STRING, REDIS_LIST, REDIS_HASH, REDIS_SET, REDIS_ZSET
  unsigned type:4;
  // 编码，记录这个对象底层使用了什么数据结构来实现
  unsigned encoding:4;
  // 指向底层实现数据结构的指针
  void *ptr;
  // 引用计数
  int refcount;
  // 对象的空转时长, 记录了最后一次被命令程序访问的时间
  unsigned lru:22;
} robj;
```

除了可以被OBJECT IDLETIME命令打印出来之外，键的空转时长还有另外一项作用：如果服务器打开了maxmemory选项，
并且服务器用于回收内存的算法为volatile-lru或者allkeys-lru，那么当服务器占用的内存数超过了maxmemory选项所设置的上限值时，
空转时长较高的那部分键会优先被服务器释放，从而回收内存。

# 9. 数据库

```c
// redis.h/redisDB
struct redisServer {
  redisDb *db;
  int dbnum;    // 默认是 16
  //...
}

typedef struct redisClient {
  //...
 redisDb *db;  // SELECT 可以选择不同的数据库
}redisClient;
```

数据库键空间：

```c
typedef struct redisDb {
  //数据库键空间，保存数据库中所有的键值对
 dict *dict;
}redisDb;
```

键的过期时间：
PEXPIREAT 和 PERSIST  可以设置和移除键过期时间


```c
typedef struct redisDb {
  //数据库键空间，保存数据库中所有的键值对
 dict *dict;
 // 过期字典保存键的过期时间, { *key, long long(毫秒精度的unix时间错) }
 dict *expires;
}redisDb;
```

如果一个键过期了，它什么时候会被删除呢？
- 定时删除: 设置过期实践的同时创建一个定时器。
  - 内存友好但是cpu时间不友好
  - redis 依赖服务器的时间事件处理，当前事件时间由无序链表实现，查找一个事件O(n)，不能高效处理大量事件。不现实

- 惰性删除：放任过期键不管，每次从键空间获取键时，检查是否过期，过期就删除，否则返回该键
  - cpu 时间最友好，内存不友好。可能出现『泄露』现象

- 定期删除(被动策略)：每个一段时间程序对数据库进行一次检查，删除里面过期的键。由算法决定检查多少个数据库删除多少个键
  - 折衷策略。难点是如何确定执行的时长和频率

redis 实际上使用了惰性和定期两种策略:

- 过期键的惰性删除策略由db.c/expireIfNeeded函数实现，所有读写数据库的Redis命令在执行之前都会调用expireIfNeeded函数对输入键进行检查：
- 过期键的定期删除策略由redis.c/activeExpireCycle函数实现，每当Redis的服务器周期性操作redis.c/serverCron函数执行时，activeExpireCycle函数就会被调用，它在规定的时间内，分多次遍历服务器中的各个数据库，从数据库的expires字典中随机检查一部分键的过期时间，并删除其中的过期键。


# 10 RDB 持久化

数据库状态：将服务器中的非空数据库以及他们的键值对统称为数据库状态。
RDB 持久化生成的 RDB 文件是一个经过压缩的二进制文件，通过该文件可以还原RDB 文件时的数据库状态。

RDB 文件的创建和载入(rdb.c/rdbSave)：
- SAVE: 阻塞 redis 服务器进程，知道 RDB 文件创建完毕为止，期间无法处理任何请求命令
- BGSAVE: 派生一个子进程负责创建RDB文件，父进程继续处理命令请求

RDB的载入是自动进行的，没有对应的命令，启动检测RDB文件存在就载入，载入期间处理阻塞状态。

另外值得一提的是，因为AOF文件的更新频率通常比RDB文件的更新频率高，所以：

·如果服务器开启了AOF持久化功能，那么服务器会优先使用AOF文件来还原数据库状态。

·只有在AOF持久化功能处于关闭状态时，服务器才会使用RDB文件来还原数据库状态。

除了执行命令外，redis 还支持设置条件自动执行 SAVE 操作。

RDB 文件结构(内容较多，详见原书):

[REDIS|db_version|databases|EOF|check_sum]

可以使用 od 命令来分析 RDB 文件。redis-check-dump 工具

# 11 AOF 持久化
Append Only File，通过保存redis 执行的写命令来记录数据库状态的

AOF 持久化功能可以分为:
- 命令追加(append): AOF功能打开时，服务器在执行完一个写命令后，会以协议格式将被执行的写命令追加到服务器状态的aof_buf
  缓冲区末尾
- 文件写入: redis服务器进程就是个事件循环，服务器每次结束一个事件循环之前，都会调用flushAppendOnlyFile
  函数，考虑是否将aof_buf缓冲区中的内容写入和保存到 AOF 文件里面
- 文件同步(sync): appendfsync 可以配置

AOF 文件重写：
将一个新的AOF文件替代现有的，新旧两个AOF文件所保存的数据库状态相同。不需要对现有aof文件进行读取，而是通过当前数据库状态实现的。

实现原理：首先从数据库读取键现在的值，然后用一条命去记录键值对，替代之前记录这个键值对的多条命令。新的aof文件只包含当前数据库状态所必须的命令，
不会浪费任何硬盘空间。

AOF后台重写：为了不阻塞主进程，用子进程处理重写，但是可能会有数据不一致的问题。为了解决这个问题redis 服务器设置了一个
AOF 重写缓冲区，这个缓冲区在服务器创建子进程之后开始使用，当redis服务器执行完一个写命令后，
它会同时将这个写命令发送给AOF缓冲区和AOF重写缓冲区。也就是 BGREWRITEAOF 命令的实现原理。

# 12 事件

redis服务器是一个事件驱动程序，服务器需要处理两类事件：
- 文件事件(file event): redis通过套接字进行连接，而文件事件就是服务器对套接字操作的抽象
- 时间事件(time event): Redis 服务器中的一些操作(serverCron)需要在给定的时间点执行，时间事件就是对这类定时操作的抽象。

### 文件事件

Redis基于Reactor模式开发了自己的网络事件处理器：这个处理器被称为文件事件处理器（file event handler）：

·文件事件处理器使用I/O多路复用（multiplexing）程序来同时监听多个套接字，并根据套接字目前执行的任务来为套接字关联不同的事件处理器。

·当被监听的套接字准备好执行连接应答（accept）、读取（read）、写入（write）、关闭（close）等操作时，与操作相对应的文件事件就会产生，这时文件事件处理器就会调用套接字之前关联好的事件处理器来处理这些事件。

文件事件是对套接字操作的抽象，每当一个套接字准备好执行连接应答（accept）、写入、读取、关闭等操作时，就会产生一个文件事件。因为一个服务器通常会连接多个套接字，所以多个文件事件有可能会并发地出现。


I/O多路复用程序可以监听多个套接字的ae.h/AE_READABLE事件和ae.h/AE_WRITABLE事件，这两类事件和套接字操作之间的对应关系如下：

·当套接字变得可读时（客户端对套接字执行write操作，或者执行close操作），或者有新的可应答（acceptable）套接字出现时（客户端对服务器的监听套接字执行connect操作），套接字产生AE_READABLE事件。

·当套接字变得可写时（客户端对套接字执行read操作），套接字产生AE_WRITABLE事件。

I/O多路复用程序允许服务器同时监听套接字的AE_READABLE事件和AE_WRITABLE事件，如果一个套接字同时产生了这两种事件，那么文件事件分派器会优先处理AE_READABLE事件，等到AE_READABLE事件处理完之后，才处理AE_WRITABLE事件。

这也就是说，如果一个套接字又可读又可写的话，那么服务器将先读套接字，后写套接字。
（具体整个流程参考原书）

### 时间事件
分为：

- 定时事件
- 周期性事件 :更新 when 属性得以循环往复

服务器将所有时间事件都放在一个无序链表中，每当时间事件执行器运行时，就遍历整个链表，查找所有已经到达的时间事件，
并调用相应的时间处理器。链表无序指的是不按照 when 属性排序，而不是说没有按照 id 排序。

### 调度
文件事件和时间事件之间是合作关系，服务器会轮流处理这两种事件，并且处理事件的过程中也不会进行抢占。
·时间事件的实际处理时间通常会比设定的到达时间晚一些。


# 13 客户端
一对多客户端程序，redis 服务器状态结构的clients 属性是一个链表

```c
struct redisServer {
  //...
  // 一个链表保存了客户端所有状态
  list *clients;
}


typedef struct  redisClient {
  // ...
  int fd;
  robj *name; // 设置客户端名字指
  int flags;   // 记录客户端角色
  sds querybuf;  // 输入缓冲区保存客户端发送的请求命令，根据输入内容动态调整
  robj **argv; // 客户端发送的请求保存到客户端状态的 querybuf 属性之后，服务器将对命令请求的内容分析，将命令行参数信息保存到 argv, argc
  int args;
  struct redisCommand *cmd; // 解析完命令(argv[0])后，根据命令类型调用对应的函数

  // 固定大小的缓冲区保存长度比较小的回复
  char buf[REDIS_REPLY_CHUNK_BYTES];
  int bufops;

  // 可变大小的缓冲区保存长度比较大的回复
  list *reply;

  // 客户端状态的 authenticated 属性用于记录客户端是否通过了身份验证
  int authenticated;

  time_t ctime; //ctime属性记录了创建客户端的时间，这个时间可以用来计算客户端与服务器已经连接了多少秒
  time_t lastinteraction; //lastinteraction属性记录了客户端与服务器最后一次进行互动（interaction）的时间，这里的互动可以是客户端向服务器发送命令请求，也可以是服务器向客户端发送命令回复。
  time_t obuf_soft_limit_reached_time; //记录了输出缓冲区第一次到达软性限制（soft limit）的时间

}redisClient;

```

服务器使用两种模式来限制客户端输出缓冲区的大小：

- 硬性限制（hard limit）：如果输出缓冲区的大小超过了硬性限制所设置的大小，那么服务器立即关闭客户端。

- 软性限制（soft limit）：如果输出缓冲区的大小超过了软性限制所设置的大小，但还没超过硬性限制，那么服务器将使用客户端状态结构的obuf_soft_limit_reached_time属性记录下客户端到达软性限制的起始时间；之后服务器会继续监视客户端，如果输出缓冲区的大小一直超出软性限制，并且持续时间超过服务器设定的时长，那么服务器将关闭客户端；相反地，如果输出缓冲区的大小在指定时间之内，不再超出软性限制，那么客户端就不会被关闭，并且obuf_soft_limit_reached_time属性的值也会被清零。


# 14 服务器
命令请求的执行过程。具体看书吧。

serverCron 函数的执行过程，每隔100 ms 执行一次，更新服务器状态信息，处理服务器接受的 SIGTERM 信号，管理客户端资源和
数据库状态，检查并执行持久化操作等。


初始化服务器步骤： 1.初始化服务器状态。2.载入配置。3.初始化服务器数据结构。4.还原数据库状态。5.最后会执行服务器的事件循环(loop)

# 15 复制
用户可以通过执行 SLAVEOF 命令或者设置 slaveof 选项，让一个服务器复制另一个服务器(master)。

## 旧版复制功能的实现
redis 复制功能分为 同步（sync）和命令传播（command propagate）


## 新版复制功能的实现
