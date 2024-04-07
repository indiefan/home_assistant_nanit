package rtmpserver

import (
	"net"
	"regexp"
	"sync"
	"time"

	"github.com/notedit/rtmp/format/rtmp"
	"github.com/rs/zerolog/log"
	"github.com/indiefan/home_assistant_nanit/pkg/baby"
)

type rtmpHandler struct {
	babyStateManager  *baby.StateManager
	broadcastersMu    sync.RWMutex
	broadcastersByUID map[string]*broadcaster
}

// StartRTMPServer - Blocking server
func StartRTMPServer(addr string, babyStateManager *baby.StateManager) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal().Str("addr", addr).Err(err).Msg("Unable to start RTMP server")
		panic(err)
	}

	log.Info().Str("addr", addr).Msg("RTMP server started")

	s := rtmp.NewServer()
	s.HandleConn = newRtmpHandler(babyStateManager).handleConnection

	for {
		nc, err := lis.Accept()
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		go s.HandleNetConn(nc)
	}
}

func newRtmpHandler(babyStateManager *baby.StateManager) *rtmpHandler {
	return &rtmpHandler{
		broadcastersByUID: make(map[string]*broadcaster),
		babyStateManager:  babyStateManager,
	}
}

var rtmpURLRX = regexp.MustCompile(`^/local/([a-z0-9_-]+)$`)

func (s *rtmpHandler) handleConnection(c *rtmp.Conn, nc net.Conn) {
	sublog := log.With().Stringer("client_addr", nc.RemoteAddr()).Logger()

	submatch := rtmpURLRX.FindStringSubmatch(c.URL.Path)
	if len(submatch) != 2 {
		sublog.Warn().Str("path", c.URL.Path).Msg("Invalid RTMP stream requested")
		nc.Close()
		return
	}

	babyUID := submatch[1]
	sublog = sublog.With().Str("baby_uid", babyUID).Logger()

	if c.Publishing {
		sublog.Info().Msg("New stream publisher connected")
		publisher := s.getNewPublisher(babyUID)

		s.babyStateManager.Update(babyUID, *baby.NewState().SetStreamState(baby.StreamState_Alive))

		for {
			pkt, err := c.ReadPacket()
			if err != nil {
				sublog.Warn().Err(err).Msg("Publisher stream closed unexpectedly")
				s.babyStateManager.Update(babyUID, *baby.NewState().SetStreamState(baby.StreamState_Unhealthy))
				s.closePublisher(babyUID, publisher)
				return
			}

			publisher.broadcast(pkt)
		}

	} else {
		sublog.Debug().Msg("New stream subscriber connected")
		subscriber, unsubscribe := s.getNewSubscriber(babyUID)

		if subscriber == nil {
			sublog.Warn().Msg("No stream publisher registered yet, closing subscriber stream")
			nc.Close()
			return
		}

		closeC := c.CloseNotify()
		for {
			select {
			case pkt, open := <-subscriber.pktC:
				if !open {
					sublog.Debug().Msg("Closing subscriber because publisher quit")
					nc.Close()
					return
				}

				c.WritePacket(pkt)

			case <-closeC:
				sublog.Debug().Msg("Stream subscriber disconnected")
				unsubscribe()
			}
		}
	}
}

func (s *rtmpHandler) getNewPublisher(babyUID string) *broadcaster {
	broadcaster := newBroadcaster()

	s.broadcastersMu.Lock()
	existingBroadcaster, hadExistingBroadcaster := s.broadcastersByUID[babyUID]
	s.broadcastersByUID[babyUID] = broadcaster
	s.broadcastersMu.Unlock()

	if hadExistingBroadcaster {
		log.Warn().Msg("Baby already has active publisher, closing existing subscribers")
		go existingBroadcaster.closeSubscribers()
	}

	return broadcaster
}

func (s *rtmpHandler) getNewSubscriber(babyUID string) (*subscriber, func()) {
	s.broadcastersMu.RLock()
	broadcaster, hasBroadcaster := s.broadcastersByUID[babyUID]
	s.broadcastersMu.RUnlock()

	if !hasBroadcaster {
		return nil, nil
	}

	sub := broadcaster.newSubscriber()

	return sub, func() { broadcaster.unsubscribe(sub) }
}

func (s *rtmpHandler) closePublisher(babyUID string, b *broadcaster) {
	s.broadcastersMu.Lock()
	if currBroadcaster, hasExistingBroadcaster := s.broadcastersByUID[babyUID]; hasExistingBroadcaster {
		if currBroadcaster == b {
			delete(s.broadcastersByUID, babyUID)
		}
	}
	s.broadcastersMu.Unlock()

	b.closeSubscribers()
}
