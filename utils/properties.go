package utils

import (
	"bufio"
	"log"
	"os"
	"strings"
)

var AppProperties map[string]string

//scanning and setting up properties from app.properties file
func GetAppPropeties() (map[string]string, error) {
	AppProperties := make(map[string]string)
	
	file, err := os.Open("app.properties")
	
	if err != nil{
		log.Fatal("error while reading file")
		log.Fatal(err)
		log.Fatal("error while reading file")
		return nil, err

	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan(){
		line := scanner.Text()
		if equal := strings.Index(line, "=") ; equal > 0{
            props := strings.Split(line, "=")
		    AppProperties[strings.TrimSpace(props[0])] = strings.TrimSpace(props[1])
		}		
		
	}

	if err := scanner.Err() ; err != nil{
		log.Fatal("error while scanning file")
		return nil, err
	}


	return AppProperties, nil

}