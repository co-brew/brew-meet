# crun  

## 编译安装crun
### 安装wasmedge
```
wget -qO- https://raw.githubusercontent.com/WasmEdge/WasmEdge/master/utils/install.sh | bash -s -- -p /usr/local
```
### 安装依赖包
```
sudo apt update
sudo apt install -y make git gcc build-essential pkgconf libtool \
   libsystemd-dev libprotobuf-c-dev libcap-dev libseccomp-dev libyajl-dev \
   go-md2man libtool autoconf python3 automake
```
### 编译crun
   编译完成以后在/usr/local/bin目录可以看到crun的二进制文件
```
git clone https://github.com/containers/crun
cd crun
./autogen.sh
./configure --with-wasmedge
make
sudo make install
```
## 安装containerd
   安装成功以后在/usr/local/bin目录可以看到containerd相关的二进制文件
```
export VERSION="1.5.7"
echo -e "Version: $VERSION"
echo -e "Installing libseccomp2 ..."
sudo apt install -y libseccomp2
echo -e "Installing wget"
sudo apt install -y wget

wget https://github.com/containerd/containerd/releases/download/v${VERSION}/cri-containerd-cni-${VERSION}-linux-amd64.tar.gz
wget https://github.com/containerd/containerd/releases/download/v${VERSION}/cri-containerd-cni-${VERSION}-linux-amd64.tar.gz.sha256sum
sha256sum --check cri-containerd-cni-${VERSION}-linux-amd64.tar.gz.sha256sum

sudo tar --no-overwrite-dir -C / -xzf cri-containerd-cni-${VERSION}-linux-amd64.tar.gz
sudo systemctl daemon-reload
sudo systemctl start containerd
```
## 对接crun
   containerd使用crun拉起一个wasm容器
### 拉取我们之前构建的wasm镜像
```
ctr i pull docker.io/wangqiongkaka/http_server_wasm:v1.0
```
### 拉起容器
   这里需要注意运行wasm容器需添加 label module.wasm.image/variant 告诉crun使用的wasm
```
ctr run --rm --net-host --runc-binary crun --runtime io.containerd.runc.v2 --label module.wasm.image/variant=compat-smart docker.io/wangqiongkaka/http_server_wasm:v1.0 http-server-example /http_server.wasm
```
### 执行验证
```
curl -d "name=crun" -X POST http://127.0.0.1:1234
```
### 使用crun运行普通容器
   拉取一个普通容器
```
ctr i pull docker.io/wangqiongkaka/nginx:latest
ctr run --rm --net-host --runc-binary crun --runtime io.containerd.runc.v2 docker.io/wangqiongkaka/nginx:latest nginx
curl 127.0.0.1
```
   跟我们使用runc运行普通容器是一样的效果
## 查看进程执行过程
   我们可以使用ebpf查看containerd拉起一个容器时的执行过程
```
apt-get -y install bpfcc-tools
```
   执行execsnoop-bpfcc
```
execsnoop-bpfcc
```
   我们新开一个窗口拉起一个wasm容器，执行过程如下(crun 执行一个wasm容器)
```
ctr run --rm --net-host --runc-binary crun --runtime io.containerd.runc.v2 --label module.wasm.image/variant=compat-smart docker.io/wangqiongkaka/http_server_wasm:v1.0 http-server-example1 /http_server.wasm
```
```
ctr              15146  14607    0 /usr/local/bin/ctr run --rm --net-host --runc-binary crun --runtime io.containerd.runc.v2 --label module.wasm.image/variant=compat-smart docker.io/wangqiongkaka/http_server_wasm:v1.0 http-server-example1 /http_server.wasm
containerd-shim  15155  14425    0 /usr/local/bin/containerd-shim-runc-v2 -namespace default -address /run/containerd/containerd.sock -publish-binary /usr/local/bin/containerd -id http-server-example1 start
containerd-shim  15163  15155    0 /usr/local/bin/containerd-shim-runc-v2 -namespace default -id http-server-example1 -address /run/containerd/containerd.sock
crun             15172  15163    0 /usr/local/bin/crun --root /run/containerd/runc/default --log /run/containerd/io.containerd.runtime.v2.task/default/http-server-example1/log.json --log-format json create --bundle /run/containerd/io.containerd.runtime.v2.task/default/http-server-example1 --pid-file /run/containerd/io.containerd.runtime.v2.task/default/http-server-example1/init.pid http-server-example1
dumpe2fs         15173  3198     0 /usr/sbin/dumpe2fs -h /dev/vda2
dumpe2fs         15174  3198     0 /usr/sbin/dumpe2fs -h /dev/vda2
crun             15176  15163    0 /usr/local/bin/crun --root /run/containerd/runc/default --log /run/containerd/io.containerd.runtime.v2.task/default/http-server-example1/log.json --log-format json start http-server-example1
crun             15179  15163    0 /usr/local/bin/crun --root /run/containerd/runc/default --log /run/containerd/io.containerd.runtime.v2.task/default/http-server-example1/log.json --log-format json delete http-server-example1
containerd-shim  15181  14425    0 /usr/local/bin/containerd-shim-runc-v2 -namespace default -address /run/containerd/containerd.sock -publish-binary /usr/local/bin/containerd -id http-server-example1 -bundle /run/containerd/io.containerd.runtime.v2.task/default/http-server-example1 delete
crun             15190  15181    0 /usr/local/bin/crun --root /run/containerd/runc/default --log /run/containerd/io.containerd.runtime.v2.task/default/http-server-example1/log.json --log-format json delete --force http-server-example1
```
   对比runc拉起一个普通容器的执行过程
