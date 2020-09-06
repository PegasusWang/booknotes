
# 获取服务器 cpu 使用情况代码

```py
# 返回服务器信息(udp) server
import socket
import psutil


def do_cpu():
    data = str(psutil.cpu_percent(0)) + '% \n'
    count = 0
    for process in psutil.process_iter():
        data = data + process.name()
        data = data + ',' + str(process.pid)
        cpu_usage_rate_process = str(process.cpu_percent(0)) + '% '
        data = data + ',' + cpu_usage_rate_process + '\n'
        count += 1
        if count == 10:
            break
    return data


s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
s.bind(('127.0.0.1', 8090))
print("bind udp on 8090...")

while True:
    info, addr = s.recvfrom(1024)
    data = do_cpu()
    s.sendto(data.encode('utf-8'), addr)
    print("The client is", addr)
    print("Sended CPU data is:", data)
```

```py
# client 获取 cpu 占用信息
import socket

s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
s_addr = ('127.0.0.1', 8090)
s.sendto(b'CPU info', s_addr)
data_b, addr = s.recvfrom(1024)

if addr == s_addr:
    data_s = data_b.decode('utf-8')
    data_list = data_s.split('\n')
    print("Cpu usage rate is ", data_list[0])
    print('% - 20s % -5s % -10s' % ('name', 'pid', 'cpu usage'))
    data_list = data_list[1:-1]
    for xx in data_list:
        yy = xx.split(',')
        print('% - 20s % -5s % -10s' % (yy[0], yy[1], yy[2]))

s.close()
```
