package tajarin

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// initialize server with n value

// wait for n subscriber

// combine data

// broadcast message to subscribers

type TajarinProducer struct {
	maxNodes int
}

func NewTajarinProducer(maxNodes int) TajarinProducer {
	return TajarinProducer{
		maxNodes,
	}
}

func (tp *TajarinProducer) ListenAndWait() {
	fmt.Println(tp.maxNodes)
	logger, _ := zap.NewProduction()
	tcpListener := NewTCPListener(
		DefaultListenAddress,
		logger,
		int64(tp.maxNodes),
	)
	tcpListener.Serve(context.Background())
}
