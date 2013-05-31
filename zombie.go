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
func okaerinasai (conn net.Conn) {
	//timeout := time.Millisecond * 5000
	//var msg []byte
	//msg := make ([]byte, len([]byte(syn)))
	msg := make([]byte, 2048)
	//conn.SetReadDeadline(time.Now().Add(timeout) )
	bytesRead, err := conn.Read(msg)
	if err != nil { conn.Write([]byte(err.Error())); return }
	if string(msg[0:bytesRead]) == syn {
		for i:=0; i<len(msg); i++ {	msg[i]=0x00; }
		conn.Write([]byte(ack))

		//conn.SetReadDeadline(time.Now().Add(timeout) )
		bytesRead, err = conn.Read(msg)
		if err != nil { conn.Write([]byte(err.Error())); return }

		fmt.Println(string(msg))
		var master Master
		err = json.Unmarshal(msg[0:bytesRead], &master)
		//master.Port (the irrelevant at this point, really), master.Uri (Also irrelevant at this point)
		if err != nil { conn.Write([]byte(err.Error())); return }
		for i:=0; i<len(msg); i++ {	msg[i]=0x00; }
		conn.Write([]byte(ack))

		//conn.SetReadDeadline(time.Now().Add(timeout) )
		bytesRead, err = conn.Read(msg)
		if err != nil { conn.Write([]byte(err.Error())); return }
		fmt.Println(string(msg))
		
		zombie := Zombie{}
		zombie.Conn = conn
		err = json.Unmarshal(msg[0:bytesRead], &zombie.Info)
		//zombie.Uri (can check if it's equal to hostname or nslookup), zombie.Ncpu (#threads)
		if err != nil { conn.Write([]byte(err.Error())); return }
		for i:=0; i<len(msg); i++ {	msg[i]=0x00; }
		conn.Write([]byte(ack))

		func () {
			runtime.GOMAXPROCS(zombie.Info.Ncpu)
		}()

		//Reset the read deadline
		//conn.SetReadDeadline(time.Time{})

		/* At this point, checkZombies is done. Start a message handler. */
		zombieHeart(zombie)
		
	}else {	conn.Write([]byte("Who are you?")); return }
}
func zombieHeart(zombie Zombie){
	msg := make ([]byte, 2028)
	invStart := make(chan int)
	for {
		bytesRead, err := zombie.Conn.Read(msg)
		if err != nil { fmt.Println (err.Error()); os.Exit(1) }
		switch (string(msg[0:bytesRead])) {
		case "!":
			for i:=0; i<len(msg); i++ {	msg[i]=0x00; }
			fmt.Println("STARTING!")
			zombie.Limbs = append(zombie.Limbs, invStart)
			//Tells the master that we're ready to receive the cmdlist
			zombie.Conn.Write([]byte(ack))

			bytesRead, err := zombie.Conn.Read(msg)
			if err != nil { fmt.Println(err); zombie.Conn.Write([]byte(err.Error())) }
			clist := string(msg[0:bytesRead])
			
			go hajime(zombie, invStart, clist)
			
		case pingMsg:
		case dieMsg:
			for _,limb := range zombie.Limbs{
				limb <- 0
			}
			fmt.Println("Dying!")
			zombie.Conn.Write([]byte(deadMsg))
			zombie.Conn.Close()
			return
		}
		for i:=0; i<len(msg); i++ {	msg[i]=0x00; }
	}
}

func hajime(zombie Zombie, handler chan int, clist string){
	msg := make ([]byte, 1024)

	var clistArray []command
	
	fmt.Println("COMMAND LIST!:", clist)
	err := json.Unmarshal([]byte(clist), &clistArray)
	if err != nil { zombie.Conn.Write([]byte(err.Error())) }
	for i:=0; i<len(msg); i++ {	msg[i]=0x00; }
	zombie.Conn.Write([]byte(ack))

	// PROBLEM HERE SOMEWHERE
	cmdList := []*command{}
	probcount := 0
	problimit := 100
	probslice := [100]*command{}
	for _, cmd := range clistArray {
		cmdList = append(cmdList, &cmd);
		fmt.Println("COMMAND!",cmd)

		for i:= 0; i < cmd.Probability; i++ {
			probslice[i] = &cmd
			probcount ++
			if probcount > problimit {
				fmt.Println("Probcount >100!!")
				//Handle this, probcount goes from 0-99
			}
		}
	}
	for _,cmd := range cmdList {
		fmt.Println("COMMAND!",*cmd)
	}
	time.Sleep(time.Second)
	done := make (chan bool)
	var numprocs int
	maxusers := runtime.GOMAXPROCS(0)
	//This should be i to max users
	for i:= 0; i < maxusers; i++ {
		rand.Seed(time.Now().Unix())
		numprocs ++
		go attack(zombie, probslice[rand.Int()%100], done)
	}
	for {
		select {
		case <- done:
			numprocs --
			if numprocs <= maxusers {
				rand.Seed(time.Now().Unix())
				go attack(zombie, probslice[rand.Int()%100], done)
			}
		case finger := <- handler:
			switch finger{
			case 0:
				close(handler)
				return
			}
		}
	}
	//Then after, do a select { case x := <- userend: decrement current-users; default: if current-users < max-users
	//go attack(probslice[rand.Int()%100]) then time.Sleep(1 * time.Second) regardless
}

func attack(zombie Zombie, cmd *command, a chan bool) {
	//Parse command list to create the variables needed to store whatever values. Going to use a map structure
	fmt.Println(cmd.Iterations, "Iterations!")
	//do the command $iterations times.
	for i:= 0; i < cmd.Iterations; i++ {
		//For each iteration, iterate through the sequence
		//Note: Come up wit ha way to have a 'global' map, use case: session id's.
		//var mapp map[string]string
		//the seqTimes is gonna be a list, where the last value is the total, so basically if a seq has 3 calls
		//it'll be [ timeforfirstcall, timeforsecondcall, timeforthirdcall, total ] <- send that back to master
		var seqTimes []int64
		seqTimes = append(seqTimes, 0)
		for _, req := range cmd.Sequence {
			/*Scan post-data for anything that looks like grok data: %{DAY:day} for example, so just do regexp %{.*}
			Then for each match, check the text inside, and make an array of pointers pointing to the proper
			result. Then for each match(again), replace the text at the index with the new text. (the mapped val) */
			
			//url = cmd.Url
			//ctype = cmd.ContentType
			//method = cmd.Method, etc
			pattern := req.Response
			//patternp := "%{DAY:s}"
			g := grok.New()
			g.AddPatternsFromFile("patterns")
			g.Compile(pattern)
			
			startTime := time.Now().UnixNano()
			//Do stuff to get response.
			endTime := time.Now().UnixNano()
			totalTime := endTime - startTime
			seqTimes[0] += totalTime
			seqTimes = append(seqTimes, totalTime)
			/*match, err := g.Match(httpresp)
            match = map[string]string
			match["DAY"] or match["DAY:day"] if resp : %{DAY} or %{DAY:day}
            mapp = match */
		}
		//strconv.FormatInt(seqTimes, 10))

		fmt.Println(seqTimes)
		zombie.Conn.Write([]byte(startStream))
		time.Sleep(time.Millisecond * 100)
		for _, val := range seqTimes{
			zombie.Conn.Write([]byte(strconv.FormatInt(val, 10)))
			time.Sleep(time.Millisecond * 100)
		}
		time.Sleep(time.Millisecond * 100)
		zombie.Conn.Write([]byte(startStream))
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
 		okaerinasai(conn)
		fmt.Println("Master killed the connection :(")
 	}
}
