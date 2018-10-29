title: 《Fluent Python》杂记
date: 2017-12-10 21:45:32
tags: python
---
> Python is a language for consenting adults. —Alan Runyan

重新看了下 《Fluent Python》，依旧还是很多东西没能消化完。简单记录下吧

<!--more-->


# 1.Python 数据模型

Python data model
可以看下 python 文档关于 data model 的讨论

# 2. 序列构成的数组

容器序列(存放引用)：list、tuple、collections.deque
扁平序列（存放值）：str、bytes、bytearray、memoryview、array.array

python2.7 列表推导有变量泄露问题，所以推导的临时变量不要和外部重名

python运行原理可视化: www.pythontutor.com

    t = (1, 2, [1,2])
    t[2] += [1,2]    # t 变成 (1, 2, [1,2,1,2]) 同时抛出异常，用dis模块查看
    # t[2].extend([1,2]) 没问题

尽量不要把可变类型放在 tuple 里;增量赋值不是原子的；
`+= *=` 对于可变和不可变对象区分对待，不可变对象会生成新对象（str除外,cpython优化过）

python 使用的排序算法 Timsort 是稳定的
内存视图：memoryview:让用户在不复制内容的情况下操作同一个数组的不同切片。

collections.deque 线程安全

# 3 字典和集合

可散列：如果一个对象是可散列的，在这个对象的生命周期中， 它的散列值是不变的。而且需要实现`__hash__`,`__eq__`

    class StrKeyDict0(dict):
        """如果一个类继承了dict，然后这个集成类提供了__missing__方法，
        那么__getitem__找不到键的时候，会自动调用它，而不是抛出Keyerror
        """

        def __missing__(self, key):
            if isinstance(key, str):    # 如果 str 的 key 还找不到就抛出 KeyError，没有这句会无限递归
                raise KeyError(key)
            return self[str(key)]

        def get(self, key, default):
            try:
                return self[key]
            except KeyError:    # 说明 __missing__ 也失败了
                return default

        def __contains__(self, key):
            """这个方法也是必须的，因为继承来的  __contains__ 没有找到也会去掉用__missing__"""
            return key in self.keys() or str(key) in self.keys()

dict 变种：

-   collections.OrderedDict: 保持 key 的顺序
-   collections.ChainMap: 容纳多个不同的映射对象
-   collections.Counter: 计数器
-   collections.UserDict : 其实就是把标准 dict 用纯 python 实习一遍


    import UserDict


    class StrKeyDict(UserDict):

        def __missing__(self, key):
            if isinstance(key, str):    # 如果 str 的 key 还找不到就抛出 KeyError，没有这句会无限递归
                raise KeyError(key)
            return self[str(key)]

        def __setitem__(self, key, item):
            self.data[str(key)] = item

        def __contains__(self, key):
            return str(key) in self.data

不可变映射类型： types.MappingProxyType  (>=python3.3)

不要在迭代字段和set 的同时修改它。可以先迭代获取需要的内容后放到一个新的dict里。
dict 实现是稀疏列表。

dict特点：

-   元素可散列
-   内存开销大
-   键查询很快
-   键次序取决于添加顺序
-   往字典里添加新键可能会改变已有键的顺序

set特点：

-   元素必须可散列
-   消耗内存
-   高效判断是否存在一个元素
-   元素次序取决于添加顺序
-   往字典里添加新元素可能会改变已有元素的次序

# 4 文本和字节序列

> 人类使用文本，计算机使用字节序列

字符的标识（码位），十进制数字，在unicode 中以4-6个十六进制数字表示
字符的具体表示取决于使用的编码，编码是在码位和字节序列之间转换时使用的算法
编码：码位-> 字节序列
解码：字节序列 -> 码位

Unicode 三明治：我们可以用一个简单的原则处理编码问题： 字节序列->字符串->字节序列。就是说程序中应当仅处理字符串，当需要保存到文件系统或者传输的时候，编码为字节序列

BOM：用来标记字节序

UnicodeEncodeError：字符串转成二进制序列。文本转成字节序列时，如果目标编码没有定义某个字符就会抛异常

UnicodeDecodeError: 二进制转成字符串。遇到无法转换的字节序列

chardet 检测文件编码

处理文本：在多系统中运行的代码需要指定打开和写入的编码，不要依赖默认的编码。除非想判断编码，否则不要在二进制模式中打开文本文件。

使用 unicodedata.normalize 函数对 unicode 规范化（标准等价物）。保存文本之前用 normalize('NFC', user_text) 清洗字符串

Unicode 排序：unicode collation algorith, UCA  使用 PyUCA 库。

双模式 API：根据接受的参数是字节序列或字符串自动处理。re 和 os 模块

cpython 16 位窄构建(narrow build)  32 位宽构建 (wild build) sys.maxunicode。窄构建无法处理 U+FFFF 以上码位

# 5 一等函数

高阶函数（higher-order function): 接受函数作为参数，或者把函数作为结果返回的函数。比如 map,filter,reduce 等（大部分可以被列表推导替代）

匿名函数：lambda 用于创建匿名函数。不过 lambda 定义体中无法赋值，也无法使用 while, try 等python 语句

可调用对象：内置的 callable() 函数判断是否可以调用

用户定义的可调用类型：任何 python 对象只要是先了 `__call__` 方法都可以表现得像函数

函数内省：使用 inspect 模块提取函数的签名、参数等

python3 函数注解： `def clip(text:str, max_len:'int > 0'=80) -> str:` 注解会存储在 函数的 `__annotations__`(一个dict) 属性中。注解对 python 解释器没有任何意义，只是给 IDE、框架、装饰器等工具使用。

