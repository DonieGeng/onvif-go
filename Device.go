package onvif_go

import (
	"encoding/json"
	"errors"
	"github.com/DonieGeng/onvif-go/device"
	"github.com/hooklift/gowsdl/soap"
	"net/url"
	"strings"
)

type Device struct {
	params    DeviceParams
	endpoints map[string]*soap.Client
	info      DeviceInfo
}

type DeviceParams struct {
	Xaddr    string
	Username string
	Password string
}

type DeviceInfo struct {
	Manufacturer    string
	Model           string
	FirmwareVersion string
	SerialNumber    string
	HardwareId      string
}

// NewDevice function construct a ONVIF Device entity
func NewDevice(params DeviceParams) (*Device, error) {
	dev := new(Device)
	dev.params = params
	dev.endpoints = make(map[string]*soap.Client)

	client := soap.NewClient("http://" + dev.params.Xaddr + "/onvif/device_service")
	device_service := device.NewDevice(client)

	servicesResp, err := device_service.GetServices(&device.GetServices{})
	if err != nil {
		return nil, errors.New("camera is not available at " + dev.params.Xaddr + " or it does not support ONVIF services")
	}

	dev.getSupportedServices(servicesResp)

	infoResp, err := device_service.GetDeviceInformation(&device.GetDeviceInformation{})
	if err != nil {
		return nil, errors.New("camera is not available at " + dev.params.Xaddr + " or it does not support ONVIF services")
	}
	data, _ := json.Marshal(infoResp)
	err = json.Unmarshal(data, &dev.info)
	if err != nil {
		return nil, err
	}

	return dev, nil

}

// GetServices return available endpoints
func (dev *Device) GetEndpoint(name string) *soap.Client {
	endpoint, _ := dev.getEndpoint(name)
	return endpoint
}

// GetServices return available endpoints
func (dev *Device) GetDeviceInfo() DeviceInfo {
	return dev.info
}

func (dev *Device) getSupportedServices(resp *device.GetServicesResponse) {

	for _, j := range resp.Service {
		addr, _ := url.Parse(string(j.XAddr))
		service := strings.TrimPrefix(addr.Path, "/onvif/")
		if service != "" {
			if service == "device_service" {
				service = "device"
			}

			dev.addEndpoint(service, string(j.XAddr))
		}

	}
}

func (dev *Device) addEndpoint(Key, Value string) {
	//use lowCaseKey
	//make key having ability to handle Mixed Case for Different vendor devcie (e.g. Events EVENTS, events)
	lowCaseKey := strings.ToLower(Key)

	// Replace host with host from device params.
	if u, err := url.Parse(Value); err == nil {
		u.Host = dev.params.Xaddr
		Value = u.String()
	}

	client := soap.NewClient(Value, soap.WithBasicAuth(dev.params.Username, dev.params.Password))

	//switch lowCaseKey {
	//case "device":
	//	service := device.NewDevice(client)
	//}
	dev.endpoints[lowCaseKey] = client
}

// getEndpoint functions get the target service endpoint in a better way
func (dev Device) getEndpoint(endpoint string) (*soap.Client, error) {

	// common condition, endpointMark in map we use this.
	if endpointClient, bFound := dev.endpoints[endpoint]; bFound {
		return endpointClient, nil
	}

	//but ,if we have endpoint like event\analytic
	//and sametime the Targetkey like : events\analytics
	//we use fuzzy way to find the best match url
	var endpointClient *soap.Client
	for targetKey := range dev.endpoints {
		if strings.Contains(targetKey, endpoint) {
			endpointClient = dev.endpoints[targetKey]
			return endpointClient, nil
		}
	}
	return endpointClient, errors.New("target endpoint service not found")
}
