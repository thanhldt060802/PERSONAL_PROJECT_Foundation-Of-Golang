package esl

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/luandnh/eslgo"
	"github.com/luandnh/eslgo/command"
)

type ESLConfig struct {
	Address  string
	Port     int
	Password string
}

func (config *ESLConfig) Connect() (*eslgo.Conn, error) {
	address := fmt.Sprintf("%v:%d", config.Address, config.Port)
	log.Printf("INFO --- Attempting to connect to ESL at %v", address)

	conn, err := eslgo.Dial(address, config.Password, func() {
		log.Printf("WARNING --- ESL disconnected from %v", address)
	})
	if err != nil {
		log.Printf("ERROR --- Connect to ESL at %v failed: %v", address, err.Error())
		return nil, fmt.Errorf("connect to ESL at %v failed: %v", address, err.Error())
	}

	log.Printf("INFO --- Connect to ESL at %v successful", address)
	return conn, nil
}

func API(conn *eslgo.Conn, cmd string) (response *eslgo.RawResponse, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Printf("INFO --- Sending ESL api command: %v", cmd)

	return conn.SendCommand(ctx, command.API{
		Background: false,
		Command:    cmd,
		Arguments:  "",
	})
}

func BgAPI(conn *eslgo.Conn, cmd string) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Printf("INFO --- Sending ESL bgapi command: %v", cmd)

	rawResponse, err := conn.SendCommand(ctx, command.API{
		Background: true,
		Command:    cmd,
		Arguments:  "",
	})
	if err != nil {
		log.Printf("ERROR --- Send ESL bgapi command %v failed: %v", cmd, err.Error())
		return err
	}

	log.Printf("INFO --- Received ESL bgapi response for %v: %v", cmd, string(rawResponse.Body))

	return
}
