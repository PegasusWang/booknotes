# 3. 启动 docker


docker run --name daemon_dave -d ubuntu /bin/sh -c "while true; do echo hello world; sleep 1; done"


# 4. 使用 docker 镜像和仓库


构建 docker image。docker build 方式。

    # 当前的文件夹叫做 build context，docker 在构建镜像时将构建上下文和上下文中的文件和目录上传到 Docker 守护进程
    mkdir static_web
    cd static_web
    touch Dockerfile


    # Veersion: 0.0.1 
    FROM ubuntu:14.04
    MAINTAINER WNN 'pegasuswang@qq.com'
    RUN apt-get update && apt-get install -y nginx 
    RUN echo 'Hi, I am your container' \
    > /usr/share/nginx/html/index.html
    EXPOSE 80

每条指令都会创建一个新的镜像层并对镜像进行提交，中间如果某条指令失败了，也能得到一个镜像，可以用来调试。

    # cd static_web
    docker build -t="wnn/static_web" .
    docker build --no-cache -t="wnn/static_web" .
    docker run -d -p 80 --name static_web wnn/static_web nginx -g "daemon off;"
    docker run -d -p 8000:80 --name static_web wnn/static_web nginx -g "daemon off;"


常用 dockerfile 命令：

EXPOSE, ENTRYPOINT, ADD, COPY, VOLUME, WORKDIR, USER, ONBUILD, LABEL, STOPSIGNAL, ARG
ENV

CMD: 指定容器启动时要运行的命令，RUN 是指定构建时要运行的命令。CMD ["bin/true"]
命令行中指定的命令会覆盖掉 Dockerfile 中的 CMD。 只能制定一个，多个只会执行最后一个。


ENTRYPOINT: 类似CMD，但是不容易 在容器启动时被覆盖。实际上 docker run命令行中指定的任何参数都会被当做参数
再次传递给ENTRYPOINT指令中指定的命令。 ENTRYPOINT ["/usr/sbin/nginx"]


WORKDIR: 从镜像创建新 容器时，在容器内部设置一个工作目录，ENTRYPOINT/CMD 指定的程序会在这个目录下运行。WORKDIR
/opt/webapp

ENV: 在镜像构建过程中 设置环境变量。 ENV RVM_PATH /home/rvm/ ,之后的 RUN 命令都可以使用这个环境变量。


USER: 指定该镜像会以什么样的用户运行。  默认 root
  - USER user
  - USER user:group
  - USER uid
  - USER uid:gid


VOLUME: 基于镜像创建的容器添加卷。一个卷可以存在于一个或者多个容器内特定的目录，这个目录可以绕过联合文件系统，
提供如下共享数据或者对数据进行持久化的功能。允许多个容器共享这些内容，及时删除了最后一个使用卷的容器，卷中数据也会持久保存。
VOLUME ["opt/project"]  这条命令将会基于此镜像创建的任何容器创建一个名为 /opt/project 的挂载点。


ADD: 用来将构建环境下的文件和目录复制到镜像中。会自动解开归档文件(gzip,bzip2,xz)
ADD会是的构建缓存变得无效，  使得后续指令无法继续使用之前的构建缓存。


COPY: 类似 COPY，只关心在构建上线文中复制本地文件，不会做文件提取和解压工作。 COPY conf.d /etc/apache2/


LABEL: 为 docker 镜像添加元数据，键值对形式。LABEL version="1.0"  通过docker inspect 查看


STOPSIGNAL: 停止容器的时候发送什么信号给容器，必须是内核中合法信号名称或者数字


ARG:  docker build 命令运行时传递给构建运行时的变量。通过 docker build --build-arg build=1234 这种形式传递进去
有一些预定义变量等。
  - HTTP_PROXY
  - NO_PROXy


ONBUILD: 为镜像添加触发器，当一个镜像被用作其他镜像的基础镜像，该镜像中的触发器将会被执行。


从容器运行registry: docker run -p 5000:5000 registry:2


# 5 在测试中使用 docker


重点是 docker 网络连接的几种方式: Docker Networking, Docker链接


# 6 使用 docker 构建服务


