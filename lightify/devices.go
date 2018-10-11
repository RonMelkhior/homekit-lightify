package lightify

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/brutella/hc/accessory"
	colorful "github.com/lucasb-eyer/go-colorful"
)

type DeviceHSV struct {
	Hue        float64 `json:"hue"`
	Saturation float64 `json:"saturation"`
	Brightness float64 `json:"brightness"`
}

type DeviceRGB struct {
	Red   uint8 `json:"red"`
	Green uint8 `json:"green"`
	Blue  uint8 `json:"blue"`
	White uint8 `json:"white"`
}

type Device struct {
	Lightbulb        *accessory.Lightbulb
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Type             string    `json:"type"`
	On               string    `json:"onOff"`
	Online           bool      `json:"online"`
	Brightness       int32     `json:"brightness"`
	ColorTemperature int32     `json:"colorTemperature"`
	HSV              DeviceHSV `json:"colorHSV"`
	RGB              DeviceRGB `json:"colorRGBW"`
	Mode             string    `json:"mode"`
	DeviceModel      string    `json:"deviceModel"`
	FirmwareVersion  string    `json:"string"`
}

type DeviceUpdateParams map[string]interface{}

func GetDevices() ([]*Device, error) {
	client := LightifyConfig.Client(context.TODO(), GetToken())
	resp, err := client.Get("https://emea.lightify-api.com/v4/devices/")
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var devices struct {
		Devices []*Device `json:"devices"`
	}

	if err = json.NewDecoder(resp.Body).Decode(&devices); err != nil {
		return nil, err
	}

	return devices.Devices, nil
}

func (d *Device) InitializeAccessory() *accessory.Accessory {
	d.Lightbulb = accessory.NewLightbulb(accessory.Info{
		Name:         d.Name,
		SerialNumber: d.ID,
		Manufacturer: "OSRAM Lightify",
		Model:        d.DeviceModel,
	})

	d.Lightbulb.Lightbulb.On.OnValueRemoteUpdate(func(on bool) {
		go d.ToggleDevice(on)
	})

	d.Lightbulb.Lightbulb.Hue.OnValueRemoteUpdate(func(v float64) {
		go d.UpdateColor()
	})

	d.Lightbulb.Lightbulb.Saturation.OnValueRemoteUpdate(func(v float64) {
		go d.UpdateColor()
	})

	d.Lightbulb.Lightbulb.Brightness.OnValueRemoteUpdate(func(v int) {
		go d.SetBrightness(int32(v))
	})

	return d.Lightbulb.Accessory
}

func (d *Device) UpdateDevice(params DeviceUpdateParams) (*Device, error) {
	jsonPayload, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PATCH", "https://emea.lightify-api.com/v4/devices/"+d.ID, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := LightifyConfig.Client(context.TODO(), GetToken())
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	device := new(Device)

	if err = json.NewDecoder(resp.Body).Decode(device); err != nil {
		return nil, err
	}

	return device, nil
}

func (d *Device) ToggleDevice(on bool) error {
	var onOff string
	if on {
		onOff = "on"
	} else {
		onOff = "off"
	}

	device, err := d.UpdateDevice(DeviceUpdateParams{
		"onOff": onOff,
	})
	if err != nil {
		return err
	}

	d.On = device.On

	return nil
}

func (d *Device) UpdateColor() error {
	R, G, B := colorful.Hsv(
		d.Lightbulb.Lightbulb.Hue.GetValue(),
		d.Lightbulb.Lightbulb.Saturation.GetValue()/100,
		float64(d.Lightbulb.Lightbulb.Brightness.GetValue())/100,
	).RGB255()

	return d.SetRGB(DeviceRGB{R, G, B, 255})
}

func (d *Device) SetRGB(rgb DeviceRGB) error {
	device, err := d.UpdateDevice(DeviceUpdateParams{
		"colorRGBW": rgb,
	})
	if err != nil {
		return err
	}

	d.RGB = device.RGB

	return nil
}

func (d *Device) SetBrightness(brightness int32) error {
	device, err := d.UpdateDevice(DeviceUpdateParams{
		"brightness": brightness,
	})
	if err != nil {
		return err
	}

	d.Brightness = device.Brightness

	return nil
}
