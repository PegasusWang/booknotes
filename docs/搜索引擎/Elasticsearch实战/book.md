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
