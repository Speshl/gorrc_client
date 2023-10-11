package mic

import (
	"fmt"
	"log"

	"github.com/Speshl/gorrc_client/internal/config"
	"github.com/Speshl/gorrc_client/internal/gst"
	"github.com/pion/webrtc/v3"
)

type Mic struct {
	AudioTrack *webrtc.TrackLocalStaticSample
	config     config.MicConfig
}

func NewMic(cfg config.MicConfig) (*Mic, error) {
	// Create a audio track
	audioTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: "audio/opus"}, "audio", "pion1")
	if err != nil {
		return nil, fmt.Errorf("error creating audio track: %w", err)
	}

	carMic := Mic{
		AudioTrack: audioTrack,
		config:     cfg,
	}

	return &carMic, nil
}

func (c *Mic) Start() {
	if c.config.Enabled {
		log.Println("creating mic pipeline")
		gst.CreateMicSendPipeline([]*webrtc.TrackLocalStaticSample{c.AudioTrack}, c.config.Device, c.config.Volume).Start()
	} else {
		log.Println("mic disabled")
	}

}
