package tajarin

import (
	"context"

	"go.uber.org/zap"
)

type TajarinProducer struct {
	maxNodes      int64
	listenAddress string
}

func NewTajarinProducer(maxNodes int64, listenAddress string) TajarinProducer {
	return TajarinProducer{
		maxNodes:      maxNodes,
		listenAddress: listenAddress,
	}
}

func (tp *TajarinProducer) ListenAndWait(ctx context.Context, logger *zap.Logger) {
	tcpListener := NewTCPListener(
		logger,
		tp.listenAddress,
		tp.maxNodes,
	)
	tcpListener.Serve(ctx)
}
