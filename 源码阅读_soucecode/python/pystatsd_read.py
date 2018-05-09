# -*- coding: utf-8 -*-


"""
各种web应用都需要打点来监控各种指标，增加报警等。

pystatsd 源码阅读, statsd 的 python 客户端
https://github.com/jsocol/pystatsd
https://statsd.readthedocs.io/

预备基础：
- socket 编程，tcp 和 udp 发送数据
- statsd 协议：非常简单的协议, https://github.com/b/statsd_spec
"""