支持函数式编程：

-   operator 模块：常用的有 attrgetter、itemgetter、methodcaller
-   functools 模块：reduce、partial（基于一个函数创建一个新的可调用对象，把原函数的某些参数固定）、dispatch、wraps

# 6 使用一等函数实现设计模式

程序设计语言会影响人们理解问题的出发点。

本章举了两个例子说明动态语言是如何简化设计模式的。（我个人感觉举的例子不是很好吧，有点过度设计的感觉）
之前曾经总结过使用 Python 实现设计模式，感兴趣的可以参考：

<http://python-web-guide.readthedocs.io/zh/latest/>

# 7 函数装饰器和闭包

#### 装饰器：可调用的对象，其参数是另一个函数。（说白了就是以函数作为参数的函数）两个特性：

-   能把被装饰的函数替换为其他函数
-   装饰器在加载模块时立即执行，通常是在导入时（即python加载模块时）。被装饰的函数只有在明确调用时运行

装饰器语法糖:

    # 等价于 target = decorate(target)
    @decorate
    def target():
        print('hehe')

#### 闭包:闭包指延伸了作用域的函数，其中包含函数定义体中引用、但是不在定义体中定义的非全局变量。比如被装饰的函数能访问装饰器函数中定义的变量（非全局的）

#### 自由变量：

    def make_averager():
        series = []

        def averager(new_value):
            # series 在 averager 中叫做自由变量(free variable)，指未在本地作用域中绑定的变量
            series.append(new_value)
            total = sum(series)
            return total / len(series)

#### nonlocal 声明：先来看个例子：

    def make_averager():
        count = 0
        total = 0

        def averager(new_value):
            # 直接运行到这里会报错，UnboundLocalError，因为对于非可变类型，会隐式创建局部变量 count，count 不是自由变量了
            count += 1
            total += new_value
            return total

使用 python3 引入的 nonlocal 刻意把变量标记为自由变量。

    def make_averager():
        count = 0
        total = 0

        def averager(new_value):
            nonlocal count, total   # python2 可以用 [count] 把需要修改的变量存储为可变对象
            count += 1
            total += new_value
            return total

#### functools.wraps 装饰器：把相关属性从被装饰函数复制到装饰器函数中

#### 标准库中的装饰器：property、classmethod、staticmehtod、functools.lru_cache、functools.singledispatch

    - lru_cache: 采用 least recent used 算法实现的缓存装饰器
    - singledispatch: 为函数提供重载功能。被其装饰的函数会成为泛函数(generic function):根据第一个参数的类型，用不同的方式执行相同操作的一组函数。替代多个 if/else isinstance 判断类型执行不同分之

#### 叠放装饰器:

    # 下边等价于 f = d1(d2(f))
    @d1
    @d2
    def f():
        print('f')

#### 参数化装饰器: 创建一个装饰器工厂函数，把参数传给它，返回一个装饰器，然后再把它应用到要装饰的函数上。

    registry = set()  # <1>

    def register(active=True):  # <2>    工厂函数
        def decorate(func):  # <3>
            print('running register(active=%s)->decorate(%s)'
                  % (active, func))
            else:
                registry.discard(func)  # <5>

            return func  # <6>
        return decorate  # <7>

    @register(active=False)  # <8>
    def f1():
        print('running f1()')

    @register()  # <9>    # 即使没有参数，工厂函数必须写成调用的形式
    def f2():
        print('running f2()')

#### 使用 class 实现装饰器：看上边多重嵌套的装饰器是不是有点不太优雅， 其实复杂的装饰器笔者更喜欢用 class 实现。还记得 `__call__` 方法吗，改写下上边这个例子

    class register(object):
        registry = set()

        def __init__(self, active=True):
            self.active = active

        def __call__(self, func):
            print('running register(active=%s)->decorate(%s)'
                % (self.active, func))
            if self.active:
                self.registry.add(func)
            else:
                self.registry.discard(func)

            return func

# 8 对象引用、可变性和垃圾回收

#### 变量：我们可以把变量理解为对象的标注（便利贴），多个标注就是别名。变量保存的是对象的引用。

#### 比较对象： 判断是同一个对象: id(obj1) == id(obj2)  或者 obj1 is obj2。比较两个对象的值用 obj1 == obj2 (obj1.**eq**(obj2))。你会发现一般我们用 `some_obj is None` 来判断一个对象是否是 None，说明 None 是个单例对象。

#### 元组(tuple)的相对不可变性：元祖的不可变指的是保存的引用不可变(tuple中不可变的是元素的标识)，与引用的对象无关。比如如果元祖的元素是个 list，我们是可以修改这个 list 的。这也会导致有些元祖无法散列

注意 str/bytes/array.array 等单一类型序列是扁平的，它们保存的不是引用，而是在连续内存中保存数据本身(字符、字节和数字)

#### 默认做浅复制：构造函数或者 [:] 方法默认是浅复制。如果元素都是不可变的，浅复制没有问题。 http://www.pythontutor.com

#### 深拷贝: copy.deepcopy 和 copy.copy 能为任意对象做深复制(副本不共享内部对象的引用)和浅复制。我们可以自定义 `__copy__()` 和 `__deepcopy__()` 控制拷贝行为

#### 函数的参数作为引用时：python 唯一支持的传参模式是共享传参（call by sharing），指函数的各个形式参数获得实参中各个引用的副本。也就是说，函数内部的形参是实参的别名。这个方案的结果就是，函数可能会修改作为参数传入的可变对象，但是无法修改那些对象的标识（即不能把一个对象替换成另一个对象）。（笔者觉得这章解释非常好，之前网上一大堆讨论python究竟是值传递还是引用传递的都是不准确的）

