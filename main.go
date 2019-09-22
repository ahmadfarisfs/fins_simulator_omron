package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/l1va/gofins/fins"
)

func main() {
	debug.SetGCPercent(100)
	ip := flag.String("ip", "127.0.0.1", "Virtual PLC IP Address")
	port := flag.Int("port", 9600, "Virtual PLC UDP port")
	network := flag.Int("net", 0, "Virtual PLC network address")
	node := flag.Int("node", 0, "Virtual PLC node address")
	unit := flag.Int("unit", 0, "Virtual PLC unit address")
	run := true
	maxTry := 3
	clientAddr := fins.NewAddress(*ip, 9601, 0, 0, 0)
	plcAddr := fins.NewAddress(*ip, *port, byte(*network), byte(*node), byte(*unit))

	s, e := fins.NewPLCSimulator(plcAddr)
	fmt.Println("Virtual PLC Running !")
	fmt.Println("IP: " + *ip)
	fmt.Println("Port: " + strconv.Itoa(*port))
	fmt.Println("Network: " + strconv.Itoa(*network))
	fmt.Println("Node: " + strconv.Itoa(*node))
	fmt.Println("Unit: " + strconv.Itoa(*unit))

	if e != nil {
		panic(e)
	}
	defer s.Close()

	c, err := fins.NewClient(clientAddr, plcAddr)
	c.SetTimeoutMs(500)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	fmt.Println("Initial Register 0 Value: ")
	try := 0
	for {
		z, err := c.ReadWords(fins.MemoryAreaDMWord, 0, 1)
		if err != nil {
			fmt.Println(z)
			if try > maxTry {
				panic(err)
			} else {
				fmt.Println("Retrying...")
				time.Sleep(1 * time.Second)
				try += 1
			}
		} else {
			fmt.Println(z)
			break
		}
	}
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Virtual PLC Shell")
	fmt.Println("Command:")
	fmt.Println("set {address} {value}")
	fmt.Println("get {address}")
	fmt.Println("end")

	fmt.Println("---------------------")

	for run == true {
		fmt.Print("-> ")
		text, _ := reader.ReadString('\n')
		//Sanitize input
		text = strings.ToLower(strings.Trim(text, " \r\n"))
		cmdText := text[0:3]
		switch cmdText {
		case "set":
			s := strings.Split(text[4:], " ")
			if len(s) != 2 {
				fmt.Println("Wrong Command")
				break
			}
			memoryArea, _ := strconv.Atoi(s[0])
			memoryValue, _ := strconv.Atoi(s[1])
			err = c.WriteWords(fins.MemoryAreaDMWord, uint16(memoryArea), []uint16{uint16(memoryValue)})
			if err != nil {
				panic(err)

			} else {
				fmt.Println("Written " + strconv.Itoa(memoryValue) + " to address " + strconv.Itoa(memoryArea))
			}
		case "get":

			s := strings.Split(text[4:], " ")
			if len(s) != 1 {
				fmt.Println("Wrong Command")
				break
			}
			memoryArea, _ := strconv.Atoi(s[0])
			z, err := c.ReadWords(fins.MemoryAreaDMWord, uint16(memoryArea), 1)
			if err != nil {
				panic(err)

			} else {
				fmt.Println("Get Value: " + strconv.Itoa(int(z[0])))
			}
		case "end":
			fmt.Println("Quiting")
			run = false
		}
	}
}
