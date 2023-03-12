package p2p

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/fzft/crypto-simple-blockchain/api"
	"github.com/fzft/crypto-simple-blockchain/core"
	"github.com/fzft/crypto-simple-blockchain/crypto"
	"github.com/fzft/crypto-simple-blockchain/types"
	"github.com/go-kit/log"
	"github.com/sirupsen/logrus"
	"net"
	"os"
	"sync"
	"time"
)

const DefaultBlockTime = 5 * time.Second

type ServerOpts struct {
	APIListenAddr string
	SeedNodes     []string
	ListenAddr    string
	ID            string
	Logger        log.Logger

	Transports    []Transport
	PrivateKey    *crypto.PrivateKey
	BlockTime     time.Duration
	RPCDecodeFunc RPCDecodeFunc
	RPCProcessor  RPCProcessor
}

type Server struct {
	ServerOpts
	TCPTransport *TCPTransport
	peerCh       chan *TCPPeer

	mu      sync.RWMutex
	peerMap map[net.Addr]*TCPPeer

	isValidator bool
	memPool     *TxPool
	blockTime   time.Duration
	chain       *core.Blockchain

	stopCh chan struct{}
	txChan chan *core.Transaction

	rpcCh chan RPC
}

func NewServer(opts ServerOpts) (*Server, error) {
	if opts.BlockTime == 0 {
		opts.BlockTime = DefaultBlockTime
	}

	if opts.RPCDecodeFunc == nil {
		opts.RPCDecodeFunc = DefaultRPCDecodeFunc
	}

	if opts.Logger == nil {
		opts.Logger = log.NewLogfmtLogger(os.Stderr)
		opts.Logger = log.With(opts.Logger, "ID", opts.ID)
	}

	chain, err := core.NewBlockchain(opts.Logger, genesisBlock())
	if err != nil {
		return nil, err
	}

	txChan := make(chan *core.Transaction, 10)

	if len(opts.APIListenAddr) > 0 {
		apiServerCfg := api.ServerConfig{
			Logger:     opts.Logger,
			ListenAddr: opts.APIListenAddr,
		}

		apiServer := api.NewServer(apiServerCfg, chain, txChan)
		go apiServer.Start()

		opts.Logger.Log("msg", "accepting API connection on", "addr", opts.APIListenAddr, "id", opts.ID)
	}

	peerCh := make(chan *TCPPeer, 10)
	tr := NewTCPTransport(opts.ListenAddr, peerCh)

	s := &Server{
		TCPTransport: tr,
		peerCh:       peerCh,
		peerMap:      make(map[net.Addr]*TCPPeer),
		ServerOpts:   opts,
		chain:        chain,
		isValidator:  opts.PrivateKey != nil,
		blockTime:    opts.BlockTime,
		memPool:      NewTxPool(1000),
		stopCh:       make(chan struct{}),
		rpcCh:        make(chan RPC, 10),
		txChan:       txChan,
	}

	if opts.RPCProcessor == nil {
		opts.RPCProcessor = s
	}

	s.ServerOpts = opts

	if s.isValidator {
		go s.validatorLoop()
	}

	return s, nil
}

func (srv *Server) Start() {
	srv.TCPTransport.Start()
	time.Sleep(time.Second * 1)
	srv.bootstrapNetwork()
	srv.Logger.Log("msg", "accepting TCP connection on", "addr", srv.ListenAddr, "id", srv.ID)
LOOP:
	for {
		select {
		case peer := <-srv.peerCh:
			srv.peerMap[peer.conn.RemoteAddr()] = peer
			go peer.readLoop(srv.rpcCh)
			if err := srv.sendGetStatusMessage(peer); err != nil {
				srv.Logger.Log("msg", "could not send get status message", "err", err)
				continue
			}
			srv.Logger.Log("msg", "peer added to the server", "outgoing", peer.Outgoing, "addr", peer.conn.RemoteAddr())
		case tx := <-srv.txChan:
			if err := srv.processTransaction(tx); err != nil {
				srv.Logger.Log("msg", "could not process transaction", "err", err)
				continue
			}
		case rpc := <-srv.rpcCh:

			msg, err := srv.RPCDecodeFunc(rpc)
			if err != nil {
				srv.Logger.Log("msg", "could not decode message", "err", err)
				continue
			}

			if err = srv.RPCProcessor.ProcessMessage(msg); err != nil {
				if err != core.ErrBlockKnown {
					logrus.Error(err)

				}
				continue
			}

		case <-srv.stopCh:
			break LOOP
		}
	}

	srv.Logger.Log("msg", "Server is shutting down")
}

