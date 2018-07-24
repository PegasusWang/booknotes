> Optimization hinders evolution. - Alan Perlis

最近扫了一本《High Performance Python》，该书从多个角度比如性能度量(profile)，内置数据结构调优，并发，多进程，多线程，使用C加速，异步，分布式，内存调优，其他网站使用经验等介绍了python优化经验，看得出几位作者的计算机体系结构和python功底很深。笔者非计算机科班出身，有些知识点超出了我的能力范围，不过多了解一些东西对于以后技术调研和选型还是很有帮助的，简要记录下涉及到的一些东西。

<!--more-->

___
## 1章： Understanding Performant Python

第一张介绍了计算机体系结构，从计算单元，内存单元，通信层次三个角度（我看起来稍微吃力，还是英文版）。最后介绍了下为什么我们选择python？虽然Cpython实现的解释器执行慢，但是python较低的学习成本，高生产力和丰富的生态还是能够解决大部分常见的问题。

___
## 2章： Profiling to Find Bottlenecks
本章介绍了很多性能度量工具，性能优化的第一步就是度量了，很多书上都说不要过早优化，而且优化之前不要仅仅靠猜测，使用profile工具我们可以对程序性能进行分析，找出瓶颈以后再有针对性地进行优化。较为常见的有以下几个

- time.time: 最直接的就是用time函数计时。
- cProfile: 内置性能分析模块。
- line_profiler: 以行为单位度量代码性能问题。
- memory_profiler: 分析内存使用。
- heapy: 以对象Objects为单位分析
- dowser: 生成变量的分析图
- dis: 分析字节码。
- unittest: 性能优化的前提是不改变代码行为，需要有足够的单元测试保证优化以后的代码行为正确。


___
## 3章：List and Tuples
最简单和容易实现的优化方式就是选用合适的数据结构和算法了，了解python内置的数据类型和实现方式能让我们避免一些常见的坑。

- list: 可变数组，O(1)访问.
- tuple: 不可变。O(1)访问，相比list更节省内存。caching resource: python对于小于20个元素的tuple会缓存下来， 不会立即进行内存回收。对于固定长度的数组元素，应该优先考虑使用tuple。

这两种内置数据结构非常常用，需要注意的是list的内存分配。当append操作超过原始分配的内存大小时，会重新开辟内存，把之前的数据复制过去，所以多次调用append有可能会有性能和内存浪费问题。

元素查找有线性查找(list.index)(On)和有序数组的二分查找(bisect模块)(Ologn)


___
## 4章:Dict and Sets
作为dict或者set的键要求是可hash的，可以hash指的是实现了 __hash__和__eq__(或__cmp__)，list这种可变类型的数据结构就无法作为key，dict和set的内部实现都是哈希表。

对于自定义类型，我们需要实现__has__和__eq__来获得期望的结果:

    class Point:
        def __init__(self, x, y):
            self.x, self.y = x, y

        def __hash__(self):
            """hash函数尽量保证较少的键冲突"""
            return hash((self.x, self.y))

        def __eq__(self, other):
            return self.x == other.x and self.y == other.y


- dict: 键值对.插入和查找都是O(1)，但是比较浪费内存。
- set: 类似数学上的集合。插入和查找O(1)，但是当超过原先分配内存的2/3时，为了维持插入和查找效率，重新分配内存。对于固定长度的集合，最好使用frozenset节省内存。

___
## 5章：Iterators and Generators

使用生成器可以帮助我们节省内存使用。 在python2中，有range和xrange两个函数，区别在于range函数返回list，xrange返回迭代器，非常节省内存。python3的一大改进就是将很多内置返回整个序列的函数改成了返回迭代器。

    def range(start, stop, step=1):
        nums = []
        while start < stop:
            nums.append(start)
            start += step
        return nums


    def xrange(start, stop, step=1):
        while start < stop:
            yield start
            start += step


