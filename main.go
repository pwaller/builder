package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

func dockerBuild(checkoutPath, name, tag string) error {
	repository := fmt.Sprintf("%s:%s", name, tag)
	cmd := Command(checkoutPath, "docker", "build", "--tag", repository, ".")
	return cmd.Run()
}

func dockerSave(name string, w io.Writer) error {
	cmd := Command(".", "docker", "save", name)
	cmd.Stdout = w
	return cmd.Run()
}

func dockerPush(name, tag string) error {
	repository := fmt.Sprintf("%s:%s", name, tag)
	cmd := Command(".", "docker", "push", repository)
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

		checkoutPath := "c/" + shortRev

		err = gitCheckout(path, checkoutPath, rev)
		if err != nil {
			log.Printf("Failed to checkout: %v", err)
			return
		}

		tagName, err := gitDescribe(path, rev)
		if err != nil {
			log.Printf("Unable to describe %v: %v", rev, err)
			return
		}

		log.Println("Checked out")

		// dockerImage := name + "-" + shortRev

		repoName := "localhost.localdomain:5000/" + name

		err = dockerBuild(path+"/"+checkoutPath, repoName, tagName)
		if err != nil {
			log.Printf("Failed to build: %v", err)
		}

		start := time.Now()
		err = dockerPush(repoName, tagName)
		log.Printf("Took %v to push", time.Since(start))

		if err != nil {
			log.Printf("Failed to push: %v", err)
		}

		// var buf bytes.Buffer
		// start := time.Now()
		// dockerSave(dockerImage, &buf)
		// log.Printf("Took %v to save %v bytes", time.Since(start), buf.Len())

		fmt.Fprintln(w, "Success:", rev)
	})

	http.HandleFunc("/ws/", serveWs)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
