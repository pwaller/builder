package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

func dockerBuild(checkoutPath, name string) error {
	cmd := Command(checkoutPath, "docker", "build", "--tag", name, ".")
	return cmd.Run()
}

func dockerSave(name string, w io.Writer) error {
	cmd := Command(".", "docker", "save", name)
	cmd.Stdout = w
	return cmd.Run()
}

func main() {

	http.HandleFunc("/build/", func(w http.ResponseWriter, r *http.Request) {
		target := r.URL.Path[len("/build/"):]

		name := path.Base(target)

		remote := fmt.Sprintf("https://%v.git", target)

		if _, useSSH := r.URL.Query()["ssh"]; useSSH {
			split := strings.SplitN(target, "/", 2)
			domain, repo := split[0], split[1]
			remote = "git@" + domain + ":" + repo
		}

		path := "./src/" + target

		gitLocalMirror(remote, path, os.Stderr)

		rev, err := gitRevParse(path, "HEAD")
		if err != nil {
			log.Printf("Unable to parse rev: %v", err)
			return
		}

		shortRev := rev[:10]

		fmt.Fprintln(w, "Here:", rev)

		checkoutPath := "c/" + shortRev

		err = gitCheckout(path, checkoutPath, rev)
		if err != nil {
			log.Printf("Failed to checkout: %v", err)
			return
		}

		log.Println("Checked out")

		dockerImage := name + "-" + shortRev

		err = dockerBuild(path+"/"+checkoutPath, dockerImage)
		if err != nil {
			log.Printf("Failed to build: %v", err)
		}

		var buf bytes.Buffer

		start := time.Now()
		dockerSave(dockerImage, &buf)
		log.Printf("Took %v to save %v bytes", time.Since(start), buf.Len())

	})

	http.HandleFunc("/ws/", serveWs)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
