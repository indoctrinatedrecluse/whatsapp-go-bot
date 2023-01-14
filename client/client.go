package client

import (
	"encoding/gob"
	"fmt"
	qrcodeTerminal "github.com/Baozisoftware/qrcode-terminal-go"
	"github.com/Rhymen/go-whatsapp"
	"os"
	"time"
)

/** ------------- globals ------------- */

// DeclaredClientCount function to keep static count of declared clients
func DeclaredClientCount() (f func() int) {
	var clientCount int
	f = func() int {
		clientCount++
		// fmt.Println(clientCount)
		return clientCount
	}
	return
}

/** ------------- globals ------------- */

/** ------------- utilities ------------- */

// GetConnection return whatsapp connection
func (client *WhatsappClientServitor) GetConnection() *whatsapp.Conn {
	return client.wac
}

// readSession cache previously saved session login data (if any)
func restoreSession() (whatsapp.Session, error) {
	session := whatsapp.Session{}
	file, err := os.Open(os.TempDir() + "/whatsappSession.gob")
	if err != nil {
		return session, err
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)
	// decode session data
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&session)
	if err != nil {
		return session, err
	}
	return session, nil
}

// saveSession save the login detail from the current session
func saveSession(session whatsapp.Session) error {
	file, err := os.Create(os.TempDir() + "/whatsappSession.gob")
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)
	// encode as session details must be encrypted
	encoder := gob.NewEncoder(file)
	err = encoder.Encode(session)
	if err != nil {
		return err
	}
	return nil
}

// WebQRLogin try to perform web QR login on a given client handle
func (client *WhatsappClientServitor) WebQRLogin() error {
	// get the connection
	wac := client.GetConnection()
	//load saved session, if any (go-whatsapp saves login sessions)
	session, err := restoreSession()
	// if we have found a saved session, restore it
	if err == nil {
		//restore session
		session, err = wac.RestoreWithSession(session)
		if err != nil {
			return fmt.Errorf("failed session restore: %v", err)
		}
	} else {
		// if we have not found a saved session, attempt regular login
		qr := make(chan string)
		go func() {
			terminal := qrcodeTerminal.New()
			terminal.Get(<-qr).Print()
		}()
		session, err = wac.Login(qr)
		if err != nil {
			return fmt.Errorf("failed login: %v", err)
		}
	}

	// save the current session's login details
	err = saveSession(session)
	if err != nil {
		return fmt.Errorf("error saving session: %v", err)
	}
	return nil
}

/** ------------- utilities ------------- */

/** ------------- initializers ------------- */

// WhatsappClientServitor define the structure of the whatsapp client interface
type WhatsappClientServitor struct {
	wac *whatsapp.Conn
}

// DeclareClient initializes a new client
func DeclareClient() *WhatsappClientServitor {
	wac, err := whatsapp.NewConn(10 * time.Second) // keep 10 second timeout duration
	if err != nil {
		_, err := fmt.Fprintf(os.Stderr, "error creating connection: %v\n", err)
		if err != nil {
			return nil
		}
		panic(err)
		return nil
	}

	err = wac.SetClientName("indoctrinated-recluse's Whatsapp Web API Client", "Whatsapp API", "")
	if err != nil {
		return nil
	}
	wac.SetClientVersion(0, 4, 1307)

	servitor := WhatsappClientServitor{wac}
	err = servitor.WebQRLogin()
	if err != nil {
		_, err := fmt.Fprintf(os.Stderr, "error logging in: %v\n", err)
		if err != nil {
			return nil
		}
		os.Exit(1)
		return nil
	}

	fmt.Println("Whatsapp Connected!")
	return &servitor
}

/** ------------- initializers ------------- */
