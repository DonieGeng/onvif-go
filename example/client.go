package main

import (
	"encoding/json"
	"fmt"
	Onvif "github.com/DonieGeng/onvif-go"
	"github.com/DonieGeng/onvif-go/device"
	"github.com/DonieGeng/onvif-go/media"
	"github.com/DonieGeng/onvif-go/media2"
	wsdiscovery "github.com/DonieGeng/onvif-go/network"
)

func Print(v any) {
	b, _ := json.MarshalIndent(v, "", "	")
	fmt.Println(string(b))
}
func main() {

	devtype := []string{"dn:NetworkVideoTransmitter"}
	message := wsdiscovery.BuildProbeMessage(devtype)
	results := wsdiscovery.SendUDPUnicast(message, "eth1")
	for ip, result := range results {
		Print(wsdiscovery.ParseProbeResp(ip, result))
	}

	//Getting an camera instance
	dev, err := Onvif.NewDevice(Onvif.DeviceParams{
		Xaddr:    "192.168.13.14:80",
		Username: "admin",
		Password: "Admin1234",
	})
	if err != nil {
		panic(err)
	}

	info := dev.GetDeviceInfo()
	Print(info)

	service, err := dev.GetService("device")
	if err != nil {
		fmt.Println(err)
		return
	}
	hostnameResp, err := device.NewDevice(service).GetHostname(&device.GetHostname{})
	if err != nil {
		fmt.Println(err)
	} else {
		Print(hostnameResp)
	}

	service, err = dev.GetService("media")
	if err != nil {
		fmt.Println(err)
		return
	}
	profilesResp, err := media.NewMedia(service).GetProfiles(&media.GetProfiles{})
	if err != nil {
		fmt.Println(err)
	} else {
		Print(profilesResp)
	}

	service, err = dev.GetService("media2")
	if err != nil {
		fmt.Println(err)
		return
	}
	encoderConfigurationsResp, err := media2.NewMedia2(service).GetVideoEncoderConfigurations(&media2.GetVideoEncoderConfigurations{})
	if err != nil {
		fmt.Println(err)
	} else {
		Print(encoderConfigurationsResp)
	}

}
