#### simple-docker 使用说明

本人使用的开发环境为：

- WSL2 (Ubuntu20.04) with aufs
- go 1.18

建议在 Linux 环境下使用。

需要在 root 目录下准备一个 `busybox.tar` 的压缩包，可以从 docker 中获取并导出：

~~~~sh
docker pull busybox
docker run -d busybox top -b
docker export -o busybox.tar <container_id>
~~~~

下面的命令演示都以 busybox 这个镜像为基础。

创建一个简单的容器：

~~~~sh
go run . run -it busybox sh
~~~~

创建一个后台运行、带名字的容器：

~~~~sh
go run . run -d --name jay busybox top
~~~~

如果不带 `--name` 选项，容器名则为一串十位数字的字符串。

查看运行中的容器：

~~~~sh
go run . ps
~~~~

停止运行中的容器：

~~~~sh
go run . stop ${containerName}
~~~~

删除停止的容器：

~~~~sh
go run . rm ${containerName}
~~~~

进入后台运行中的容器：

````sh
go run . exec ${runningContainer} sh
````

打包镜像：

````sh
go run . commit ${containerName} ${imageName}
````

第一个参数是正在运行的容器名，第二个参数是想要保存的镜像名，默认存放在 `/root/${imageName}.tar`

创建网桥：

~~~~sh
go run . network create --driver bridge --subnet 192.168.10.1/24 testbridge
~~~~

启动一个带网桥的容器：

````sh
go run . run -it -net testbridge busybox sh
````