#### 不要使用可变类型作为参数的默认值：这个坑写 py 写多的人应该都碰到过。函数默认值在定义函数时计算（通常是加载模块时），因此默认值变成了函数对象的属性。如果默认值是可变对象，而且修改了它的值，后续的函数调用就会受影响。一般我们用 None 作为占位符。

    def func(l=None):    # 不要写 def func(l=[]):
        # 使用 None 作为占位符（pylint 默认会提示可变类型参数作为默认值，所以俺经常安利用 pylint 检测代码，防范风险）
        l = None or []

所以，一般对于一个函数，要么确认是要修改可变参数，要么返回新的值（使用参数的拷贝），请不要两者同时做。(笔者在小书 web guide 中明确提醒过)

#### del 和 垃圾回收: del 语句删除名称，而不是对象(删除引用而不是对象）。只有对象变量保存的是对象的最后一个引用的时候，才会被回收。Cpython 中垃圾回收主要使用的是引用计数。不要轻易自定义 `__del__` 方法，很难用对。

#### 弱引用(weakref)：有时候需要引用对象，而又不让对象存在的时间超过所需时间，经常用在缓存中。弱引用不会增加对象的引用数量，不会妨碍所指的对象被当做垃圾回收

-   WeakValueDictionary: 一种可变映射，值是对象的弱引用。还有 WeakKeyDictionary、WeakSet、finalize。

#### Python 对不可变类型施加的把戏（CPython 实现细节）

-   对于元祖 t, t[:] 和 tuple(t) 不会创建副本，返回的是引用（这点和list 不同）。str, bytes 和 frozenset 也有这种行为

# 9 符合 Python 风格的对象

鸭子类型(duck typing): 只需按照预定行为实现对象所需的方法即可。

#### 对象的表示形式: repr() 让开发者理解的方式返回对象的字符串表示。str() 用户理解的方式返回对象的字符串表示

#### classmethod 和 staticmehtod: classmethod 定义操作类而不是操作实例的方法，第一个参数是类本身，最常见的用途是定义定义备选构造方法(返回 cls(\*))。staticmehtod 方法就是普通函数，只是碰巧在类的定义体中。

#### Python的私有属性和『受保护』属性：python没有 private修饰符，可以通过双下划线  `__attr` 的形式定义，python 的子类会在存储属性名的时候在前面加上一个下划线和类名。这个语言特性成为名称改写（name mangling）。通常受保护的属性使用一个下划线作为前缀，算是一种命名约定，调用者不应该在类外部访问这种属性。

python 没有访问控制和 java 设计迥然不同，本章最后的杂谈讨论了这两种设计。在 python 中，我们可以先使用公开属性，等需要时再变成特性。

#### 使用`__slots__`类属性节省空间：默认情况下，python在各个中名为`__dict__`的字典存储实力属性，当生成大量对象时字典会消耗大量内存（底层是稀疏数组），通过`__slots__`类属性，让解释器在元祖而不是字典中存储实例属性，能大大节省内存。(不支持继承)

-   每个子类都需要定义 `__slots__`，解释器会忽略继承的 `__slots__`属性
-   实例只能拥有 `__slots__` 属性，除非把 `__dict__` 也加到 `__slots__` 里（这样就失去了节省内存的功效）

#### 覆盖类属性：python有个独特的特性，类属性可以为实例属性提供默认值

# 10 序列的修改、散列和切片

#### 协议和鸭子类型: python中我们刻意创建序列类型而无需使用继承，只需实现符合序列协议的方法。

鸭子类型： 在面向对象编程中，协议是非正式的接口，只在文档中定义，在代码中不定义。例如 python 序列协议只需要实现 `__lens__`  和 `__getitem__` 两个方法。只关心行为而不关心类型。

我们可以模仿 python 对象的内置方法来编写符合 python 风格的类。（具体的大家还是看下书中的代码示例吧，这一章举得例子不错）

# 11 接口：从协议到抽象基类

#### 使用猴子补丁在运行时实现协议：

运行时修改类或者模块，而不改动源码。可以在运行时让类实现协议

#### 抽象基类：collections.abc 模块

-   Iterable，Container 和 Sized:：Iterable 通过 `__iter__` 支持迭代，Container 通过  `__contains__` 支持 in 操作符, Sized 通过 `__len__` 支持 len() 函数
-   Sequence， Mapping and Set ：不可变集合类型
-   MappingView: 映射方法 .items()，.kesy()，.values() 返回的对象分别是 ItemsView,KeysView 和 ValuesView 实例
-   Callable 和 Hashable: 主要作用市委内置 isinstance 提供支持，以一种安全的方式判断对象能不能调用或散列。python 提供了callable 内置函数却没有提供 hashable() ，用 isinstance(obj, Hashable) 判断
-   Iterator
-   numbers包：Number, Complex，Real，Rational，Integral

#### 定义并使用抽象基类

    import abc
    class Base(abc.ABC):    # py3， py2 中使用  __metaclass__ = abc.ABCMeta
        @abc.abstractmethod    # 该装饰器应该放在最里层
        def some_method(self):  # 这里可以只有 docstring 省略函数体
            """抽象方法，在抽象基类出现之前抽象方法用 Raise NotImplementedError 语句表明由子类实现"""

#### 使用 register 方法注册虚拟子类:

在抽象基类上调用 register 方法注册其虚拟子类，issubclass 和 isinstance 都能识别，但是注册的的类不会从抽象基类中继承任何方法和属性。查看虚拟子类的 `__mro__` 会发现抽象基类不在其中(没继承其属性和方法)

