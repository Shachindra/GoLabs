package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("File Upload Endpoint Hit")

	// Parse our multipart form, 25 << 20 specifies a maximum of 25 MB file size
	const size25K = (1 << 10) * 25
	err := r.ParseMultipartForm(size25K)
	if err != nil {
		http.Error(w, "Error in Parsing Multipart Form: "+err.Error(), http.StatusInternalServerError)
		fmt.Println("Error in Parsing Multipart Form: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// FormFile returns the first file for the given name `fileSelection`
	// It also returns the FileHeader so we can get the Filename, Header and the Size of the file
	file, fileHeader, err := r.FormFile("fileSelection")
	if err != nil {
		http.Error(w, "Error parsing uploaded file: "+err.Error(), http.StatusBadRequest)
		fmt.Println("Error parsing uploaded file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create a file within our directory
	outfile, err := os.Create("./uploads/" + fileHeader.Filename)
	if err != nil {
		http.Error(w, "Unable to create the file for writing. Check your write access privileges. Error: "+err.Error(), http.StatusBadRequest)
		fmt.Println("Unable to create the file for writing. Check your write access privileges. Error: " + err.Error())
		return
	}
	defer outfile.Close()

	fmt.Printf("Uploaded File: %+v\n", fileHeader.Filename)
	fmt.Printf("File Size: %+v\n", fileHeader.Size)
	fmt.Printf("MIME Header: %+v\n", fileHeader.Header)

	// Write the content from POST to the file
	written, err := io.Copy(outfile, file)
	if err != nil {
		http.Error(w, "Error saving file: "+err.Error(), http.StatusInternalServerError)
		fmt.Println("Error saving file:" + err.Error())
	}

	// return that we have successfully uploaded our file!
	fmt.Fprintf(w, "Successfully Uploaded File\nLength :+length:"+strconv.Itoa(int(written)))
}

func setupRoutes() {
	// Handle File Upload Request - POST Only
	http.HandleFunc("/upload", uploadHandler)

	// Handle Any Request
	port := flag.String("p", "8080", "Port to serve on")
	directory := flag.String("d", "./httpdocs", "Directory of static HTML files to host")
	flag.Parse()

	fileServer := http.FileServer(http.Dir(*directory))
	http.Handle("/", fileServer)

	log.Printf("Serving %s on HTTP port: %s\n", *directory, *port)
	err := http.ListenAndServe(":"+*port, nil)
	if err != nil {
		panic(err)
	}
	log.Fatal(err)
}

func main() {
	setupRoutes()
}
