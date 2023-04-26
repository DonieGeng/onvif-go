package wsdiscovery

import (
	"bytes"
	"encoding/xml"
	"github.com/hooklift/gowsdl/soap"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/net/ipv4"
	"log"
	"net"
	"time"
)

const (
	bufSzie                        = 8192
	discoveryTimeout time.Duration = 900
	waitTimeout                    = discoveryTimeout + 450
)

type Action struct {
	XMLName xml.Name `xml:"a:Action"`
	Text    string   `xml:",chardata"`
}

type MessageID struct {
	XMLName xml.Name `xml:"a:MessageID"`
	Text    string   `xml:",chardata"`
}

type ReplyTo struct {
	XMLName xml.Name `xml:"a:ReplyTo"`
	Address string   `xml:"a:Address"`
}

type To struct {
	XMLName xml.Name `xml:"a:MessageID"`
	Text    string   `xml:",chardata"`
}

type Probe struct {
	XMLName xml.Name `xml:"Probe"`
	Text    string   `xml:",chardata"`
	Xmlns   string   `xml:"xmlns,attr"`
	Types   struct {
		Text    string `xml:",chardata"`
		XmlnsD  string `xml:"xmlns:d,attr"`
		XmlnsDn string `xml:"xmlns:dn,attr"`
	} `xml:"d:Types"`
}

type EndpointReferenceType struct {
	XMLName xml.Name `xml:"EndpointReference"`
	Address string   `xml:"Address"`
}

type ProbeMatchType struct {
	XMLName           xml.Name              `xml:"ProbeMatch"`
	EndpointReference EndpointReferenceType `xml:"EndpointReference"`
	Types             string                `xml:"Types"`
	Scopes            string                `xml:"Scopes"`
	XAddrs            string                `xml:"XAddrs"`
	MetadataVersion   int                   `xml:"MetadataVersion"`
}

type ProbeMatches struct {
	XMLName    xml.Name         `xml:"ProbeMatches"`
	ProbeMatch []ProbeMatchType `xml:"ProbeMatch"`
	Host       string
}

func BuildProbeMessage(types []string) *bytes.Buffer {
	action := Action{
		Text: "http://schemas.xmlsoap.org/ws/2005/04/discovery/Probe",
	}
	messageid := MessageID{
		Text: "uuid:" + uuid.Must(uuid.NewV4(), nil).String(),
	}
	replyto := ReplyTo{
		Address: "http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous",
	}
	to := To{
		Text: "urn:schemas-xmlsoap:ws:2005:04:discovery",
	}

	header := soap.SOAPHeader{Headers: []interface{}{action, messageid, replyto, to}}

	var typeString string
	for _, j := range types {
		typeString += j
		typeString += " "
	}

	probe := Probe{
		Xmlns: "http://schemas.xmlsoap.org/ws/2005/04/discovery",
		Types: struct {
			Text    string `xml:",chardata"`
			XmlnsD  string `xml:"xmlns:d,attr"`
			XmlnsDn string `xml:"xmlns:dn,attr"`
		}{
			Text:    typeString,
			XmlnsD:  "http://schemas.xmlsoap.org/ws/2005/04/discovery",
			XmlnsDn: "http://www.onvif.org/ver10/network/wsdl",
		},
	}
	body := soap.SOAPBody{Content: probe}
	probeMessage := soap.SOAPEnvelope{
		XmlNS:  "http://www.w3.org/ws/2003/05/soap-envelope",
		Header: &header,
		Body:   body,
	}

	message := new(bytes.Buffer)
	buffer := new(bytes.Buffer)
	var encoder soap.SOAPEncoder
	encoder = xml.NewEncoder(buffer)
	if err := encoder.Encode(probeMessage); err != nil {
		return nil
	}
	message.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>")
	message.Write(buffer.Bytes())
	return message
}

