笔者虽然不做运维，大部分时间关注业务代码和工程，但是了解一些运维工具对于排插问题还是有帮助的。稍微记录一下，至少碰到问题了知道有啥工具。


## 1章：系统基础信息模块详解

通过第三方模块获取服务器的基本性能、块设备、网卡接口、网络地址库等信息。

### 1.1 系统性能模块psutil:获取系统性能信息、内存信息、磁盘信息、网络信息、用户信息等。
### 1.2 IP地址处理模块IPy: 处理IP地址，网段等。
### 1.3 DNS处理模块dnspython: 实现dns服务监控以及解析结果的校验，替代nslookup及dig等工具。比如查询CNAME记录等。DNS解析可以做简单的负载均衡。

## 2章：业务服务监控详解

### 2.1 文件内容差异比对方法：difflib模块。也可以用diff工具，meld等。
### 2.2 文件与目录差异对比方法：filecmp。meld也可以进行目录比较。
### 2.3 邮件发送模块：smtplib。笔者经常使用flask_mail插件和pandas给运营人员生成报表，还是相当方便的。
### 2.4 探测web服务质量: pycurl。我觉得可能用http，curl，requests等更方便。

## 3章：定制业务质量报表详解

### 3.1 数据报表之Excel操作模块；XlsxWriter，笔者更常使用pandas，处理报表excel等比较方便，pandas.DataFrame提供了很多功能。
### 3.2 Python与rrdtool的结合模块：python-rrdtool，rrdtool（round brobin database）工具为环状数据库的存储格式，round robin是一种处理定量数据及当前元素指针的技术。 比如实现网卡流量图绘制，很多监控工具都用到了该工具。
### 3.3 生成动态路由轨迹: scapy 强大的交互式数据包处理程序，能对数据包进行伪造或者解包，包括发送数据包、包嗅探、应答和反馈匹配等功能。比如使用traceroute函数实现生成路由轨迹图。

## 4章：Python与系统安全
#### 4.1 构建集中式的病毒扫描机制：pyClamad，让python直接使用ClamAV病毒扫描守护进程clamd。
#### 4.2 实现高效的端口扫描器：高危端口暴露在网站有被入侵风险。使用python-nmap实现高效的端口扫描。


## 5章：系统批量运维管理器pexpect
pexpect可以理解成Linux下的expect的python封装，通过pexpect可以实现对ssh、ftp、passwd、telnet等命令的自动交互，而无需人工干涉达到自动化的目的。
核心组建包括spawn类、run函数以及派生类pxssh等。
- spawn: 启动和控制子应用程序
- run：调用外部命令的函数，可以同时获得命令的输出结果及命令的退出状态。
- pxssh类：操作ssh

## 6章：系统批量运维管理器paramiko
paramiko是基于python实现的ssh2远程安全连接，支持认证及密钥方式，可以实现远程命令执行，文件传输、中金ssh代理等功能，相对于pexpect封装的层次更高，更贴近ssh协议的功能。paramiko包含俩核心组件，SSHClient类和SFTPClient类。
- SSHClient类：ssh服务会话的高级表示，封装了传输(transport)、通道(channel)及SFTPClient的校验、建立的方法，通常用于执行远程命令。
- SFTPClient类：SFTP客户端对象，根据ssh传输协议的sftp会话，实现远程文件操作，比如文件上传、下载、权限、状态等操作。


## 7章：系统批量运维管理器Fabric详解
Fabric基于python实现的ssh命令行工具，简化了ssh的应用程序部署及系统管理任务，提供了系统基础的操作组件，可以实现本地或远程shell命令，包括命令执行、文件上传、下载及完整执行日志输出等功能。详细使用还是看官方文档，典型使用场景有文件上传与校验、环境部署、代码发布。


## 9章：集中化管理平台Ansible详解
Ansible一种集成IT系统的配置管理、应用部署、执行特定任务的开源平台。Ansible提供了一个在线Playbook分享平台，汇聚了各类常用功能的角色。Ansible配置文件以YAML格式存在。

## 10章：集中化管理平台Saltstack详解
Saltstack是一个服务器基础架构集中化管理平台，具备配置管理、远程执行、监控等功能，可以理解成简化版的puppet。


## 11章：统一网络控制器Func详解
Fedora Unified Network Controller:Fedora平台统一构建的网络控制器。

## 12章：Python大数据应用详解
