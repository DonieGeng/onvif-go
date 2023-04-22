package main

import (
	"encoding/json"
	"fmt"
	Onvif "github.com/DonieGeng/onvif-go"
	"github.com/DonieGeng/onvif-go/device"
	"github.com/DonieGeng/onvif-go/media"
)

func Print(v any) {
	b, _ := json.MarshalIndent(v, "", "	")
	fmt.Println(b)
}
func main() {

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

	hostnameResp, err := device.NewDevice(dev.GetEndpoint("device")).GetHostname(&device.GetHostname{})
	if err != nil {
		fmt.Println(err)
	} else {
		Print(hostnameResp)
	}

	profilesResp, err := media.NewMedia(dev.GetEndpoint("media")).GetProfiles(&media.GetProfiles{})
	if err != nil {
		fmt.Println(err)
	} else {
		Print(profilesResp)
	}

}
