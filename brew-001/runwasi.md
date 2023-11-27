# runwasi
# containerd 对接 runwasi
## build runwasi shim
  编译成功以后在target/x86_64-unknown-linux-gnu/debug目录可以看到wasm相关的shim二进制文件，将shim文件拷贝至/usr/local/bin/下
```
git clone https://github.com/containerd/runwasi.git
cd runwasi
./scripts/setup-linux.sh
make build
```
## 拉起wasm容器
  使用wasmedge|wasmtime|wasmer拉起一个wasm容器，我们这里使用wasmedge-shim拉起一个容器。如需使用wasmtime-shim，需安装wasmtime
```
ctr run --rm --runtime=io.containerd.wasmedge.v1 docker.io/wangqiongkaka/http_server_wasm:v1.0 http-server-wasmtime /http_server.wasm
```
## 查看wasmtime执行过程
```
ctr              30340  16803    0 /usr/local/bin/ctr run --rm --runtime=io.containerd.wasmedge.v1 docker.io/wangqiongkaka/http_server_wasm:v1.0 http-server-wasmtime /http_server.wasm
containerd-shim  30348  30182    0 /usr/local/bin/containerd-shim-wasmedge-v1 -namespace default -address /run/containerd/containerd.sock -publish-binary /usr/local/bin/containerd -id http-server-wasmtime start
containerd-shim  30349  30348    0 /usr/local/bin/containerd-shim-wasmedge-v1 -namespace default -id http-server-wasmtime -address /run/containerd/containerd.sock
containerd-shim  30365  30182    0 /usr/local/bin/containerd-shim-wasmedge-v1 -namespace default -address /run/containerd/containerd.sock -publish-binary /usr/local/bin/containerd -id http-server-wasmtime -bundle /run/containerd/io.containerd.runtime.v2.task/default/http-server-wasmtime delete
```
