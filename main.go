package main

import (
	"flag"
	"log"
	"strconv"
	"strings"
	"sync"
)

var network, server, serverDesc, subnet, clientListens, serverListens string
var networkCa, certificate, privateKey string

func init() {
	flag.StringVar(&network, "network", "", "Name of the IRC network to which this server belongs")
	flag.StringVar(&server, "server", "", "Name of this server")
	flag.StringVar(&serverDesc, "server_desc", "", "Description of this server")
	flag.StringVar(&subnet, "default_subnet", "", "Name of the default subnet for this server")
	flag.StringVar(&clientListens, "client_listens", "", "Comma separated set of host:port combinations for accepting client connections")
	flag.StringVar(&serverListens, "server_listens", "", "Comma separated set of host:port combinations for accepting server connections")
	flag.StringVar(&networkCa, "tls_network_ca", "", "Path to the Certificate Authority (CA) certificate for the network")
	flag.StringVar(&certificate, "tls_certificate", "", "Path to the TLS certificate for this server")
	flag.StringVar(&privateKey, "tls_private_key", "", "Path to the private key for the TLS certificate")
}

func main() {
	flag.Parse()
	if !validate() {
		return
	}

	var wg sync.WaitGroup

	ircd := NewIrcd(network, server, serverDesc, subnet, &wg)

	ircd.LoadTls(networkCa, certificate, privateKey)

	if len(clientListens) > 0 {
		listens := strings.Split(clientListens, ",")
		for _, listen := range listens {
			pieces := strings.Split(listen, ":")
			if len(pieces) < 2 {
				log.Printf("Invalid listen specification: %s", listen)
				continue
			}

			count := len(pieces)
			host := strings.Join(pieces[0:count-1], ":")
			portStr := pieces[count-1]
			tls := false
			if portStr[0] == '*' {
				tls = true
				portStr = portStr[1:]
			}
			port64, err := strconv.ParseUint(portStr, 10, 16)
			if err != nil {
				log.Printf("Invalid listen specification: %s (%s)", listen, err)
				continue
			}
			port := uint16(port64)
			tlsStr := ""
			if tls {
				tlsStr = "TLS "
			}
			log.Printf("Listening for %sclient connections on %s port %d", tlsStr, host, port)
			ircd.NewListener(host, port, tls)
		}
	}

	if len(serverListens) > 0 {
		listens := strings.Split(serverListens, ",")
		for _, listen := range listens {
			pieces := strings.Split(listen, ":")
			if len(pieces) < 2 {
				log.Printf("Invalid listen specification: %s", listen)
				continue
			}

			count := len(pieces)
			host := strings.Join(pieces[0:count-1], ":")
			port64, err := strconv.ParseUint(pieces[count-1], 10, 16)
			if err != nil {
				log.Printf("Invalid listen specification: %s (%s)", listen, err)
				continue
			}
			port := uint16(port64)
			log.Printf("Listening for server connections on %s port %d", host, port)
			ircd.NewLinkListener(host, port)
		}
	}

	log.Printf("Starting ircd...")
	ircd.Run()
	wg.Wait()
}

func validate() (valid bool) {
	valid = true
	if network == "" {
		valid = false
		log.Printf("Must specify --network")
	}
	if server == "" {
		valid = false
		log.Printf("Must specify --server")
	}
	if subnet == "" {
		valid = false
		log.Printf("Must specify --default_subnet")
	}
	if serverDesc == "" {
		log.Printf("--server_desc not specified, description will be empty")
	}

	if networkCa == "" {
		valid = false
		log.Printf("Must specify --tls_network_ca")
	}
	if certificate == "" {
		valid = false
		log.Printf("Must specify --tls_certificate")
	}
	if privateKey == "" {
		valid = false
		log.Printf("Must specify --tls_private_key")
	}
	return
}
