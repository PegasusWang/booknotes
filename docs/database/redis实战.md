# 《Redis 实战》英文民《Redis In Action》笔记

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
