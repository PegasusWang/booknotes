
# 应用1：分布式锁


redis2.8 之后加入了set 指令的扩展参数， 使得 setnx 和 expire 可以一起执行（保证原子性）

可重入性：指的是线程持有锁的情况下再次请求加锁，如果一个锁 支持同一个线程的多次加锁，就是可重入的。


# 应用2：缓兵之计-延时队列

使用 list 作为异步消息队列（你对可靠性没有极致追求）
blpop,brpop 阻塞操作，如果list 没有数据就会进入休眠。这里需要注意长时间连接断开，需要处理超时异常并重试。


# 应用3：节衣缩食：位图

bit 数组，bytes 数组，使用 get/set/getbit/setbit/bitcount/bitops 操作

redis 位图自动扩展，如果某个偏移位置超出了现有的内容范围，自动将位图数组进行零扩充。

比如可以用来计算用户的签到时间，非常节省内存。


# 应用4: 四两拨千斤：HyperLogLog

解决统计问题的，比如统计UV(unique visit)，pv 好统计，直接可以用hash(key,val) 计数。

提供不精确的去重计数方案，标准误差 0.81%.

pfadd,pfcount 类似 sadd,scard。 (pf 是其HyperLogLog 发明人首字母缩写)

HyperLogLog  占据 12k 存储空间


# 应用5：布隆过滤器

比如用户推荐系统的去重复。布隆过滤器说某个值存在，这个值可能不存在，但是若确定不存在，则肯定不存在。

redis4.0 提供了插件功能之后才有布隆过滤器功能。

bf.add, bf.exists

bloom filter calculator

# 应用6：断尾求生-简单限流

思想是使用 zset 记录用户在一个时间窗口之内的操作数量。
这种方式适合小规模的限流，比如用户发帖等。


```py
import time
import redis


client = client.StrictRedis()


def is_action_allowed(user_id, action_key, period, max_count):
    key = "hist:%s:%s" % (user_id, action_key)
    now_ts = int(time.time() * 1000)  # 毫秒时间戳
    with client.pipeline() as pipe:
        # 记录行为，这里第一个 now_ts 没啥意义，用 uuid 之类的也可以
        pipe.zadd(key, now_ts, now_ts)
        # 移除时间窗口之前的行为记录，剩下的都是时间窗口之内的
        pipe.zremrangbyscore(key, 0, now_ts - period * 1000)
        # 获取时间窗口内的行为数量
        pipe.zcard(key)
        pipe.expire(key, perid + 1)  # 设置 zset 过期时间，主要是为了处理冷用户持续占用内存
        _, _, current_count = pipe.execute()
    return current_count <= max_count


for i in range(20):
    print(is_action_allowed('laowang', 'reply', 60, 5))
```

# 应用7：一毛不拔-漏斗限流


```py
# 单机漏斗算法

import time
# 漏斗的剩余空间代表当前行为可以持续进行的数量
# 漏斗的流水速度代表系统允许该行为的最大频率


class Funnel:
    def __init__(self, capacity, leaking_rate):
        self.capacity = capacity  # 漏斗容量
        self.leaking_rate = leaking_rate  # 流水速率
        self.left_quota = capacity
        self.leaking_ts = time.tiem()  # 上一次漏水时间

    def make_space(self):
        now_ts = time.time()
        delta_ts = now_ts - self.leaking_ts  # 距离上一次漏水过了多久
        delta_quota = delta_ts * self.leaking_rate  # 又可以腾出来的空间
        if delta_quota < 1:  # 腾出来的空间太少，等下一次
            return
        self.left_quota += delta_quota
        self.leaking_ts = now_ts  # 更新漏水时间
        if self.left_quota > self.capacity:
            self.left_quota = self.capacity  # 不能多余容量

    def watering(self, quota):
        self.make_space()
        if self.left_quota >= quota:  # 判断剩余空间是否足够
            self.left_quota -= quota
            return True
        return False


funnels = {}  # 所有漏斗


def is_action_allowed(user_id, action_key, capacity, leaking_rate):
    key = '%s:%s' % (user_id, action_key)
    funnel = funnels.get(key)
    if not funnel:
        funnel = Funnel(capacity, leaking_rate)
        funnels[key] = funnel
    return funnel.watering(1)

for i in range(20):
    print(is_action_allowed("laoqian", "reply", 15, 0.5))
```

