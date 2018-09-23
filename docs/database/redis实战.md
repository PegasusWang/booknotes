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

# 4. 数据安全与性能保障

### 4.1 持久化选项

- 快照(snapshotting):讲某一时刻所有数据都写入硬盘。如果系统崩溃，可能丢失最近一次快照生成之后更改的所有数据。
- 追加文件(AOF, append-only file): 执行写命令时，被执行的写命令复制到硬盘里

### 4.2 复制

主从模式，redis 不支持主-主复制
slave of 命令

从服务器在进行同步时，会清空自己的所有数据，并被替换为主服务器发来的数据。
使用复制和AOF持久化，增强redis 对系统崩溃的抵抗能力。

### 4.3 处理系统故障

验证快找文件和AOF文件：

- redis-check-aof
- redis-check-dump

更换故障主服务器

### 4.4 Redis 事务

pipeline: 一次性发送多个命令，然后等待所有回复出现

redis使用了乐观锁的方式：redis为了尽可能减少客户端的等待时间，并不会在执行WATCH命令的时候对数据加锁。
相反，redis只会在数据已经被其他客户端抢先修改了的情况下，通知执行了watch 命令的客户端，这种做法被称为乐观锁。
关系数据库实际执行的加锁操作则被称为悲观锁(pessimistic locking)。

乐观锁使客户端不用等待第一个取得锁的客户端，只需要在自己的事务执行失败的时候重试就可以了。

### 4.5 非事务型流水线

除了使用批量命令，还可以使用非事务型流水线。
python客户端 redispy 在 pipe = conn.pipeline(False) 传入False
参数，可以让客户端会像执行事务那样收集用户要执行的所有命令，但是不会使用 MULTI 和 EXEC 包裹这些命令。

### 4.6 关于性能方面的注意事项

一个不使用流水线的python客户端性能大概只有 redis-benchmark 的 50%~60%。
大部分客户端库都提供了连接池。

# 5. 使用 Redis 构建支持程序

### 5.1 使用 Redis 来记录日志

linux/unix中两种常见记录日志的方法：

- 日志记录到文件。
- syslog

### 5.2 计数器和统计数据

实现时间序列计数器

最好用现成的 Graphite

### 5.3 查找 IP 所属城市以及国家

预先载入 ip 数据和地址数据到redis，存储的时候把点分十进制格式的 ip 转成一个整数分值。
把数据转换成整数并搭配有序集合进行操作。

```
def ip_to_score(ip_address):
    score = 0
    for v in ip_address.split('.'):
        score = score * 256 + int(v, 10)
    return score
```

### 5.4 服务的发现与配置

将配置存储在redis里，编写应用程序获取配置


# 6. 使用 Redis 构建应用程序组件

### 6.1 自动补全

- 主动补全最近联系人: 通过list 存储元素（元素数量较小），然后在 python 代码里进行 filter
- 通讯录自动补全：使用zset，所有分值置为0，通过插入带查找元素的前缀和后缀元素的方式确定待查找元素的范围

### 6.2 分布式锁

redis WATCH 实现的是乐观锁（只有通知功能）。由WATCH, MULTI EXEC
组成的事务并不具有可扩展性，程序在尝试完成一个事务的时候，可能会因为事务执行失败反复重试。

```py
def acquire_lock(conn, lockname, acquire_timeout=10):
    identifier = str(uuid.uuid4())
    end = time.time() + acquire_timeout
    while time.time() < end:
        if conn.setnx('lock:' + lockname, identifier):  # setnx 如果没有key就会设置
            return identifier
        time.sleep(0.001)
    return False

def release_lock(conn, lockname, identifier):
    pipe = conn.pipeline(True)
    lockname = 'lock:' + lockname
    while True:
        try:
            pipe.watch(lockname)    # 检查进程是否仍然持有锁
            if pipe.get(locname) == identifier:
                pipe.multi()   # 开始释放锁
                pipe.delete(lockname)
                pipe.execute()
                return True
            pipe.unwathc()
            break
        except redis.exceptions.WatchError:  # 有其他客户端修改了锁，重试
            pass
    return False  # 进程已经失去了锁
```

细粒度的锁能够提升程序性能，但是过细粒度可能导致死锁问题。

上边的实现在持有者崩溃的时候不会自动释放，会导致锁一直处于被获取的状态，下边加上超时功能。

```py
def acquire_lock_with_timeout(conn, lockname, acquire_timeout=10, lock_timeout=10):
    identifier = str(uuid.uuid4())
    lockname = 'lock:' + lockname
    lock_timeout = int(math.ceil(lock_timeout))
    end = time.time() + acquire_timeout
    while time.time() < end:
        if conn.setnx(lockname, identifier):
            conn.expire(lockname, lock_timeout)
            return identifier
        elif not conn.ttl(lockname):
            conn.expire(lockname, lock_timeout)
        time.sleep(.001)
    return False
```
