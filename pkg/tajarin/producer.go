package tajarin

import (
	"context"

	"github.com/gnolang/tajarin/pkg/tcp"
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
	tcpListener := tcp.NewTCPListener(
		logger,
		tp.listenAddress,
		tp.maxNodes,
	)
	tcpListener.Serve(ctx)
}
