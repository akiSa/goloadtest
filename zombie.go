package main

import (
	"fmt"
	"os"
	"net"
	"strconv"
	"runtime"
	"encoding/json"
	"time"
	"math/rand"
//	"github.com/blakesmith/go-grok/grok"
	"github.com/rckclmbr/gogrok/grok"
)
func okaerinasai (conn net.Conn, cfg config) {
	timeout := time.Millisecond * 5000
	msg := make ([]byte, 2048)
	conn.SetReadDeadline(time.Now().Add(timeout) )
	_, err := conn.Read(msg)
	if err != nil { conn.Write([]byte(err.Error())); return }

	if string(msg) == syn {
		for i:=0; i<len(msg); i++ {	msg[i]=0x00; }
		conn.Write([]byte(ack))

		conn.SetReadDeadline(time.Now().Add(timeout) )
		_, err = conn.Read(msg)
		if err != nil { conn.Write([]byte(err.Error())); return }

		var master Master
		err = json.Unmarshal(msg, master)
		//master.Port (the irrelevant at this point, really), master.Uri (Also irrelevant at this point)
		if err != nil { conn.Write([]byte(err.Error())); return }
		for i:=0; i<len(msg); i++ {	msg[i]=0x00; }
		conn.Write([]byte(ack))

		conn.SetReadDeadline(time.Now().Add(timeout) )
		_, err = conn.Read(msg)
		if err != nil { conn.Write([]byte(err.Error())); return }

		zombie := Zombie{}
		zombie.Conn = conn
		err = json.Unmarshal(msg, zombie.Info)
		//zombie.Uri (can check if it's equal to hostname or nslookup), zombie.Ncpu (#threads)
		if err != nil { conn.Write([]byte(err.Error())); return }
		for i:=0; i<len(msg); i++ {	msg[i]=0x00; }
		conn.Write([]byte(ack))

		func () {
			runtime.GOMAXPROCS(zombie.Info.Ncpu)
		}()

		//Reset the read deadline
		conn.SetReadDeadline(time.Time{})

		/* At this point, checkZombies is done. Start a message handler. */
		zombieHeart(zombie, cfg)
		
	}else {	conn.Write([]byte("Who are you?")); return }
}
func zombieHeart(zombie Zombie, cfg config){
	msg := make ([]byte, 1024)
	for {
		zombie.Conn.Read(msg)
		switch (string(msg)) {
		case "!":
			invStart := make(chan int)
			zombie.Limbs = append(zombie.Limbs, invStart)
			go hajime(zombie, invStart, cfg)
		case pingMsg:
		case dieMsg:
			for _,limb := range zombie.Limbs{
				limb <- 0
			}
			zombie.Conn.Write([]byte(deadMsg))
			os.Exit(1)
		}
		for i:=0; i<len(msg); i++ {	msg[i]=0x00; }		
	}
}

func hajime(zombie Zombie, handler chan int, cfg config){
	msg := make ([]byte, 4096)
	zombie.Conn.Write([]byte(ack))

	_, err := zombie.Conn.Read(msg)
	if err != nil { zombie.Conn.Write([]byte(err.Error())) }
	err = json.Unmarshal(msg, cfg.CommandList)
	if err != nil { zombie.Conn.Write([]byte(err.Error())) }
	for i:=0; i<len(msg); i++ {	msg[i]=0x00; }
	zombie.Conn.Write([]byte(ack))

	//Parse the command list and start the invasion!
	cmdList := []*command{}
	probcount := 0
	problimit := 100
	probslice := [100]*command{}
	for _, cmd := range cfg.CommandList {
		cmdList = append(cmdList, &cmd);
	 	for i:= 0; i < cmd.Probability; i++ {
			probslice[i] = &cmd
			probcount ++
			if probcount >= problimit {
				//Handle this, probcount goes from 0-99
			}
		}
	}
	for i:= 0; i < runtime.GOMAXPROCS(0); i++ {
		rand.Seed(time.Now().Unix())
		
		go attack(probslice[rand.Int()%100])
	}

	select {
	case finger := <- handler:
		switch finger {
		case 0:
			close (handler)
			return
		}

	}
}

func attack(cmd *command) {
	//Parse command list to create the variables needed to store whatever values. Going to use a map structure
	
	//do the command $iterations times.
	for i:= 0; i < cmd.Iterations; i++ {
		//For each iteration, iterate through the sequence
		for _, req := range cmd.Sequence {
			//url = cmd.Url
			//ctype = cmd.ContentType
			//method = cmd.Method, etc
			pattern := &req.Response
			//pattern := "%{DAY:s}"
			g := grok.New()
			g.AddPatternsFromFile("patterns")
			g.Compile(*pattern)
			

			//Do stuff to get response.
			
			/*match, err := g.Match(httpresp)
            match = map[string]string
			match["DAY"] or match["DAY:day"] if resp : %{DAY} or %{DAY:day}*/
		}
	}
}

func zombieStart(cfg config) {
 	ln, err := net.Listen("tcp", ":"+strconv.Itoa(cfg.Port))
 	if err != nil { fmt.Println (err.Error()); os.Exit(1) }
 	print("Listening for a master on: "); print(ln.Addr()); print("\n");
 	for {
 		conn, err := ln.Accept()
 		if err != nil { fmt.Println (err.Error()) }
 		//Only one master per zombie.
 		okaerinasai(conn, cfg)
 	}
}
