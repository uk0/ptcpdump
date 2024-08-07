package writer

import (
	"fmt"

	"github.com/gopacket/gopacket"
	"github.com/gopacket/gopacket/pcapgo"
	"github.com/mozillazg/ptcpdump/internal/event"
	"github.com/mozillazg/ptcpdump/internal/log"
	"github.com/mozillazg/ptcpdump/internal/metadata"
)

type PcapNGWriter struct {
	pw     *pcapgo.NgWriter
	pcache *metadata.ProcessCache

	noBuffer bool
}

func NewPcapNGWriter(pw *pcapgo.NgWriter, pcache *metadata.ProcessCache) *PcapNGWriter {
	return &PcapNGWriter{pw: pw, pcache: pcache}
}

func (w *PcapNGWriter) Write(e *event.Packet) error {
	payloadLen := len(e.Data)
	info := gopacket.CaptureInfo{
		Timestamp:      e.Time.Local(),
		CaptureLength:  payloadLen,
		Length:         e.Len,
		InterfaceIndex: e.Device.Ifindex,
	}
	p := w.pcache.Get(e.Pid, e.MntNs, e.NetNs, e.CgroupName)

	opts := pcapgo.NgPacketOptions{}
	if p.Pid != 0 {
		log.Debugf("found pid from cache: %d", e.Pid)
		opts.Comments = append(opts.Comments,
			fmt.Sprintf("PID: %d\nCmd: %s\nArgs: %s",
				e.Pid, p.Cmd, p.FormatArgs()),
		)
		opts.Comments = append(opts.Comments,
			fmt.Sprintf("ParentPID: %d\nParentCmd: %s\nParentArgs: %s",
				p.Parent.Pid, p.Parent.Cmd, p.Parent.FormatArgs()),
		)
	}
	if p.Container.Id != "" {
		opts.Comments = append(opts.Comments,
			fmt.Sprintf("ContainerName: %s\nContainerId: %s\nContainerImage: %s\nContainerLabels: %s",
				p.Container.TidyName(), p.Container.Id, p.Container.Image, p.Container.FormatLabels()),
		)
	}
	if p.Pod.Name != "" {
		opts.Comments = append(opts.Comments,
			fmt.Sprintf("PodName: %s\nPodNamespace: %s\nPodUID: %s\nPodLabels: %s\nPodAnnotations: %s",
				p.Pod.Name, p.Pod.Namespace, p.Pod.Uid, p.Pod.FormatLabels(), p.Pod.FormatAnnotations()),
		)
	}

	if err := w.pw.WritePacketWithOptions(info, e.Data, opts); err != nil {
		return fmt.Errorf("writing packet: %w", err)
	}
	if w.noBuffer {
		w.pw.Flush()
	}

	return nil
}

func (w *PcapNGWriter) Flush() error {
	return w.pw.Flush()
}

func (w *PcapNGWriter) Close() error {
	return nil
}

func (w *PcapNGWriter) WithNoBuffer() *PcapNGWriter {
	w.noBuffer = true
	return w
}
