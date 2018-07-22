# 6章 I/O 复用：select 和 poll 函数

unix 可用的五种 I/O 模型：
- 阻塞式 I/O
- 非阻塞式 I/O, nonblocking
- I/O 复用(select poll), multiplexing
- 信号驱动式 I/O (SIGIO), signal-driven，内核告诉我们何时可以启动一个 I/O 操作
- 异步 I/O (POSIX 的 aio_系列函数), asynchronous I/O，内核告诉我们 I/O 操作何时完成

一个输入操作通常包括两个不同阶段：
1. 等待数据准备好
2. 从内核向进程复制数据


POSIX 定义术语：
- 同步 I/O 操作(synchronous I/O opetation) 导致请求进程阻塞，直到 I/O 操作完成。
- 异步 I/O 操作(asynchronous I/O opetation) 不导致请求进程阻塞。前面五种只有异步 I/O 与 POSIX  定义的异步 I/O 匹配。

select 函数：允许进程指示内核等待多个事件中的任何一个发生，并只在有一个或多个事件发生或经历一段指定的时间后才唤醒它。
