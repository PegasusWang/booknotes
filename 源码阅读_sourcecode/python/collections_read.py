# -*- coding: utf-8 -*-

"""
仿写python 内置模块 collections
简化版；功能；原理；单测
"""


class Node(object):
    __slots__ = 'prev', 'next', 'key'

    def __init__(self, prev=None, next=None, key=None):
        self.prev, self.next, self.key = prev, next, key


class LinkList(object):
    def __init__(self):
        root = Node()
        root.prev, root.next, root.key = root, root, None
        self.root = root

    def append(self, node):
        lastnode = self.root if self.root.next is self.root else self.root.next
        lastnode.next = node
        node.prev = lastnode
        node.next = self.root
        return node

    def __iter__(self):
        node = self.root
        while node.next is not self.root:
            yield node


class OrderedDict(dict):
    """
    功能：实现有序字典 1.字典的顺序是赋值的顺序
    实现原理: 用一个循环双向列表记录访问 key 的顺序
    """

    def __init__(*args, **kwargs):
        self = args[0]
        args = args[1:]
        try:
            self._root
        except AttributeError:
            self._root = []  # LinkList()
            self._map = {}
        self._update(*args, **kwargs)

    def _update(self, *args, **kwargs):
        # from collections import MutableMapping
        # update = MutableMapping.update
        for k, v in args[0]:
            self.__setitem__(k, v)

    def __setitem__(self, k, v):
        if k not in self._map:
            node = Node(key=k)
            node = self._root.append(node)
            self._map[k] = node
        dict.__setitem__(self, k, v)    # 注意这里 call dict method

    def __iter__(self):
        for node in self._root:
            yield node.key

    def keys(self):
        """keys in order"""
        return [node.key for node in self._root]

    def values(self):
        return [self[node.key] for node in self._root]


def test_OrderedDict():
    items = (
        ['a', 1],
        ['b', 2],
        ['c', 3],
        ['a', 3]
    )
    o = OrderedDict(items)
    assert o
    assert o['a'] == 3
    assert o.keys() == ['a', 'b', 'c']
    assert o.values() == [3, 2, 3]


class Counter(dict):
    """计数器
    继承了dict，value 值保存了对应的次数
    most_common: 如果是获取所有直接排序输出，否则构造 (heap) 返回
    update: 更新元素。三种参数  items [(a,b)]  ; dict;   kwargs
    """
    def __init__(*args, **kwargs):
        self = args[0]
        args = args[1:]
        super(Counter, self).__init__()
        self.update(*args, **kwargs)

    def update(*args, **kwargs):
        from collections import Mapping
        self = args[0]
        args = args[1:]
        iterable = args[0] if args else None
        if iterable is not None:
            if isinstance(args, Mapping):
                for elem, count in args.iteritems():
                    self[elem] = self.get(elem, 0) + count
            else:
                for elem in iterable:
                    self[elem] = self.get(elem, 0) + 1

        if kwargs:
            self.update(kwargs)   # kwargs 变成了dict，然后走了上边Mapping 流程

    def most_common(self, n=None):
        if n is None:
            return sorted(self.iteritems(), key=lambda kv: kv[1], reverse=True)
        else:
            import heapq
            heapq.nlargest(n, self.iteritems(), key=lambda kv: kv[1])

    def __missing__(self, key):
        return 0


def test_Counter():
    c = Counter()
    c['a'] = 1
    c['a'] += 1
    assert c['a'] == 2
    assert c['no_exist'] == 0


if __name__ == '__main__':
    test_Counter()
