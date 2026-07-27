package notify

import (
	"context"
	"sync"

	sharedtypes "crona/shared/types"
)

type companionDeliveryResult struct {
	notificationAccepted bool
	soundAccepted        bool
}

type companionSubscriber struct {
	clientID     string
	capabilities sharedtypes.AlertDeliveryCapability
	deliveries   chan sharedtypes.AlertDelivery
	done         chan struct{}
	closeOnce    sync.Once
}

func (s *companionSubscriber) close() {
	s.closeOnce.Do(func() {
		close(s.done)
		close(s.deliveries)
	})
}

type companionBroker struct {
	mu      sync.Mutex
	active  *companionSubscriber
	pending map[string]pendingCompanionDelivery
}

type pendingCompanionDelivery struct {
	result              chan companionDeliveryResult
	notificationOffered bool
	soundOffered        bool
}

func newCompanionBroker() *companionBroker {
	return &companionBroker{pending: make(map[string]pendingCompanionDelivery)}
}

func (b *companionBroker) subscribe(
	ctx context.Context,
	capabilities sharedtypes.AlertDeliveryCapability,
) (<-chan sharedtypes.AlertDelivery, func()) {
	subscriber := &companionSubscriber{
		clientID:     capabilities.ClientID,
		capabilities: capabilities,
		deliveries:   make(chan sharedtypes.AlertDelivery, 16),
		done:         make(chan struct{}),
	}

	b.mu.Lock()
	previous := b.active
	b.active = subscriber
	b.mu.Unlock()
	if previous != nil {
		previous.close()
	}

	var once sync.Once
	unsubscribe := func() {
		once.Do(func() {
			b.mu.Lock()
			if b.active == subscriber {
				b.active = nil
			}
			b.mu.Unlock()
			subscriber.close()
		})
	}
	go func() {
		<-ctx.Done()
		unsubscribe()
	}()
	return subscriber.deliveries, unsubscribe
}

func (b *companionBroker) isActive() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.active != nil
}

func (b *companionBroker) offer(
	delivery sharedtypes.AlertDelivery,
) (<-chan companionDeliveryResult, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.active == nil {
		return nil, false
	}
	delivery.DeliverNotification =
		delivery.DeliverNotification && b.active.capabilities.Notifications
	delivery.PlaySound = delivery.PlaySound && b.active.capabilities.Sounds
	if !delivery.DeliverNotification && !delivery.PlaySound {
		return nil, false
	}
	result := make(chan companionDeliveryResult, 1)
	select {
	case b.active.deliveries <- delivery:
		b.pending[delivery.ID] = pendingCompanionDelivery{
			result:              result,
			notificationOffered: delivery.DeliverNotification,
			soundOffered:        delivery.PlaySound,
		}
		return result, true
	default:
		return nil, false
	}
}

func (b *companionBroker) acknowledge(ack sharedtypes.AlertDeliveryAck) bool {
	b.mu.Lock()
	pending, ok := b.pending[ack.DeliveryID]
	if ok {
		delete(b.pending, ack.DeliveryID)
	}
	b.mu.Unlock()
	if !ok {
		return false
	}
	pending.result <- companionDeliveryResult{
		notificationAccepted: pending.notificationOffered && ack.NotificationAccepted,
		soundAccepted:        pending.soundOffered && ack.SoundAccepted,
	}
	return true
}

func (b *companionBroker) abandon(deliveryID string) {
	b.mu.Lock()
	delete(b.pending, deliveryID)
	b.mu.Unlock()
}
