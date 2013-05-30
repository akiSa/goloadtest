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
	syn = "HELLO LITTLE ONE"
	ack = "BLARGH"
	pingMsg = "ピング"
	dieMsg = "死んでくれる"
	deadMsg = "死んでる"
)
/* Structs */
type request struct {
	Url string `json:"url"`
	ContentType string `json:"content-type"`
	Method string `json:"method"`

	/* For post, not necessary.
     * Basically, for response you just write how the data -should- be formatted.
     * Response: variable to store the response in, dump it otherwise.
     * Response syntax: string: str(var) -- Catch all //greedy
     * int: int(var)
     * List: [var1,var2...]var (can just use []var for variable sized response)
     * json: {}var -- ex: {str(var11)}var1 (var1.var11 = "the string")
     *
     * Example response:
     * { "userid": 1010201, "sessionid": 10000 }
     * response: {int(userid), int(sessionid)}resp
     * Access via: resp.userid or resp.sessionid
     */
	PostData string `json:"post-data"`
	Response string `json:"resp"`

	/* Can actually perform functionality on the response data.
     * ex: response = []thelist
     * func: [ "for range(thelist)x: if x<1 then x=1 else nil;" ]
     */
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