如何实现分布式限流呢？这里其实可以把 dict 替换成hash，但是要保证从 hash 取值/内存计算/取出字段的原子性。
Redis4.0 提供了一个限流 redis 模块，redis-cell，提供了漏斗算法和原子限流指令。

```
# 表示user_id的回复行为频率每60秒最多30次。
cl.throttle user_id:reply 15 30 60 1

15 capacity 漏斗容量
30 operations/60seconds 漏水速率
1 是可选quota, 默认值1
```

# 应用8：近水楼台-GeHash

redis3.2 以后增加了地理位置 GEO，可以实现附近的餐馆这种功能。

地图元素的位置数据使用二维经纬度表示，经度范围(-180, 180]，纬度范围（-90，90]

比如指定一个半径r，使用 sql 可以圈出来。如果用户不满意，可以扩大半径继续筛选。

```
select id from positions where x0-r <x< x0+r and y0-r<y<y0+r
```
一般为了性能加上双向符合索引 (x,y)。但是在高并发场景不是好的选择。


### GEOHash 算法，地理位置距离排序

原理：把地球看成二维平面，划分成一系列的正方形方格（类似棋盘）。
所有地图元素坐标都放置在唯一的方格中，方格越小越精确。
然后对这些方格整数编码，越是靠近方格的编码越是接近。比如两刀切蛋糕，可以用 00,01,10,11四个二进制数字
表示。继续切下去正方形会越来越小，二进制正数也会越来越长，精度更高。
编码之后每个地图元素的坐标都是一个整数，通过整数可以快速还原出坐标。

redis 使用52位的整数进行编码，放到 zset 里边，value 是元素 key，score 是 GeoHash 的52位整数值。

- 增加：geoadd company 116.48015 39.996794 juejin
- 计算距离：geodist company juejin ireader km
- 获取元素位置(轻微误差，不影响附近的人功能)：geopos company juejin ireader
- 获取元素 hash(geohash 52位编码)：geohash company ireader   # http://gohash.org/XXXX 可以获取位置
- 附近的公司: georadiusbymember copany ireader 20 km count 3 asc
  - 范围20公里以内最多的3个元素按照距离正序排序，不会排除自身
- 根据坐标查询：georadius company 116.514202 39.905409 20 km withdist count 3 asc

集群环境中单个 zset key 数据量不宜超过 1M，否则迁移集群出现卡顿。
可以根据国家，省份，市区等进行拆分，显著降低 zset 集合大小


# 应用9：大海捞针-Scan

从海量 key 找出特定前缀的key列表。
redis keys 简单粗暴列出所有 满足特定正则的key.  `keys codehole*`
缺点：
- 没有 limit， offet，刷屏
- O(n)，千外级别以上的 key导致 redis 卡顿

redis2.8 加入了 scan 用来大海捞针

- 复杂度虽然也是O(n)，但是通过游标分布进行，不会阻塞线程
- limit 参数，limit 返回的只是 hint，结果可大可小
- 同 keys 有模式匹配
- 返回结构可能重复，需要客户端去重
- 如果遍历过程期间有修改，改动后的数据能否遍历到不确定
- 通过返回的游标是否为0决定是否遍历结束，而不是返回的个数

避免大 key 产生。如果你观察 redis 内存大起大落，很有可能是大 key 导致的。
定位到 key 然后改进业务代码。

redis提供了大 key 扫描功能。

- redis-cli -h 127.0.0.1 -p 7001 --bigkeys -i 0.1
- 0.1 表示每隔100条scan 休眠 0.1 ，防止 ops 剧烈抬升，扫描时间会变久


# 原理1：线程 IO 模型

redis是单线程程序，一定要小心使用O(n)的指令，防止 redis 卡顿

