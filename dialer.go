package p2pgrpc

import (
	"context"
	"github.com/libp2p/go-libp2p-core/peer"
	ps "github.com/libp2p/go-libp2p-peerstore"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"net"
)

// GetDialOption returns the WithDialer option to dial via libp2p.
// note: ctx should be the root context.
func (p *GRPCProtocol) GetDialOption() grpc.DialOption {
	return grpc.WithContextDialer(func(ctx context.Context, peerIdStr string) (net.Conn, error) {
		id, err := peer.Decode(peerIdStr)
		if err != nil {
			return nil, errors.WithMessage(err, "grpc tried to dial non peer-id")
		}

		err = p.host.Connect(ctx, ps.PeerInfo{
			ID: id,
		})
		if err != nil {
			return nil, err
		}

		stream, err := p.host.NewStream(ctx, id, Protocol)
		if err != nil {
			return nil, err
		}

		return &streamConn{Stream: stream}, nil
	})
}

// Dial attempts to open a GRPC connection over libp2p to a peer.
// Note that the context is used as the **stream context** not just the dial context.
func (p *GRPCProtocol) Dial(
	ctx context.Context,
	peerID peer.ID,
	dialOpts ...grpc.DialOption,
) (*grpc.ClientConn, error) {
	dialOpsPrepended := append([]grpc.DialOption{p.GetDialOption(ctx)}, dialOpts...)
	return grpc.DialContext(ctx, peerID.Pretty(), dialOpsPrepended...)
}
