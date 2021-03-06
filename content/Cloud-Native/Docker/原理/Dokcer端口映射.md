---
title: Dokcer端口映射
date: 2020-04-14T10:09:14.122627+08:00
draft: false
---

多个服务组件容器共同协作提供服务，需要多个容器之间有能够相互访问到对方的服务。

除了通过网络访问外，Docker提供了两个功能来满足服务访问的需求：

1. 映射容器内应用的服务端口到本地宿主机
2. 互联机制实现多个容器见通过容器名快速访问

# 进程端口映射实现容器访问

### 从外部访问容器应用

在启动容器的时候，如果不指定对应的参数，在容器外部是无法通过网络来访问容器内的网络应用和服务的。

使用-P/-p参数来指定端口映射：

- 大写： Docker会随机映射主机的49000~49900的端口到容器的进程端口
- 小写：自己手动指定需要映射的端口地址，宿主机的一个端口只能绑定一个容器的端口

**小写的时候，支持如下三种格式：**

1. ++IP：HostPort：ContainerPort++：映射到指定接口的指定端口
2. ++IP：：ContainerPort++：映射到指定接口的任意端口，此时宿主机会自动分配一个端口
3. ++HostPort：ContainerPort++：  默认会绑定宿主机所有接口（网卡）上的所有地址，多次使用-p可以绑定多个端口

```
docker port //查看当前映射的端口配置
docker inspect + 容器ID  //查看容器自己的内部网络和IP地址
```

### 互联机制实现便捷互访

容器的互联是一种让多个容器中应用进行快速交互的方式，它会在源和接收容器之间创建连接关系，接收容器可以通过容器名快速访问到源容器，而不需要指定具体的IP地址

### 自定义容器名

连接系统依据容器的名称来执行，因此需要定义一个好记忆的容器名字，自定义名字的好处：

1. 好记
2. 当要连接其他容器时，即便重启，也可以使用容器名而不用改变

**容器的名字是唯一的**

### 容器互联

使用--link参数可以让容器之间安全地进行交互
--link参数的格式：--link name：alias  其中name是要源容器的名字，alias是这个link的别名

```
docker run -d --name db training/postgres
docker run -d -P --name web --link db:db training/webapp python app.py
//web容器连接了db容器，两者之间的这个link的别名是db
//允许web容器访问db容器
```

这相当于在两个互联的容器之间创建了虚拟通道，而且不用映射他们的端口到宿主机上，在启动容器的时候并没有使用-p参数标记，这样也不会暴露数据库服务器的端口到外部网络。

**Docker通过两种方式为容器公开连接信息**：

1. 更新环境变量
2. 更新/etc/hosts文件

使用env命令来查看web容器的环境变量：

```
docker run --rm --name web2 --link db:db training/webapp env
...
DB_NAME=/web2/db
DB_PORT=tcp://172.17.0.5:5432
...
```

其中DB_开头的环境病历是供web容器连接db容器使用的（前缀采用别名的大写）
除了环境变量，docker还提供host信息到源容器的/etc/hosts文件

```
docker run -it --rm --link db:db  training/webapp /bin/bash
root@aed84ee21bde:/opt/webapp# cat /etc/hosts
172.17.0.7  aed84ee21bde   //web容器自己
172.17.0.2  db             //db容器
```

可以同ping命令测试容器的连通性，可以连接多个子容器到源容器。
