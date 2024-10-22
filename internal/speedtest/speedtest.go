package main

import (
	"bufio"
	"fmt"
	"github.com/schollz/progressbar/v3"
	"io"
	"net/http"
	"os"
	"time"
)

func main() {
	resp, err := http.Get(`https://nbg1-speed.hetzner.com/100MB.bin`)
	if err != nil {
		panic(err)
	}
	fmt.Println("sent request")
	defer resp.Body.Close()

	buffer := bufio.NewReaderSize(resp.Body, 1_000_000)

	f, err := os.CreateTemp("", "speedtest")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	defer os.Remove(f.Name())

	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		"downloading",
	)

	fmt.Println("started download")
	start := time.Now()
	fmt.Println(io.Copy(io.MultiWriter(f, bar), buffer))
	//fmt.Println(io.Copy(f, buffer))
	fmt.Println(time.Since(start))
}