```
read_events, write_events = select(read_fds, write_fds, timeout)
for event in read_events:
    handle_read(event.fd)
for event in write_events:
    handle_write(enent.fd)
handle_others() # 处理其他任务，比如定时任务
```
###  指令队列
redis 为每个客户端维护了一个指令队列，先来先服务

### 响应队列
redis 同样也为每个客户端套接字关联一个响应队列，redis服务器通过响应队列来讲指令的结果返回给客户端。
如果队列为空，意味着连接暂时空闲，不需要获取写17:26:02，可以把当前客户端 socket 从write_fds移出来。
等到队列有数据了再放进去，避免select系统调用立即返回写事件，结果发现没什么数据可写，线程飙高 cpu。

### 定时任务
redis定时任务记录在一个最小堆。快要执行的任务放在堆顶，每个循环周期，redis 都会把最小堆里已经到点的任务
立即进行处理。处理完毕后，把最快要执行的任务还需要的时间记录下来，这个时间就是 select 的 timeout 参数。
redis 知道未来 timeout 时间段内，没有其他定时任务需要处理，可以安心睡眠 timeout 的时间。


# 原理2：交头接耳- 通信协议

RESP(Redis Serialization Protocol): redis序列化协议

resp 把传输的数据结构分成5种最小单元类型，单元结束后统一加上回车换行符 \r\n

- 单行字符串以+开头
  - +hello\r\n
- 多行字符串以$开头，后缀字符串长度
  - $11\r\nhello world\r\n
- 整数值以: 开头，跟上字符串形式
  - :1024\r\n
- 错误消息，以-开头
  - WRONGTYPE Operation against a key holding the wrong kind of value
- 数组，以 * 开头，后跟数组的长度
  - *3\r\n:1\r\n:2\r\n:3\r\n

特殊：

- NULL 用多行字符串，长度-1
  - $-1\r\n
- 空串，用多行字符串表示，长度0。注意这里的俩\r\n 中间隔的是空串
  - $0\r\n\r\n

### 客户端->服务器

只有一种格式，多行字符串数组

### 服务器->客户端

也是5种基本类型组合

- 单行字符串响应
- 错误响应
- 整数响应
- 多行字符串响应
- 数组响应
- 嵌套


# 原理3：未雨绸缪-持久化

### 快照，一次全量备份，内存数据的二进制序列化格式
- redis使用操作系统的多进程 COW(copy on write) 机制来实现快照持久化
- glibc fork产生一个子进程，快照持久化完全交给子进程处理。
- cow 机制进行数据段页面的分离，数据段由很多操作系统的页面组合而成。
  当父进程对其中一个页面的数据修改时，会将被共享的页面复制一份出来，然后对这个复制的页面修改。
  子进程还是 fork 瞬间的数据。

```
pid = os.fork()
if pid > 0:
    handle_client_requests() #父进程继续处理 client 请求
if pid == 0:
    handle_snapshot_write() # 子进程快照写磁盘
if pid < 0:
    # fork error
```

### AOF日志，连续的增量备份，记录内存数据修改的指令记录版本。

只记录对内存进行修改的指令记录。通过『重放』恢复 Redis 当前实例的内存数据结构状态。

AOF 重写：gbrewriteaof 指令用于对 AOF 日志瘦身。开辟一个子进程对内存遍历转换成一系列 Redis 操作指令，
序列化到一个新的 AOF 日志文件中。序列化完毕后再把操作期间发生的增量 AOF 日志追到薪的 AOF 日志文件中，
追加完毕立即替换旧的 AOF 文件。瘦身完成。

linux glibc提供了 fsync(int fd) 函数可以将指定文件内容强制从内核缓存刷到磁盘。
只要 redis 进程实时调用 fsync 可以保证 aof日志不丢失。


运维：通常不在 master 持久化，而是从节点进行。做好主从监控。


### 混合持久化
Redis4.0 增加了混合持久化。rdb文件内容和增量的 AOF 日志文件存到一起，AOF 不再是全量的的日志，
而是自持久化开始 到 持久化结束 这段时间的增量 AOF 日志，通常会比较小。

