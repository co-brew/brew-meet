## 构建wasm镜像
### 安装rust  
```  
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
```
### 编译 .wasm 文件  
  编译完成后会生成一个http_server.wasm二进制文件 
```
apt-get update && apt-get -y install build-essential && rustup target add wasm32-wasi  
git clone https://github.com/second-state/wasmedge_wasi_socket.git  
cd wasmedge_wasi_socket/examples/http_server  
cargo build --target wasm32-wasi --release  
cd target/wasm32-wasi/release
```  
### 构建镜像 
```
FROM scratch  
ADD http_server.wasm  
CMD ["/http_server.wasm"]
```
```
docker build . -t docker.io/wangqiongkaka/http_server_wasm:v1.0
```
  crun容器运行时可以启动基于WebAssembly的容器镜像。但它需要在容器镜像上添加module.wasm.image/variant=compat-smart注释，以表明它是一个没有客户操作系统的WebAssembly应用程序。要在容器镜像中添加module.wasm.image/variant=compat-smart，目前Docker不支持此功能，需要使用最新的buildah
```
buildah build . --annotation "module.wasm.image/variant=compat-smart" -t docker.io/wangqiongkaka/http_server_wasm:v1.1
```
## 参考文档  
  https://github.com/second-state/wasmedge-containers-examples/blob/main/http_server_wasi_app.md  
  https://www.rust-lang.org/tools/install
