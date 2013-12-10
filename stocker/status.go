package stocker

import (
	"github.com/dotcloud/docker"
	// "log"
	// "sort"
)

func ConfigsEqual(a, b *docker.Config) bool {

	// Hostname        string
	// if a.Hostname != b.Hostname {
	// 	return false
	// }
	// Domainname      string
	if a.Domainname != b.Domainname {
		return false
	}
	// User            string
	if a.User != b.User {
		return false
	}
	// Memory          int64
	if a.Memory != b.Memory {
		return false
	}
	// MemorySwap      int64
	if a.MemorySwap != b.MemorySwap {
		return false
	}
	// CpuShares       int64
	if a.CpuShares != b.CpuShares {
		return false
	}
	// AttachStdin     bool
	if a.AttachStdin != b.AttachStdin {
		return false
	}
	// AttachStdout    bool
	if a.AttachStdout != b.AttachStdout {
		return false
	}
	// AttachStderr    bool
	if a.AttachStderr != b.AttachStderr {
		return false
	}
	// PortSpecs       []string
	// ExposedPorts    map[Port]struct{}
	// Tty             bool
	if a.Tty != b.Tty {
		return false
	}
	// OpenStdin       bool
	if a.OpenStdin != b.OpenStdin {
		return false
	}
	// StdinOnce       bool
	if a.StdinOnce != b.StdinOnce {
		return false
	}
	// Env             []string
	// Cmd             []string
	// Dns             []string
	// Image           string
	// Volumes         map[string]struct{}
	// VolumesFrom     string
	if a.VolumesFrom != b.VolumesFrom {
		return false
	}
	// WorkingDir      string
	if a.WorkingDir != b.WorkingDir {
		return false
	}
	// Entrypoint      []string
	// NetworkDisabled bool
	if a.NetworkDisabled != b.NetworkDisabled {
		return false
	}

	return true
}
