## 主题：eBPF简述

### 时间：2023年12月23日 14:00 - 17:00

### 主要内容：
本文介绍eBPF相关的概念与相关工具，行业应用场景，编程实战，还有安全问题。共计9个部分
- 社区暴论与初步理解
- 历史背景简介
- 相关项目应用介绍
- BPF基本概念介绍与[`bpftool`](https://github.com/libbpf/bpftool)使用简介
- [`bpftrace`](https://github.com/iovisor/bpftrace/blob/master/INSTALL.md)/[`bcc-tools`](https://github.com/iovisor/bcc/blob/master/INSTALL.md)等命令行工具简介
- bpftrace为例子介绍bpf加载运行流程
- 基于XDP从零实现一个四层LB(部署[`docker`](https://docs.docker.com/engine/install/ubuntu/)之上)
- BPF网络栈与[`Cilium`](https://github.com/cilium/cilium/)工作流程简介
- 安全挑战与能力局限

### 学习建议：
- 不太建议一开始上手就尝试工程化撸代码，先玩熟 `bpftrace`/`bcc-tools` 已经够充分入门理解
- 可以在工具化实践中，可尝试多修改运行一下 `bcc-tools` 相关工具加强成就感，毕竟报错相对友好，同时在工程化过程中，[可以使用bcc加载宏参数debug=4](https://just4coding.com/2022/03/22/ebpf-c/)来输出宏展开后的原始C语言代码，辅助和简化编写纯C代码的步骤
- 理解加载流程之后尝试工程化时可以考虑稍微读一下[`cilium/ebpf`](https://github.com/cilium/ebpf)的代码，特别是git历史，对理解项目演化和避坑，还有理解内核verifier报错时莫名其妙的输出会有帮助

### 主持人：
- Martin Ho
- 张晓辉 @addozhang

### 其他参与人员：[共8个](https://community.cncf.io/events/details/cncf-cloud-native-guangzhou-presents-brew-meet/)