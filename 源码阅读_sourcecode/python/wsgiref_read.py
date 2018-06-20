# -*- coding: utf-8 -*-

"""
阅读内置的 wsgi server 的实现，了解 http web server 的实现原理，并自己仿照实现简单的 wsgi server


知识点：
- python3
- tcp socket 编程
- http 协议
- WSGI 规范


"""


def application(environ, start_response):
    """一个简单的 wsgi 规范的 app 函数，编写 wsgi server 跑起来它"""
    status = '200 OK'
    headers = [('Content-Type', 'text/html; charset=utf8')]

    start_response(status, headers)
    return [b"<h1>Hello</h1>"]


"""
先编写 BaseServer，之后我们会编写 TCPServer 继承它
"""
