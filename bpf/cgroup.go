package bpf

import (
	"fmt"
	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/mozillazg/ptcpdump/internal/log"
)

func (b *BPF) AttachCgroups(cgroupPath string) error {
	if cgroupPath == "" {
		b.skipAttachCgroup = true
	}
	if b.skipAttachCgroup {
		log.Info("skipping cgroup attach")
		return nil
	}

	log.Info("attaching cgroup/sock_create")
	lk, err := link.AttachCgroup(link.CgroupOptions{
		Path:    cgroupPath,
		Attach:  ebpf.AttachCGroupInetSockCreate,
		Program: b.objs.PtcpdumpCgroupSockCreate,
	})
	if err != nil {
		return fmt.Errorf("attach cgroup/sock_create: %w", err)
	}
	b.links = append(b.links, lk)

	log.Info("attaching cgroup/sock_release")
	lk, err = link.AttachCgroup(link.CgroupOptions{
		Path:    cgroupPath,
		Attach:  ebpf.AttachCgroupInetSockRelease,
		Program: b.objs.PtcpdumpCgroupSockRelease,
	})
	if err != nil {
		return fmt.Errorf("attach cgroup/sock_release: %w", err)
	}
	b.links = append(b.links, lk)

	return nil
}

func (b *BPF) disableCgroupSkb() {
	for k, v := range b.spec.Programs {
		if v.Type == ebpf.CGroupSKB {
			delete(b.spec.Programs, k)
		}
	}
}

func (b *BPF) AttachCgroupSkb(cgroupPath string, egress, ingress bool) error {
	if cgroupPath == "" {
		b.skipAttachCgroup = true
	}
	if b.skipAttachCgroup {
		return nil
	}

	if ingress {
		log.Info("attaching cgroup_skb/ingress")
		lk, err := link.AttachCgroup(link.CgroupOptions{
			Path:    cgroupPath,
			Attach:  ebpf.AttachCGroupInetIngress,
			Program: b.objs.PtcpdumpCgroupSkbIngress,
		})
		if err != nil {
			return fmt.Errorf("attach cgroup_skb/ingress: %w", err)
		}
		b.links = append(b.links, lk)
	}
	if egress {
		log.Info("attaching cgroup_skb/egress")
		lk, err := link.AttachCgroup(link.CgroupOptions{
			Path:    cgroupPath,
			Attach:  ebpf.AttachCGroupInetEgress,
			Program: b.objs.PtcpdumpCgroupSkbEgress,
		})
		if err != nil {
			return fmt.Errorf("attach cgroup_skb/egress: %w", err)
		}
		b.links = append(b.links, lk)
	}

	return nil
}