#### `__subclasshook__` : 即使不注册，抽象基类也能把一个类识别为虚拟子类。定义 `__subclasshook__` 方法动态识别子类。参考 abc.Sized 源码

#### 强类型和弱类型：

如果一门语言很少隐式转换类型，说明它是强类型语言(java/c++/python)。如果经常这么做，是弱类型语言(php,javascript,perl)。强类型能及早发现缺陷

#### 静态和动态类型：

在编译时期检查类型的语言是静态语言，运行时检查类型的语言是动态语言。静态类型需要类型声明（有些现代语言使用类型推导避免部分类型声明）。静态类型便于编译器和 IDE
及早分析代码、找出错误和提供其他服务（优化、重构等）。动态类型便于代码重用，代码行数更少，而且能让接口自然成为协议而不提早实行。

# 12 继承的优缺点

#### 子类化内置类型：

内置类型的方法不会调用子类覆盖的方法。不要子类化C语言实现的内置类型（list,dict等），用户自定义的类应该继承自 collections
模块。collections.UserDict, UserList and UserString

#### 多重继承和方法解析顺序

任何支持多重继承的语言都要处理潜在的明明冲突问题，菱形继承问题。python  会按照 方法解析顺序MRO(method resolution order)
遍历继承图。类都有一个  `__mro__` 属性，它的值是一个tuple，按照顺序列出各个超类，直到 object 类。MRO 根据 C3 算法计算

#### 处理多重继承

多重继承增加了可选方案和复杂度

-   把接口继承和实现继承区分开。明确一开始为什么创建子类。1.继承接口，创建子类型，实现『是什么』关系。2.继承实现，重用代码。通过继承重用代码是实现细节，通常可换成组合和委托。接口继承是框架的支柱
-   使用抽象基类显示表示接口
-   通过 mixin 重用代码。mixin 不定义新类型, 只是打包方法，便于重用。mixin 类绝对不能实例化，应该提供某方面的特定行为，只是实现少量关系非常紧密的方法
-   明确指名 mixin。类应该以 mixin 后缀
-   抽象基类可以作为 mixin，但是反过来不成立
-   不要子类化多个具体类。具体类的超类中除了一个具体类，其他都应该是抽象基类或者 mixin


    class MyConcreteClass(Alpha, Beta, Gamma):
        """ 如果 Alpha 是具体类，Beta 和 Gamma 必须是抽象基类或者 mixin"""
        pass

-   创建聚合类。django 中的 ListView，tinker中的 Widget


    class Widget(BaseWidget, Pack, Place, Grid):
        pass

-   优先使用组合而非继承。子类化是一种紧耦合，不要过度使用

# 13  正确重载运算符

python 不允许用户随意创建运算符，禁止重载内置类型的运算符。python支持运算符重载是其在科学计算领域使用广泛的原因。

-   一元运算符：始终返回一个新对象。
-   NotImplemented 是个特殊的单例值，如果中缀运算符特殊方法不能处理给定的操作数，要把它返回给解释器。NotImplementedError 是一种异常，抽象类中的方法把它 raise 出，提醒子类必须覆盖。

          def __add__(self, other):
              try:
                  pairs = itertools.zip_longest(self, other, fillvalue=0.0)
                  return Vector(a + b for a, b in pairs)
              except TypeError:
                  # 返回 NotImplemented 解释器会尝试调用 反向运算符方法 __radd__
                  return NotImplemented

          def __radd__(self, other):
              return self + other

-   增量赋值运算符不会修改不可变目标，而是新建实例，然后重新绑定。

# 14 可迭代对象、迭代器和生成器

解释器需要迭代对象 x 时，会自动调用 iter(x)。内置的 iter 有以下作用：

-   检查对象是否实现了 `__iter__` ，如果实现了就调用它获取一个迭代器
-   如果没有实现 `__iter__` 方法，但是实现了 `__getitem__` 方法，python 会创建一个迭代器，尝试按照顺序（从索引0）获取元素
-   如果尝试失败，抛出 TypeError 异常

#### 可迭代对象:

如果对象实现了能返回迭代器的 `__iter__` 方法，就是可迭代的。序列都可以迭代；实现了  `__getitem__` 方法，而且其参数是从 0 开始的索引，这种对象也可以迭代。

#### 标准迭代器接口有两个方法：

-   `__next__`: 返回下一个可用的元素，没有元素抛出 StopIteration
-   `__iter__`: 返回 self，以便在应该使用可迭代对象的地方使用迭代器，例如 for 循环中

检查对象是否是迭代器的最好方法是调用 isinstance(x, abc.Iterator)

#### 迭代器:

实现了无参数的 `__next__` 方法，返回序列中下一个元素；如果没有元素了，抛出 StopIteration 异常。python中的迭代器还实现了
`__iter__` 方法，因此迭代器也可以迭代。

#### 二者区别：

迭代器可以迭代，但是可迭代的对象不是迭代器。可迭代的对象一定不能是自身的迭代器。也就是说，可迭代对象必须实现  `__iter__`
，但是不能实现 `__next__`

#### 生成器函数

只要 python 的函数体中有 yield 关键字，该函数就是生成器函数。调用生成器函数会返回一个生成器对象。生成器函数是生成器工厂

#### 标准库的中生成器函数

-   用于过滤的生成器函数: itertools.takewhile/compress/dropwhile/filter/filterfalse/islice/
-   用于映射的生成器函数: 内置的 enumerate/map  itertools.accumulate/starmap
-   用于合并的生成器函数：itertools.chain/from_iterable/product/zip_longest  内置的 zip
-   从一个元素产生多个值，扩展输入的可迭代对象: itertools.combinations/combinations_with_replacement/count/cycle/permutations/repeat
-   产出输入可迭代对象的全部元素，以某种方式排列：itertools.groupby/tee     内置的 reversed

