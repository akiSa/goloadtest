package main

import (
	"runtime"
//	"io"
	"fmt"
	"os"
	"io/ioutil"
	"flag"
	"encoding/json"
	"net"
)

const (
	syn = "syn"
	ack = "ack"
	pingMsg = "ピング"
	dieMsg = "死んでくれる"
	deadMsg = "死んでる"
	startStream = "%"
)
/* Structs */
type request struct {
	Url string `json:"url"`
	ContentType string `json:"content-type"`
	Method string `json:"method"`

	PostData string `json:"post-data"`
	Response string `json:"resp"`

	Func []string `json:"func"`
}
type command struct {
	Probability int `json:"probability"`
	Iterations int `json:"iterations"`
	Sequence []request `json:"sequence"`
}
type client struct {
	Uri string `json:"uri"`
	Ncpu int `json:"ncpu"`
}
type config struct {
	Port int `json:"port"`
	Master client `json:"master"`
	Zombies []client `json:"zombies"`
	CommandList []command `json:"commandlist"`
}
type zombieCmd struct {
	CommandList []command `json:"commandlist"`
}
type Master struct {
	Port int `json:"port"`
	Uri string `json:"uri"`
}
type Zombie struct {
	/* Limbs refer to process branches, yes; I -will- stick to this theme.
     * The Limbs slice is how the master passes messages to the subproc's.
     */
	Info client
	Conn net.Conn
	Limbs []chan int
}
type Error struct {
	err error
	context string
}
/* End Structs */

var file = flag.String("f", "config.json", "The config file")
var mode = flag.String("m", "", "Mode: master or zombie")

func init() {
	
}
func parseConfig(file string, conf chan config, e chan error) {
	cfg, err := ioutil.ReadFile(file)
	if err != nil { e <- err; return}
	var c config
	err = json.Unmarshal(cfg, &c)
	if err != nil { e <- err; return}

	conf <- c
}

func main() {
	flag.Parse()
	switch {
	case *(mode) == "master":
		fmt.Println ("Time to grab my whip...")
	case *(mode) == "zombie":
		fmt.Println ("Brains...")
	default:
		flag.PrintDefaults()
		os.Exit(1)
	}
	var cfg config
	
	/* Create the log file, close it at the very end. */
	log, err := os.Create("log")
	if err != nil { print(err.Error()); os.Exit(1) }
	defer func() {
		if err := log.Close(); err != nil {
			panic(err)
		}
	}()

	/* Parse the config in a go routine, collect either the result or the error using the channels. */
	func (){
		conf := make (chan config)
		e := make (chan error)
		go parseConfig(*(file), conf, e)
		select {
		case cfg = <- conf:
		case err = <- e :
			fmt.Println ("Config error: " + err.Error());
			os.Exit(1);
		}
	}()

	func () {
		runtime.GOMAXPROCS(cfg.Master.Ncpu)
	}()
	
	switch {
	case *(mode) == "master":
		masterStart(cfg)
	case *(mode) == "zombie":
		zombieStart(cfg)
	}
}
