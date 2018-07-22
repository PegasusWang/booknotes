《Python Microservices Development》


# 1 Understading Microservices

传统的单体架构（LAMP/LNMP）随着代码的膨胀很快就不可控。

A microservice is a lightweight application, which provides a narrowed list of features with a well-defined contract. It's a component with a single responsibility , which can be developed and deployed independently.

微服务架构的好处：
- Separation of concerns: 可以被单独的小组维护（康威定律）；单一职责，解耦
- Smaller projects: 拆分项目复杂度；减少风险（比如一个服务可以尝试改用新框架）
- Scaling and deployment: 容易扩展，但是也会带来新问题

微服务架构的缺点：没有银弹
- Illogical splitting: 不成熟的切分是万恶之源，拆分远比合并容易
- More network interactions
- Data storing and sharing: 减少不同微服务数据重复是一大挑战
- Compatibility issues: 良好的 api 设计一定程度上避免此类问题
- Testing

使用 Python 实现微服务：
- wsgi: 最大的不支持异步特性
- Greenlet and Gevent: 通过event loop 在不同 greenlet 切换，gevent 简化了 greenlet 使用
- Twisted and Tornado:
- asyncio


# 2 Discovering Flask
微框架不意味着只能做小应用，而是把自由交给开发者，可以随意组合，缺点就是容易做出错误决定(选错依赖包)，缺少最佳实践。

flask url_for function, REPL(Read-Eval-Print Loop)

# 3 Coding, Testing, and Documenting - the Virtuous Cycle


### Different kinds of tests
- Unit tests: 仅在 IO 操作/CPU密集操作/特定复现行为 才使用 mock。
- Functional tests: sending http requests and asserting the http responses. 测试能按照预期工作/测试非正常行为被修复且不会再出现
- Integration tests: functional test without any mocking. 测试真实部署的应用. curl 脚本等。WebTest
- Load tests: 发现瓶颈，不过早优化。 工具：Boom, locust, Molotov, flask-profiler
- End-to-end tests: need real user client


### Using WebTest

使用  webtest 来进行 functional test，http 接口级别的测试

### Using pytest and Tox

在 TestClass 中写测试可以保证各种测试发现库兼容， pytest，nose
tox 支持不同版本的 python 测试

### Developer documentation
- how it's design
- how to install it
- how to run tests
- what are the exposed APIs and what data comes in and out, and so on

Sphinx, Autodoc extension

### Continuous Integration
Travis-CI
ReadTheDocs
Coveralls

# 4 Designing Runnerly

本章讲了拆解一个 flask app 为多个微服务的过程

https://github.com/Runnerly

Connexion: swagger support for flask


# 5 Interacting with Ohter Services

## Synchronous calls
http: focus resource
rpc: focus action
网络交互需要处理 超市、连接错误等异常

#### http connection pooling
注意 requests session 在 flask 多线程中无法做到线程安全

#### Http cache headers
Etag:  使用时间计算 Etag 需要处理始同步的问题，使用 hash 是比较耗费 cpu 的

#### Improving data transfer
json 虽然可读，但是比较冗余，浪费带宽。可以采用数据压缩或者采用二进制传输协议。

###### Gzip 压缩
nginx/apache 支持，尽量不要在 python 做。requests 库自动会解压 gzip 压缩的数据。
如果需要压缩发送的数据，使用 gzip 模块并且指定 {'Content-Encoding': 'gzip'} header，需要server端处理。

##### Binary payloads
两种广泛使用的二进制协议是 Protocol Buffers(protobuf) and MessagePack，但是压缩方面不太好。
除非尽可能序列化提速，否则还是坚持用 json 吧。

## Asynchronous calls

### Task queues
celery use push-pull tasks queue。可以使用 RabibitMQ 作为 message brokers 支持消息持久化

### Topic queues
采用针对某个 topic 的订阅模式, RabibitMQ 实现了 AMQP(Advanced Message Queuing Protocol)，三个概念：
- queues: 消息接受者，并且等待消费者从中获取消息
- exchanges: 是发布者向系统添加新消息的入口
- bindings: 定义了消息如何从 exchanges 路由到 queues

