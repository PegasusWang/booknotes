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

redis基本事务(basic transaction):让一个客户端在不被其他客户端打断的情况下执行多个命令，和关系数据库可以执行过程中回滚的事务不同，
redis 里被 multi 命令和exec 命令包围的所有命令会一个接一个执行，直到所有命令执行完毕，redis 才会处理其他客户端命令。

redis 事务在 python client 上使用 pipeline 实现，客户端自动使用multi和exec，客户端会存储事务包含的多个命令，一次性把所有命令发送给redis。
移除竞争条件；减少通信次数提升性能。redis 原子操作指的是在读取或者修改数据的时候，其他客户端不能读取或修改相同数据。

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

(乐观锁)
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
            pipe.unwatch()
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

### 6.3 计数信号量

限制一个资源最多同时被多少个进程访问，限定能够同时使用的资源数量。
和锁不同的是，锁通常在客户端获取锁失败的时候等待，而当客户端获取信号量失败的时候，客户端通常会立即返回失败结果。

```py
def acquire_semaphore(conn, semname, limit, timeout=10):
    identifier = str(uuid.uuid4())
    now  = time.time()

    pipeline = conn.pipeline(True)
    pipeline.zremrangebyscore(semname, '-inf', now-timeout) # 清理过期的信号量持有者
    pipeline.zadd(semname, identifier, now)   # 尝试获取信号量
    pipeline.zrank(semname, identifier)
    if pipeline.execute()[-1] < limit:  # 检查是否成功获取了信号量
        return identifier
    conn.zrem(semname, identifier)  # 获取信号量失败后，删除之前添加的标识符
    return None


def release_semphore(conn, semname, identifier):
    return conn.zrem(semname, identifier)
```

这个信号量简单快速，但是有个问题，就是它假设每个进程访问到的系统事件都是相同的,
每当锁或者信号量因为系统始终的细微不同导致锁的获取结果出现剧烈变化时，这个锁或者信号量就是不公平的。(unfair)

如何实现公平信号量：给信号量实现添加一个计数器以及一个有序集合。计数器通过持续执行自增操作，创建出一个类似于计时器的机制，
确保最先对计数器执行自增操作的客户端能够获得信号量。另外，为了满足『最先对计数器执行自增操作的客户端能够获得信号量』这一要求，
程序会将计数器生成的值用作分值，存储到一个『信号量拥有者』有序集合里，然后通过检查客户端生成的标志符在有序集合里的排名判断客户端是否取得了信号量。


```py
def acquire_faire_lock_semaphore(conn, semname,  limit, timeout=10):
    identifier = str(uuid.uuid4())
    czset = semname + ':owner'
    ctr = semname + ':counter'

    now = time.time()
    pipeline = conn.pipeline(True)
    pipeline.zremrangebyscore(semname, '-inf', now - timeout)
    pipeline.zinterstore(czset, {czeset: 1, semname: 0})    # 删除超时的信号量

    pipeline.incr(ctr)
    counter = pipeline.execute()[-1]

    pipeline.zadd(semname, identifier, now)   # NOTE :注意这个客户端不是StrictRedis，参数顺序不一样
    pipeline.zadd(czset, identifier, counter)

    pipeline.zrank(czset, identifier)
    if pipeline.execute()[-1] < limit:  # 通过检查排名来判断客户端是否取得信号量
        return identifier

    pipeline.zrem(semname, identifier)  # 未能获取信号量清理数据
    pipeline.zrem(czset, identifier)
    pipeline.execute()
    return None


def release_faire_semaphore(conn, semname, identifier):
    pipeline = conn.pipeline(True)
    pipeline.zrem(semname, identifier)
    pipeline.zrem(semname + ':owner', identifier)
    return pipeline.execute()[0]    # 返回True表示信号量已经释放，False表示想要释放的信号量因为超时被删除了
```

这里注意如果是频繁大量使用信号量的情况下，32位计数器的值大约2小时就会溢出一次，最好切到64位平台。
这里实现依然需要控制各个主机的差距系统事件在1 秒之内。

对信号量进行刷新，防止过期：

```py
def refresh_fair_semaphore(conn, semname, identifier):
    if conn.zadd(semname, identifier, time.time()):  # 更新客户端持有的信号量
        release_fair_semaphore(conn, semname, identifier)
        return False  # 告知调用者已经失去了信号量
    return True  # 客户端仍持有信号量
```