func (srv *Server) bootstrapNetwork() {
	for _, addr := range srv.SeedNodes {
		go func(add string) {
			tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
			if err != nil {
				srv.Logger.Log("msg", "could not resolve seed node", "err", err)
			}

			conn, err := net.DialTCP("tcp", nil, tcpAddr)
			if err != nil {
				srv.Logger.Log("msg", "could not connect to seed node", "err", err)
				return
			}

			if err = conn.SetReadBuffer(4096); err != nil {
				srv.Logger.Log("msg", "could not set read buffer", "err", err)
				return
			}

			if err = conn.SetWriteBuffer(4096); err != nil {
				srv.Logger.Log("msg", "could not set write buffer", "err", err)
				return
			}

			srv.peerCh <- &TCPPeer{
				conn: conn,
			}

			return
		}(addr)

	}
}

func (srv *Server) validatorLoop() {
	srv.Logger.Log("msg", "validator loop started", "blockTime", srv.blockTime)
	ticker := time.NewTicker(srv.blockTime)
	for {
		<-ticker.C
		err := srv.createNewBlock()
		if err != nil {
			logrus.Error(err)
		}

	}
}

func (srv *Server) ProcessMessage(msg *DecodedMessage) error {
	switch t := msg.Data.(type) {
	case *core.Transaction:
		return srv.processTransaction(t)
	case *core.Block:
		return srv.processBlock(msg.From, t)
	case *GetStatusMessage:
		return srv.processGetStatusMessage(msg.From)
	case *StatusMessage:
		return srv.processStatusMessage(msg.From, t)
	case *GetBlocksMessage:
		return srv.processGetBlocksMessage(msg.From, t)
	case *BlocksMessage:
		return srv.processBlocksMessage(msg.From, t)
	default:
		return fmt.Errorf("invalid message type: %s", t)
	}
}

func (srv *Server) broadcast(msg []byte) error {
	srv.mu.RLock()
	defer srv.mu.RUnlock()

	for addr, peer := range srv.peerMap {
		if err := peer.SendMessage(msg); err != nil {
			fmt.Printf("could not send message to peer[%s] =>: %+v\n", addr.String(), err)
			return err
		}
	}
	return nil
}

func (srv *Server) processTransaction(tx *core.Transaction) error {
	hash := tx.Hash(core.TxHasher{})
	if srv.memPool.Contains(hash) {
		return nil
	}
	if err := tx.Verify(); err != nil {
		return err
	}

	tx.SetFirstSeen(time.Now().UnixNano())

	go srv.broadcastTx(tx)

	srv.memPool.Add(tx)

	return nil
}

func (srv *Server) broadcastBlock(b *core.Block) error {
	buf := &bytes.Buffer{}
	if err := b.Encode(core.NewGobBlockEncoder(buf)); err != nil {
		return err
	}

	msg := NewMessage(MessageTypeBlock, buf.Bytes())
	return srv.broadcast(msg.Bytes())
}

func (srv *Server) broadcastTx(tx *core.Transaction) error {
	buf := &bytes.Buffer{}
	if err := tx.Encode(core.NewGobTxEncoder(buf)); err != nil {
		return err
	}
	msg := NewMessage(MessageTypeTx, buf.Bytes())
	return srv.broadcast(msg.Bytes())
}

func (srv *Server) createNewBlock() error {
	currentHeader, err := srv.chain.GetHeader(srv.chain.Height())
	if err != nil {
		return err
	}

	txx := srv.memPool.Pending()

	block, err := core.NewBlockFromHeader(currentHeader, txx)
	if err != nil {
		return err
	}

	err = block.Sign(*srv.PrivateKey)
	if err != nil {
		return err
	}

	if err = srv.chain.AddBlock(block); err != nil {
		return err
	}

	srv.memPool.ClearPending()

	go srv.broadcastBlock(block)
	return nil
}

