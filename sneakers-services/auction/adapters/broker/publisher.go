package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"hotsneakers/auction/core"

	"github.com/nats-io/nats.go"
)

type Broker struct {
	Conn *nats.Conn
}

func NewBroker(brokerURL string) (*Broker, error) {
	nc, err := nats.Connect(brokerURL)
	if err != nil {
		return nil, err
	}

	return &Broker{
		Conn: nc,
	}, nil
}

func (b Broker) PublishBidPlaced(ctx context.Context, bid core.Bid) error {
	data, err := json.Marshal(bid)
	if err != nil {
		return err
	}

	subject := fmt.Sprintf("auction.bids.%v", bid.ID)

	return b.Conn.Publish(subject, data)
}
func (b Broker) Close() error {
	b.Conn.Close()
	return nil
}