前面介绍的实现会出现不正确的竞争条件。如果AB俩进程都在尝试获取一个信号量时，即使A首先对计数器执行了自增操作，
但是B只要能抢先把自己的标识符添加到有序集合里，并检查标志符在有序集合中的排名，B就可以成功获取信号量。
之后当A也将自己的标志符添加到有序集合里时，并检查标志符在有序集合中的排名时，A将『偷走』B已经取得的信号量，
而B只有在尝试释放或者刷新的时候才能发现。

```py
def acquire_semaphore_with_lock(conn, semname, limit, timeout=10):
    identifier = acquire_lock(conn, semname, acquire_timeout=.01)
    if identifier:
        try:
            return acquire_fair_semaphore(conn, semname, limit, timeout)
        finally:
            release_lock(conn, semname, identifier)
```

信号量可以用来限制同时可运行的 api 调用数量，针对数据库的并发请求等。

### 6.4 任务队列

FIFO 队列：借助 redis list 的 blpop rpush 来实现一个队列
延迟任务：借助 redis zset 实现

### 6.5 消息拉取

使用 Redis 的 PUBLISH SUBSCRIBE 的问题时，如果客户端因为某些原因无法一直在线，消息推送可能会出问题。
可以采用拉取模式


### 6.6 redis 文件分发

当NFS，Samba、文件复制、mapreduce 不适合的时候。


# 7. 基于搜索的应用程序

### 7.1 使用 redis 进行搜索

首先需要根据文档生成倒排索引。从文档提取单词的过程称为语法分析(parsing)和标记化(tokenize)。
标记化的一个常见附加步骤是移除内容中的非用词(stop word)，在文档中出现频繁但是却没有提供相应信息量的单词。
然后存储到 redis 的 set 里，每个集合记录该单词出现的所有文档。

关联度：搜索程序在获得多个文档之后，还需要根据每个文档的重要性排序(SORT 命令)。一种方式是看哪个文档更新事件最接近当前时间。

### 7.2 有序索引

使用zset 实现多个分值的复合排序操作。ZINTERSTORE

如何使用有序集合进行非数值排序：把字符串通过一种方式转成数值类型，书中给了一种方式（最多6个字符，受限于浮点数范围）
这样就能基于字符串前缀实现搜索，使用 zrangebyscore 命令查找长度不超过6个字符的前缀。

### 7.3 广告定向

广告服务器：每当用户访问一个带有广告的web页面的时候，web服务器和用户的web浏览器都会向远程服务器发送请求以获取广告，
广告服务器会接受各种信息，并根据这些信息找出能够通过点击、浏览、动作获取最大经济收益的广告。

web 页面广告通常有三类型： 按展示次数(cost per view), 按点击数(cost per click)，按动作执行次数(cost per action)
动作执行次数又称按购买次数（cost per acquisitoin）。按展示次数计费的广告又称为CPM 广告或千次计费(cost per mille)广告。

### 7.4 职位搜索

首先把每个职位需要的技能数量添加到有序集合里，首先计算出求职者对于每个职位的得分，然后使用胜任这些职位所需的总分减去求职者在这些职位上的得分，
在最后得出的结果有序集合里，分值为0 的职位就是求职者能够胜任的职位。


# 8. 构建简单的社交网站
本章构建一个和 twitter 后端类似的社交网站

### 8.1 用户和状态
使用 hash 存储用户的信息， name 'user:id'  {'id': login, id, name, folllowers, following, posts, signup}
hash 存储状态消息： 'status:id' {'id': message, posted, id, uid, login}

### 8.2 最常见的状态消息列表，用户主页时间线
用户登录的情况下访问twitter 时，主页看到的是自己的主页时间线。时间线是一个列表，由用户以及用户正在关注的人发布的状态组成。
使用zset 实现时间线。 zset  home:12234 {'status_id': timestamp} 个人时间线 zset  profile:12234 {'status_id': timestamp} 个

### 8.3 关注者列表和正在关注列表

使用 zset 存储关注者和被关注。 zset followers:id  {id: timestamp}
当关注或者停止关注的时候，需要对两个关注的zset 和用户的关注数量、被关注数量进行更新。
如果用户执行的是关注操作，还需要从被关注的用户时间线里，复制一些状态消息 id 到执行关注操作的主页时间线里，
从而使得用户在关注另一个用户之后， 可以立即看见被关注用户所发的状态消息。

### 8.4 状态消息的发布与删除

对于关注者比较少的用户，可以直接把用户发送的状态同步到每个关注者的时间线里，但是如果关注者很多(大V)， 需要使用延迟方式。
这种问题涉及到 feed 流系统中的『推拉模型』，到底是选择推还是拉取。经常是推拉结合。