for 语句的原理:

    object_iterator = iter(object)
    while True:
        try:
            i = object_iterator.next()
            do_work(i)
        except StopIteration:
            break


使用itertools模块，内置的itertools模块提供了很多方便的函数处理生成器序列：

    itertools.islice: 对生成器切片操作
    itertools.chain: 连接多个生成器
    itertools.takewhile: 给生成器加上终结条件
    itertools.cycle: 生成循环的的序列

___
## 6章：Matrix and Vector Computation

这一章举得例子有点高深，涉及到比较多计算机体系的问题，以及perf工具分析程序的执行。简单来说就是涉及到向量或者矩阵运算的时候，还是使用numbpy,pandas等比较成熟优化过的库。笔者公司报表的处理就是使用pandas来处理的，pandas提供了很多方便的结构来操作Excel。


___
## 7章：Compling to C
本章介绍了很多使用C语言来加速python的方案，适合cpu密集的代码。不过笔者的职业生涯还没见过使用c来加速代码的项目，因为做的是web应用，一般性能瓶颈出现在IO或者数据库访问上，用c语言的话维护成本也会大大增加。

1. Profile to understand program's behavior
2. Improve algorithm based on evidence
3. Use a compiler to achieve quick wins
4. Beware diminishing returns with extedned effort

对于一些CPU密集的代码，我们可以用Cython将python代码转成c代码提高执行效率或者使用PyPy来执行。书中还提到了其他几个小众的玩意Shed Skin, Numba，Pythran，不过似乎生产环境下不怎么用，甚至Cython和PyPy生成环境下使用的也很少，之前clone过reddit的代码，发现他们在一些计算密集的小函数上使用了Cython。