```
ctr run --rm --net-host --runc-binary runc --runtime io.containerd.runc.v2 docker.io/wangqiongkaka/nginx:latest nginx
```
```
ctr              15274  14607    0 /usr/local/bin/ctr run --rm --net-host --runc-binary runc --runtime io.containerd.runc.v2 docker.io/wangqiongkaka/nginx:latest nginx
containerd-shim  15283  14425    0 /usr/local/bin/containerd-shim-runc-v2 -namespace default -address /run/containerd/containerd.sock -publish-binary /usr/local/bin/containerd -id nginx start
containerd-shim  15291  15283    0 /usr/local/bin/containerd-shim-runc-v2 -namespace default -id nginx -address /run/containerd/containerd.sock
runc             15302  15291    0 /usr/local/sbin/runc --root /run/containerd/runc/default --log /run/containerd/io.containerd.runtime.v2.task/default/nginx/log.json --log-format json create --bundle /run/containerd/io.containerd.runtime.v2.task/default/nginx --pid-file /run/containerd/io.containerd.runtime.v2.task/default/nginx/init.pid nginx
exe              15312  15302    0 /proc/self/exe init
runc             15321  15291    0 /usr/local/sbin/runc --root /run/containerd/runc/default --log /run/containerd/io.containerd.runtime.v2.task/default/nginx/log.json --log-format json start nginx
docker-entrypoi  15314  15291    0 /docker-entrypoint.sh nginx -g daemon off;
find             15328  15314    0 /usr/bin/find /docker-entrypoint.d/ -mindepth 1 -maxdepth 1 -type f -print -quit
sort             15331  15314    0 /usr/bin/sort -V
find             15330  15314    0 /usr/bin/find /docker-entrypoint.d/ -follow -type f -print
10-listen-on-ip  15333  15332    0 /docker-entrypoint.d/10-listen-on-ipv6-by-default.sh
basename         15334  15333    0 /usr/bin/basename /docker-entrypoint.d/10-listen-on-ipv6-by-default.sh
touch            15335  15333    0 /usr/bin/touch /etc/nginx/conf.d/default.conf
grep             15336  15333    0 /bin/grep -q listen  \[::]\:80; /etc/nginx/conf.d/default.conf
dpkg-query       15338  15337    0 /usr/bin/dpkg-query --show --showformat=${Conffiles}\n nginx
cut              15340  15337    0 /usr/bin/cut -d  -f 3
grep             15339  15337    0 /bin/grep etc/nginx/conf.d/default.conf
md5sum           15342  15333    0 /usr/bin/md5sum -c -
sed              15343  15333    0 /bin/sed -i -E s,listen       80;,listen       80;\n    listen  [::]:80;, /etc/nginx/conf.d/default.conf
20-envsubst-on-  15344  15332    0 /docker-entrypoint.d/20-envsubst-on-templates.sh
basename         15345  15344    0 /usr/bin/basename /docker-entrypoint.d/20-envsubst-on-templates.sh
env              15348  15347    0 /usr/bin/env
cut              15349  15347    0 /usr/bin/cut -d= -f1
30-tune-worker-  15350  15332    0 /docker-entrypoint.d/30-tune-worker-processes.sh
basename         15351  15350    0 /usr/bin/basename /docker-entrypoint.d/30-tune-worker-processes.sh
nginx            15314  15291    0 /usr/sbin/nginx -g daemon off;
runc             15354  15291    0 /usr/local/sbin/runc --root /run/containerd/runc/default --log /run/containerd/io.containerd.runtime.v2.task/default/nginx/log.json --log-format json kill nginx 2
runc             15362  15291    0 /usr/local/sbin/runc --root /run/containerd/runc/default --log /run/containerd/io.containerd.runtime.v2.task/default/nginx/log.json --log-format json delete nginx
containerd-shim  15369  14425    0 /usr/local/bin/containerd-shim-runc-v2 -namespace default -address /run/containerd/containerd.sock -publish-binary /usr/local/bin/containerd -id nginx -bundle /run/containerd/io.containerd.runtime.v2.task/default/nginx delete
runc             15375  15369    0 /usr/local/sbin/runc --root /run/containerd/runc/default --log /run/containerd/io.containerd.runtime.v2.task/default/nginx/log.json --log-format json delete --force nginx
```
   对比crun拉起一个普通容器的执行过程
