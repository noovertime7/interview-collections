// Code generated by bpf2go; DO NOT EDIT.
//go:build 386 || amd64

package docker

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"

	"github.com/cilium/ebpf"
)

// loadMydocker returns the embedded CollectionSpec for mydocker.
func loadMydocker() (*ebpf.CollectionSpec, error) {
	reader := bytes.NewReader(_MydockerBytes)
	spec, err := ebpf.LoadCollectionSpecFromReader(reader)
	if err != nil {
		return nil, fmt.Errorf("can't load mydocker: %w", err)
	}

	return spec, err
}

// loadMydockerObjects loads mydocker and converts it into a struct.
//
// The following types are suitable as obj argument:
//
//	*mydockerObjects
//	*mydockerPrograms
//	*mydockerMaps
//
// See ebpf.CollectionSpec.LoadAndAssign documentation for details.
func loadMydockerObjects(obj interface{}, opts *ebpf.CollectionOptions) error {
	spec, err := loadMydocker()
	if err != nil {
		return err
	}

	return spec.LoadAndAssign(obj, opts)
}

// mydockerSpecs contains maps and programs before they are loaded into the kernel.
//
// It can be passed ebpf.CollectionSpec.Assign.
type mydockerSpecs struct {
	mydockerProgramSpecs
	mydockerMapSpecs
}

// mydockerSpecs contains programs before they are loaded into the kernel.
//
// It can be passed ebpf.CollectionSpec.Assign.
type mydockerProgramSpecs struct {
	DockerNet *ebpf.ProgramSpec `ebpf:"docker_net"`
}

// mydockerMapSpecs contains maps before they are loaded into the kernel.
//
// It can be passed ebpf.CollectionSpec.Assign.
type mydockerMapSpecs struct {
	IpMap *ebpf.MapSpec `ebpf:"ip_map"`
}

// mydockerObjects contains all objects after they have been loaded into the kernel.
//
// It can be passed to loadMydockerObjects or ebpf.CollectionSpec.LoadAndAssign.
type mydockerObjects struct {
	mydockerPrograms
	mydockerMaps
}

func (o *mydockerObjects) Close() error {
	return _MydockerClose(
		&o.mydockerPrograms,
		&o.mydockerMaps,
	)
}

// mydockerMaps contains all maps after they have been loaded into the kernel.
//
// It can be passed to loadMydockerObjects or ebpf.CollectionSpec.LoadAndAssign.
type mydockerMaps struct {
	IpMap *ebpf.Map `ebpf:"ip_map"`
}

func (m *mydockerMaps) Close() error {
	return _MydockerClose(
		m.IpMap,
	)
}

// mydockerPrograms contains all programs after they have been loaded into the kernel.
//
// It can be passed to loadMydockerObjects or ebpf.CollectionSpec.LoadAndAssign.
type mydockerPrograms struct {
	DockerNet *ebpf.Program `ebpf:"docker_net"`
}

func (p *mydockerPrograms) Close() error {
	return _MydockerClose(
		p.DockerNet,
	)
}

func _MydockerClose(closers ...io.Closer) error {
	for _, closer := range closers {
		if err := closer.Close(); err != nil {
			return err
		}
	}
	return nil
}

// Do not access this directly.
//
//go:embed mydocker_x86_bpfel.o
var _MydockerBytes []byte
