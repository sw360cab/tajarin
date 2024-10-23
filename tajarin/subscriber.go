package tajarin

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	"go.uber.org/zap"
)

type TajarinSubscriber struct {
	logger        *zap.Logger
	listenAddress string
}

func (ts *TajarinSubscriber) Subscribe(req JsonTajarinRequest, listenAddress string, logger *zap.Logger) error {
	ts.logger = logger

	// Connect to the server
	conn, err := net.Dial("tcp", listenAddress)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Marshall request
	reqj, err := json.Marshal(req)
	if err != nil {
		return err
	}

	// Sending Data
	_, err = conn.Write(reqj)
	if err != nil {
		return err
	}

	buf := make([]byte, 2048)
	conn.SetReadDeadline(time.Now().Add(5 * time.Minute))
	// Read the incoming connection into the buffer.
	lenRead, err := conn.Read(buf)
	if err != nil {
		return err
	}
	return ts.MarshallSuppressEmptyFields(buf[:lenRead], req)
}

func (ts *TajarinSubscriber) MarshallSuppressEmptyFields(buf []byte, req JsonTajarinRequest) error {
	outputFilename := fmt.Sprintf("%s-output.json", req.Name)
	// Unmarshal the buffer into a map to manipulate null values
	var data map[string]interface{}
	err := json.Unmarshal(buf, &data)
	if err != nil {
		ts.logger.Sugar().Error("Error marshalling:", err)
		return err
	}

	// Replace nil (null) values with empty strings
	for key, value := range data {
		if value == nil {
			data[key] = ""
		}
	}

	// Marshal the modified data back to JSON
	modifiedJSON, err := json.Marshal(data)
	if err != nil {
		ts.logger.Sugar().Error("Error marshalling:", err)
		return err
	}

	// Save the modified JSON to a file
	err = os.WriteFile(outputFilename, modifiedJSON, 0644)
	if err != nil {
		ts.logger.Sugar().Error("Error marshalling:", err)
		return err
	}

	ts.logger.Sugar().Info("JSON data saved successfully to %s", outputFilename)
	return nil
}
