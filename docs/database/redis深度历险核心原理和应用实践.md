
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
