package ethclient

import (
	"context"
	"testing"
)

func TestClient_ChainID(t *testing.T) {
	client, err := Dial("http://locahost:8545")
	if err != nil {
		t.Errorf("error:%v", err)
		return
	}

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		t.Errorf("error:%v", err)
		return
	}

	t.Logf("chainID:%v", chainID)
}
