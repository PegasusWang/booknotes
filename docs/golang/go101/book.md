
# 21. Channels - The Go Way to do concurrency synchronizations

Don't (let computations) communicate by sharing memroy, (let them) share memory by communicating (through channels).

五种操作(all these operations are already synchronized)：

- close(ch), ch must not be a receive-only channel.
- `ch <- v`, send a value
- `<-ch`, receive a value from the channel
- cap(ch), value buffer capacity, return int
- len(ch), query current number of values in the value buffer

We can think of each channel as maintaing 3 queues:
- the receiving goroutine queue
- the sending goroutine queue
- the value buffer queue

![](./channel.png)
