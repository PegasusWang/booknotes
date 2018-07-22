> Architecture is a hypothesis, that needs to be proven by implementation and measurement. - Tom Bilb

# 1. What is Design and Architecture ?

> The only way to go fast, is to go well.

软件构建的目标是尽可能减小构建和维护软件系统所需的人力资源。
不要一开始就匆匆写代码（制造混乱）然后妄图以后重新设计，推翻重来会重走老路。

# 2. A Tale of Two Values

每个软件系统都提供两种价值：

-   behavior: 让机器实现需求(urgent but not always particularly important)
-   structure: software must be soft ,easy to change.(important but never particularly urgent)

大部分商业领导者和开发人员不能把紧急却不重要的特性和紧急且重要的特性区分开，这种错误导致人们更重视不重要的
特性而不是重要的系统架构。开发人员面临的窘境是商业领导者不具备评估架构重要性的能力，这种责任落到了软件开发 team 的身上。

# 3. Paradigm Overview

本章讲了三种编程范式：结构化、面向对象、函数式(你会发现都是几十年前发明的)

-   Structured Programming: Dijkstra 在 1968 发明，他提出随意使用 goto 语句对程序结构是有害的。

    > Structured programming imposes discipline on direct transfer of control.

-   Object-Oriented Programming: 1966 年由 Johan Dahl 和 Kristen Nvgaard 发明。

    > Object-oroented programming imposes discipline on indirect transfer of control.

-   Functional Programming: 1936 年在 Alonzo Church(阿隆佐·邱奇) 在 Lambda calculus (lambda演算)中提出。

    > Functional programming imposes discipline upon assignment.

三种编程范式很好地对应了架构中的三种概念：函数、组件分离、数据管理

# 4. Structured Programming

本章讲了 Dijkstra 提出结构化编程的过程。
所有程序都可以用三种基本结构表示：顺序、选择、循环

# 5. Object-Oriented Programming

Encapsulation, Inheritance, Polymorphism(提供了依赖倒置的能力)

To the software artchitect, the answer is clear: OO is the ability, through the use of polymorphism, to gain absolute
control over every source code dependency in the system. It allows the artchitect to create a plugin architecture, in
which modules that contain high-level policies are independent of modules that contain low-level details.

# 6. Functional Programming

函数式编程概念甚至早于编程本身。变量在函数式编程语言中不会变化(variables in functional languages do not vary)

Immutablility and Architecture: 为什么架构要考虑不可变性。答案很简单：所有的竞争条件、死锁、并发更新问题都源于可变变量。
架构师应该尽可能把处理放在不可变组件上。

虽然工具在变、软件在变，但是软件的本质几十年来并没有发生改变。

# 7. SRP: The Single Responsibility Principle

A module should be responsible to one, and only one, actor.

Actor 可以理解为一个抽象实体，一个类的方法应该仅仅属于相关的实体，不要把本应该属于其他实体的类放到一个类里。

# 8. OCP: The Open-Closed Principle

A software artifact should be open for extension but closed for modification.

通过合理分层，使得模块在需要扩充功能的时候，增加而不是修改它。低层次的组件修改的时候要保护高层次的组件不受影响。
一般通过把系统进行组件拆分，然后重组这些组件的依赖等级来保护高层次的组件不被低层次组件改变受到影响。

# 9. LSP: The Liskov Substitution Principle

所有使用基类的地方都可以使用子类替换。Anywhere you use a base class, you should be able to use a subclass and not know it.
要遵守Liskov替换原则，相对基类的对应方法，派生类服务（方法）应该不要求更多，不承诺更少。

# 10. ISP: The Interface Segregation Principle

不要强制客户端使用他们不需要的接口。
It is Harmful to depend on modules that contain more than you need.

# 11. DIP: The Dependency Inversion Principle

对于灵活的系统，源码应该仅仅依赖于抽象，而不是细节。在静态类型语言如 java 中，意味着 use，import include 语句应该仅引入
包含接口、抽象类、或者其他抽象定义的模块。

-   Don't refer to volatile concrete classes. Refer to abstract interfaces instead.
-   Don't derve from volatile concrete classes.
-   Don't override concrete functions.
-   Never mention the name of anything concrete and volatile.

    为了满足以上限制，经常使用抽象工厂创建对象

# 12. Components

能被独立部署的基本单元。(jar files, DLLS)
动态链接的文件，可以在运行时组合在一起，就是我们所说的架构中的软件组件(sofeware components)

# 13. Component Cohesion

Three principles of component cohesion:

-   REP: The Resuse/Release Equivalence Principle: The granule of resuse is the granule of release. (重用粒度等价与发布粒度)
    这样可以通过发布版本号来更新重用的组件。从软件设计的角度说，这个原则意味着组成一个组件的类和模块应该是内聚的，
    并且能被一起发布。


-   CCP: The Common Closure Principle: Gather into components those classes that chagne for the same reasons and at the
    same times.


