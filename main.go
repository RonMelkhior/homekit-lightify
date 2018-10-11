package main

import (
	"log"
	"os"

	"github.com/brutella/hc"

	"github.com/RonMelkhior/homekit-lightify/lightify"
	"github.com/brutella/hc/accessory"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if err := lightify.Init(); err != nil {
		log.Fatal(err)
	}

	devices, err := lightify.GetDevices()
	if err != nil {
		log.Fatal(err)
	}

	var accessories []*accessory.Accessory

	for _, device := range devices {
		if device.Type == "GATEWAY" {
			continue
		}

		accessories = append(accessories, device.InitializeAccessory())
	}

	config := hc.Config{
		Port:        os.Getenv("HOMEKIT_PORT"),
		Pin:         os.Getenv("HOMEKIT_PIN"),
		StoragePath: "./db",
	}

	t, err := hc.NewIPTransport(config, accessories[0], accessories[1:]...)
	if err != nil {
		log.Fatal(err)
	}

	hc.OnTermination(func() {
		<-t.Stop()
	})

	t.Start()
}