func genesisBlock() *core.Block {
	header := &core.Header{
		Version:   1,
		DataHash:  types.Hash{},
		Timestamp: 000000,
		Height:    0,
	}

	b := &core.Block{Header: header}

	prvKey := crypto.GeneratePrivateKey()
	if err := b.Sign(prvKey); err != nil {
		panic(err)
	}

	return b
}

func (srv *Server) processBlock(from net.Addr, b *core.Block) error {
	//srv.Logger.Log("msg", "received BLOCKS!!!!!!!!", "from", from)

	if err := srv.chain.AddBlock(b); err != nil {
		return err
	}

	go srv.broadcastBlock(b)

	return nil
}

func (srv *Server) sendGetStatusMessage(peer *TCPPeer) error {
	var (
		getStatusMsg = new(GetStatusMessage)
		buf          = new(bytes.Buffer)
	)
	if err := gob.NewEncoder(buf).Encode(getStatusMsg); err != nil {
		return err
	}
	msg := NewMessage(MessageTypeGetStatus, buf.Bytes())
	return peer.SendMessage(msg.Bytes())
}

func (srv *Server) processGetStatusMessage(from net.Addr) error {
	srv.Logger.Log("msg", "received getStatus message", "from", from)
	statusMessage := &StatusMessage{
		CurrentHeight: srv.chain.Height(),
		ID:            srv.ID,
	}

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(statusMessage); err != nil {
		return err
	}
	msg := NewMessage(MessageTypeStatus, buf.Bytes())

	srv.mu.RLock()
	defer srv.mu.RUnlock()
	peer, ok := srv.peerMap[from]
	if !ok {
		return fmt.Errorf("peer %s not known", peer.conn.RemoteAddr())
	}

	return peer.SendMessage(msg.Bytes())
}

func (srv *Server) processStatusMessage(from net.Addr, data *StatusMessage) error {
	srv.Logger.Log("msg", "received STATUS message", "from", from)
	if data.CurrentHeight <= srv.chain.Height() {
		srv.Logger.Log("msg", "cannot sync block from peer", "peer", from.String(), "peerHeight", data.CurrentHeight, "ourHeight", srv.chain.Height())
		return nil
	}

	go srv.requestBlocksLoop(from)

	return nil
}

func (srv *Server) processGetBlocksMessage(from net.Addr, data *GetBlocksMessage) error {
	srv.Logger.Log("msg", "received GETBLOCKS message", "from", from)
	var (
		blocks    = []*core.Block{}
		outHeight = srv.chain.Height()
	)
	if data.To == 0 {
		for i := data.From; i <= outHeight; i++ {
			block, err := srv.chain.GetBlock(i)
			if err != nil {
				return err
			}
			blocks = append(blocks, block)
		}
	}

	srv.mu.RLock()
	defer srv.mu.RUnlock()
	peer := srv.peerMap[from]
	blocksMsg := &BlocksMessage{
		Blocks: blocks,
	}
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(blocksMsg); err != nil {
		return err
	}
	return peer.SendMessage(NewMessage(MessageTypeBlocks, buf.Bytes()).Bytes())
}

func (srv *Server) requestBlocksLoop(from net.Addr) error {
	ticker := time.NewTicker(3 * time.Second)

	for {
		//srv.Logger.Log("msg", "requesting blocks", "from", from)

		getBlockMessage := &GetBlocksMessage{
			From: srv.chain.Height() + 1,
			To:   0,
		}
		buf := new(bytes.Buffer)
		if err := gob.NewEncoder(buf).Encode(getBlockMessage); err != nil {
			return err
		}

		srv.mu.RLock()
		peer, ok := srv.peerMap[from]
		if !ok {
			return fmt.Errorf("peer %s not known", peer.conn.RemoteAddr())
		}

		msg := NewMessage(MessageTypeGetBlocks, buf.Bytes())
		if err := peer.SendMessage(msg.Bytes()); err != nil {
			srv.Logger.Log("msg", "failed to send get blocks message", "err", err)
		}

		srv.mu.RUnlock()
		<-ticker.C

	}
}

func (srv *Server) processBlocksMessage(from net.Addr, t *BlocksMessage) error {
	srv.Logger.Log("msg", "received BLOCKS message", "from", from)

	for _, block := range t.Blocks {
		if err := srv.chain.AddBlock(block); err != nil {
			srv.Logger.Log("msg", "failed to add block", "err", err)
			return err
		}
	}
	return nil
}
