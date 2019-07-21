

# 1 An Introduction to Concurrency


critical section（临界区）：for a section of your program needs exclusive access to a shared resource.


### Deadlocks, Livelocks, and Starvation

Deadlocks conditions:

- Mutual Exclusion
- Wait For Condition
- No Preemption
- Circular Wait

Livelocks: 活锁。想象两个人迎面走来，一个人向一遍转，然后另一个人也同方向转，一直僵持谁都过不去。

Starvation: 饥饿 ，一个并发的进程无法获取所有需要工作的资源。一个贪心的进程阻止其他进程获取执行资源。

### Determining Concurrency Safety

- who is responsible for the Concurrency?
- how is the problem space mapped onto concurrency primitives?
- who is responsible for the synchronization?

```go
func CalculatePi(begin,end int64, pi *Pi)
func CalculatePi(begin,end int64) []int64
func CalculatePi(bengin,end int64) <-chan uint
```

### Simplicity in the Face of Complexity


# 2 Modeling Your Code: Communicating Sequential Processes

# Concurrency vs Parallelism

Concurrency is a property of the code; parallelism is a property of the runnning programm.

# What Is CAP?
