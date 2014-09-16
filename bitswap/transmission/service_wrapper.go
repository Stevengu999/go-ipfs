package transmission

import (
	context "github.com/jbenet/go-ipfs/Godeps/_workspace/src/code.google.com/p/go.net/context"

	bsmsg "github.com/jbenet/go-ipfs/bitswap/message"
	inet "github.com/jbenet/go-ipfs/net"
	netmsg "github.com/jbenet/go-ipfs/net/message"
	netservice "github.com/jbenet/go-ipfs/net/service"
	peer "github.com/jbenet/go-ipfs/peer"
)

// NewServiceWrapper handles protobuf marshalling
func NewServiceWrapper(ctx context.Context, r Receiver) Sender {
	h := &handlerWrapper{r}
	s := netservice.NewService(h)
	s.Start(ctx)
	return &senderWrapper{s}
}

// handlerWrapper is responsible for marshaling/unmarshaling NetMessages. It
// delegates calls to the BitSwap delegate.
type handlerWrapper struct {
	bitswapDelegate Receiver
}

// HandleMessage marshals and unmarshals net messages, forwarding them to the
// BitSwapMessage receiver
func (wrapper *handlerWrapper) HandleMessage(
	ctx context.Context, incoming netmsg.NetMessage) (netmsg.NetMessage, error) {

	received, err := bsmsg.FromNet(incoming)
	if err != nil {
		return nil, err
	}

	bsmsg, p, err := wrapper.bitswapDelegate.ReceiveMessage(ctx, incoming.Peer(), received)
	if err != nil {
		return nil, err
	}
	if bsmsg == nil {
		return nil, nil
	}

	outgoing, err := bsmsg.ToNet(p)
	if err != nil {
		return nil, err
	}

	return outgoing, nil
}

type senderWrapper struct {
	serviceDelegate inet.Sender
}

func (wrapper *senderWrapper) SendMessage(
	ctx context.Context, p *peer.Peer, outgoing bsmsg.Exportable) error {
	nmsg, err := outgoing.ToNet(p)
	if err != nil {
		return err
	}
	return wrapper.serviceDelegate.SendMessage(ctx, nmsg)
}

func (wrapper *senderWrapper) SendRequest(ctx context.Context,
	p *peer.Peer, outgoing bsmsg.Exportable) (bsmsg.BitSwapMessage, error) {

	outgoingMsg, err := outgoing.ToNet(p)
	if err != nil {
		return nil, err
	}
	incomingMsg, err := wrapper.serviceDelegate.SendRequest(ctx, outgoingMsg)
	if err != nil {
		return nil, err
	}
	return bsmsg.FromNet(incomingMsg)
}