func SendUDPUnicast(msg *bytes.Buffer, ip string) map[string]*bytes.Buffer {
	socket, err := net.ListenUDP("udp", nil)
	if err != nil {
		return nil
	}
	defer socket.Close()

	addr := net.UDPAddr{IP: net.ParseIP(ip), Port: 3702}
	_, err = socket.WriteTo(msg.Bytes(), &addr)
	if err != nil {
		return nil
	}

	message := make(chan map[string]*bytes.Buffer)
	done := make(chan bool)
	defer close(message)
	go func() {
		b := make([]byte, bufSzie)
		for {
			select {
			case <-done:
				return
			default:
				n, src, err := socket.ReadFromUDP(b)
				if err != nil {
					message <- nil
					return
				}
				buf := new(bytes.Buffer)
				buf.Write(b[0:n])
				var result = make(map[string]*bytes.Buffer)
				result[src.String()] = buf
				message <- result
				return
			}
		}
	}()

	for {
		select {
		case i := <-message:
			return i
		case <-time.After(time.Millisecond * waitTimeout):
			close(done)
			return nil
		}
	}
}

func SendUDPMulticast(msg *bytes.Buffer, interfaceName string) map[string]*bytes.Buffer {
	var result = make(map[string]*bytes.Buffer)
	data := msg.Bytes()
	iface, err := net.InterfaceByName(interfaceName)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	group := net.IPv4(239, 255, 255, 250)
	broad := net.IPv4(255, 255, 255, 255)

	interfaceIps, _ := iface.Addrs()
	for _, tip := range interfaceIps {
		var ip net.IP
		switch v := tip.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		}
		ip = ip.To4()
		if ip == nil {
			continue //not an ipv4 address
		}

		c, err := net.ListenPacket("udp", ip.String()+":1024")
		if err != nil {
			log.Fatal(err)
			continue
		}
		defer c.Close()

		p := ipv4.NewPacketConn(c)
		if err := p.JoinGroup(iface, &net.UDPAddr{IP: group}); err != nil {
			log.Fatal(err)
			continue
		}
		dst := &net.UDPAddr{IP: group, Port: 3702}
		dst_b := &net.UDPAddr{IP: broad, Port: 3702}
		for _, ifi := range []*net.Interface{iface} {
			if err := p.SetMulticastInterface(ifi); err != nil {
				log.Fatal(err)
				continue
			}
			p.SetMulticastTTL(2)
			//2次组播, 增加搜索成功率
			for i := 0; i < 2; i++ {
				if _, err := p.WriteTo(data, nil, dst); err != nil {
					log.Fatal(err)
				}
			}
			if _, err := p.WriteTo(data, nil, dst_b); err != nil {
				log.Fatal(err)
			}
		}

		if err := p.SetReadDeadline(time.Now().Add(time.Second * 1)); err != nil {
			log.Fatal(err)
		}

		for {
			//Todo 优化
			b := make([]byte, bufSzie)
			n, _, src, err := p.ReadFrom(b)
			if err != nil {
				log.Fatal(err)
				break
			}
			buf := new(bytes.Buffer)
			buf.Write(b[0:n])
			result[src.String()] = buf
		}
	}
	return result
}

type wsSOAPEnvelopeResponse struct {
	XMLName xml.Name `xml:"soap:Envelope"`
	XmlNS   string   `xml:"xmlns:soap,attr"`
	Header  *soap.SOAPHeaderResponse
	Body    soap.SOAPBodyResponse
}

func ParseProbeResp(ip string, result *bytes.Buffer) interface{} {
	response := new(ProbeMatches)
	response.Host = ip
	response.ProbeMatch = make([]ProbeMatchType, 0)
	respEnvelope := new(wsSOAPEnvelopeResponse)
	respEnvelope.XmlNS = "http://www.w3.org/2003/05/soap-envelope"
	respEnvelope.Body = soap.SOAPBodyResponse{
		Content: response,
	}

	var dec soap.SOAPDecoder
	dec = xml.NewDecoder(result)
	if err := dec.Decode(respEnvelope); err != nil {
		log.Fatal(err)
		return nil
	}
	return respEnvelope.Body.Content
}
