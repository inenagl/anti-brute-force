package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/cristalhq/acmd"
	"github.com/inenagl/anti-brute-force/internal/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	host string
	port string
)

var usage = func(cfg acmd.Config, cmds []acmd.Command) {
	cfg.Output.Write(
		[]byte(`Commands:
	help                            get help
	auth <login> <password> <ip>    check auth possibility for given params
	reset <login> <password> <ip>   reset auth attempts count for each of given params
	add white <ipNet>               add subnet to white list
	add black <ipNet>               add subnet to black list
	remove white <ipNet>            remove subnet from white list
	remove black <ipNet>            remove subnet from black list

Examples:
	auth username password 123.23.45.200
	reset login "pa\\ss\"w0rd" "123.15.65.1"
	reset "" "" 123.33.53.6
	add white 1.12.54.0/24
	add black 192.168.1.25
	remove black 125.0.0.0/16
	remove white 255.13.235.0/24
`),
	)
}

func getConnStr() string {
	b := strings.Builder{}
	b.WriteString(host)
	if port != "" {
		b.WriteString(":")
		b.WriteString(port)
	}
	return b.String()
}

func main() {
	flag.StringVar(&host, "h", "localhost", "ABF service host. Default is localhost")
	flag.StringVar(&port, "p", "8889", "ABF service port. Default is 8889")
	flag.Parse()

	conn, err := grpc.Dial(getConnStr(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		printStderr(err.Error())
		return
	}
	defer conn.Close()

	client := api.NewAntiBruteForceClient(conn)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	defer stop()

	commands := []acmd.Command{
		AuthCmd(os.Stdout, client),
		ResetCmd(os.Stdout, client),
		AddCmd(os.Stdout, client),
		RemoveCmd(os.Stdout, client),
		ExitCmd(stop),
	}

	handler := NewHandler(ctx, stop, os.Stdin, os.Stdout, commands)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		handler.Handle()
	}(&wg)

	wg.Wait()
}

func printStderr(msg string) {
	_, err := fmt.Fprintln(os.Stderr, msg)
	if err != nil {
		log.Fatalln(err)
	}
}
