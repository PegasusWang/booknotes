# 1. Elasticsearch 介绍

### 解决搜索问题
Elasticsearch 建立在 Apache Lucene 之上的开源分布式搜索引擎。
计算文档相关性使用的是 TF-IDF(term frequency-inverse document frequency)

- 词频
- 逆文档词频: 如果某个单词在所有文档中比较少见，该词权重越高


### 使用案例

- 作为主要的后端系统。不支持事务
- 添加到现有系统。保持数据同步
- 和现有工具一同使用。ELK

### 组织数据

Elasticsearch 以文档方式存储数据

### 安装

安装完成之后 打开 http://localhost:9200


# 2. 深入功能


- 逻辑设计
- 物理设计。配置决定了集群的性能、可扩展性和可用性

### 文档、类型和索引

索引-类型-ID 的组合唯一确定了某篇文档

文档: 自包含，层次型，结构灵活，无模式。

类型：是文档的逻辑容器。不同类型最好放入不同结构的文档 。每个类型中字段的定义称为映射。

索引：是映射类型的容器。索引由一个或多个称为分片的数据块组成


# 3. 索引，更新和删除数据

核心类型：字符串；数值；日期(默认解析ISO 8601)；布尔

字符串如果索引类型是 analyzed，会转化为小写并且分词。

### 3.5 更新数据

更新文档包括:

- 检索现有文档。必须打开 `_source`
- 进行指定的修改
- 删除旧的文档，在其原有位置索引新的文档

使用更新 API:

- 发送部分文档

```sh
curl -XPOST 'localhost:9200/get-together/group/2/_update' -d '{
　"doc": {
　　"organizer": "Roy"
　}
}'
```

- 使用 upsert 创建不存在的文档。默认脚本语言是 Groovy

ES 通过 version 乐观锁来进行并发更新的控制：

- 冲突发生的时候可以使用重试操作。retry_on_conflict 参数可以让 es 自动重试
- 索引文档的时候使用版本号

### 3.6 删除数据

- 通过 id 删除单个文档
- 单个请求删除多个文档
- 删除映射类型，包括其中的文档
- 删除匹配某个查询的所有文档

删除索引。

除了删除索引还可以关闭。 `curl -XPOST 'localhost:9200/online-shop/_close'`


# 4 搜索数据

确定搜索范围，尽量限制在最小范围和类型，增加响应速度。

请求搜索的基本模块：

- query。使用查询 DSL和过滤 DSL 配置
- size: 返回文档数量
- from：分页
- _source: 指定_source 字段如何返回
- sort: 默认排序基于文档得分。sort 可以额外控制

```sh
curl 'localhost:9200/get-together/group/_search' -d '
{
  "query": {
    "match_all": {}
  },
  "from" 0,
  "size": 10,
  "_source": ["name", "organizer"],
  "sort": ["created_on": "desc"]
}'
```

### 4.2 查询和过滤 DSL

由于不计算得分，过滤器处理更少，并且可以被缓存。

```sh
curl 'localhost:9200/get-together/_search' -d '
{
  "query": {
    "filtered": {
      "query" {
        "match": {
          "title": "hadoop"
        }
      }
    }
  },
  "filter": {
    "term": {
      "host": "andy"
    }
  }
}'
```

- match_all 查询
- query_string 查询
- term 查询和 term 过滤器
- terms 查询，搜索某个文档那个字段中的多个词条
- match 查询和 term 过滤器
- phrase_prefix 查询，词组中最后一个词条进行前缀匹配。对于提供搜索框里的自动完成功能很有用，输入词条就可以提示。最好用
  max_expansions 限制最大的前缀扩展数量
  - multiple_match: 可以搜索多个字段中的值

### 4.3 组合查询或者复合查询

- bool 查询:允许在单独的查询组合任意数量的查询，指定的查询子句表名哪些部分是 must/should/must_not

### 4.4 超越 match 和过滤器查询

范围查询：

```sh
curl 'localhost:9200/get-together/_search' -d '
{
  "query": {
    "range": {
      "created_on": {
        "gt": "2011-06-01",
        "lt": "2012-06-01"
      }
    }
  }
}'
```

前缀查询：(如果需要可以先转成小写)

```sh
curl 'localhost:9200/get-together/_search' -d '
{
  "query": {
      "prefix": {
        "title": "liber"
      }
    }
  }
}'
```

wildcard 查询: 类似 shell globbing 的工作方式

### 4.5 使用过滤器查询字段的存在性

- exists 过滤器
- missing 过滤器

### 4.6 为任务选择最好的查询

![](./查询选择.png)


# 5. 分析数据

什么是分析： 文档被发送并加入到倒排索引之前，es 对其进行的操作。

- 字符过滤：字符过滤器转变字符
- 文本切分为分词
- 分词过滤：分词过滤器转变每个分词
- 分词索引：分词存储到索引中

为文档使用分析器：

- 创建索引的时候，为特定的索引进行设置
- 在 es 配置文件设置全局分析器
