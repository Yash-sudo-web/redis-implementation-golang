package db

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

func LoadRDBFile() {
	dir := Rdbconfig["dir"]
	filename := Rdbconfig["dbfilename"]

	if dir == "" || filename == "" {
		fmt.Println("RDB file directory or filename is not set.")
		return
	}

	path := dir + "/" + filename
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("Error opening RDB file: %s\n", err.Error())
		return
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	// Validate header
	header := make([]byte, 9)
	_, err = io.ReadFull(reader, header)
	fmt.Printf("Header: %s\n", string(header))
	if err != nil || string(header) != "REDIS0011" {
		fmt.Println("Invalid RDB file header")
		return
	}

	var buffer []byte // Buffer to hold data between markers
	for {
		byteRead, err := reader.ReadByte()
		if err != nil {
			if err == io.EOF {
				break // End of file
			}
			fmt.Printf("Error reading byte: %v\n", err)
			return
		}

		if byteRead == 0xFA { // Marker 0xFA found
			if len(buffer) > 0 {
				fmt.Printf("Start of metadata section: %s\n", string(buffer))
				buffer = nil // Reset buffer
			}
		} else if byteRead == 0xFE { // End of metadata section, Marker for Database Section
			if len(buffer) > 0 {
				fmt.Printf("Start of metadata section: %s\n", string(buffer))
				buffer = nil // Reset buffer
				fmt.Println("End of metadata section")
				break
			}
		} else {
			buffer = append(buffer, byteRead)
		}
	}

	for {
		marker, err := reader.ReadByte()
		if err != nil {
			if err == io.EOF {
				break // End of file, exit the loop
			}
			fmt.Printf("Error reading marker: %v\n", err)
			return
		}
		var hashTableSizeNumber int
		if marker == 0x00 {
			nextMarker, err := reader.ReadByte()
			if err != nil {
				fmt.Printf("Error reading next marker: %v\n", err)
				return
			}
			var hashTableExpirySize byte
			if nextMarker == 0xfb { // Indicates that hash table size information follows
				fmt.Println("Hash table size information follows:")
				hashTableSize, err := reader.ReadByte()
				hashTableSizeNumber = int(hashTableSize)
				if err != nil {
					fmt.Printf("Error reading hash table size: %v\n", err)
					return
				}
				fmt.Printf("Hash table size: %d\n", hashTableSize)
				hashTableExpirySize, err = reader.ReadByte()
				if err != nil {
					fmt.Printf("Error reading hash table expiry size: %v\n", err)
					return
				}
				fmt.Printf("Hash table expiry size: %d\n", hashTableExpirySize)
			}
			expTimeBuffer := make([]byte, 8)
			nextMarker, err = reader.ReadByte()
			var expTimeUnixMsUInt64 uint64
			var expTimeUnixMs int64
			if nextMarker == 0xFC {
				_, err = reader.Read(expTimeBuffer)
				nextMarker, err = reader.ReadByte()
				expTimeUnixMsUInt64 = binary.LittleEndian.Uint64(expTimeBuffer)
				expTimeUnixMs = int64(expTimeUnixMsUInt64)
			}
			if nextMarker == 0x0 {
				nextMarker, err = reader.ReadByte()
				if err != nil {
					fmt.Printf("Error reading next marker: %v\n", err)
					return
				}
				if nextMarker == 0x9 {
					fmt.Printf("Horizontal Tab Found \n")
				}
				var elementCount int = 1
				for {
					if nextMarker == 0xFC {
						expTimeBuffer := make([]byte, 8)
						_, err = reader.Read(expTimeBuffer)
						nextMarker, err = reader.ReadByte()
						nextMarker, err = reader.ReadByte()
						expTimeUnixMsUInt64 = binary.LittleEndian.Uint64(expTimeBuffer)
						expTimeUnixMs = int64(expTimeUnixMsUInt64)
					}
					nextMarker, err = reader.ReadByte()
					if nextMarker == 0x09 || nextMarker == 0x06 || nextMarker == 0x05 || nextMarker == 0x04 || nextMarker == 0x03 || nextMarker == 0x0A {
						var key string
						// var value string
						key = string(buffer)
						buffer = nil
						fmt.Printf("Key: %s\n", key)
						var valueMarker []byte
						if hashTableSizeNumber != elementCount {
							for {

								nextMarker, err = reader.ReadByte()
								if nextMarker == 0x0 || nextMarker == 0xFC {
									break
								}
								valueMarker = append(valueMarker, nextMarker)
							}
						} else {
							valueMarker, err = reader.ReadBytes(0xFF)
							valueMarker = valueMarker[:len(valueMarker)-1]
						}

						if err != nil {
							fmt.Printf("Error reading for marker: %v\n", err)
							return
						}
						value := string(valueMarker)
						fmt.Printf("Value: %s\n", value)

						fmt.Printf("Expiration Time (Unix ms): %d\n", expTimeUnixMs)

						Db[key] = map[string]interface{}{
							"value":   string(value),
							"expTime": expTimeUnixMs,
						}
						expTimeBuffer = nil
						if hashTableSizeNumber == elementCount {
							break
						}
						elementCount++
						if nextMarker == 0xFC {
							continue
						}
						nextMarker, err = reader.ReadByte()
					} else {
						buffer = append(buffer, nextMarker)
					}
				}
			}
		}
	}
}
