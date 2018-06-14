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
