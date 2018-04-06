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