-   CRP: The Common Resuse Principle: Don't force users of a component to depend on things they don't need.
    倾向于在一起被重复使用的类和模块应该放在一个组件里。

# 14. Component Coupling

-   The Acyclic Dependencies Principle: Allow no cycles in the component dependency graph.  (避免循环依赖，各个组件的依赖应该是有向无环图)
    如果有循环依赖，可以通过依赖倒置原则创见新的接口来打破循环依赖。(动态语言和静态语言实现上不同)

-   The Stable Dependencies Principle: Depend in the direction of stability.
    任何易变的组件都不应该依赖难以修改的组件。如何衡量一个组件的稳定性：一种方式是计算有多少组件依赖和被依赖(Fan-in, Fan-out)该组件

-   The Stable Abstractions Principle: A component shoud be as abstract as it is stable.
    一个稳定的组件应该是抽象的，而不稳定的组件应该是具体的。

# 15. What is Architecture ?

-   development: easy to develop
-   deployment: easily deployed with a single action
-   operation: A good software architecture communicates the operational needs of the system.
-   maintenance: separate the system in to components, and isolate those component through stable interfaces.

# 16. Independence

-   Use Cases: the artchitecture of the system must support the intent of the system
-   Operation:  Architecture plays a more substantial, and less cosmetic, role in supporting the operation of the system.
-   development: 康威定律
-   deployment: immediate deplyment

# 17. Boundaries: Drawing Lines

要在软件架构中划清楚界限，首先需要把系统分成组件。其中一些是核心业务逻辑、其他是包含必要函数但是不直接和业务逻辑关联的组件。
然后重组组件中的代码使得他们都指向同一个方向：业务逻辑。

# 18. Boundary Anatomy

软件界限划分的几种形式：单一文件、可单独部署的组件、线程、本地进程、服务

# 19. Policy And Level

-   Level: the distance from the inputs and outputs.

    Lower-level component should plug in to higher-level components.

# 20. Business Rules

-   Entities: An Entity is an object within our computer system that embodies a small set of critical business rules
    operating on Critical Business Data.

-   Use case: A use case is an object. It has one or more functions that implement the application-specific business
    rules.(比如数据校验层)

-   Request And Response Modules: The use case class accepts simple request data structures for its input, and returns
    simple response data structures as its output.

# 21. Screaming Architecture

-   The purpose of an architecture: 好的软件架构允许我们延迟选择所使用的框架、数据库、web 服务器等。

# 22. The Clean Architecture

   The Dependency Rule: source code dependencies must point only inward, toward higher-level policies.

![clean](https://pic1.zhimg.com/v2-07b2e4403c83a8b377ad14ab3589044c.jpg)

-   Entities: 实体封装业务领域规则，可以是包含方法的对象，或者一系列数据结构和函数。
-   Use Cases: 包含应用层业务逻辑。
-   Interface Adapters: convert data from the format most convenient for the use cases and entities, to the format most
    convenient for some external agency such as the database or the web.
-   Frameworks and Drivers: is generally composed of frameworks and tools such as the database and the web framework.

# 23 Presenters and Humble Objects

-   Humble Object Pattern: separation of the behaviors into testable and non-testable parts.

# 24 Partial Boundaries

three simple ways to partially implement an architectural boundary.

-   skip the last step: do all the work necessary to create independently compilable and deployable compoments, and then simple keep them together in the same component.
-   One-Dimensional Boundaries: Use Strategy pattern
-   Facades

# 25. Layers And Boundaries

You must weight the costs and determine where the architectural boundaries lie, and which should be fully implemented,
and which should be partially implemented and which should be ignored.

# 26. The Main Component

Main is a dirty low-level module in the outermost circle of the clean artchitecture. It loads everything up for the high level system, and then hands control over to it.

# 27. Services: Great and Small

 服务化

-   解耦谬论：服务化只实现了变量级的解耦，他们依然被共享处理器、网络甚至数据耦合。
-   独立开发和部署的谬论：历史证明基于组件式系统也能被很好部署；解耦谬论决定了开发和部署不能够完全独立。

服务必须被设计为其内部组件遵循『Dependency Rule』

# 28. The Test Boundary

-   Design for Testability: Don't Depend on volatile things. 比如不要引入 GUI 到测试里。
-   The Tesging API: decouple the structure of the tests from the structure of the application.

# 29. Clean Embedded Architecture

App-Titude Test:

# 30. The Database is a Detail

# 31. The Web is a Detail

# 32. Frameworks are details

Keep the framework behind an architectural boundary if at all possible, for as long as possible.

# 33. The Missing Chapter

-   Package By Layer: 起初最简单的设计是水平层次的设计。web 代码一层、业务逻辑一层、持久化一层。
-   Package By Feature: 垂直切分。放到一个 package 里， 根据被分组的概念来命名
-   Ports and Adapters
-   Package By Component: 把职责相关的代码放到一个粗粒度的组件里
