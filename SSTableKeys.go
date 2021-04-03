package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/golang/leveldb/table"
)

var concurrentWorkers = 4

var re = regexp.MustCompile(`^\d{16}$`)

var mutexProto = &sync.Mutex{}
var mutexPort = &sync.Mutex{}
var mutexIP4 = &sync.Mutex{}
var mutexIP6 = &sync.Mutex{}
var mutexSize = &sync.Mutex{}

var protocolSetCount = make(map[int]int)
var portSetCount = make(map[int]int)
var ipv4SetCount = make(map[string]int)
var ipv6SetCount = make(map[string]int)
var totalSize = int64(0)

var wg sync.WaitGroup

const majorVersionNumber = 2

func worker(id int, jobs <-chan string, folderPath string, dataFolderPath string, startDate int, endDate int) {
	for filename := range jobs {
		if re.MatchString(filename) {
			if startDate != -1 && endDate != -1 {
				fileNameShort, err := strconv.Atoi(filename[:10])
				if err == nil && fileNameShort >= startDate-60 && endDate+60 >= fileNameShort {
					readIndexFile(filename, folderPath, dataFolderPath)
				}
			} else {
				readIndexFile(filename, folderPath, dataFolderPath)
			}
		}
		wg.Done()
	}
}

func readIndexFile(filename string, folderPath string, dataFolderPath string) {

	filePath := fmt.Sprintf("%s/%s", folderPath, filename)
	fh, fhErr := os.Open(filePath)

	if fhErr != nil {
		fh.Close()
		return
	}
	ss := table.NewReader(fh, nil)
	if versions, err := ss.Get([]byte{0}, nil); err != nil {
		fh.Close()
		return
	} else if len(versions) != 8 {
		fh.Close()
		return
	} else if major := binary.BigEndian.Uint32(versions[:4]); major != majorVersionNumber {
		fh.Close()
		return
	}

	iter := ss.Find([]byte{}, nil)

	for iter.Next() {
		foundKey := iter.Key()
		ttype := foundKey[0]

		if ttype == 1 {
			proto := int(foundKey[1])
			mutexProto.Lock()
			protocolSetCount[proto] += len(iter.Value()) / 4
			mutexProto.Unlock()
		} else if ttype == 2 {
			port := int(binary.BigEndian.Uint16([]byte{foundKey[1], foundKey[2]}))
			mutexPort.Lock()
			portSetCount[port] += len(iter.Value()) / 4
			mutexPort.Unlock()
		} else if ttype == 4 {
			ipv4 := net.IP{foundKey[1],
				foundKey[2],
				foundKey[3],
				foundKey[4]}
			mutexIP4.Lock()
			ipv4SetCount[ipv4.String()] += len(iter.Value()) / 4
			mutexIP4.Unlock()
		} else if ttype == 6 {
			ipv6 := net.IP{
				foundKey[1], foundKey[2], foundKey[3], foundKey[4],
				foundKey[5], foundKey[6], foundKey[7], foundKey[8],
				foundKey[9], foundKey[10], foundKey[11], foundKey[12],
				foundKey[13], foundKey[14], foundKey[15], foundKey[16],
			}
			mutexIP6.Lock()
			ipv6SetCount[ipv6.String()] += len(iter.Value()) / 4
			mutexIP6.Unlock()
		}
	}
	iter.Close()
	fh.Close()

	// get PKT0 filesize
	dataFileStat, err := os.Stat(fmt.Sprintf("%s/%s", dataFolderPath, filename))
	if err == nil {
		mutexSize.Lock()
		totalSize += dataFileStat.Size()
		mutexSize.Unlock()
	}
}

func main() {
	if len(os.Args) != 2 && len(os.Args) != 4 {
		log.Fatal("missing arguments")
	}

	folderPath := os.Args[1]
	dataFolderPath := strings.Replace(folderPath, "IDX0", "PKT0", 1)

	startDate := -1
	endDate := -1

	inputTimestampArgs := regexp.MustCompile(`^\d{10}$`)

	if len(os.Args) == 4 {
		if !inputTimestampArgs.MatchString(os.Args[2]) || !inputTimestampArgs.MatchString(os.Args[3]) {
			log.Fatal("wrong timestamp input")
		}

		startDate, _ = strconv.Atoi(os.Args[2])
		endDate, _ = strconv.Atoi(os.Args[3])

		if endDate <= startDate {
			log.Fatal("start timestamp needs to be smaller than end timestamp")
		}
	}

	files, err := ioutil.ReadDir(folderPath)
	if err != nil {
		log.Fatal(err)
	}

	jobs := make(chan string, 100)

	for w := 1; w <= concurrentWorkers; w++ {
		go worker(w, jobs, folderPath, dataFolderPath, startDate, endDate)
	}

	for _, file := range files {
		wg.Add(1)
		jobs <- file.Name()
	}

	close(jobs)
	wg.Wait()

	protocolsOut, _ := json.Marshal(protocolSetCount)
	portsOut, _ := json.Marshal(portSetCount)
	ipv4Out, _ := json.Marshal(ipv4SetCount)
	ipv6Out, _ := json.Marshal(ipv6SetCount)

	out := fmt.Sprintf(`{"totalSize": %d, "protocols":%s,"ports":%s,"ipv4":%s,"ipv6":%s}`,
		totalSize,
		string(protocolsOut),
		string(portsOut),
		string(ipv4Out),
		string(ipv6Out))

	fmt.Println(out)
}