```
ctr run --rm --net-host --runc-binary crun --runtime io.containerd.runc.v2 docker.io/wangqiongkaka/nginx:latest nginx
```
```
ctr              15383  14607    0 /usr/local/bin/ctr run --rm --net-host --runc-binary crun --runtime io.containerd.runc.v2 docker.io/wangqiongkaka/nginx:latest nginx
containerd-shim  15392  14425    0 /usr/local/bin/containerd-shim-runc-v2 -namespace default -address /run/containerd/containerd.sock -publish-binary /usr/local/bin/containerd -id nginx start
containerd-shim  15398  15392    0 /usr/local/bin/containerd-shim-runc-v2 -namespace default -id nginx -address /run/containerd/containerd.sock
crun             15409  15398    0 /usr/local/bin/crun --root /run/containerd/runc/default --log /run/containerd/io.containerd.runtime.v2.task/default/nginx/log.json --log-format json create --bundle /run/containerd/io.containerd.runtime.v2.task/default/nginx --pid-file /run/containerd/io.containerd.runtime.v2.task/default/nginx/init.pid nginx
dumpe2fs         15410  3198     0 /usr/sbin/dumpe2fs -h /dev/vda2
dumpe2fs         15411  3198     0 /usr/sbin/dumpe2fs -h /dev/vda2
crun             15413  15398    0 /usr/local/bin/crun --root /run/containerd/runc/default --log /run/containerd/io.containerd.runtime.v2.task/default/nginx/log.json --log-format json start nginx
dumpe2fs         15414  3198     0 /usr/sbin/dumpe2fs -h /dev/vda2
docker-entrypoi  15412  15398    0 /docker-entrypoint.sh nginx -g daemon off;
dumpe2fs         15417  3198     0 /usr/sbin/dumpe2fs -h /dev/vda2
find             15418  15412    0 /usr/bin/find /docker-entrypoint.d/ -follow -type f -print
find             15415  15412    0 /usr/bin/find /docker-entrypoint.d/ -mindepth 1 -maxdepth 1 -type f -print -quit
sort             15419  15412    0 /usr/bin/sort -V
10-listen-on-ip  15421  15420    0 /docker-entrypoint.d/10-listen-on-ipv6-by-default.sh
basename         15422  15421    0 /usr/bin/basename /docker-entrypoint.d/10-listen-on-ipv6-by-default.sh
touch            15423  15421    0 /usr/bin/touch /etc/nginx/conf.d/default.conf
grep             15424  15421    0 /bin/grep -q listen  \[::]\:80; /etc/nginx/conf.d/default.conf
dpkg-query       15426  15425    0 /usr/bin/dpkg-query --show --showformat=${Conffiles}\n nginx
grep             15427  15425    0 /bin/grep etc/nginx/conf.d/default.conf
cut              15428  15425    0 /usr/bin/cut -d  -f 3
md5sum           15430  15421    0 /usr/bin/md5sum -c -
sed              15431  15421    0 /bin/sed -i -E s,listen       80;,listen       80;\n    listen  [::]:80;, /etc/nginx/conf.d/default.conf
20-envsubst-on-  15432  15420    0 /docker-entrypoint.d/20-envsubst-on-templates.sh
basename         15433  15432    0 /usr/bin/basename /docker-entrypoint.d/20-envsubst-on-templates.sh
env              15436  15435    0 /usr/bin/env
cut              15437  15435    0 /usr/bin/cut -d= -f1
30-tune-worker-  15438  15420    0 /docker-entrypoint.d/30-tune-worker-processes.sh
basename         15439  15438    0 /usr/bin/basename /docker-entrypoint.d/30-tune-worker-processes.sh
nginx            15412  15398    0 /usr/sbin/nginx -g daemon off;
crun             15442  15398    0 /usr/local/bin/crun --root /run/containerd/runc/default --log /run/containerd/io.containerd.runtime.v2.task/default/nginx/log.json --log-format json kill nginx 2
crun             15443  15398    0 /usr/local/bin/crun --root /run/containerd/runc/default --log /run/containerd/io.containerd.runtime.v2.task/default/nginx/log.json --log-format json delete nginx
dumpe2fs         15444  3198     0 /usr/sbin/dumpe2fs -h /dev/vda2
dumpe2fs         15446  3198     0 /usr/sbin/dumpe2fs -h /dev/vda2
containerd-shim  15447  14425    0 /usr/local/bin/containerd-shim-runc-v2 -namespace default -address /run/containerd/containerd.sock -publish-binary /usr/local/bin/containerd -id nginx -bundle /run/containerd/io.containerd.runtime.v2.task/default/nginx delete
crun             15455  15447    0 /usr/local/bin/crun --root /run/containerd/runc/default --log /run/containerd/io.containerd.runtime.v2.task/default/nginx/log.json --log-format json delete --force nginx
dumpe2fs         15456  3198     0 /usr/sbin/dumpe2fs -h /dev/vda2
dumpe2fs         15458  3198     0 /usr/sbin/dumpe2fs -h /dev/vda2
```
## 结论
   crun既能执行wasm容器也可以执行普通容器，crun与runc都支持 containerd-shim-runc-v2，调用过程都是一样的，都是先create然后start接着delete，理论上我们可以直接使用crun代替runc
