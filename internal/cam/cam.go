package cam

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/Speshl/gorrc_client/internal/config"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
)

const DefaultLevel = "4.2"
const DefaultFPS = 30

type Cam struct {
	VideoTrack   *webrtc.TrackLocalStaticSample
	videoChannel chan []byte
	cfg          config.CamConfig
}

func NewCam(cfg config.CamConfig) (*Cam, error) {
	// Create a video track
	videoTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, "video", "pion")
	if err != nil {
		return nil, fmt.Errorf("error creating first video track: %w", err)
	}

	carCam := Cam{
		VideoTrack:   videoTrack,
		videoChannel: make(chan []byte, 5),
		cfg:          cfg,
	}
	carCam.cfg.Level = DefaultLevel
	return &carCam, nil
}

func (c *Cam) Start(ctx context.Context) error {
	go c.StartVideoDataListener(ctx)
	return c.StartStreaming(ctx)
}

func (c *Cam) StartVideoDataListener(ctx context.Context) {
	fps, err := strconv.ParseInt(c.cfg.Fps, 10, 32)
	if err != nil {
		fps = DefaultFPS
	}

	duration := int(1000 / fps)
	for {
		select {
		case <-ctx.Done():
			log.Println("video data listener done due to ctx")
			return
		case data, ok := <-c.videoChannel:
			if !ok {
				log.Println("video data channel closed, stopping")
				return
			}

			err := c.VideoTrack.WriteSample(media.Sample{Data: data, Duration: time.Millisecond * time.Duration(duration)})
			if err != nil {
				log.Printf("error writing sample to track: %s\n", err.Error())
				return
			}
		}
	}
}
