# 2. 简单动态字符串

SDS: simple dynamic string，用作 redis 默认字符串表示，还被用作buffer，AOF 缓冲区和客户端状态中的输入缓冲区。
sds能减少缓冲区溢出漏洞。

SDS是二进制安全的，不仅可以保存文本，还可以是任意二进制数据

```c
struct sdshdr {
  int len; //保存字符串的长度，获取长度变成了 O(1)
  int free; //未使用的字节数量，实现了空间预分配和惰性空间释放两种优化策略
  char buf[]; //保存字符串
}
```

# 3. 链表

redis 实现的实际上是 无环的双端链表

```c
// adlist.h/listNode
typedef struct listNode {
  struct listNode *prev;
  struct listNode *next;
  void *value;
}listNode;

//adlist.h/list
typedef stuct list {
  listNode *head;
  listNode *tail;
  unsigned long len;
  void *(*dup)(void *ptr);
  void (*free)(void *ptr);
  int (*match)(void *ptr, void *key);
}list;
```
