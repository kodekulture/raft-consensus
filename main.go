package main

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
}

func main() {
	id, err := strconv.Atoi(os.Getenv("NODE_ID"))
	if err != nil {
		log.Fatal().Msg("failed to parse NODE_ID")
	}

	heartBeat, err := time.ParseDuration(os.Getenv("HEART_BEAT"))
	if err != nil {
		log.Fatal().Msg("failed to parse HEART_BEAT")
	}

	peers := strings.Split(os.Getenv("CLUSTER_PEERS"), ",")
	port := os.Getenv("PORT")

	node, err := NewNode(id, port, peers, heartBeat)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create node")
	}

	node.start(port)
}