#### 可迭代的规约函数

归约函数：接受一个可迭代的对象，返回单个结果。all/any/max/mini/functools.reduce/sum    all/any 有短路特性

#### 把生成器当协程

`.send()` 方法致使生成器前进到下一个 yield 语句，还允许使用生成器的客户把数据发给自己，不管传给 send 方法什么参数，
那个参数都会成为生成器函数定义体中对应的 yield 表达式的值。

# 15 上下文管理器和 else 块

#### EAFP vs LBYL

-   EAFP: easier to ask for forgiveness than permission
-   LBYL: look before you leap

#### 上下文管理器和 with 块

with 语句用来简化 try/finally 模式。经常用在管理事务，维护锁、条件和信号，给对象打补丁等。

    class LookingGlass:

        def __enter__(self):  # <1>
            import sys
            self.original_write = sys.stdout.write  # <2>
            sys.stdout.write = self.reverse_write  # <3>
            return 'JABBERWOCKY'  # <4>

        def reverse_write(self, text):  # <5>
            self.original_write(text[::-1])

        def __exit__(self, exc_type, exc_value, traceback):  # <6>
            import sys  # <7>
            sys.stdout.write = self.original_write  # <8>
            if exc_type is ZeroDivisionError:  # <9>
                print('Please DO NOT divide by zero!')
                return True  # <10>

#### contextlib 模块中的实用工具

`@contextmanager` 装饰器能减少创建上下文管理器的样板代码。只需要实现一个 yield 语句的生成器，生成想让 `__enter__`
方法返回的值。

    @contextlib.contextmanager
    def looking_glass():
        import sys
        original_write = sys.stdout.write

        def reverse_write(text):
            original_write(text[::-1])

        sys.stdout.write = reverse_write
        msg = ''  # <1>
        try:
            yield 'JABBERWOCKY'  # 产出一个值，这个值会绑定到with语句中的 as 子句的目标变量上
        except ZeroDivisionError:  # <2>
            msg = 'Please DO NOT divide by zero!'
        finally:
            sys.stdout.write = original_write  # <3>
            if msg:
                print(msg)  # <4>

# 16 协程

句法上看，协程和生成器类似，都是定义体中包含 yield 关键字的函数。但在协程中，yield 通常出现在表达式右边(datum =
yield)，可以产出值，也可以不产出。如果yield 关键字后边没有表达式，那么生成器产出 None。调用方可以用 send
方法把数据提供给协程。从根本上把 yield 视作控制流程的方式。

#### 生成器如何进化成协程

python2.5 之后yield 关键在能在表达式中使用，而且生成器 api 中增加了 `.send(value)` 方法。生成器的调用方可以用 send
发送数据，发送的数据会成为生成器函数中 yield 表达式的值。因此生成器能当做协程使用。协程是指一个过程，这个过程与调用方协作，产出由调用方提供的值。

协程使用 next 函数预激(prime)，即让协程向前执行到第一个 yield 表达式。

#### 预激（prime）协程的装饰器

启动协程之前需要 prime，方法是调用 send(None) 或者 next() 。为了简化协程的语法，有时候会使用一个 预激 装饰器。
比如 tornado.gen 装饰器。yield from 调用协程会自动 预激

    from functools import wraps

    def coroutine(func):
        """向前执行到第一个 yield 表达式，预激 func """
        @wraps(func)
        def primer(*args, **kwargs):
            gen = func(*args, **kwargs)    # 获取生成器对象
            next(gen)    # prime
            return gen
        return primer

#### 终止协程和异常处理

协程中未处理的异常会向上冒泡

-   generator.throw(exc_type)
-   generator.close()

#### 让协程返回值

协程中 return 表达式的值会偷偷传给调用方，赋值给 StopIteration 异常的一个属性 value

    try:
        coro.send(None)
    except StopIteration as exc:
        result = exc.value

#### yield from(python3)

RESULT = yield from EXPR  等效代码如下，比较烧脑， 但是能帮助我们理解 yield from 如何工作

    _i = iter(EXPR)  # EXPR 是任何可迭代对象
    try:
        _y = next(_i)  # 预激(prime) 子生成器
    except StopIteration as _e:
        _r = _e.value  # 如果抛出 StopIteration 获取 value 属性（返回值）
    else:
        while 1:  # 运行这个循环时，委派生成器会阻塞，只作为调用方和子生成器之间的通道
            try:
                _s = yield _y  # 产出字生成器当前产出元素；等待调用方发送 _s 中保存的值
            except GeneratorExit as _e:  # 用于关闭委派生成器和子生成器
                try:
                    _m = _i.close
                except AttributeError:   # 子生成器是任何可迭代对象，所以可能没有 close 方法
                    pass
                else:
                    _m()
                raise _e
            except BaseException as _e:  # 处理调用方通过 throw 方法传入的异常
                _x = sys.exc_info()
                try:
                    _m = _i.throw
                except AttributeError:    # 子生成器是任何可迭代对象，所以可能没有 throw 方法
                    raise _e
                else:  # 如果子生成器有 throw 方法，调用它并传入调用方发来的异常
                    try:
                        _y = _m(*_x)
                    except StopIteration as _e:
                        _r = _e.value
                        break
            else:  # 如果产出值时没有异常
                try:  尝试让子生成器向前执行
                    if _s is None:  # <11>
                        _y = next(_i)
                    else:
                        _y = _i.send(_s)
                except StopIteration as _e:  # <12>
                    _r = _e.value
                    break

    RESULT = _r  # 返回的值是 _r，即整个 yield from 表达式的值

