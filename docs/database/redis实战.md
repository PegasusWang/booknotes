# 《Redis 实战》英文名《Redis In Action》笔记

# 1. 初识Redis

5种基本类型：string, list, set, hash, zset
内存存储、远程、持久、可扩展（主从复制和分片）

# 2. 使用 Redis 构建 web 应用

典型使用场景：

- 登录和 cookie 缓存。签名(signed)cookie和令牌(token)cookie
- 用 redis 实现购物车，存储商品 id 和商品订购数量之间的映射
- 网页缓存，缓存不经常变动的网页结果
- 缓存部分关系数据库中的行
- 缓存部分页面：通过页面压缩、Edge Side includes 技术、移除模板无用字符等优化

# 3. Redis 命令

### 3.1 字符串

redis 里字符串可以存储三种类型的值，字节串(byte string),整数,浮点数

- incr
- decr
- incrby
- decrby
- incrbyfloat
- append：追加到值的末尾
- getrange: 获取子串
- setrange
- getbit: 将字节串看作是二进制位串(bit string)，并返回串中偏移量为offset的二进制位的值
- setbit
- bitcount: 统计二进制串位里值为1的二进制位的数量
- bitop: 对一个或者多个二进制位串执行位运算操作

### 3.2 列表

允许从两端推入或者弹出元素

- rpush
- lpush
- rpop
- lpop
- lindex
- lrange
- ltrim

阻塞式的列表弹出命令以及在列表之间移动元素的命令。常用在消息传递(messaging)和任务队列(task queue)

- blpop
- brpop
- rpoplpush
- brpoplpush

### 3.3 集合

无序存储多个不同的元素

- sadd
- srem
- sismember
- scard
- smembers
- srandmember
- spop
- smove

组合和关联多个集合

- sdiff
- sdiffstore: 差集
- sinter
- sinterstore
- sunion
- sunionstore

### 3.4 散列

多个键值对存储到一个 redis 键里

- hmget
- hmset
- hdel
- hlen
- hexists
- hkeys
- hvals
- hgetall
- hincrby
- hincrybyfloat

### 3.5 有序集合

根据分值大小有序获取(fetch)或扫描(scan)成员和分值

- zadd
- zrem
- zcard
- zincrby
- zcount
- zrank
- zscore
- zrange

范围命令，并集和交集命令。rev 逆序表示分值从大到小排列

- zrevrank: 分值从大到小排列
- zrevrange
- zrangebyscore
- zrevrangebyscore
- zremrangebyrank
- zremrangebyscore
- zinterstore
- zunionstore

### 3.6 发布订阅

- subscribe
- unsubscribe
- publish
- psusbscribe: 订阅与给定模式匹配的所有频道
- punsubscribe

redis可能无法很好处理客户端失联、消息积压等

### 3.7 其他命令

##### 3.7.1 排序

redis sort 能对多种数据类型排序

##### 3.7.2 基本的 redis 事务

redis基本事务(basic transaction):让一个客户端在不被其他客户端打断的情况下执行多个命令，和关系数据库可以执行过程中回滚的事务不同，redis 里被 multi 命令和exec 命令包围的所有命令会一个接一个执行，直到所有命令执行完毕，erdis 才会处理其他客户端命令。

redis 事务在 python client 上使用 pipeline 实现，客户端自动使用multi和exec，客户端会存储事务包含的多个命令，一次性把所有命令发送给redis。移除竞争条件；减少通信次数提升性能。redis 原子操作指的是在读取或者修改数据的时候，其他客户端不能读取或修改相同数据。

##### 3.7.3 键过期时间

自动删除过期键减少内存使用，

- persist: 移除键的过期时间
- ttl: 查看给定键离过期时间还有多少秒
- expire: 将给定键的过期时间设置为给定 unix 时间戳
- pttl: 查看给定键过期时间还有多少毫秒
- pexpire: 给定键指定毫秒后过期
- pexpireat

