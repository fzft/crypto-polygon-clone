package core

type Storage interface {
	Put(*Block) error
}


type MemoryStore struct {

}

func (s *MemoryStore) Put(block *Block) error {
	return nil
}