#include "vmlinux.h"

#include "bpf_helpers.h"
#include "bpf_core_read.h"
#include "bpf_endian.h"
#include "bpf_tracing.h"

char __license[] SEC("license") = "GPL";

char name1[] = "conn";
char name2[] = "hshk";
char name3[] = "newc";
char name4[] = "fake";

struct {
    __uint(type, BPF_MAP_TYPE_PROG_ARRAY);
    __uint(key_size, 4);
    __uint(value_size, 4);
    __uint(max_entries, 1);
} progs SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
    __type(key, __u32);
    __type(value, __u64);
    __uint(max_entries, 1);
} socks SEC(".maps");

typedef struct event {
    __be32 saddr, daddr;
    __be16 sport, dport;
} __attribute__((packed)) event_t;

struct {
    __uint(type, BPF_MAP_TYPE_PERF_EVENT_ARRAY);
} events SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_PERF_EVENT_ARRAY);
} logs SEC(".maps");

SEC("kprobe/handle_new")
int handle_new_connection(void *ctx)
{
    __u32 key = 0;
    struct sock **skp = bpf_map_lookup_elem(&socks, &key);
    if (!skp)
        return 0;

    bpf_map_delete_elem(&socks, &key);

    struct sock *sk = *skp;
    event_t ev = {};
    ev.saddr = BPF_CORE_READ(sk, __sk_common.skc_rcv_saddr);
    ev.daddr = BPF_CORE_READ(sk, __sk_common.skc_daddr);
    ev.sport = BPF_CORE_READ(sk, __sk_common.skc_num);
    ev.dport = bpf_ntohs(BPF_CORE_READ(sk, __sk_common.skc_dport));

    bpf_perf_event_output(ctx, &logs, BPF_F_CURRENT_CPU, &name3, sizeof(name3));
    bpf_perf_event_output(ctx, &events, BPF_F_CURRENT_CPU, &ev, sizeof(ev));

    return 0;
}

SEC("kprobe/fake_new")
int fake_new_connection(void *ctx)
{
    __u32 key = 0;
    struct sock **skp = bpf_map_lookup_elem(&socks, &key);
    if (!skp)
        return 0;

    bpf_map_delete_elem(&socks, &key);

    struct sock *sk = *skp;
    event_t ev = {};
    ev.saddr = BPF_CORE_READ(sk, __sk_common.skc_rcv_saddr);
    ev.daddr = BPF_CORE_READ(sk, __sk_common.skc_daddr);
    ev.sport = BPF_CORE_READ(sk, __sk_common.skc_num);
    ev.dport = bpf_ntohs(BPF_CORE_READ(sk, __sk_common.skc_dport));

    bpf_perf_event_output(ctx, &logs, BPF_F_CURRENT_CPU, &name4, sizeof(name4));
    bpf_perf_event_output(ctx, &events, BPF_F_CURRENT_CPU, &ev, sizeof(ev));

    return 0;
}

SEC("kprobe/tcp_connect")
int k_tcp_connect(struct pt_regs *ctx)
{
    bpf_perf_event_output(ctx, &logs, BPF_F_CURRENT_CPU, &name1, sizeof(name1));
    struct sock *sk;
    sk = (typeof(sk))PT_REGS_PARM1(ctx);

    __u32 key = 0;
    bpf_map_update_elem(&socks, &key, &sk, BPF_ANY);

    bpf_tail_call_static(ctx, &progs, 0);

    return 0;
}

SEC("kprobe/inet_csk_complete_hashdance")
int k_icsk_complete_hashdance(struct pt_regs *ctx)
{
    bpf_perf_event_output(ctx, &logs, BPF_F_CURRENT_CPU, &name2, sizeof(name2));
    struct sock *sk;
    sk = (typeof(sk))PT_REGS_PARM2(ctx);

    __u32 key = 0;
    bpf_map_update_elem(&socks, &key, &sk, BPF_ANY);

    bpf_tail_call_static(ctx, &progs, 0);

    return 0;
}