package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"

	chclient "github.com/jpillora/chisel/client"
	chserver "github.com/jpillora/chisel/server"
	"github.com/jpillora/chisel/share/cnet"
	cos "github.com/jpillora/chisel/share/cos"
)

func usage() {
	fmt.Printf("usage: (client|server)")
}

var serverVersion = "1.9.1"

func main() {
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()

	cmd := ""
	if len(args) > 0 {
		cmd = args[0]
		args = args[1:]
	}

	var cmdErr error
	switch cmd {
	case "server":
		cmdErr = server(args)
	case "client":
		cmdErr = client(args)
	default:
		usage()
		os.Exit(1)
	}
	if cmdErr != nil {
		log.Fatal(cmdErr)
	}
}

func server(args []string) error {
	flagset := flag.NewFlagSet("server", flag.ExitOnError)
	verbose := flagset.Bool("v", false, "verbose")
	host := flagset.String("h", "0.0.0.0", "host")
	port := flagset.String("p", "8080", "port")
	flagset.Parse(args)

	ctx := cos.InterruptContext()

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return err
	}
	reverseServer := cnet.NewHTTPServer()
	h := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		referrerUrl, _ := url.Parse(r.Referer())
		referrerUrl.RawQuery = ""
		referrerUrl.Path += fmt.Sprintf("proxy/%s", *port)

		for _, cookie := range r.Cookies() {
			if cookie.Name == "code-server-session" {
				clientCmd := fmt.Sprintf(`chisel client --header 'Cookie: %s' %s  # add local port`, cookie.String(), referrerUrl.String())
				w.Write([]byte(clientCmd))
				fmt.Println(clientCmd)
			}
		}
	}))
	reverseServer.GoServe(ctx, listener, h)

	config := &chserver.Config{}
	config.Proxy = fmt.Sprintf("http://%s/reverse", listener.Addr().String())

	s, err := chserver.NewServer(config)
	if err != nil {
		return err
	}
	s.Debug = *verbose

	if err := s.StartContext(ctx, *host, *port); err != nil {
		log.Fatal(err)
	}
	if err := s.Wait(); err != nil {
		log.Fatal(err)
	}
	return nil
}

func client(args []string) error {
	flagset := flag.NewFlagSet("client", flag.ExitOnError)
	verbose := flagset.Bool("v", false, "verbose")
	flagset.Parse(args)

	config := &chclient.Config{Headers: http.Header{}}
	c, err := chclient.NewClient(config)
	if err != nil {
		return err
	}
	c.Debug = *verbose

	ctx := cos.InterruptContext()
	if err := c.Start(ctx); err != nil {
		return err
	}
	if err := c.Wait(); err != nil {
		return err
	}
	return nil
}
