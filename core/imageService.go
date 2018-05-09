package core

import (
	"encoding/gob"
	"net"

	gs "github.com/Toyz/GlitchyImageService/engine"
)

func SendImageToService(data []byte, mime string, isGif bool, expressions []string) gs.Packet {
	return imageServiceClient(data, mime, isGif, expressions)
}

func imageServiceClient(data []byte, mime string, isGif bool, expressions []string) gs.Packet {
	conn, _ := net.Dial("tcp", GetEnv("IMAGE_SERVICE", "127.0.0.1:1200"))
	defer conn.Close()

	encoder := gob.NewEncoder(conn)
	decoder := gob.NewDecoder(conn)

	encoder.Encode(gs.Packet{
		ID: 0,
		To: gs.Glitch{
			Name:        "testing_image.jpg", // goes unsused currently
			Mime:        mime,
			IsGif:       isGif,
			Expressions: expressions,
			File:        data,
		},
	})

	var packet gs.Packet
	decoder.Decode(&packet)

	return packet
}
