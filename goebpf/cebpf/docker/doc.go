package docker

//_go:generate bpf2go  -cc $BPF_CLANG -cflags $BPF_CFLAGS -target amd64 mydocker docker.bpf.c -- -I $BPF_HEADERS

//go:generate bpf2go  -cc $BPF_CLANG -cflags $BPF_CFLAGS -target amd64 mydockertc dockertc.bpf.c -- -I $BPF_HEADERS
