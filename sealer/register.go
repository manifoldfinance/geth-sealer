package sealer

import (
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/rpc"
)

func Register(stack *node.Node, backend *eth.Ethereum, cfg *SealerConfig) error {

	sealerService := newSealer(backend)

	stack.RegisterAPIs([]rpc.API{
		{
			Namespace:     "sealer",
			Version:       "1.0",
			Service:       sealerService,
			Public:        true,
			Authenticated: !cfg.Insecure,
		},
	})
	return nil
}
