package tajarin

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"
)

// receive command

// parse command

// connect to producer

// send data

func Subscribe(req JsonTajarinRequest) {
	// Connect to the server
	conn, err := net.Dial("tcp", DefaultListenAddress)
	defer conn.Close()
	if err != nil {
		fmt.Println(err)
		return
	}

	// Marshall request
	reqj, err := json.Marshal(req)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Send some data to the server
	_, err = conn.Write(reqj)
	if err != nil {
		fmt.Println(err)
		return
	}

	buf := make([]byte, 2048)
	conn.SetReadDeadline(time.Now().Add(5 * time.Minute))
	// Read the incoming connection into the buffer.
	lenRead, err := conn.Read(buf)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(string(buf))
	MarshallEmpty(buf[:lenRead], req)
}

func MarshallEmpty(buf []byte, req JsonTajarinRequest) {
	// Unmarshal the buffer into a map to manipulate null values
	var data map[string]interface{}
	err := json.Unmarshal(buf, &data)
	if err != nil {
		fmt.Println("Error unmarshalling:", err)
		return
	}

	// Replace nil (null) values with empty strings
	for key, value := range data {
		if value == nil {
			data[key] = ""
		}
	}

	// Marshal the modified data back to JSON
	modifiedJSON, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling:", err)
		return
	}

	// Save the modified JSON to a file
	err = os.WriteFile(fmt.Sprintf("%s-output.json", req.Name), modifiedJSON, 0644)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}

	fmt.Println("JSON data saved successfully to output.json")
}
