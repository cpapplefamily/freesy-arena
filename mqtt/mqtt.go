package mqtt

import (
	"encoding/json"
	"github.com/Team254/cheesy-arena/field"
	"github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
	"github.com/mochi-mqtt/server/v2/packets"
	"log"
)

// RequestPayload defines the structure of the eStop state payload.
type RequestPayload struct {
	Channel int  `json:"channel"`
	State   bool `json:"state"`
}

// Broker holds the MQTT server instance and arena reference.
type Broker struct {
	server *mqtt.Server
	arena  *field.Arena
}

// loggingHook is a custom hook to log messages and handle eStop state updates.
type loggingHook struct {
	mqtt.HookBase
	arena *field.Arena
}

// ID returns the hook identifier.
func (h *loggingHook) ID() string {
	return "logging-hook"
}

// Provides indicates which hook methods are implemented.
func (h *loggingHook) Provides(b byte) bool {
	return b == mqtt.OnPublish
}

// OnPublish logs messages and processes eStop state updates for freezy/eStopState.
func (h *loggingHook) OnPublish(cl *mqtt.Client, pk packets.Packet) (packets.Packet, error) {
	log.Printf("Received message on topic '%s': %s", pk.TopicName, string(pk.Payload))

	if pk.TopicName == "freezy/eStopState" && h.arena != nil {
		var payload []RequestPayload
		if err := json.Unmarshal(pk.Payload, &payload); err != nil {
			log.Printf("Failed to parse eStop payload: %v", err)
			return pk, nil // Continue processing the message
		}

		for _, item := range payload {
			h.arena.Plc.SetAlternateIOStopState(item.Channel, item.State)
		}
		log.Printf("eStop state updated successfully for %d channels", len(payload))
	}

	return pk, nil
}

// NewBroker creates and configures a new MQTT broker with arena reference.
func NewBroker(arena *field.Arena) *Broker {
	// Create a new MQTT server
	server := mqtt.New(nil)

	// Allow all connections (no authentication for simplicity)
	_ = server.AddHook(new(auth.AllowHook), nil)

	// Add custom logging hook with arena reference
	_ = server.AddHook(&loggingHook{arena: arena}, nil)

	// Create a TCP listener configuration
	tcpConfig := listeners.Config{
		ID:      "t1",
		Address: ":1883", // Default MQTT port
	}

	// Add TCP listener
	tcp := listeners.NewTCP(tcpConfig)
	if err := server.AddListener(tcp); err != nil {
		log.Fatalf("Failed to add MQTT listener: %v", err)
	}

	return &Broker{server: server, arena: arena}
}

// Start begins the MQTT broker in a goroutine.
func (b *Broker) Start() {
	go func() {
		if err := b.server.Serve(); err != nil {
			log.Fatalf("MQTT broker failed: %v", err)
		}
	}()
	log.Println("MQTT broker started on :1883")
}

// Stop shuts down the MQTT broker.
func (b *Broker) Stop() {
	log.Println("Shutting down MQTT broker")
	b.server.Close()
}