# 17 使用 concurrent.futures 处理并发

python3.2 后引入了 concurrent.futers 模块用来处理并发。该模块引入了 TreadPoolExecutor 和 ProcessPoolExecutor
类，这两个类实现的接口能分别在不同的线程和进程中执行可调用的对象。

    from concurrent import futures

    from flags import save_flag, get_flag, show, main  # <1>

    MAX_WORKERS = 20  # <2>


    def download_one(cc):  # <3>
        image = get_flag(cc)
        show(cc)
        save_flag(image, cc.lower() + '.gif')
        return cc


    def download_many(cc_list):
        workers = min(MAX_WORKERS, len(cc_list))  # <4>
        with futures.ThreadPoolExecutor(workers) as executor:  # <5>
            res = executor.map(download_one, sorted(cc_list))  # <6>

        return len(list(res))  # <7>


    if __name__ == '__main__':
        main(download_many)  # <8>

#### Future(期物)(中文版翻译感觉这个名字怪怪的)

concurrent.futures.Future: Feature 类的实例都表示可能已经完成或者尚未完成的延迟计算，可以调用它的 result() 方法获取结果

    def download_many(cc_list):
        cc_list = cc_list[:5]  # <1>
        with futures.ThreadPoolExecutor(max_workers=3) as executor:  # <2>
            to_do = []
            for cc in sorted(cc_list):  # <3>
                future = executor.submit(download_one, cc)  # <4>
                to_do.append(future)  # <5>
                msg = 'Scheduled for {}: {}'
                print(msg.format(cc, future))  # <6>

            results = []
            for future in futures.as_completed(to_do):  # <7>
                res = future.result()  # <8>
                msg = '{} result: {!r}'
                print(msg.format(future, res)) # <9>
                results.append(res)

        return len(results)

#### 阻塞型 IO 和 GIL

GIL 一次只允许一个线程执行 python 字节码。但是标准库中所有执行阻塞型 I/O 操作的函数，在等待操作系统返回的结果时都会释放
GIL，这意味着python 在这个层次上能使用多线程，一个 python 线程等待网络请求时，阻塞型 I/O 会释放(sleep 函数也会)
GIL，运行另一个线程。因此尽管有 GIL，python 线程还是能在 IO 密集型应用中发挥作用。

#### concurrent.futures.ProcessPoolExecutor 绕开 GIL

# 18 使用 asyncio 处理并发

#### asyncio 使用事件循环驱动的协程实现并发

    import asyncio

    import aiohttp  # <1>

    from flags import BASE_URL, save_flag, show, main  # <2>


    @asyncio.coroutine  # <3>
    def get_flag(cc):
        url = '{}/{cc}/{cc}.gif'.format(BASE_URL, cc=cc.lower())
        resp = yield from aiohttp.request('GET', url)  # <4>
        image = yield from resp.read()  # <5>
        return image


    @asyncio.coroutine
    def download_one(cc):  # <6>
        image = yield from get_flag(cc)  # <7>
        show(cc)
        save_flag(image, cc.lower() + '.gif')
        return cc


    def download_many(cc_list):
        loop = asyncio.get_event_loop()  # <8>
        to_do = [download_one(cc) for cc in sorted(cc_list)]  # <9>
        wait_coro = asyncio.wait(to_do)  # <10>
        res, _ = loop.run_until_complete(wait_coro)  # <11>
        loop.close() # <12>

        return len(res)


    if __name__ == '__main__':
        main(download_many)

#### 避免阻塞型调用

两种方式避免阻塞型调用中止整个应用程序的进程：

-   在单独的线程中运行各个阻塞型操作(线程内存消耗达到兆字节，取决于操作系统)
-   把每个阻塞型调用操作转成非阻塞的异步调用

##### 在 asyncio 中使用 Executor 对象，防止阻塞事件循环

python 访问本地文件系统会阻塞，硬盘IO 阻塞会浪费几百万个 cpu 周期。解决方法是使用时间循环对象的 run_in_executor 方法。

    @asyncio.coroutine
    def download_one(cc, base_url, semaphore, verbose):
        try:
            with (yield from semaphore):
                image = yield from get_flag(base_url, cc)
        except web.HTTPNotFound:
            status = HTTPStatus.not_found
            msg = 'not found'
        except Exception as exc:
            raise FetchError(cc) from exc
        else:
            loop = asyncio.get_event_loop()  # 获取事件循环对象的引用
            loop.run_in_executor(None,  # None 使用默认的 TrreadPoolExecutor 实例
                    save_flag, image, cc.lower() + '.gif')  # 传入可调用对象
            status = HTTPStatus.ok
            msg = 'OK'

        if verbose and msg:
            print(cc, msg)

        return Result(status, cc)
    asyncio 的事件循环背后维护一个 ThreadPoolExecutor 对象，我们可以调用 run_in_executor 方法， 把可调用的对象发给它执行。

#### 从回调到 Futures 和 协程

回调地狱：如果一个操作需要依赖之前操作的结果，那就得嵌套回调。

python 中的回调地狱：

    def stage1(response1):
        request2 = step1(response1)
        api_call2(request2, stage2)


    def stage2(response2):
        request3 = step2(response2)
        api_call3(request3, stage3)


    def stage3(response3):
        step3(response3)


    api_call1(request1, step1)

