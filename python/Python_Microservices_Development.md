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
