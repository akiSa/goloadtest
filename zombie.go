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
	fmt.Println("Heart Start!")
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
			hajimedone := make (chan bool)
			go hajime(zombie, invStart, clist, hajimedone)
			<- hajimedone
			
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

func hajime(zombie Zombie, handler chan int, clist string, hajimedone chan bool){
	msg := make ([]byte, 1024)

	var clistArray []command
	
	fmt.Println("COMMAND LIST!:", clist)
	err := json.Unmarshal([]byte(clist), &clistArray)
	if err != nil { zombie.Conn.Write([]byte(err.Error())) }
	for i:=0; i<len(msg); i++ {	msg[i]=0x00; }
	zombie.Conn.Write([]byte(ack))

	var probsum int
	var pc int
	probslice := [100]*command{}
	probmap := [100]int{}
	for iter, cmd := range clistArray {
		probsum += cmd.Probability
		
		for i:=0; i < cmd.Probability; i++ {
			probmap[pc] = iter
			probslice[pc] = &cmd

			pc++
		}
	}
	//Handle this, probably quit.
	if probsum != 100 {	fmt.Println("probabilities don't match up")	}

	// for _,cmd := range cmdList {
	// 	fmt.Println("COMMAND!",*cmd)
	// }
	time.Sleep(time.Second)
	var numprocs int
	maxusers := runtime.GOMAXPROCS(0)
	killHandler := []chanHandler{}

	done := make(chan chanHandler)
	valStream := make(chan results)
	//This should be i to max users
	for i:= 0; i < maxusers; i++ {
		rand.Seed(time.Now().UnixNano())
		numprocs ++
		randcmd := rand.Int()%100
		rand.Seed(time.Now().UnixNano())
		killHandler = append(killHandler, chanHandler{ make (chan bool), i, rand.Int()%100000, probmap[randcmd] } )
		go attack(zombie, probslice[randcmd], killHandler[i], valStream, done)
	}
	hajimedone <- true
	for {
		select {
		case <- done://val := <- done:
			//This is when one of the procs are done, need to re-randomize val.sid
			numprocs --
			//rand.Seed(time.Now().UnixNano())
			//newval := chanHandler{val.Chan, val.Val, rand.Int()%100000 }
			// if numprocs <= maxusers {
			// 	rand.Seed(time.Now().UnixNano())
			// 	go attack(zombie, probslice[rand.Int()%100], newval, valStream, done)
			//}
			//fmt.Println(val, done)
			if numprocs == 0 {
				fmt.Println("Sending deadMsg")
				
				killAllRoutines(killHandler)
				//panic("fdsadsa")
				//zombie.Conn.Close()
				zombie.Conn.Write([]byte(deadMsg))
				return
			}
		case finger := <- handler:
			switch finger{
			case 0:
				close(handler)
				fmt.Println("GOT KILL REQ")
				zombie.Conn.Close()
				killAllRoutines(killHandler)
				return
			}
		case stream := <- valStream:
			/*STATE
             *State where one value will be sent at a time until startStream again
             */
			sendString := startStream+strconv.Itoa(stream.CHandler.Sid)
			sendString += "|"+strconv.Itoa(stream.CHandler.Cmd)
			for _, val := range stream.SeqTimes{
				sendString += "|"+strconv.FormatInt(val, 10)
			}
			sendString+=startStream
			fmt.Println(sendString)
			zombie.Conn.Write([]byte(sendString))
		}
	}
	//Then after, do a select { case x := <- userend: decrement current-users; default: if current-users < max-users
	//go attack(probslice[rand.Int()%100]) then time.Sleep(1 * time.Second) regardless
}

func killAllRoutines(killHandler []chanHandler) {
	for _, kill := range killHandler {
		go killChan(kill)
	}
}
func killChan(kill chanHandler) {
	go func() {
		kill.Chan <- true
	}()
}
func attack(zombie Zombie, cmd *command, kill chanHandler, valStream chan results, done chan chanHandler) {
	//Parse command list to create the variables needed to store whatever values. Going to use a map structure
	
	fmt.Println(cmd.Iterations, "Iterations!")
	allDone := make (chan bool)
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

		//fmt.Println(seqTimes, kill.Val)

		/*
         * Before writing, check to see if it has received a kill message.
         */
		//timeout := make(chan bool, 1)
		// go func() {
		// 	time.Sleep(1 * time.Nanosecond)
		// 	timeout <- true
		// }()
		select {
		case <- kill.Chan:
			return
		case <- time.After(100)://timeout:
			//continue
		}
		//Results
		go func() {
			result := results{kill, seqTimes}
			valStream <- result
			allDone <- true
		}()
	}
	for n:=0; n < cmd.Iterations; n++ {
		<- allDone
	}
	done <- kill
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
