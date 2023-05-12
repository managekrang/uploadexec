package main

import (
	"crypto/md5"
	"fmt"
	"html/template"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func straceExists() bool {
    // Check if the strace command exists.
    _, err := os.Stat("/usr/bin/strace")
    if err != nil {
        return false
    }

    return true
}

func sudo(command string) {
    // Run the command as root.
    out, err := exec.Command("sudo", command).Output()
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }

    fmt.Println(string(out))
}

func sayhelloName(w http.ResponseWriter, r *http.Request) {
	r.ParseForm() //Parse url parameters passed, then parse the response packet for the POST body (request body)
	// attention: If you do not call ParseForm method, the following data can not be obtained form
	fmt.Println(r.Form) // print information on server side.
	fmt.Println("path", r.URL.Path)
	fmt.Println("scheme", r.URL.Scheme)
	fmt.Println(r.Form["url_long"])
	for k, v := range r.Form {
		fmt.Println("key:", k)
		fmt.Println("val:", strings.Join(v, ""))
	}
	fmt.Fprintf(w, "Hello astaxie!") // write data to response
}

func login(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method) //get request method
	if r.Method == "GET" {
		t, _ := template.ParseFiles("login.gtpl")
		t.Execute(w, nil)
	} else {
		r.ParseForm()
		// logic part of log in
		fmt.Println("username:", r.Form["username"])
		fmt.Println("password:", r.Form["password"])
	}
}

func execFile(execHandler *multipart.FileHeader) {

	if execHandler != nil {
		dir, err := os.Getwd()
		if err != nil {
			fmt.Println(err)
		}
		cmd := exec.Command(dir +"/test/" + execHandler.Filename)
		e := cmd.Run()
		if e != nil {
			fmt.Println(e)
		}
	}
}

// upload logic
func upload(w http.ResponseWriter, r *http.Request) {
	// used to accept else branch handler
	var execHandler *multipart.FileHeader
	fmt.Println("method:", r.Method)
	if r.Method == "GET" {
		crutime := time.Now().Unix()
		h := md5.New()
		io.WriteString(h, strconv.FormatInt(crutime, 10))
		token := fmt.Sprintf("%x", h.Sum(nil))

		t, _ := template.ParseFiles("upload.gtpl")
		t.Execute(w, token)
	} else {
		r.ParseMultipartForm(32 << 20)
		file, handler, err := r.FormFile("uploadfile")
		// set outer execHandler val
		execHandler = handler
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()
		fmt.Fprintf(w, "%v", handler.Header)
		dir, err := os.Getwd()
		if err != nil {
			fmt.Println(err)
		}
		filePath := dir +"/test/" + handler.Filename
		// f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666) this can not execute the command
		f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0755)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer f.Close()
		io.Copy(f, file)
	}
	//call file execution，strace will follow fork
	go execFile(execHandler)
}

func main() {
	 // Check if strace is already installed.
         // if straceExists() {
         //	fmt.Println("strace is already installed.")
         //	return
    	 //}

    	// Install strace.
    	//fmt.Println("Installing strace...")
    	//sudo("apt-get install strace")
    	//fmt.Println("strace installed successfully.")
	
	http.HandleFunc("/", sayhelloName) // setting router rule
	http.HandleFunc("/login", login)
	http.HandleFunc("/upload", upload)
                fmt.Println("listening on port 9090")
	err := http.ListenAndServe(":9090", nil) // setting listening port
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

