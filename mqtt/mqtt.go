package mqtt

import (
	"github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
	"github.com/mochi-mqtt/server/v2/packets"
	"log"
)

// Broker holds the MQTT server instance.
type Broker struct {
	server *mqtt.Server
}

// loggingHook is a custom hook to log received messages.
type loggingHook struct {
	mqtt.HookBase
}

// ID returns the hook identifier.
func (h *loggingHook) ID() string {
	return "logging-hook"
}

// Provides indicates which hook methods are implemented.
func (h *loggingHook) Provides(b byte) bool {
	return b == mqtt.OnPublish
}

// OnPublish logs the received MQTT message.
func (h *loggingHook) OnPublish(cl *mqtt.Client, pk packets.Packet) (packets.Packet, error) {
	log.Printf("Received message on topic '%s': %s", pk.TopicName, string(pk.Payload))
	return pk, nil
}

// NewBroker creates and configures a new MQTT broker.
func NewBroker() *Broker {
	// Create a new MQTT server
	server := mqtt.New(nil)

	// Allow all connections (no authentication for simplicity)
	_ = server.AddHook(new(auth.AllowHook), nil)

	// Add custom logging hook
	_ = server.AddHook(new(loggingHook), nil)

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

	return &Broker{server: server}
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