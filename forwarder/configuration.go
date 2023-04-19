package forwarder

import (
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

//Configuration for the forwarder. Since it is capitalized, should be accessible outside package.
type Configuration struct {
	srcPort        int    //source where incoming connections to forward are listened to
	dstPort        int    //port on destination host where to forward incoming data/connections to. data coming back from destination is forwarded back using the source connection.
	dstHost        string //destination host ip/name
	mirrorUpPort   int    //port on mirror host where to also forward incoming data/connection upstream traffic (from source to destination)
	mirrorUpHost   string //upstream mirror host ip/name
	mirrorDownPort int    //port on mirror host where to also forward incoming data/connection downstream traffic (from destination to source)
	mirrorDownHost string //downstream mirror host ip/name
	dataDownFile   string //if defined, write downstream data passed through to this file
	dataUpFile     string //if defined, write upstream data passed through to this file
	logFile        string //if defined, write log to this file
	logToConsole   bool   //if we should log to console
	bufferSize     int    //size to use for buffering read/write data
	bandwidth      int64
}

type PortConfig struct {
	srcPort   int
	dstHost   string
	dstPort   int
	bandwidth int64
}

var ListenerList []PortConfig

var StartListenersFunc func([]PortConfig)

//this is how go defines variables, so the actual configurations are stored here
var Config Configuration

//ParseConfig reads the command line arguments and sets the global Configuration object from those. Also checks the arguments make basic sense.
func ParseConfig() {

	srcPortPtr := 0

	//here I was trying to create a shorter form of source port config option, while having the previous one be longer. the default package does not really support that, so dropped it.
	//	flagSet.IntVar(srcPortPtr, "sp", -1, "Source port for incoming connections.")

	dstPortPtr := 0
	dstHostPtr := ""

	mirrorUpPortPtr := 0

	mirrorUpHostPtr := ""

	mirrorDownPortPtr := 0
	mirrorDownHostPtr := ""

	dataUpFilePtr := ""
	dataDownFilePtr := ""

	logFilePtr := ""

	logToConsolePtr := false
	bufferSizePtr := 1024

	Config.srcPort = srcPortPtr
	Config.dstPort = dstPortPtr
	Config.dstHost = dstHostPtr
	Config.mirrorUpPort = mirrorUpPortPtr
	Config.mirrorUpHost = mirrorUpHostPtr
	Config.mirrorDownPort = mirrorDownPortPtr
	Config.mirrorDownHost = mirrorDownHostPtr
	Config.dataDownFile = dataDownFilePtr
	Config.dataUpFile = dataUpFilePtr
	Config.logFile = logFilePtr
	Config.logToConsole = logToConsolePtr
	Config.bufferSize = bufferSizePtr

	//var errors = ""
	if Config.srcPort < 1 || Config.srcPort > 65535 {
		//errors += "You need to specify source port in range 1-65535.\n"
	}
	if len(Config.dstHost) == 0 {
		//errors += "You need to specify destination host.\n"
	}
	if Config.dstPort < 1 || Config.dstPort > 65535 {
		//errors += "You need to specify destination port in range 1-65535.\n"
	}
	if Config.bufferSize < 1 {
		//errors += "Buffer size needs to be >= 1.\n"
	}
	if len(Config.mirrorUpHost) > 0 {
		if Config.mirrorUpPort < 1 || Config.mirrorUpPort > 65535 {
			//errors += "When upstream mirror host is defined, its port must be defined in range 1-65535.\n"
		}
	} else {
		if Config.mirrorUpPort != 0 {
			//errors += "Mirror-up port defined but no mirror-up host. Mirror host is required if mirror is enabled.\n"
		}
	}
	if len(Config.mirrorDownHost) > 0 {
		if Config.mirrorDownPort < 1 || Config.mirrorDownPort > 65535 {
			//errors += "When downstream mirror host is defined, its port must be defined in range 1-65535.\n"
		}
	} else {
		if Config.mirrorDownPort != 0 {
			//errors += "Mirror-down port defined but no mirror-down host. Mirror host is required if mirror is enabled.\n"
		}
	}

	go watchFileAndRun("portsforward.conf", updatePortsForward)

	// if len(errors) > 0 {
	// 	fmt.Print(errors)
	// 	fmt.Println()
	// 	fmt.Print("Usage: " + os.Args[0] + " [options]")
	// 	os.Exit(1)
	// }
}

/*
func updateBandwidthLimit() {

	ticker := time.NewTicker(10 * time.Second)
	quit := make(chan struct{})

	go func() {
		for {
			select {
			case <-ticker.C:

				f, err := os.OpenFile("bandwidth.conf", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

				if err != nil {
					//panic(err)
				}

				defer f.Close()

				bc := make([]byte, 10)

				_, err = f.Read(bc)

				bc = bytes.Trim(bc, "\x00")

				content := string(bc)

				if content == "" {
					content = "512"
					bc = []byte(content)
					f.Write(bc)
				}

				Config.bandwidth, _ = strconv.ParseInt(content, 10, 20)

			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	return
}
*/

func updatePortsForward() {

	bc, _ := ioutil.ReadFile("portsforward.conf")

	content := strings.TrimSpace(string(bc))

	if content == "" {
		content = "22001,0.0.0.0,22000,1500\n"
		ioutil.WriteFile("portsforward.conf", []byte(content), 0666)

	}

	rows := make([]string, 1)

	rows = strings.Split(content, "\n")

	for i := range rows {
		row := rows[i]
		if row != "" {
			parts := strings.Split(row, ",")
			pc := PortConfig{}

			pc.srcPort, _ = strconv.Atoi(parts[0])
			pc.dstHost = parts[1]
			pc.dstPort, _ = strconv.Atoi(parts[2])
			pc.bandwidth, _ = strconv.ParseInt(parts[3], 10, 20)

			ListenerList = append(ListenerList, pc)

		}

	}

	StartListenersFunc(ListenerList)

	return
}

func watchFileAndRun(filePath string, fn func()) {
	defer func() {
		r := recover()
		if r != nil {
			//logCore("ERROR", "Error:watching file:", r)
		}
	}()
	fn()
	initialStat, _ := os.Stat(filePath)
	//checkErr(err)
	for {
		stat, _ := os.Stat(filePath)
		//checkErr(err)
		if stat.Size() != initialStat.Size() || stat.ModTime() != initialStat.ModTime() {

			log.Print("portsforward.conf changed.")

			fn()
			initialStat, _ = os.Stat(filePath)
			//checkErr(err)
		}
		time.Sleep(10 * time.Second)
	}

}
