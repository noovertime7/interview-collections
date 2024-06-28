package xdp

//go:generate bpf2go  -cc $BPF_CLANG -cflags $BPF_CFLAGS myxdp xdp.bpf.c -- -I $BPF_HEADERS
