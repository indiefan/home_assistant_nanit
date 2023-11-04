package rtmpserver

import (
	"sync"

	"github.com/notedit/rtmp/av"
)

type subscriber struct {
	initialized bool
	pktC        chan av.Packet
}

type broadcaster struct {
	headerPkts  []av.Packet
	subscribers sync.Map
}

func newBroadcaster() *broadcaster {
	return &broadcaster{}
}

func (b *broadcaster) newSubscriber() *subscriber {
	sub := &subscriber{
		initialized: false,
		pktC:        make(chan av.Packet, 10),
	}

	b.subscribers.Store(sub, sub)
	return sub
}

func (b *broadcaster) unsubscribe(sub *subscriber) {
	b.subscribers.Delete(sub)
}

func (b *broadcaster) broadcast(pkt av.Packet) {
	// Audio / Video packets
	if pkt.Type <= 2 {
		b.subscribers.Range(func(key, value interface{}) bool {
			sub := value.(*subscriber)

			// Send header packets before sending any data
			if !sub.initialized {
				sub.initialized = true
				for _, headerPkt := range b.headerPkts {
					sub.pktC <- headerPkt
				}
			}

			sub.pktC <- pkt

			return true
		})
	} else {
		// Header packets
		b.headerPkts = append(b.headerPkts, pkt)
	}
}

func (b *broadcaster) closeSubscribers() {
	b.subscribers.Range(func(key, value interface{}) bool {
		sub := value.(*subscriber)
		close(sub.pktC)
		return true
	})
}
