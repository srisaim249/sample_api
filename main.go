package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sample_api/config"
	"sample_api/connections"
	"sample_api/file"
	"sample_api/globals"
	"sample_api/handlers/userHandler"
	logger "sample_api/log"
	"strconv"
	"syscall"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// main ... start of the application process
func main() {
	input := os.Args //reading the inputs from user

	if len(input) < 2 {
		log.Println("please provide the config file information")
		os.Exit(0)
	}

	//configurations taking from the input json file
	filePath, err := file.GetFilePathFromArgs(input)
	if err != nil {
		log.Println("error while fetching the file path", err)
		os.Exit(0)
	}
	//storing json file information to global object
	err = config.ReadConfigFromFile(filePath)
	if err != nil {
		log.Printf("Error occured while reading the configuration file,err:%+v\n", err.Error())
		os.Exit(0)
	}
	_, filePath, lineNo, _ := runtime.Caller(0)
	logger.InsertMsgIntoLogFile(filePath, strconv.Itoa(lineNo+1), "info", "creating db Connections")

	//creating the db connections , if it fails then stops the application
	err = connections.PrepareUserDbConnection()
	if err != nil {
		_, filePath, lineNo, _ := runtime.Caller(0)
		logger.InsertMsgIntoLogFile(filePath, strconv.Itoa(lineNo+1), "error", fmt.Sprintf("Error occured while preparing the connections:%s", err.Error()))
		os.Exit(0)
	}

	/* STEP - 2 */
	// Start listening for endpoint requests
	_, filePath, lineNo, _ = runtime.Caller(0)
	logger.InsertMsgIntoLogFile(filePath, strconv.Itoa(lineNo+1), "info", "starting HTTP Server")

	listenandServe()

	_, filePath, lineNo, _ = runtime.Caller(0)
	logger.InsertMsgIntoLogFile(filePath, strconv.Itoa(lineNo+1), "info", ("Closing the Application"))

}

// listenAndServe ... listen on the specified port and serve the request from the specified address
func listenandServe() {
	_, filePath, lineNo, _ := runtime.Caller(0)
	logger.InsertMsgIntoLogFile(filePath, strconv.Itoa(lineNo+1), "info", fmt.Sprintf("Application is running on port %s", globals.CONFIG_OBJ.APP_CONFIG.PORT))

	router := mux.NewRouter()

	//user sign up
	router.Handle("/userLogin", http.HandlerFunc(userHandler.UserLoginHandler)).Methods(http.MethodPost)
	//user registration
	router.Handle("/userRegistration", http.HandlerFunc(userHandler.UserRegistrationHandler)).Methods(http.MethodPost)

	//creating the http server
	server := &http.Server{
		Addr: globals.CONFIG_OBJ.APP_CONFIG.PORT,
		Handler: handlers.CORS(
			handlers.AllowedMethods([]string{"POST"}),
			handlers.AllowedOrigins([]string{"*"}))(router),
	}

	//safeExit ... for the ctrl+c singnal handling
	safeExit(server)

	_, filePath, lineNo, _ = runtime.Caller(0)
	logger.InsertMsgIntoLogFile(filePath, strconv.Itoa(lineNo+1), "info", fmt.Sprintf("Application:%s is ready", globals.CONFIG_OBJ.APP_CONFIG.APP_NAME))

	if err := server.ListenAndServe(); err != nil {
		_, filePath, lineNo, _ = runtime.Caller(0)
		logger.InsertMsgIntoLogFile(filePath, strconv.Itoa(lineNo+1), "info", fmt.Sprintf("Application :%s is closing", globals.CONFIG_OBJ.APP_CONFIG.APP_NAME))
		fmt.Println(err)
	}
}

// safeExit  ... for the ctrl+c singnal handling
func safeExit(server *http.Server) {
	signalChannel := make(chan os.Signal, 1) //signalChannel ... holds the signal when user interrupts
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	go func() { //check for user interruption
		_, ok := <-signalChannel
		if ok { //when we got signal from user , then userinterrupted flag will be true, by default it is false
			globals.USER_INTERRUPTED = true
			err := globals.USER_DB_CONN.Close()
			if err != nil {
				_, filePath, lineNo, _ := runtime.Caller(0)
				logger.InsertMsgIntoLogFile(filePath, strconv.Itoa(lineNo+1), "error", fmt.Sprintf("error occured while closing the connections,err:%+v", err))
			}
			//gracefully shuts down the server without interrupting any active connections
			err = server.Shutdown(context.Background())
			if err != nil {
				log.Println(fmt.Sprint(err))
			}
			os.Exit(0)
		}
	}()
}
