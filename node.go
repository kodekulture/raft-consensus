package main

import (
	"context"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/kodekulture/raft-consensus/pb"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

var r *rand.Rand

func init() {
	r = rand.New(rand.NewSource(time.Now().UnixNano()))
}

type logStore interface {
	Add(context.Context, ...*pb.Log) (uint64, error)
}

type State int

const (
	Leader    State = 0
	Candidate State = 1
	Follower  State = 2
)

// Node - entity
type node struct {
	logger zerolog.Logger

	id int // current node id

	peer         []string
	clients      map[string]pb.NodeServiceClient
	followerLogs map[int]logMetadata

	// hearBeat is the frequence of sending heartbeats to the followers.
	heartBeat time.Duration
	// timeout is the wait time for followers to expect a heartbeat
	// from the leader node. If this timeout is exceeded, they become candidates.
	timeout time.Duration

	state     State
	stateMu   sync.RWMutex
	stateChan chan State

	logData logMetadata
	*pb.UnimplementedNodeServiceServer
}

func NewNode(id int, port string, peers []string, heartBeat time.Duration) (*node, error) {
	n := &node{
		id:        id,
		heartBeat: heartBeat,

		state: Follower,

		timeout: randomTimeout(),
		logData: logMetadata{},

		logger: log.With().Int("node_id", id).Logger(),
	}
	return n, nil
}

func (n *node) start(port string) {
	// Create a gRPC server
	server := grpc.NewServer()
	pb.RegisterNodeServiceServer(server, n)
	reflection.Register(server)

	// Create a listener on the port
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal().Err(err).Msgf("failed to listen on port: %s", port)
	}
	log.Info().Msgf("starting gRPC server on port: %s", port)

	// Start the server
	go func() {
		err = server.Serve(lis)
		if err != nil {
			log.Fatal().Err(err).Msgf("failed to serve on port %s", port)
		}
	}()

	n.stateChan = make(chan State)

	var (
		cancelOld func()                // the cancel callback for the last loop
		runner    func(context.Context) // the looping function for this node
	)

	for state := range n.stateChan {
		ctx, cancel := context.WithCancel(context.Background())

		if cancelOld != nil {
			cancelOld()
		}

		cancelOld = cancel

		switch state {
		case Leader:
			runner = n.lead
		case Candidate:
			runner = n.vote
		case Follower:
			runner = n.follow
		}
		go runner(ctx)
	}
}

// lead is the loop function for the leader node
func (n *node) lead(ctx context.Context) {
	cctx, cancel := context.WithCancel(ctx) //TODO: cctx is for clients
	defer cancel()

	_ = cctx // TODO: remove
	// Handle requests to the client
	for {
		select {
		// TODO: handling different clocks for each follower, and sends them messages of logs[followerLogs.last:]
		case <-ctx.Done():
			return
		}
	}
}

// vote is the loop function for the candidate node
func (n *node) vote(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	_ = ctx // TODO: remove
}

// follow is the loop function for the follower node
func (n *node) follow(ctx context.Context) {

}

// listen listens to the leader
// listen should be called in a goroutine only when the node is a follower
// TODO: implement
func (n *node) listen(ctx context.Context) {
	// set up the server and start listening
	for {
		select {
		case <-time.After(n.timeout):
			n.setState(Candidate)
		case <-ctx.Done():
			return
		}
	}
}

// leaderSend is called by the leader to send heartbeats to the followers
// and send new log entries.
// Calls to followers should be in a separate goroutine.
func (n *node) leaderSend(data []*pb.Log) {
	req := &pb.AppendEntriesRequest{
		Term: n.logData.Term,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	for _, peer := range n.clients {
		go func(peer pb.NodeServiceClient) {
			resp, err := peer.AppendEntries(ctx, req)
			if err != nil {
				log.Error().Err(err).Msg("failed to send heartbeat")
				return
			}

			if resp.Success {
				return
			}

			// 1- Handle error (timeoout/connection broken)
			// 2- Handle higher term from
			if resp.Term > n.logData.Term {
				n.logData.Term = resp.Term
				log.Info().Msgf("new term from client: %d", n.logData.Term)
			}
		}(peer)

	}
}

// AppendEntries is called by the leader to append new log entries
// to the current node's log
// TODO: implement
func (n *node) AppendEntries(ctx context.Context, req *pb.AppendEntriesRequest) (*pb.AppendEntriesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AppendEntries not implemented")
}

// RequestVote is called by a candidate to request votes from the followers
// TODO: implement
func (n *node) RequestVote(ctx context.Context, req *pb.RequestVoteRequest) (*pb.RequestVoteResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RequestVote not implemented")
}

func (n *node) setState(s State) {
	n.stateMu.Lock()
	n.state = s
	n.stateMu.Unlock()
}

func (n *node) getState() State {
	n.stateMu.RLock()
	defer n.stateMu.RUnlock()
	return n.state
}

func randomTimeout() time.Duration {
	return time.Duration(r.Intn(200)+150) * time.Millisecond
}
