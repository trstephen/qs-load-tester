package main

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"sync/atomic"
	"time"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	// Globals
	quoteServerTCPAddr *net.TCPAddr
	quoteCount         uint64
	requestCount       uint64
)

func main() {
	// Seed random to get non-deterministic stock names
	rand.Seed(time.Now().Unix())

	loadConfig()
	resolveTCPAddresses()

	// Recall: Tickers fire every interval and repeat.
	//         Timers fire once and finish.
	spawnRequest := time.NewTicker(time.Millisecond * time.Duration(config.delay))
	sessionExpiry := time.NewTimer(time.Second * time.Duration(config.length))

	for {
		select {
		case <-spawnRequest.C:
			go getQuote()
		case <-sessionExpiry.C:
			// stops execution
			printStats()
			return
		}
	}
}

var config struct {
	host   string
	port   int
	delay  int
	length int
}

func loadConfig() {
	app := kingpin.New("quoteserver load tester", "Requests quotes at a fixed rate")

	app.Flag("host", "Quote server host address").
		Short('h').
		Default("quoteserve.seng.uvic.ca").
		StringVar(&config.host)

	app.Flag("port", "Port to make requests from quoteserver").
		Short('p').
		Default("4440").
		IntVar(&config.port)

	app.Flag("delay", "Delay between quote requests, in ms").
		Short('d').
		Default("100").
		IntVar(&config.delay)

	app.Flag("length", "How long to request quotes, in sec").
		Short('l').
		Default("60").
		IntVar(&config.length)

	kingpin.MustParse(app.Parse(os.Args[1:]))
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func resolveTCPAddresses() {
	// TCP connections need to know the specific IP for the destination.
	// We can do the lookup in advance since destinations are fixed.
	quoteServerAddress := fmt.Sprintf("%s:%d", config.host, config.port)

	var err error
	quoteServerTCPAddr, err = net.ResolveTCPAddr("tcp", quoteServerAddress)
	failOnError(err, "Could not resolve TCP addr for "+quoteServerAddress)
}

func getQuote() {
	request := fmt.Sprintf("%s,jdoe\n", getRandStock())

	quoteServerConn, err := net.DialTCP("tcp", nil, quoteServerTCPAddr)
	failOnError(err, "Could not connect to "+quoteServerTCPAddr.String())
	defer quoteServerConn.Close()

	// Fail if we can't read/write to the quoteserver
	quoteServerConn.SetDeadline(time.Now().Add(time.Second * 10))

	_, err = quoteServerConn.Write([]byte(request))
	failOnError(err, "Problem sending to "+quoteServerTCPAddr.String())
	atomic.AddUint64(&requestCount, 1)

	respBuf := make([]byte, 1024)
	_, err = quoteServerConn.Read(respBuf)
	failOnError(err, "Problem reading from "+quoteServerTCPAddr.String())

	// Clean up unused space in buffer and line break from quote server
	respBuf = bytes.Trim(respBuf, "\x00")
	respBuf = bytes.Trim(respBuf, "\n")

	fmt.Println(string(respBuf))

	atomic.AddUint64(&quoteCount, 1)
}

const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

func getRandStock() string {
	stock := make([]byte, 3)
	for i := range stock {
		stock[i] = letters[rand.Intn(len(letters))]
	}
	return string(stock)
}

func printStats() {
	totalQuotes := atomic.LoadUint64(&quoteCount)
	fmt.Fprintf(os.Stderr, "Quotes: %d\n", totalQuotes)

	totalRequests := atomic.LoadUint64(&requestCount)
	fmt.Fprintf(os.Stderr, "Requests: %d\n", totalRequests)

	fmt.Fprintf(os.Stderr, "Req per sec: %.2f\n",
		float64(totalRequests)/float64(config.length),
	)
}