Redis重启，先加载 rdb 内容，然后重放增量 AOF 日志就可以完全替代之前的 AOF 全量文件重放，大幅提升重启效率。


# 原理4：雷厉风行-管道
管道本质是 client 提供的，不是 redis server.

# 原理5：同舟共济-事务

redis 事务模型不严格。
传统事务又 begin,commit,rollback，redis 是 multi,exec,discard，分别是事务开始，执行和丢弃。
所有指令在 exec 之前不执行，而是缓存在服务器的一个事务队列中，服务器一旦收到 exec，
开始执行整个事务队列，并且期间不会被其他指令打扰。
如果遇到一个失败了，后续指令还是会继续执行，不能算是原子性，仅仅是满足了事务的隔离性。

discard 用于丢弃事务缓冲队列种的所有指令，在 exec 执行之前。
通常结合 pipeline 使用减少网络IO。

```
pipe = redis.pipeline(transaction=True)
pipe.multi()
pipe.incr('a')
pipe.incr('a')
values = pipe.execute()
```

### redis Watch

乐观锁，解决并发修改。

```
while True:
    do_watch()
    commands()
    multi()
    send_commands()
    try:
        exec()
        break
    except WatchError:
        continue
```

watch 监控的变量如果事务执行的时候发现被修改了（包括当前客户端），
exec 指令返回 null 告诉 client 执行失败，一般由 client 重试。
```
# 使用 redis-py 实现余额加倍

import redis


def key_for(user_id):
    return 'account_{}'.format(user_id)


def double_account(client, user_id):
    key = key_for(user_id)
    while True:
        client.watch(key)
        value = int(client.get(key))
        value *= 2
        pipe = client.pipeline(transaction=True)
        pipe.multi()
        pipe.set(key, value)
        try:
            pipeline.execute()
            break  # 成功，退出
        except redis.WatchError:
            continue  # 重试
    return int(client.get(key))  # 重新获余额


client = client.StrictRedis()
user_id = 1
client.setnx(ke_for(user_id), 5)  # 初始化
print(double_account(client, user_id))
```

#  原理6：Pubsub

发布订阅，不过无法持久化消息队列，一般没啥用。

redis5.0 加入了 stream 数据结构，给redis 带来了持久化消息队列。
从此 pubsub 可以消失了。


# 原理7：开源节流-小对象压缩

32 vs 64 bit: 如果使用不超过4g，可以考虑32bit编译，内部所有数据结构所使用的指针空间占用少一半。

小对象压缩存储 : ziplist, intset


内存回收机制：redis 并不总是可以将空闲内存立即归还给操作系统。原因os 回收内存以page 为单位，
如果这个页上只要有一个 key 还在使用，那么它就不能被回收。但是 Redis 会对这些内存重用。


内存分配算法：redis 将内存分配细节丢给了第三方，目前可以用 jemalloc(facebook)库来管理内存，
或者切换到tcmalloc(google)，默认 redis 使用了性能更好的 jemalloc。

info memory 命令查看redis 内存信息。


# 原理8：有备无患-主从同步

增量同步：同步指令流，主节点会把那些对自己的状态产生修改性影响的指令记录在本地内存 buffer，
然后异步将 buffer 中的指令同步到从节点。使用的 buffer 是环形 buffer，存在覆盖丢失数据危险


快照同步：非常耗费资源的操作，首先在主库上进行一次bgsave将当前内存的数据全部快照到磁盘文件中，
然后再将快照文件的内容全部传送到从节点。从节点将快照文件接收完毕后， 立即清空数据然后全量加载。
加载完毕后同事主节点继续增量同步（之所以需要继续增量是同步期间还有新的执行指令需要同步到从节点）
(快照同步可能遇到死循环，配置一个合适的 buffer 大小)

增加从节点：从节点刚刚加入到集群先要进行一次快照同步，同步完成后再继续进行增量同步。

Wait指令：redis 复制是异步进行的，wait 可以让异步复制变身同步复制，确保强一致性（不严格）

小结：如果只用来做缓存可以不用复制功能，否则就要保证主从复制。


# 集群1：Sentinel
