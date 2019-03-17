# 1. 跑起来


pod是容器集合， 共享 ip 地址和Port空间，在一个network和namespace中。
是 k8s 调度最小单位。


    kubectl get pos


# 2. 概念

1. Cluster
Cluster是计算、存储和网络资源的集合，Kubernetes利用这些资源运行各种基于容器的应用。

2. Master
Master是Cluster的大脑，它的主要职责是调度，即决定将应用放在哪里运行。Master运行Linux操作系统，可以是物理机或者虚拟机。为了实现高可用，可以运行多个Master。

3. Node
Node的职责是运行容器应用。Node由Master管理，Node负责监控并汇报容器的状态，同时根据Master的要求管理容器的生命周期。Node运行在Linux操作系统上，可以是物理机或者是虚拟机。

4. Pod
Pod是Kubernetes的最小工作单元。每个Pod包含一个或多个容器。Pod中的容器会作为一个整体被Master调度到一个Node上运行。

5. Controller
Kubernetes通常不会直接创建Pod，而是通过Controller来管理Pod的。Controller中定义了Pod的部署特性，比如有几个副本、在什么样的Node上运行等。为了满足不同的业务场景，Kubernetes提供了多种Controller，包括Deployment、ReplicaSet、DaemonSet、StatefuleSet、Job等

6. Service
答案是Service。
Kubernetes Service定义了外界访问一组特定Pod的方式。Service有自己的IP和端口，Service为Pod提供了负载均衡。
Kubernetes运行容器（Pod）与访问容器（Pod）这两项任务分别由Controller和Service执行。

7. Namespace
如果有多个用户或项目组使用同一个Kubernetes Cluster，如何将他们创建的Controller、Pod等资源分开呢？
答案就是Namespace。

# 3. 部署

kubelet运行在Cluster所有节点上，负责启动Pod和容器。
kubeadm用于初始化Cluster。
kubectl是Kubernetes命令行工具。通过kubectl可以部署和管理应用，查看各种资源，创建、删除和更新各种组件。


# 4. k8s架构

Master是Kubernetes Cluster的大脑，运行着的Daemon服务包括kube-apiserver、kube-scheduler、kube-controller-manager、etcd和Pod网络（例如flannel），如图4-1所示。

Node是Pod运行的地方，Kubernetes支持Docker、rkt等容器Runtime。Node上运行的Kubernetes组件有kubelet、kube-proxy和Pod网络（例如flannel）


# 5. 运行应用




# 6. 通过 service 创建 pod

service ip 不变
Cluster IP是一个虚拟IP，是由Kubernetes节点上的iptables规则管理的。

iptables 将访问到 service 的流量转发到后端 Pod,采用了类似负载轮询的策略


# 7. Rolling Update

滚动更新。采用渐进的方式逐步替换旧的 pod，也可以回滚操作恢复到更新前状态。
一般配合健康检查完成。

maxSurge/maxUnavailable 参数配置


# 8. Health Check

默认通过容器启动进程是否返回0判断是否探测成功。

Liveness探测：用户可以自定义判断容器是否健康的条件，探测失败 k8s会重启容器。

Readiness探测：什么时候可以将容器加入到Service负载均衡池中，对外提供服务。相当于将容器设置为不可用
扩容的时候，可以通过Readiness探测判断容器是否就绪（比如加载缓存、连接数据库），
避免将请求发送到还没有准备好的backend。


# 9. 数据管理

Volume: 容器销毁时，内部文件系统中的数据会清除，为了持久化可以使用Volumns。

emptyDir: 对于容器来说持久，于  pod 不是，生命周期和pod一致。

hostPath: 将docker host 文件系统中已经存在的目录 mount 给Pod的容器。pod 被销毁也会保留。
不过一旦 Host 崩溃，hostPath 无法访问。

外部storage provider: ESB volumne, Ceph, GlusterFS等主流分布式存储系统


# 10. Secret & Configmap

启动过程中的敏感数据如果直接保存在镜像中不妥，k8s 提供了Secret来解决。
Secret会以密文的方式存储数据，避免了直接在配置文件中保存敏感信息。Secret会以Volume的形式被mount到Pod，容器可通过文件的方式使用Secret中的敏感数据；此外，容器也可以环境变量的方式使用这些数据。

4种方式可以创建secret，常见的使用yaml文件。

Pod可以通过 Volume 或者环境变量的方式使用Secret。

Secret可以提供密码、私钥、token 等敏感信息。非敏感配置可以使用 ConfigMap(明文存放)


# 11. Helm-Kubernetes 的包管理器

k8s 能够很好地组织和编排容器，但是缺少一个更高层次的应用打包工具(类似apt)。

Helm有俩重要概念：
  - chart: 创建一个应用的信息集合，包括各种k8s对象的配置模板、参数定义、依赖关系、文档说明等。
    只包含逻辑单元，可以想象成apt,yum中的软件安装包。
  - release: 是 chart 的运行实例，代表一个正在运行的应用。chart每次安装都是一个 release

Helm 包含两个组件：Helm 客户端和 Tiller服务器。

Helm Client ----->(grpc)  ---------> Tiller ---------> Kube API

Helm 客户端负责管理 chart，Tiller 服务器负责管理 release。


# 12 网络

k8s采用基于扁平地址空间的网络模型，集群中每个 pod 有自己的 ip 地址，不要配置 NAT 就能直接通信。
k8s 采用  Container Networking INterface(CNI)规范


# 13 Dashboard

k8s 同时提供了 dashboard web 界面工具操作


# 14 k8s 集群监控

weave scope  是 docker 和k8s 可视化监控工具。

Heapster: k8s 原生的集群监控方案。

Prometheus Operator: 监控集群本身的运行状态


# 15 k8s 集群日志管理

k8s开发了一个 Elasticsearch  附加组件实现集群日志管理。

k8s -> Fluentd(搜集日志并发送给Elasticsearch) ->Elasticsearch(搜索引擎，存储日志并提供查询接口) -> Kibana(Web
GUI浏览和搜索 es 的日志)
