## 代码运行说明

- `bootstrap`/`tcpdump_demo` 此二示例的操作方法留意程序注释
- `tcpdump_demo`为tcpdump的ebpf原始字节码指令加载使用示例，代码注释中包含每步的字节码含义
- ebpf编译依赖llvm11+/clang11+，请确保已安装
- `tail_call`与`xdp_lb`程序依赖cilium/ebpf框架，所以需要先在这两个目录下调用`go generate main.go`生成ebpf二进制以及相关胶水代码文件，然后才能继续编译，详情可参考[bpf2go](https://github.com/cilium/ebpf/tree/main/cmd/bpf2go)包
- `ebpf_headers` 包中包含大多数bpf编程需要的头文件，其中`vmlinux.h`包含内核所有的数据结构但不包含对应的宏，它由`btf`生成，即`bpftool btf dump file /sys/kernel/btf/vmlinux format c > vmlinux.h`，不支持btf的内核需要手动找到对应的头文件导入，helper函数则视乎环境支持增减，与`/usr/include/bpf/bpf_helper_defs.h`保持一致即可
- `xdp_lb`示例中利用了`docker`的IP分配特性，即Mac地址最后一位跟IP地址最后一位一致，所以写的比较简单粗暴，实际用例通常需要BPF Map来存储映射关系