___
## 8章：Concurrency
本章介绍了提升程序并发性能的几种方案，包括使用gevent，tornado，asyncio等，不过目前貌似异步框架在web应用中使用的比例仍然不高。python3引入的asyncio或许是python异步的未来，已经有基于asyncio和aiohttp的异步web框架[sanic](https://github.com/channelcat/sanic/)，性能接近Golang，非常强悍， 不过目前生态圈和案例太少，没有生产环境下的经验能借鉴。异步数据库orm我搜了一下貌似只有一个async-peewee，貌似也没多少人用，感觉python异步web框架还有很多路要走。Instagram，Reddit，Youtube等没有使用异步框架也能撑起巨大的访问，可能有时候可扩展的架构更加重要。

异步编程中的一些概念:

Event Loop: 实质上就是一系列需要被执行的函数列表

    from Queue import Queue

    eventloop = None
    class EventLoop(Queue):
        def start(self):
            while True:
                function = self.get()
                function()

    def do_hello():
        global eventloop
        print('Hello')
        eventloop.put(do_world)

    def do_world():
        global eventloop
        print('world')
        eventloop.put(do_hello)

    if __name__ == "__main__":
        eventloop = EventLoop()
        eventloop.put(do_hello)
        eventloop.start()


我们可以将事件循环和异步IO操作结合，这些操作是非阻塞的，意味着如果我们在一个异步函数上做IO写操作，它会立即返回，即使写操作还没完成。当写操作完成时，事件通知程序写操作完成。这允许我们可以在IO等待时做其他操作。
一般实现事件循环有两种形式，callbacks或futures.


    # callbacks实现
    from functools import partial
    def save_value(value, callback):
        print('Saving {} to datavase'.format(value))
        # save_result_to_db是一个异步函数，立即返回
        save_result_to_db(result, callback)


    def print_response(db_response):
        print('Response from database: {}'.format(db_response))


    if __name__ == '__main__':
        eventloop.put(
            partial(save_value, "Hello world", print_response)
        )


对于future实现，一个异步函数返回一个future结果的promise而非真正的结果。我们必须等待future完成填充我们需要的结果，期间可以进行其他计算。。


    @coroutine
    def save_value(value, callback):
        print('Saving {} to database'.format(value))
        db_response = yield save_result_to_db(result, callback)
        print('Response from database: {}'.format(db_response))


这种实现下save_result_to_db返回future对象，通过yield我们可以暂停它执行直到返回值准备好了恢复它的执行。python里协程使用生成器实现的，因为生成器是python内置的并且支持暂停和恢复操作。我们需要做的就是yield一个future对象，然后事件循环等待它计算完成。一旦完成，事件循环将恢复函数的执行，发送结果到future对象。
在python2里边基于future实现的协程有点不优雅，因为我们试图将协程作为真正的函数使用，但是实现协程的生成器是无法返回值的，例如在tornado里python2要 ` raise gen.Return(json_decode(response.body))` 来返回值。从python3.4以后允许协程直接返回值。

通过爬虫的例子来演示gevent，tornado，asyncio实现异步操作。一个线性运行的爬虫:

    import requests
    import string
    import random


    def generate_urls(base_url, num_urls):
        """
        We add random characters to the end of the URL to break any caching
        mechanisms in the requests library or the server
        """
        for i in xrange(num_urls):
            yield base_url + "".join(random.sample(string.ascii_lowercase, 10))


    def run_experiment(base_url, num_iter=500):
        response_size = 0
        for url in generate_urls(base_url, num_iter):
            response = requests.get(url)
            response_size += len(response.text)
        return response_size

    if __name__ == "__main__":
        import time
        delay = 100
        num_iter = 50
        base_url = "http://127.0.0.1:8080/add?name=serial&delay={}&".format(delay)

        start = time.time()
        result = run_experiment(base_url, num_iter)
        end = time.time()
        print("Result: {}, Time: {}".format(result, end - start))


gevent: monkey-patch标准IO函数使之变成异步的。，它有一个被称为『Greenlet』的对象用来执行并发操作。gevent使用事件循环在IO等待期间对greenlet进行调度，直到所有greenlet执行完成。

    from gevent import monkey
    monkey.patch_socket()

    import gevent
    from gevent.coros import Semaphore
    import urllib2
    from contextlib import closing
    import string
    import random


    def generate_urls(base_url, num_urls):
        for i in xrange(num_urls):
            yield base_url + "".join(random.sample(string.ascii_lowercase, 10))


    def download(url, semaphore):
        with semaphore, closing(urllib2.urlopen(url)) as data:
            return data.read()


    def chunked_requests(urls, chunk_size=100):
        semaphore = Semaphore(chunk_size)    # 用来控制并发数
        requests = [gevent.spawn(download, u, semaphore) for u in urls]
        for response in gevent.iwait(requests):    # 事件循环只在iwait调用时执行
            yield response


    def run_experiment(base_url, num_iter=500):
        urls = generate_urls(base_url, num_iter)
        response_futures = chunked_requests(urls, 100)
        response_size = sum(len(r.value) for r in response_futures)
        return response_size

    if __name__ == "__main__":
        import time
        delay = 100
        num_iter = 500
        base_url = "http://127.0.0.1:8080/add?name=gevent&delay={}&".format(delay)

        start = time.time()
        result = run_experiment(base_url, num_iter)
        end = time.time()
        print("Result: {}, Time: {}".format(result, end - start))

下边是使用tornado的异步HTTPClient爬虫示例:

    from tornado import ioloop
    from tornado.httpclient import AsyncHTTPClient
    from tornado import gen

    from functools import partial
    import string
    import random

    AsyncHTTPClient.configure(
        "tornado.curl_httpclient.CurlAsyncHTTPClient", max_clients=100)


    def generate_urls(base_url, num_urls):
        for i in xrange(num_urls):
            yield base_url + "".join(random.sample(string.ascii_lowercase, 10))


    @gen.coroutine
    def run_experiment(base_url, num_iter=500):
        http_client = AsyncHTTPClient()
        urls = generate_urls(base_url, num_iter)G
        responses = yield [http_client.fetch(url) for url in urls]
        response_sum = sum(len(r.body) for r in responses)
        raise gen.Return(value=response_sum)

    if __name__ == "__main__":
        import time
        delay = 100
        num_iter = 500
        _ioloop = ioloop.IOLoop.instance()
        run_func = partial(
            run_experiment,
            "http://127.0.0.1:8080/add?name=tornado&delay={}&".format(delay),
            num_iter)

        start = time.time()
        result = _ioloop.run_sync(run_func)
        end = time.time()
        print result, (end - start)

最后一个是使用asyncio的示例，不过asyncio提供的api比较偏低层，我们使用aiphttp来发送http请求:

    #!/usr/bin/env python3

    import asyncio
    import aiohttp
    import random
    import string


    def generate_urls(base_url, num_urls):
        for i in range(num_urls):
            yield base_url + "".join(random.sample(string.ascii_lowercase, 10))


    def chunked_http_client(num_chunks):
        semaphore = asyncio.Semaphore(num_chunks)

        @asyncio.coroutine
        def http_get(url):
            nonlocal semaphore
            with (yield from semaphore):
                response = yield from aiohttp.request('GET', url)
                body = yield from response.content.read()
                yield from response.wait_for_close()
            return body
        return http_get


    def run_experiment(base_url, num_iter=500):
        urls = generate_urls(base_url, num_iter)
        http_client = chunked_http_client(100)
        tasks = [http_client(url) for url in urls]
        responses_sum = 0
        for future in asyncio.as_completed(tasks):
            data = yield from future
            responses_sum += len(data)
        return responses_sum

    if __name__ == "__main__":
        import time
        loop = asyncio.get_event_loop()
        delay = 100
        num_iter = 500

        start = time.time()
        result = loop.run_until_complete(
            run_experiment(
                "http://127.0.0.1:8080/add?name=asyncio&delay={}&".format(delay),
                num_iter))
        end = time.time()
        print("{} {}".format(result, end - start))


___
## 9章：The Multiprocessing Module
本章主要介绍多进程模块multiprocessing，它提供了以下几个主要的部分：

- Process: 当前进程的复刻(fork)
- Pool: 封装了Process或者threading.Thread API到一个方便的工作池（进程池）
- Queue: 支持多个生产者和消费者的队列
- Pipe: 在两个进程之间单向或者双向通信的管道
- Manager: 进程之间共享python对象的高级接口
- ctypes: 允许在进程被forked后共享原始类型(integers, floats and bytes等)
- Synchronization primitives: 同步原语,Locks或者semaphores

python3.2以后引入了concurrent.futures模块，提供了更加简洁的api来操作进程或者线程，但是不如multiprocessing灵活一些。

    # Monte Carlo方法来模拟pi的计算
    import math
    import random
    import time


    def y_is_in_circle(x, y):
        """Test if x,y coordinate lives within the radius of the unit circle"""
        circle_edge_y = math.sin(math.acos(x))
        return y <= circle_edge_y


    def estimate_nbr_points_in_circle(nbr_samples):
        nbr_in_circle = 0
        for n in xrange(nbr_samples):
            x = random.uniform(0.0, 1.0)
            y = random.uniform(0.0, 1.0)
            if y_is_in_circle(x, y):
                nbr_in_circle += 1
        return nbr_in_circle

    nbr_samples = int(1e7)
    t1 = time.time()
    nbr_in_circle = estimate_nbr_points_in_circle(nbr_samples)
    print "Took {}s".format(time.time() - t1)
    pi_estimate = float(nbr_in_circle) / nbr_samples * 4
    print "Estimated pi", pi_estimate
    print "Pi", math.pi


    import math
    import random
    import time
    #from multiprocessing.dummy import Pool
    from multiprocessing import Pool


    def y_is_in_circle(x, y):
        """Test if x,y coordinate lives within the radius of the unit circle"""
        circle_edge_y = math.sin(math.acos(x))
        return y <= circle_edge_y


    def estimate_nbr_points_in_circle(nbr_samples):
        nbr_in_circle = 0
        for n in xrange(nbr_samples):
            x = random.uniform(0.0, 1.0)
            y = random.uniform(0.0, 1.0)
            if y_is_in_circle(x, y):
                nbr_in_circle += 1
        return nbr_in_circle


    pool = Pool()


    nbr_samples = int(1e7)
    nbr_parallel_blocks = 4
    map_inputs = [nbr_samples] * nbr_parallel_blocks
    t1 = time.time()
    results = pool.map(estimate_nbr_points_in_circle, map_inputs)
    # pool.close()
    print results
    print "Took {}s".format(time.time() - t1)
    nbr_in_circle = sum(results)
    combined_nbr_samples = sum(map_inputs)

    pi_estimate = float(nbr_in_circle) / combined_nbr_samples * 4
    print "Estimated pi", pi_estimate
    print "Pi", math.pi


___
## 10章：Clusters and Job Queues

本章介绍了python实现集群的几种方案，笔者并没有使用经验，不过等业务场景需要的时候可以再去深入调研，现在就先了解一下。上一章讨论了多进程只能利用一台计算机的资源，如果有多台计算机就需要用到集群方案，下边是本章介绍的三个方案：

- Parallel Python Module: for simple local clusters
- IPython Parallel: support research
- NSQ: for robust production clustering

其他集群工具:

- Celery: 一个广泛使用的异步消息队列。
- Gearman： 支持多个平台的任务处理系统
- PyRes：python和redis实现的轻量级任务管理器
- Amazon's Simple Queue Service(SQS): 亚马逊云服务提供的任务处理系统

___
## 11章：Using Less RAM
本章介绍的一些技术帮助我们减少内存使用：

- 分块加载，如使用 memory-mapped file
- 使用array模块相比list存储原生类型（integers, floats, characters等，not complex numbers or classses）能节省很多内存使用，或者使用numpy的array
- 使用memit(ipython)，memory_profiler衡量内存使用, 内置的sys.getsizeof对于容器类型的计算不准确
- python2里的unicode对象比较耗费内存，在python3中没有这个问题
- 如果需要存储大量字符串到内存：使用DAWG(Directed asyclic word graph)和trie(Marisa trie,Datrie,HAT trie)等结构比存储到list或者set里节省大量内存，DAWG等结构通过公用字符串前缀或后缀节省存储，github上有相应实现。
- 嵌入式系统可以用"Micro Python", http://micropython.org

Probabilistic Data Structures:如果允许一定概率的精度损失，我们可以使用一些概率数据结构

- Morris Counter: 仅使用一个字节完成近似计数
- K-minimum Values: 只用很少内存完成集合操作
- Bloom Filters: 固定内存完成判重操作，比如写爬虫的时候有海量url需要判重，就可以使用Bloom Filters节省内存
- LogLog Counter: 限定内存计数

___
## 12章：Lessons from the Field
本章汇集了一些公司使用python的成功案例和经验，其实看看Instagram的技术博客有不少干货

    - lyst.com(fashion recommendation engine):django/Amazon EC2/Redis/PyRes/supervisord/Elasticsearch/Graphite/Sentry/Jenkins/Cython/numpy/scipy/

    - sme.sh(social media analytics):django/flask/pyramid/celery/Boto/PyMongo/MongoEngine/redis-py/Pyscopy/Graphite/Sentry/docker/

    - adaptivelab.com(social media analytics): Elasticsearch/Celery/redis/Mozilla's Circus(替代crontab)/SaltStack/Fabric/Vagrant/敏捷/Amazon's Elastic Compute Cloud(EC2)/

    - radimrehurek.com(Deep Learning): Stream your data, watch your memory./ Take advantage of Python's rich ecosystem. / Profile and compile hotspots./ Parallelization and multiple cores. /Static memory allocations. / Problem-specific optimizations.