使用 协程 和 yield from 结构做异步编程，无需用回调

    @asyncio.coroutine
    def three_stages(request1):
        response1 = yield from api_call1()
        request2 = step1(response1)
        response2 = yield from api_call2(request2)
        request3 = step2(response2)
        response3 = yield from api_call3(request3)
        step3(response3)

    # 协程不能直接调用，必须用事件循环显示指定协程的执行时间，或者在其他排定了执行时间的协程中使用 yield from 表达式把它激活
    loop.create_task(three_stages(request1))

何时使用 yield from：基本原则很简单，yield from 只能用于 协程 和 asyncio.Future 实例（包括 Task
实例）。有些肆意混淆了协程和普通函数的 api 比较棘手。

驱动协程：只有驱动协程，协程才能做事，而驱动 asyncio.coroutine 装饰的协程有两种方法，要么使用 yield from，要么传给
asyncio 包中某个参数为协程或者 Futures 的函数，例如 run_until_complete

#### 使用 asyncio 包编写服务器

可以使用 asyncio 编写 tcp/udp 服务器，使用 aiohttp 编写 web 服务器。具体看各自的文档吧。

# 19 动态属性(attribute)和特性(property)

python 中，数据的属性和处理数据的方法统称为属性（attribute），方法是可调用的属性。特性（property）是不改变类接口的前提下，使用存取方法（读值和设值）修改数据属性。

统一访问原则：不管服务是由存取还是计算实现的，一个模块提供的所有服务都应该统一的方式使用。

## 使用动态属性转换数据

#### 使用动态属性访问数据

    from collections import abc

    class FrozenJSON:
        """A read-only façade for navigating a JSON-like object
        using attribute notation
        """

        def __init__(self, mapping):
            self.__data = dict(mapping)  # <1>

        def __getattr__(self, name):  # <2>
            if hasattr(self.__data, name):
                return getattr(self.__data, name)  # <3>
            else:
                return FrozenJSON.build(self.__data[name])  # <4>

        @classmethod
        def build(cls, obj):  # <5>
            if isinstance(obj, abc.Mapping):  # <6>
                return cls(obj)
            elif isinstance(obj, abc.MutableSequence):  # <7>
                return [cls.build(item) for item in obj]
            else:  # <8>
                return obj

#### 处理无效属性名

    def __init__(self, mapping):
        self.__data = {}
        for key, value in mapping.items():
            if keyword.iskeyword(key):  # <1>
                key += '_'    # 和 python 重名的关键字加上下划线
            self.__data[key] = value

#### 使用 `__new__` 以灵活的方式创建对象

实际上用来构建对象的方法是 `__new__`，`__init__` 是初始化方法。`__new__` 必须返回一个实例，作为 `__init__`
方法的第一个参数。

    def __new__(cls, arg):  # <1>
        if isinstance(arg, abc.Mapping):
            return super().__new__(cls)  # <2>
        elif isinstance(arg, abc.MutableSequence):  # <3>
            return [cls(item) for item in arg]
        else:
            return arg

## 使用特性验证属性

    class LineItem:

        def __init__(self, description, weight, price):
            self.description = description
            self.weight = weight  # <1>
            self.price = price

        def subtotal(self):
            return self.weight * self.price

        @property  # <2>
        def weight(self):  # <3>
            return self.__weight  # <4>

        @weight.setter  # <5>
        def weight(self, value):
            if value > 0:
                self.__weight = value  # <6>
            else:
                raise ValueError('value must be > 0')  # <7>

## 解析 property

property 签名

    class property(fget=None, fset=None, fdel=None, doc=None)

-   特性会覆盖实例属性。特性都是【类属性】，但是特性管理的其实是实例属性的存取。obj.attr 这样的表达式不会从 obj 开始寻找
    attr，而是从 `obj.__class__` 开始，且仅当类中没有 attr 的属性时， python 才会在 obj 实例中寻找。

## 定义一个特性工厂函数

    def quantity(storage_name):  # <1>

        def qty_getter(instance):  # <2>
            return instance.__dict__[storage_name]  # <3>

        def qty_setter(instance, value):  # <4>
            if value > 0:
                instance.__dict__[storage_name] = value  # <5>
            else:
                raise ValueError('value must be > 0')

        return property(qty_getter, qty_setter)  # <6>


    class LineItem:
        weight = quantity('weight')  # <1>
        price = quantity('price')  # <2>

        def __init__(self, description, weight, price):
            self.description = description
            self.weight = weight  # <3>
            self.price = price

        def subtotal(self):
            return self.weight * self.price  # <4>

## 处理属性删除操作

    class BlackKnight:

      def __init__(self):
          self.members = ['an arm', 'another arm',
                          'a leg', 'another leg']
          self.phrases = ["'Tis but a scratch.",
                          "It's just a flesh wound.",
                          "I'm invincible!",
                          "All right, we'll call it a draw."]

      @property
      def member(self):
          print('next member is:')
          return self.members[0]

      @member.deleter
      def member(self):
          text = 'BLACK KNIGHT (loses {})\n-- {}'
          print(text.format(self.members.pop(0), self.phrases.pop(0)))

## 处理属性的重要属性和函数

##### 影响属性处理方式的特殊属性

-   `__class__`: 对象所属类的引用。 `obj.__class__` 与 type(obj) 作用相同。python的某些特殊方法比如
    `__getattr__`，只在对象的类中寻找，而不在实例中寻找
-   `__dict__`: 存储对象或者类的可写属性。
-   `__slots__`: 字符串tuple，限制允许有的属性。

#### 处理属性的内置函数

