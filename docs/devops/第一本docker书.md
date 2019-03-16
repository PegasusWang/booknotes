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
