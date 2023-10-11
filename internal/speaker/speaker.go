package speaker

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/Speshl/gorrc_client/internal/config"
	"github.com/Speshl/gorrc_client/internal/gst"
	"github.com/pion/webrtc/v3"
)

var soundMap = map[string]string{
	"startup":             "./internal/speaker/audio/startup.wav",
	"shutdown":            "./internal/speaker/audio/shutting_down.wav",
	"client_connected":    "./internal/speaker/audio/connected.wav",
	"client_disconnected": "./internal/speaker/audio/disconnected.wav",
}

type Speaker struct {
	soundChannel chan string
	cfg          config.SpeakerConfig
}

func NewSpeaker(cfg config.SpeakerConfig, soundChannel chan string) *Speaker {
	return &Speaker{
		soundChannel: soundChannel,
		cfg:          cfg,
	}
}

func (s *Speaker) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			log.Println("speaker done due to ctx")
			return nil
		case data, ok := <-s.soundChannel:
			if !ok {
				log.Println("speaker channel closed, stopping")
				return nil
			}

			go func() {
				err := s.Play(ctx, data)
				if err != nil {
					log.Printf("failed to play sound - %s\n", err.Error())
				}
			}()
		}
	}
}

func (s *Speaker) Play(ctx context.Context, sound string) error {
	if !s.cfg.Enabled {
		log.Printf("warning: speaker disabled, not playing %s sound\n", sound)
		return nil
	}

	log.Printf("start playing %s sound\n", sound)
	defer log.Printf("finished playing %s sound\n", sound)

	soundPath, ok := soundMap[sound]
	if !ok {
		return fmt.Errorf("error: sound not found")
	}

	//TODO: configure this to specify device
	args := []string{
		// 	"-E",
		"aplay",
		// 	//"-D", "hw:CARD=wm8960soundcard,DEV=0", //use default
		soundPath,
	}
	cmd := exec.CommandContext(ctx, "sudo", args...)
	//cmd := exec.CommandContext(ctx, "aplay", soundPath)
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("error starting audio playback - %w", err)
	}
	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("error during audio playback - %w", err)
	}
	return nil
}

func (s *Speaker) TrackPlayer(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
	if !s.cfg.Enabled {
		log.Println("warning: speaker disabled, not playing user audio")
		return
	}

	log.Println("start playing client track")
	defer log.Println("done playing client track")
	codecName := strings.Split(track.Codec().RTPCodecCapability.MimeType, "/")[1]
	log.Printf("Track has started, of type %d: %s \n", track.PayloadType(), codecName)
	pipeline := gst.CreateRecievePipeline(track.PayloadType(), strings.ToLower(codecName), s.cfg.Device, s.cfg.Volume)
	pipeline.Start()
	defer pipeline.Stop()

	buf := make([]byte, 1400)
	for {
		i, _, err := track.Read(buf)
		if err != nil {
			log.Printf("stopping client audio - error reading client audio track buffer - %s\n", err)
			return
		}
		//log.Printf("Pushing %d bytes to pipeline", i)
		pipeline.Push(buf[:i])
	}
}
