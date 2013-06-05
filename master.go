package main

import (
	"net"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"errors"
	"time"
	"os/signal"
)
/*
checkZombies connects to the zombies via tcp, and sends the first syn "HELLO LITTLE ONE"
*/
func checkZombies (zombie client, zombchan chan Zombie, zomberr chan error, master Master){
	//timeout := time.Millisecond * 5000	

	fmt.Println ("Connecting to: " + zombie.Uri)
	zombconn, err := net.Dial("tcp", zombie.Uri+":"+strconv.Itoa(master.Port))
	if err != nil { zomberr <- err; return}
	_, err = zombconn.Write([]byte(syn))
	if err != nil { zomberr <- err; return}
	//defer zombconn.Close()
	
	zomb := Zombie{zombie, zombconn, []chan int{}}

	msg := make ([]byte, 1024)
	//zomb.Conn.SetReadDeadline(time.Now().Add(timeout) )
	bytesRead, err := zomb.Conn.Read(msg)
	if err != nil { zomberr <- err; return}
	if string(msg[0:bytesRead]) == ack {
		// Write master info, port/uri
		masterJson, err := json.Marshal(master)
		if err != nil { zomberr <- err; return}
		_, err = zomb.Conn.Write(masterJson)
		if err != nil { zomberr <- err; return}

		//Clear msg buffer and read for the ack
		for i:=0; i<len(msg); i++ {	msg[i]=0x00; }
		//zomb.Conn.SetReadDeadline(time.Now().Add(timeout) )
		bytesRead, err = zomb.Conn.Read(msg)
		if err != nil { zomberr <- err; return}
		if string(msg[0:bytesRead]) != ack { zomberr <- errors.New(string(msg)); return }

		// Write zombie info, uri/ncpu
		zombJson, err := json.Marshal(zomb.Info)
		if err != nil { zomberr <- err; return}
		_, err = zomb.Conn.Write(zombJson)
		if err != nil { zomberr <- err; return}

		//Clear msg buffer and read for the ack
		for i:=0; i<len(msg); i++ {	msg[i]=0x00; }
		//zomb.Conn.SetReadDeadline(time.Now().Add(timeout) )

		bytesRead, err = zomb.Conn.Read(msg)
		//Reset the read deadline
		//zomb.Conn.SetReadDeadline(time.Time{})
		if err != nil { zomberr <- err; return}
		if string(msg[0:bytesRead]) != ack { zomberr <- errors.New(string(msg)); return }

		zombchan <- zomb
	}else {
		zomberr <- errors.New(string(msg))
	}
}
func masterStart(cfg config){
	/* Just to pretty-print the config file. * /
	jsonIndent, err := json.MarshalIndent(cfg, "", " ")
	if err != nil { fmt.Println ("Json Error: " + err.Error()) }
	fmt.Println (string(jsonIndent))
	// */

	clist, err := json.Marshal(cfg.CommandList)
	if err != nil { fmt.Println ("Json Error: " + err.Error()) }
	//fmt.Println(string(clist))

	master := Master{cfg.Port, cfg.Master.Uri}

	zombies := []Zombie{}
	zombchan := make (chan Zombie, len(cfg.Zombies))
	zomberr := make (chan error, len(cfg.Zombies))

	for _, zombie := range cfg.Zombies {
		//fmt.Println(string(masterJson))//Send masterJson to zombie
		go checkZombies(zombie, zombchan, zomberr, master)
	}

	for _,_ = range cfg.Zombies{
		select {
		case newzomb := <-zombchan:
			zombies = append(zombies, newzomb)
			//zombies[conn.RemoteAddr().String()] = conn
		case err = <- zomberr:
			fmt.Println (err.Error())
			fmt.Println("One of the zombies is not in-line! Will not proceed without the army!")
			//Tell each zombie to die.n
			for _, zombie := range zombies{	zombie.Conn.Write([]byte(dieMsg)) }
			os.Exit(1)
		}
	}
	//Send commandlist and start the invasion

	var flaggu string
	fmt.Println ("AWAITING YOUR COMMAND TO START THE INVASION: ! to start, q to quit.")
	for {
		fmt.Scan(&flaggu)
		if flaggu == "!" {
			/* Release the kraken
			 * Keep a channel that all zombies will use to communicate result values.
			 */
			vals := make (chan string)
			for _,zombi := range zombies{ go startInvasion(zombi, clist, vals) };
			theWorld(cfg, zombies, vals)
		} else if flaggu == "q" {
			killZombies(zombies); break
		} else { fmt.Println("TRY AGAIN: ! to start, q to quit."); fmt.Scan(&flaggu) }
	}
}
func theWorld(cfg config, zombies []Zombie, vals chan string) {
	/* The world is where the user interaction will be handled, grabbing inputs.
     * It also instantiates and controls the values pushed
     */
	passer := make (chan string)
	go inputHandler (passer)

	sigint := make (chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	
	for {
		select {
		case meirei := <- passer:
			switch meirei {
			case "h":
				displayHelp()
			case "q":
				fmt.Println("Quitting!")
				killZombies(zombies)
				time.Sleep(1 * time.Second)
				os.Exit(1)
			case "i":
				fmt.Println(cfg)
			}
		case val := <- vals:
			//Got some vals from a zombie, handle here; for some reason not getting vals.
			fmt.Println(val)
		case <- sigint:
			//C-c from user
			fmt.Println("bai bai")
			killZombies(zombies)
			time.Sleep(1 * time.Second)
			os.Exit(1)
		}
	}
}
func displayHelp() {
	fmt.Println("h - help\nq - quit\ni = info")
}
func inputHandler(passer chan string){
	var m string
	for {
		fmt.Println ("命令は?")
		fmt.Scan(&m)
		passer <- m
	}
}
//TODO: Better error handling in startInvasion
func startInvasion(zombie Zombie, cmdList []byte, vals chan string){
	msg := make([]byte, 1024)
	_, err := zombie.Conn.Write([]byte("!"))
	if err != nil { fmt.Println(err.Error()) }
	bytesRead, err := zombie.Conn.Read(msg)
	if string(msg[0:bytesRead]) != ack {
		fmt.Println("NOT ACK",string(msg[0:bytesRead]));
		os.Exit(1)
	}

	fmt.Println("Sending command list!", string(cmdList))
	_, err = zombie.Conn.Write(cmdList)
	
	if err != nil { fmt.Println(err.Error()); os.Exit(1) }
	bytesRead, err = zombie.Conn.Read(msg)
	if err != nil { fmt.Println(err.Error()); os.Exit(1) }
	if string(msg[0:bytesRead]) != ack { fmt.Println("NOT ACK",string(msg[0:bytesRead])); os.Exit(1) }

	zombie.Limbs = append(zombie.Limbs, make(chan int))
	zombie.Limbs = append(zombie.Limbs, make(chan int))
	// Creates the process tree for each zombie.
	go ping(zombie, zombie.Limbs[0])
	go start(zombie, zombie.Limbs[1], vals)
}
func readConn(zombie Zombie, vals chan string){
	msg := make([]byte, 1024)
	fmt.Println("Reading conn")
	// for{
	// 	br, err := zombie.Conn.Read(msg)
	// 	if err != nil {
	// 		fmt.Println (err.Error) //Needs better error handling
	// 	}else{
	// 		fmt.Println("Message:",string(msg[0:br]))
	var state int
	var data []byte
	/* State info
     * State 0: Default
     * State 1: Reading chars
     */
	for {
		br, err := zombie.Conn.Read(msg)
		//fmt.Println("Message:",string(msg))
		if err != nil {	fmt.Println(err.Error()); break }
		for i:=0; i<br;i++ {
			//fmt.Println("State:", state, "NextByte:",string(msg[i]))
			switch state {
			case 0:
				switch string(msg[i]) {
				case startStream:
					state = 1
				case deadMsg:
					return
				}
			case 1:
				switch string(msg[i]) {
				case startStream:
					state = 0
					if len(data) > 0 {
						go parseData(data);
						data = []byte{}
					}
				default:
					data = append(data, msg[i])
				}
			}
		}
		//fmt.Println("reading again")

	}
	// 		case 1:
	// 		}
	// 		switch string(msg) {
	// 		case deadMsg:
	// 			//zombie.Conn.Close()
	// 			return
	// 		default:
	// 			/* STATE */
	// 			val := string(msg[0:br])
	// 			for i:=0; i<len(msg); i++ {	msg[i]=0x00; }
	// 			vals <- val
	// 		}
	// 	}
	// }
}
func parseData (data []byte){
	fmt.Println(string(data))
}
func start(zombie Zombie, limb chan int, vals chan string){
	go readConn(zombie, vals)
	for {
		select {
		case finger := <- limb:
			switch finger{
			case 0:
				//Cleanup
				close(limb)
				return
			}
		}
	}
}
/* Just something to ensure connectivity. */
func ping(zombie Zombie, limb chan int){
	tick := time.NewTicker(5 * time.Second)
	select {
	case <- tick.C:
		/* Should only have one thing reading from the connection. */
		_, err := zombie.Conn.Write([]byte(pingMsg))
		if err != nil { fmt.Println(err.Error()) }
	case finger := <- limb:
		switch finger {
		case 0:
			/* Kill all zombies case. */
			close(limb)
			return
		}
	}
}
func killZombies(zombies []Zombie) {
	for _, zomb := range zombies {
		zomb.Conn.Write([]byte(dieMsg))
		for _, limb := range zomb.Limbs { limb <- 0 } // Send kill message to all limbs.
	}
}
