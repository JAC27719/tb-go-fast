package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	//"strconv"

	"solace.dev/go/messaging"
	"solace.dev/go/messaging/pkg/solace"
	"solace.dev/go/messaging/pkg/solace/config"
	"solace.dev/go/messaging/pkg/solace/resource"
)

// Define Topic Prefix
const TopicPrefix = "solace/weather"

type TemperatureSensor struct {
	T float32 //Temp in degrees C
	H float32 //Humidity in percentage
}

type GenericMessageHandler struct {
	directPublisher  solace.DirectMessagePublisher
	messagingService solace.MessagingService
	msgSeqNum        int
}

func (h *GenericMessageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sensor := new(TemperatureSensor)
	sensor.T = 15
	sensor.H = 35
	messageBody, err := json.Marshal(sensor)
	if err != nil {
		panic(err)
	}
	messageBuilder := h.messagingService.MessageBuilder().
		WithProperty("application", "go_pub").
		WithProperty("language", "go")
	message, err := messageBuilder.BuildWithByteArrayPayload(messageBody)
	if err != nil {
		panic(err)
	}

	topic := resource.TopicOf(TopicPrefix)

	// Publish on dynamic topic with dynamic body
	publishErr := h.directPublisher.Publish(message, topic)
	if publishErr != nil {
		panic(publishErr)
	}

	fmt.Fprintf(w, "Hello from %v:%v", SERVER_HOST, SERVER_PORT)
	h.msgSeqNum++
}

func ReconnectionHandler(e solace.ServiceEvent) {
	e.GetTimestamp()
	e.GetBrokerURI()
	err := e.GetCause()
	if err != nil {
		fmt.Println(err)
	}
}

func getEnv(key, def string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return def
}

var SERVER_HOST = os.Getenv("SERVER_HOST")
var SERVER_PORT = os.Getenv("SERVER_PORT")

func main() {

	// Configuration parameters
	brokerConfig := config.ServicePropertyMap{
		config.TransportLayerPropertyHost:                getEnv("SOLACE_HOST", "tcp://solace:55555"),
		config.ServicePropertyVPNName:                    getEnv("SOLACE_VPN", "default"),
		config.AuthenticationPropertySchemeBasicPassword: getEnv("SOLACE_PASSWORD", "admin"),
		config.AuthenticationPropertySchemeBasicUserName: getEnv("SOLACE_USERNAME", "admin"),
	}

	messagingService, err := messaging.NewMessagingServiceBuilder().FromConfigurationProvider(brokerConfig).Build()

	if err != nil {
		panic(err)
	}

	// Connect to the messaging serice
	numberOfRetries := 1

	i := 1
	for i <= numberOfRetries {
		fmt.Printf("Attempting to connect to server. Attempt %v/%v\n", i, numberOfRetries)
		if err := messagingService.Connect(); err != nil {
			if i == numberOfRetries {
				fmt.Printf("Failed to connect after %v retires. Exiting the program!\n", numberOfRetries)
				panic(err)
			}
			fmt.Println(err)
			time.Sleep(time.Duration(i) * time.Second * 2)
			i++
			continue
		}
		break
	}

	fmt.Println("Connected to the broker? ", messagingService.IsConnected())

	//  Build a Direct Message Publisher
	directPublisher, builderErr := messagingService.CreateDirectMessagePublisherBuilder().Build()
	if builderErr != nil {
		panic(builderErr)
	}

	startErr := directPublisher.Start()
	if startErr != nil {
		panic(startErr)
	}

	fmt.Println("Direct Publisher running? ", directPublisher.IsRunning())

	fmt.Println("\n===Interrupt (CTR+C) to stop publishing===\n")

	//  Prepare outbound message payload and body

	if SERVER_HOST == "" {
		fmt.Println("SERVER_HOST is not set")
	}

	if SERVER_PORT == "" {
		fmt.Println("SERVER_PORT is not set")
	}

	handler := new(GenericMessageHandler)
	handler.directPublisher = directPublisher
	handler.messagingService = messagingService
	http.Handle("/", handler)

	fmt.Printf("Server listening on %v:%v...\n", SERVER_HOST, SERVER_PORT)
	http.ListenAndServe(":"+SERVER_PORT, nil)
}
