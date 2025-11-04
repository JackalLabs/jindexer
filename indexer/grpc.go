package indexer

import (
	"crypto/tls"
	"regexp"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	HTTPProtocols = regexp.MustCompile("https?://")
)

func CreateGrpcConnection(address string) (*grpc.ClientConn, error) {
	var transportCredentials credentials.TransportCredentials
	if strings.HasPrefix(address, "https") {
		transportCredentials = credentials.NewTLS(&tls.Config{MinVersion: tls.VersionTLS12})
	} else {
		transportCredentials = insecure.NewCredentials()
	}

	address = HTTPProtocols.ReplaceAllString(address, "")
	return grpc.Dial(address, grpc.WithTransportCredentials(transportCredentials))
}
