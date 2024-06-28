// Code generated by bpf2go; DO NOT EDIT.
//go:build 386 || amd64

package sys

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"

	"github.com/cilium/ebpf"
)

// loadSys returns the embedded CollectionSpec for sys.
func loadSys() (*ebpf.CollectionSpec, error) {
	reader := bytes.NewReader(_SysBytes)
	spec, err := ebpf.LoadCollectionSpecFromReader(reader)
	if err != nil {
		return nil, fmt.Errorf("can't load sys: %w", err)
	}

	return spec, err
}

// loadSysObjects loads sys and converts it into a struct.
//
// The following types are suitable as obj argument:
//
//	*sysObjects
//	*sysPrograms
//	*sysMaps
//
// See ebpf.CollectionSpec.LoadAndAssign documentation for details.
func loadSysObjects(obj interface{}, opts *ebpf.CollectionOptions) error {
	spec, err := loadSys()
	if err != nil {
		return err
	}

	return spec.LoadAndAssign(obj, opts)
}

// sysSpecs contains maps and programs before they are loaded into the kernel.
//
// It can be passed ebpf.CollectionSpec.Assign.
type sysSpecs struct {
	sysProgramSpecs
	sysMapSpecs
}

// sysSpecs contains programs before they are loaded into the kernel.
//
// It can be passed ebpf.CollectionSpec.Assign.
type sysProgramSpecs struct {
	HandleProc            *ebpf.ProgramSpec `ebpf:"handle_proc"`
	HandleTaskSwitch      *ebpf.ProgramSpec `ebpf:"handle_task_switch"`
	UretprobeBashReadline *ebpf.ProgramSpec `ebpf:"uretprobe_bash_readline"`
}

// sysMapSpecs contains maps before they are loaded into the kernel.
//
// It can be passed ebpf.CollectionSpec.Assign.
type sysMapSpecs struct {
	EventMap *ebpf.MapSpec `ebpf:"event_map"`
	ProcMap  *ebpf.MapSpec `ebpf:"proc_map"`
}

// sysObjects contains all objects after they have been loaded into the kernel.
//
// It can be passed to loadSysObjects or ebpf.CollectionSpec.LoadAndAssign.
type sysObjects struct {
	sysPrograms
	sysMaps
}

func (o *sysObjects) Close() error {
	return _SysClose(
		&o.sysPrograms,
		&o.sysMaps,
	)
}

// sysMaps contains all maps after they have been loaded into the kernel.
//
// It can be passed to loadSysObjects or ebpf.CollectionSpec.LoadAndAssign.
type sysMaps struct {
	EventMap *ebpf.Map `ebpf:"event_map"`
	ProcMap  *ebpf.Map `ebpf:"proc_map"`
}

func (m *sysMaps) Close() error {
	return _SysClose(
		m.EventMap,
		m.ProcMap,
	)
}

// sysPrograms contains all programs after they have been loaded into the kernel.
//
// It can be passed to loadSysObjects or ebpf.CollectionSpec.LoadAndAssign.
type sysPrograms struct {
	HandleProc            *ebpf.Program `ebpf:"handle_proc"`
	HandleTaskSwitch      *ebpf.Program `ebpf:"handle_task_switch"`
	UretprobeBashReadline *ebpf.Program `ebpf:"uretprobe_bash_readline"`
}

func (p *sysPrograms) Close() error {
	return _SysClose(
		p.HandleProc,
		p.HandleTaskSwitch,
		p.UretprobeBashReadline,
	)
}

func _SysClose(closers ...io.Closer) error {
	for _, closer := range closers {
		if err := closer.Close(); err != nil {
			return err
		}
	}
	return nil
}

// Do not access this directly.
//
//go:embed sys_x86_bpfel.o
var _SysBytes []byte