-   dir: 列出对象的大多数属性
-   getattr: 从 obj 对象中获取对应名称的属性。获取的属性可能来自对象所属的类或者超类。
-   hasattr: 判断对象中存在指定的属性
-   setattr: 创建新属性或者覆盖现有属性
-   vars： 返回对象的 `__dict__` 属性

#### 处理属性的特殊方法

-   `__delatttr__(self, name)` 使用 del 删除属性就会调用这个方法
-   `__dir__(self)`: 把对象传给 dir 函数时候调用
-   `__getattr__`: 仅当获取指定的属性失败，搜索过 obj、Class、和超类之后调用
-   `__getattribute__`: 尝试获取指定的属性时总会调用这个方法，寻找的属性是特殊属性或者特殊方法时候除外。为了防止获取 obj
    的属性无限递归， `__getattribute__` 方法的实现要使用  `super().__getattribute__(obj, name)`
-   `__setattr__`: 尝试设置指定的属性总会调用
    # 20 属性描述符

描述符是对多个属性运用相同存储逻辑的一种方式。例如 orm 中的字段类型是描述符。描述符是实现了特定协议的类，这个协议包括
`__get__ __set__ __delete__` 方法。描述符的用法是创建一个实例，作为另一个类的类属性。

    class Quantity:  # <1>

        def __init__(self, storage_name):
            self.storage_name = storage_name  # <2>

        def __set__(self, instance, value):  # <3>
            if value > 0:
                instance.__dict__[self.storage_name] = value  # <4>
            else:
                raise ValueError('value must be > 0')


    class LineItem:
        weight = Quantity('weight')  # <5>
        price = Quantity('price')  # <6>

        def __init__(self, description, weight, price):  # <7>
            self.description = description
            self.weight = weight
            self.price = price

        def subtotal(self):
            return self.weight * self.price

覆盖型描述符：实现 `__set__`  方法的描述符属于覆盖型描述符。
非覆盖型描述符：没有实现 `__set__` 方法的描述符。

方法是描述符

#### 描述符用法建议：

-   使用 property 以保持简单: 内置的 property 实现的其实是覆盖型描述符
-   只读描述符必须有 `__set__` 方法: 如果使用描述符类实现只读属性， `__get__ __set__` 两个方法必须定义，否则实例的同名属性会遮盖描述符。只读属性的 `__set__` 只需抛出 AttributeError 异常，并提供合适的错误消息
-   用于验证的描述符可以只有 `__set__`
-   只有 `__get__` 方法的描述符可以实现高效缓存
-   非特殊的方法可以被实例属性覆盖

# 21 类元编程

元编程是指在运行时创建或者定制类的技术。除非开发框架，否则不要编写元类
类装饰器能以较为简单的的方式做到需要使用元类去做的事情-创建时定制类。缺点是无法继承

#### 导入时和运行时比较

python 中的 import 不只是声明，进程首次导入模块时，还会运行所导入模块中的全部顶层代码。导入时，解释器会执行执行每个类的定义体.
（原书有段代码示例非常好地解释了导入的问题）

#### 元类基础

感觉这一章写得不如笔者之前写的一篇博客《简单的python元类》好理解。

    class EntityMeta(type):
        """Metaclass for business entities with validated fields"""

        def __init__(cls, name, bases, attr_dict):
            super().__init__(name, bases, attr_dict)  # <1>
            for key, attr in attr_dict.items():  # <2>
                if isinstance(attr, Validated):
                    type_name = type(attr).__name__
                    attr.storage_name = '_{}#{}'.format(type_name, key)

    class Entity(metaclass=EntityMeta):  # <3>
        """Business entity with validated fields"""

#### 元类的特殊方法 `__prepare__`

某些应用中可能想知道属性定义的顺序，解决办法是使用 python3 引入的 `__prepare__` 。这个特殊方法只在元类中有用，而且必须是类方法。解释器调用元类的 `__new__` 之前会先调用 `__prepare__`，使用类定义体中的属性创建映射。
元类构建新类时， `__prepare__` 方法返回的映射会传给 `__new__` 的最后一个参数，然后再传给  `__init__` 方法。

    class EntityMeta(type):
        """Metaclass for business entities with validated fields"""

        @classmethod
        def __prepare__(cls, name, bases):    # py3, must be a class method
            return collections.OrderedDict()  # <1> return empty OrderedDict, where the class attritubes will be stored
        def __init__(cls, name, bases, attr_dict):
            super().__init__(name, bases, attr_dict)
            cls._field_names = []  # <2>
            for key, attr in attr_dict.items():  # <3>    # in order
                if isinstance(attr, Validated):
                    type_name = type(attr).__name__
                    attr.storage_name = '_{}#{}'.format(type_name, key)
                    cls._field_names.append(key)  # <4>


    class Entity(metaclass=EntityMeta):
        """Business entity with validated fields"""

        @classmethod
        def field_names(cls):  # <5>
            for name in cls._field_names:
                yield name

#### 元类使用场景

-   验证属性
-   一次把装饰器依附到多个方法上
-   序列化对象或者转换数据
-   对象关系映射（ORM框架）
-   基于对象的持久存储
-   动态转换使用其他语言编写的类结构

#### 类作为对象

-   `cls.__bases__`: 类的基类组成的元祖
-   `cls.__qualname__`: py3 引入，值是类或函数的限定名称，即从模块的全局作用域到类的点分路径
-   `cls.__subclasses__()`： 返回一个list，包含类的直接子类。其实现使用弱引用，防止超类和子类之间出现循环引用。这个方法返回的列表是内存里现存的子类。
-   `cls.mro()`: 构建类时，如果需要获取存储在类属性 `__mro__` 中的超类元组，解释器会调用这个方法。元类可以覆盖这个方法。
