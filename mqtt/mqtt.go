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

// RequestEstopPayload defines the structure for the freezy/eStopState topic payload.
type RequestEstopPayload struct {
	Channel int  `json:"channel"`
	State   bool `json:"state"`
}

// TopicHandler defines a function type for handling MQTT messages for a specific topic.
type TopicHandler func(arena *field.Arena, payload json.RawMessage) error

// Broker holds the MQTT server instance and arena reference.
type Broker struct {
	server *mqtt.Server
	arena  *field.Arena
}

// loggingHook is a custom hook to log messages and handle specific topics.
type loggingHook struct {
	mqtt.HookBase
	arena   *field.Arena
	handlers map[string]TopicHandler
}

// ID returns the hook identifier.
func (h *loggingHook) ID() string {
	return "logging-hook"
}

// Provides indicates which hook methods are implemented.
func (h *loggingHook) Provides(b byte) bool {
	return b == mqtt.OnPublish
}

// OnPublish logs messages and processes specific topics.
func (h *loggingHook) OnPublish(cl *mqtt.Client, pk packets.Packet) (packets.Packet, error) {
	log.Printf("Received message on topic '%s': %s", pk.TopicName, string(pk.Payload))

	// Look up handler for the topic
	if handler, exists := h.handlers[pk.TopicName]; exists && h.arena != nil {
		if err := handler(h.arena, pk.Payload); err != nil {
			log.Printf("Failed to process topic '%s': %v", pk.TopicName, err)
		} else {
			log.Printf("Successfully processed topic '%s'", pk.TopicName)
		}
	}

	return pk, nil
}

// NewBroker creates and configures a new MQTT broker with arena reference.
func NewBroker(arena *field.Arena) *Broker {
	// Create a new MQTT server
	server := mqtt.New(nil)

	// Allow all connections (no authentication for simplicity)
	_ = server.AddHook(new(auth.AllowHook), nil)

	// Define topic handlers
	handlers := map[string]TopicHandler{
		"freezy/eStopState": func(arena *field.Arena, payload json.RawMessage) error {
			var eStopPayload []RequestEstopPayload
			if err := json.Unmarshal(payload, &eStopPayload); err != nil {
				return err
			}
			for _, item := range eStopPayload {
				arena.Plc.SetAlternateIOStopState(item.Channel, item.State)
			}
			log.Printf("eStop state updated successfully for %d channels", len(eStopPayload))
			return nil
		},
		// TEMPLATE: Add new topic handlers here
		// "<topic_name>": func(arena *field.Arena, payload json.RawMessage) error {
		//     // Define your payload structure if needed
		//     type CustomPayload struct {
		//         Field string `json:"field"`
		//         // Add other fields
		//     }
		//     var customPayload CustomPayload
		//     if err := json.Unmarshal(payload, &customPayload); err != nil {
		//         return err
		//     }
		//     // Add logic to process payload, e.g., call arena methods
		//     log.Printf("Processed topic '%s' with payload: %v", "<topic_name>", customPayload)
		//     return nil
		// },
	}

	// Add custom logging hook with arena and handlers
	_ = server.AddHook(&loggingHook{arena: arena, handlers: handlers}, nil)

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