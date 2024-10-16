package writer

import (
	"bytes"
	"fmt"
	"github.com/mozillazg/ptcpdump/internal/types"
	"sync"

	"github.com/gopacket/gopacket"
	"github.com/gopacket/gopacket/pcapgo"
	"github.com/mozillazg/ptcpdump/internal/event"
	"github.com/mozillazg/ptcpdump/internal/log"
	"github.com/mozillazg/ptcpdump/internal/metadata"
)

type PcapNGWriter struct {
	pw         *pcapgo.NgWriter
	pcache     *metadata.ProcessCache
	interfaces []pcapgo.NgInterface
	pcapFilter string

	noBuffer bool
	lock     sync.Mutex
	keylogs  bytes.Buffer
}

func NewPcapNGWriter(pw *pcapgo.NgWriter, pcache *metadata.ProcessCache,
	interfaces []pcapgo.NgInterface) *PcapNGWriter {
	return &PcapNGWriter{pw: pw, pcache: pcache, interfaces: interfaces, lock: sync.Mutex{}}
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

	if err := w.writeTLSKeyLogs(); err != nil {
		return err
	}

	if err := w.pw.WritePacketWithOptions(info, e.Data, opts); err != nil {
		return fmt.Errorf("writing packet: %w", err)
	}
	if w.noBuffer {
		w.pw.Flush()
	}

	return nil
}

func (w *PcapNGWriter) WriteTLSKeyLog(line string) error {
	w.lock.Lock()
	defer w.lock.Unlock()

	w.keylogs.WriteString(line)

	return nil
}

func (w *PcapNGWriter) writeTLSKeyLogs() error {
	w.lock.Lock()
	defer w.lock.Unlock()

	lines := w.keylogs.Bytes()
	if len(lines) == 0 {
		return nil
	}

	if err := w.pw.WriteDecryptionSecretsBlock(pcapgo.DSB_SECRETS_TYPE_TLS, lines); err != nil {
		return fmt.Errorf("writing tls key log: %w", err)
	}

	w.keylogs.Reset()

	return nil
}

func (w *PcapNGWriter) AddDev(dev types.Device) {
	w.lock.Lock()
	defer w.lock.Unlock()

	log.Infof("new dev: %+v, currLen: %d", dev, len(w.interfaces))
	if len(w.interfaces) > dev.Ifindex {
		return
	}

	for i := len(w.interfaces); i <= dev.Ifindex; i++ {
		var intf pcapgo.NgInterface
		if i == dev.Ifindex {
			intf = metadata.NewNgInterface(dev, w.pcapFilter)
		} else {
			intf = metadata.NewDummyNgInterface(i)
		}
		log.Debugf("add interface: %+v", intf)
		if _, err := w.pw.AddInterface(intf); err != nil {
			log.Errorf("error adding interface %s: %+v", intf.Name, err)
		}
		w.interfaces = append(w.interfaces, intf)
	}
}

func (w *PcapNGWriter) WithPcapFilter(filter string) *PcapNGWriter {
	w.pcapFilter = filter
	return w
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
