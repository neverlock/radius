package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/layeh/radius"
)

const usage = `
Sends an Access-Request RADIUS packet to a server and prints the result.
`

var ARG0 = os.Args[0]
var host, port, hostport string
var fARG4, fARG0, fARG1, fARG3 string
var NASPORT int
var timeout *time.Duration

func main() {
	var wg sync.WaitGroup
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [flags] <user> <password> <radius-server>[:port] <nas-port-number> <secret> <loopno>\n", ARG0)
		flag.PrintDefaults()
		fmt.Fprint(os.Stderr, usage)
	}
	timeout = flag.Duration("timeout", time.Second*10, "timeout for the request to finish")
	flag.Parse()
	if flag.NArg() != 6 {
		flag.Usage()
		os.Exit(1)
	}

	fARG4 = flag.Arg(4)
	fARG0 = flag.Arg(0)
	fARG1 = flag.Arg(1)
	fARG3 = flag.Arg(3)
	start := time.Now()
	loopNo, _ := strconv.Atoi(flag.Arg(5))

	for i := 1; i <= loopNo; i++ {
		wg.Add(1)
		go runTest(&wg)
	}
	wg.Wait()

	elapsed := time.Since(start)
	fmt.Println("Time use =", elapsed)
}
func runTest(wg *sync.WaitGroup) {
	defer wg.Done()
	host, port, err := net.SplitHostPort(flag.Arg(2))
	if err != nil {
		host = flag.Arg(2)
		port = "1812"
	}
	hostport := net.JoinHostPort(host, port)

	packet := radius.New(radius.CodeAccessRequest, []byte(fARG4))
	packet.Add("User-Name", fARG0)
	packet.Add("User-Password", fARG1)
	NASPORT, _ = strconv.Atoi(fARG3)
	packet.Add("NAS-Port", uint32(NASPORT))

	client := radius.Client{
		DialTimeout: *timeout,
		ReadTimeout: *timeout,
	}
	received, err := client.Exchange(packet, hostport)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var status string
	if received.Code == radius.CodeAccessAccept {
		status = "Accept"
	} else {
		status = "Reject"
	}
	if msg, ok := received.Value("Reply-Message").(string); ok {
		status += " (" + msg + ")"
	}

	fmt.Println(status)

	if received.Code != radius.CodeAccessAccept {
		os.Exit(2)
	}

}
