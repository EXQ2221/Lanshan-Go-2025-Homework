 
# 一、容器到底是不是虚拟机？
![[c440d2c19b6922c8b7eb9060881a8f23.png]]
## 1.  启动一个 mysql 容器：
运行
```bash 
docker exec -it example-mysql /bin/sh
```
这里运行ps失败，说明docker启动的mysql的docker镜像只包含mysql需要的最小用户态环境，并没有完整的linux系统

## 2.启动一个 ubuntu 容器
运行
```bash
ps aux | head -n 5
```
容器内看到的PID1是bash，说明这个容器并没有一套完整的操作系统，只是把bash这个普通的进程当成了容器的第一个进程
这里kill 1杀不掉，我问了一下AI说是Linux里面对PID1进行了特殊处理让他不会响应kill
## 3.在宿主机内查看
![[79376e74cc6eb58447d6ec43a06ed2cd.png]]
这里运行
```bash
docker top d291c94c7580
```
看到这个ubuntu容器能被宿主机看到，且是一个普通进程, PID为1553，我运行ps查看进程发现报错了，后面发现要进去linux环境才能看，这说明docker本质就是linux进程加上内核隔离

## 4.总结
这次实验说明 Docker 容器不是虚拟机。
在容器里看到的PID1是 bash，而不是 init/systemd，说明容器里并没有启动一整套操作系统，只是把一个普通进程当成了第一个进程。 在宿主机的wsl上还能直接看到这个进程，说明容器本质上就是宿主机上的 Linux 进程。
Docker 只是通过 namespace 隔离视图、用 cgroup 限制资源，让这个进程看起来像一台独立的机器。

# 二 、用docker-compose构建一个最小可用系统
## 1.构建并查看
```yaml
version: "3.8"

services:

  web:

    image: nginx:latest

    container_name: demo-web

    ports:

      - "8080:80"

    networks:

      - demo-net

  db:

    image: mysql:8.0

    container_name: demo-db

    environment:

      MYSQL_ROOT_PASSWORD: root

      MYSQL_DATABASE: testdb

    volumes:

      - mysql-data:/var/lib/mysql

    networks:

      - demo-net

networks:

  demo-net:

    driver: bridge

volumes:

  mysql-data:
```
![[a262f38e386f583c99c8b9763de0a7f5 1.png]]
运行
```bash
docker compose up
```
可以看到构建成功了
接下来运行
```bash
docker ps
```
![[eb59ffc4698ec467acf3f0c3db6199e4.png]]
可以看到web和db两个容器正常运行，并且处于同一个自定义网络中

## 2.结论
docker-compose把相关的容器写在一个配置文件里，一条命令一起启动
建立bridge之后，容器之间可以互相访问
volume做到单独把数据存出来，可以保留数据

# 3.Docker 容器在 Linux 中留下了什么痕迹
![[79376e74cc6eb58447d6ec43a06ed2cd 1.png]]
这里就可以看到docker 在linux上就是一个普通的进程，他只是被namespace和cgroup管理了