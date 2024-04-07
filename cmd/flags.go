package main

import "fmt"

func help() {
	text := `
	__________  ____  ____ 
	/_  __/ __ \/ __ \/ __ \
	 / / / / / / / / / / / /
	/ / / /_/ / /_/ / /_/ / 
   /_/  \____/_____/\____/  

	go-todo is a minimalistic todo REST api written in go.

	The following configuration options can be provided through CLI flags:
	--env : the name of the environment the server is being run on (eg. prod)
	--port : the port number (should be between 1024 and 65535)
	 --log_level : the minimum logging level (should be one of the values listed below)
		Available values: -4 (debug), 0 (info), 4 (warn), 8 (error)
	 --log_output : the path to the logger's output file
		default :  stdout 
	--secret : jwt secret key, used in signing and validating jwt tokens
	--token_exp : duration of jwt validity (should be a number)
		default :  30 minutes
	--conn_str : database connection string
	--payload_size : maximum permitted payload size (should be a number)
		default : 1 Mb
	`
	fmt.Println(text)
}