### Publish/subscribe
如果希望一个消息可以被多个 worker 消费，就需要使用pub/sub

### RPC over AMQP
AMQP 同样实现了同步的 request/response 模式

## Testing
functional tests 需要隔离网络调用，通常使用 mock 的方式

### Mocking synchronous calls
requests: requests-mock 项目实现了 mock adapter

### Mocking asynchronous calls
讲了如何mock celery 任务


# Monitoring Your Services

### Centralizing logs
中心化日志方便收集和查询。

Sentry 是python社区内知名的错误日志收集工具(知乎也在用)，主要用于收集错误异常，而不是通用的日志查询工具
Graylog 是一个开源的通用日志收集工具，可以使用内置的收集器或者 fluentd，配合 ES 搜索

### Performance metrics
如果内存使用过高可能就被 out-of-memory killer(oomkiller)杀掉
- 应用有内存泄露。经常出现在一些 c 扩展忘记对象解引用
- 代码使用了太多内存。比如无限增长的内存缓存
- 服务的内存不够用，服务器接受太多了请求

##### System metrics
psutil: 跨平台的获取系统信息的库
system-metrics

##### Code metrics
最简单的方式写个装饰器记录耗时然后发给 Graylog

#### Web server metrics
nginx syslog


# Securing Your Services
OAuth2 authorization protocol

### The OAuth2 protocol
CCG: client credentials grant

##### Token-based authentication
token 可以携带 authentication (认证) 和 authorization(授权)信息
OAuth2 使用 JWT 生成token

##### The JWT standard
JSON Web Token(JWT) 在 RFC 7519 描述, base64 encoded 所以可以被用在 query string
- Header
- Payload
- Signature

PyJWT 模块

##### X.509 certificate-based authentication

### Web application firewall
common attacks:
- SQL injection: use sqlalchemy and avoid raw sql
- Cross Site Scripting(XSS)
- Cross-Site Request Forgery(XSRF/CSRF)

Open Web Application Security Project(OWASP)

##### OpenResty - Lua and nginx
REPL(Read Eval Print Loop)

##### Rate and concurrency limting
lua-resty-limit-traffic, you can use it in a acces_by_lua_block_section
lua-resty-waf

### Securing Your code
两个原则：
- 每个来自外部的请求在对你的应用或数据操作之前都应该估算
- 每个应用在系统上的操作都应该有良好定义和限制的域

#### Asserting incoming data
Server-Side TemplateInjection(SSTI): 不要直接使用用户传来的数据渲染模板

#### Limiting your application scope
- 限定访问的http 方法
- 限定能访问的资源

web 服务不要用 root 用户启动，尽量避免在 web 服务中执行外部进程

#### Using Bandit linter
Openstack 开源了一个检测代码安全性的工具 Bandit


# Bringing It All Together
### Building a ReactJS dashboard

### Flask and ReactJS

##### Cross-origin resource sharing
浏览器同源策略
CORS

### Authentication and authorication


# Packaging and Running Runnerly

### Packaing
打包 Python 项目：
- setup.py: 一个特殊的模块，用来驱动一切
- requirements.txt：依赖文件列表
- MANIFEST.in: 列出需要发布文件


### Process management
preform model: uwsgi, gunicorn

Circus: 进程管理器


# Containerized Services

### What is Docker?
Docker project is a container platform, which lets you run your application in isolated environments.

### Docker101


### Docker Compose
Docker compose simplifies the task by letting you define multiple containers configuration in a single configuration file.

### Introduction to Clustering and Provisioning
The collection of containers running the same image is called cluster.

Service discovery and sharing configuration can be done by tools like Consul or Etcd and
Docker's swarm mode can be configured to interact with those tools.

Provisioning: this term describes the process of creatin new hosts, and therefore, clusters, given the description of the stack you are deploying in some declarative form.

Ansible or Salt provide a  Devops-friendly environment to deploy and manage hosts.

Kubernetes is yet another tool, can be used to deply clusters containers on hosts.
