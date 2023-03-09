package p2p

type StatusMessage struct {
	CurrentHeight uint32
	ID            string
}

type GetStatusMessage struct {
}