删除消息：删除该消息在状态消息的散列(hash status:id)，更新已经发送的消息数量。然后清理用户时间线残留的消息 id

### 8.5 流API

有时候我们想知道网站正在发生的事情，比如每个小时发布了多少条状态，最热门的恩主题是什么等。可以通过专门执行一些调用，
或者函数内部记录这些信息。
还有一种方法就是构建一些函数来广播(broadcast)事件(event)，然后由负责进行数据分析的事件监听器(event listener)接受处理这些事件。

- 公开哪些事件？
- 是否访问限制?何种方式
- 提供哪些过滤选项


# 9. 降低内存占用

### 9.1 使用短结构

redis为列表、集合、散列和有序集合提供了可以配置选项，让 redis 以更节约控件的的方式存储长度较短的结构（短结构）
在列表(双链表)、散列(散列表)和有序集合(散列表+跳跃表)长度较短或者体积较小的时候，redis 可以选择使用 压缩列表(ziplist)的紧凑存储方式存储这些结构。

list-max-zip-entries 512
list-max-zip-value 64

可以用 DEBUG OBJECT 来观察

当整数包含的成员都能解释成十进制整数，并且又处于平台的有符号整数范围之内，集合使用『整数集合(intset)』，有有序数组方式存储集合。

尽量减少键的长度


### 9.2 分片结构
分片(sharding)，本质上就是基于简单的规则吧数据划分更小的部分，然后根据数据所属的部分来决定将数据发送到哪个位置上。

- 对列表分片。11 章介绍 lua 脚本实现方案
- 对有序集合分片。无法想普通的有序集合那样快速实现 zrange 等，分片作用不大

对于 namespace:id 形式的值可以考虑存储到分片散列里。降低内存占用

### 9.3 打包存储二进制位和字节

存储的是简短并且长度固定的连续 id，可以将数据打包存储在字符串键里边。


# 10. 扩展 redis

### 10.1 扩展读性能

回顾下提升性能的几个途径：

- 短结构
- 合适的数据结构
- 大体积对象压缩(lz4/gzip/bzip2)
- 流水线和连接池

为一个 redis 主服务器配置多个从服务器，但只对主服务器写入。主从同步问题

### 10.2 扩展写性能和内存容量

通过预先分片写入不同的服务器

### 10.3 扩展复杂的查询

对每个分片进行查询(并行)，最后统一聚合查询结果并且排序输出。


# 11. Redis Lua 脚本编程

### 11.1 在不编写C代码的情况下添加功能

```py
def script_load(script):
    sha = [None]

    def call(conn, keys=[], args=[], force_eval=False):
        if not force_eval:
            if not sha[0]:
                sha[0] = conn.execute_command("SCRIPT", "LOAD", script, parse="LOAD")
            try:
                return conn.execute_command("EVALSHA", sha[0], len(keys), *(keys + args))
            except redis.exceptions.ResponseError as msg:
                if not msg.arg[0].startswith("NOSCRIPT"):
                    raise

        return conn.execute_command("EVAL", script, len(keys), *(keys + args))
    return call
```

### 11.2 使用 Lua 重写锁和信号量

```py
def acquire_lock_with_timeout(conn, lockname, acquire_timeout=10, lock_timeout=10):
    identifier = str(uuid.uuid4())
    lockname = 'lock:' + lockname
    lock_timeout = int(math.ceil(lock_timeout))

    acquired = False
    end = time.time() + acquire_timeout
    while time.time() < end and not acquired:
        acquired = acquire_lock_with_timeout_lua(conn, [lockname], [lock_timeout, identifier]) == 'OK'
        time.sleep(.001 * (not acquired))
    return acquired and identifier

# 注意 lua 下标 1 开始
acquire_lock_with_timeout_lua = script_load('''
if redis.call('exists', KEYS[1]) == 0 then
    return redis.call('setex', KEYS[1], unpack(ARGV))
end
''')


def release_lock(conn, lockname, identifier):
    lockname = 'lock' + lockname
    return release_lock_lua(conn, [lockname], [identifier])

release_lock_lua=script_load('''
if redis.call('get', KEYS[1]) == ARGV[1] then
    return redis.call('del', KEYS[1]) or true
end
''')
```

### 11.3 移除 WATCH/MULTI/EXEC 事务

注意：运行在redis内部的lua脚本只能访问位于 lua 脚本或者 redis 数据库之内的数据， 锁或WATCH/MULTI/EXEC 事务没有这个限制

### 11.4 使用 Lua 对列表进行分片


