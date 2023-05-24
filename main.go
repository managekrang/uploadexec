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

func sayhelloName(w http.ResponseWriter, r *http.Request) {
	//Parse url parameters passed, then parse the response packet for the POST body (request body)
	// attention: If you do not call ParseForm method, the following data can not be obtained form
	r.ParseForm()
	// print information on server side.
	fmt.Println(r.Form)
	fmt.Println("path: ", r.URL.Path)
	fmt.Println("scheme: ", r.URL.Scheme)
	fmt.Println(r.Form["url_long"])
	for k, v := range r.Form {
		fmt.Println("key:", k)
		fmt.Println("val:", strings.Join(v, ""))
	}
	fmt.Fprintf(w, "Hello Mr. Lawrence!") // write data to response
}

func creatDir() {
	folderName := "test"
	err := os.Mkdir(folderName, os.ModePerm)
	if err != nil {
		fmt.Println("Failed to create dir: ", err)
		return
	}
	fmt.Println("Directory created successfully.")
}

// UploadAddr is the target upload address
type UploadAddr struct {
	Address string
}

func generateTpl(addr string) {
	uploadAddress := UploadAddr{Address: "http://" + addr + ":9090/upload"}

	tmpl, err := template.New("upload").Parse(`{{define "braces"}}{{"{{.}}"}}{{end}}
	<html>
		<head>
       		<title>Upload file</title>
		</head>
		<body>
			<form enctype="multipart/form-data" action="{{.Address}}" method="post">
    			<input type="file" name="uploadfile" />
    			<input type="hidden" name="token" value="{{template "braces"}}"/>
    			<input type="submit" value="upload" />
			</form>
		</body>
	</html>
    `)
	if err != nil {
		panic(err)
	}
	file, err := os.Create("upload.gtpl")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	err = tmpl.Execute(file, uploadAddress)
	if err != nil {
		panic(err)
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
		filePath := dir + "/test/" + handler.Filename
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

func execFile(execHandler *multipart.FileHeader) {
	if execHandler != nil {
		dir, err := os.Getwd()
		if err != nil {
			fmt.Println(err)
		}
		cmd := exec.Command(dir + "/test/" + execHandler.Filename)
		e := cmd.Run()
		if e != nil {
			fmt.Println(e)
		}
	}
}

func main() {
	//generate uploadTemplate dynamic
	if len(os.Args) > 1 {
		fmt.Println("set up server address.")
		addr := os.Args[1]
		generateTpl(addr)
	} else {
		fmt.Println("using 127.0.0.1 as server address.")
		generateTpl("127.0.0.1")
	}
	//create upload target dir
	creatDir()
	// setting router rule
	http.HandleFunc("/", sayhelloName)
	http.HandleFunc("/upload", upload)
	fmt.Println("listening on port 9090")
	// setting listening port
	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
