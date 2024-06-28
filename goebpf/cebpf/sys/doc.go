package sys

//_go:generate bpf2go  -cc $BPF_CLANG -cflags $BPF_CFLAGS sys sys.bpf.c -- -I $BPF_HEADERS
//go:generate bpf2go  -cc $BPF_CLANG -cflags $BPF_CFLAGS -target amd64 sys sys.bpf.c -- -I $BPF_HEADERS
