package api

import (
	"encoding/gob"
	"encoding/hex"
	"github.com/fzft/crypto-simple-blockchain/core"
	"github.com/fzft/crypto-simple-blockchain/types"
	"github.com/go-kit/kit/log"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

type TxResponse struct {
	TxCount uint
	Hashes  []string
}

type Error struct {
	Error string
}

type Block struct {
	Version       uint32
	DataHash      string
	PrevBlockHash string
	Height        uint32
	Timestamp     int64
	Validator     string
	Signature     string
	Hash          string

	Transactions []TxResponse
}

type ServerConfig struct {
	Logger     log.Logger
	ListenAddr string
}

type Server struct {
	ServerConfig
	bc     *core.Blockchain
	txChan chan *core.Transaction
}

func NewServer(config ServerConfig, bc *core.Blockchain, txChan chan *core.Transaction) *Server {
	return &Server{
		ServerConfig: config,
		bc:           bc,
		txChan:       txChan,
	}
}

func (s *Server) Start() error {
	e := echo.New()
	e.GET("/block/:hashorid", s.handleGetBlock)
	e.GET("/tx/:hash", s.handleGetTx)
	e.POST("/tx", s.handlePostTx)
	return e.Start(s.ListenAddr)
}

func (s *Server) handleGetBlock(c echo.Context) error {
	hashOrId := c.Param("hashorid")

	height, err := strconv.Atoi(hashOrId)
	//
	if err == nil {
		block, err := s.bc.GetBlock(uint32(height))
		if err != nil {
			return c.JSON(http.StatusNotFound, Error{Error: err.Error()})
		}

		jsonBlock := intoJSONBlock(block)
		return c.JSON(http.StatusOK, jsonBlock)
	}

	b, err := hex.DecodeString(hashOrId)
	if err != nil {
		return c.JSON(http.StatusNotFound, Error{Error: err.Error()})
	}
	block, err := s.bc.GetBlockByHash(types.HashFromBytes(b))

	if err == nil {
		jsonBlock := intoJSONBlock(block)
		return c.JSON(http.StatusOK, jsonBlock)

	}

	return c.JSON(http.StatusOK, map[string]any{"msg": "hello"})
}

func (s *Server) handleGetTx(c echo.Context) error {
	hash := c.Param("hash")

	b, err := hex.DecodeString(hash)
	if err != nil {
		return c.JSON(http.StatusNotFound, Error{Error: err.Error()})
	}

	tx, err := s.bc.GetTxByHash(types.HashFromBytes(b))
	if err != nil {
		return c.JSON(http.StatusNotFound, Error{Error: err.Error()})
	}
	return c.JSON(http.StatusOK, tx)
}

func (s *Server) handlePostTx(c echo.Context) error {
	tx := core.NewTransaction(nil)
	if err := gob.NewDecoder(c.Request().Body).Decode(tx); err != nil {
		return c.JSON(http.StatusBadRequest, Error{Error: err.Error()})
	}

	s.txChan <- tx
	return c.JSON(http.StatusOK, tx)
}

func intoJSONBlock(b *core.Block) Block {
	txCount := len(b.Transactions)
	txReponse := TxResponse{
		TxCount: uint(txCount),
		Hashes:  make([]string, txCount),
	}

	for i := 0; i < txCount; i++ {
		txReponse.Hashes[i] = b.Transactions[i].Hash(core.TxHasher{}).String()
	}

	return Block{
		Version:       b.Header.Version,
		DataHash:      b.DataHash.String(),
		PrevBlockHash: b.PrevBlockHash.String(),
		Height:        b.Height,
		Timestamp:     b.Header.Timestamp,
		Validator:     b.Validator.Address().String(),
		Signature:     b.Signature.String(),
		Hash:          b.Hash(core.BlockHasher{}).String(),
		Transactions:  []TxResponse{txReponse},
	}
}
