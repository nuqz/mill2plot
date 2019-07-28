package main

import (
	"bufio"
	"bytes"
	"flag"
	"io"
	"log"
	"os"
	"path/filepath"
)

var (
	inputPath     = flag.String("if", "", "Path to input gcode file.")
	inputFilename = filepath.Base(*inputPath)
	outputPath    = flag.String("of", "plot.gc", "Path to output gcode file.")

	skipCommands = [][]byte{
		[]byte("M3 M8"),
		[]byte("M9 M5"),
		[]byte("F150"),
		[]byte("T1 M6"),
		[]byte("S6000"),

		// also skip comment lines
		[]byte("("),
	}

	ifd, ofd *os.File
	err      error
	rdwr     *bufio.ReadWriter
	inline   []byte
	outline  string
)

func init() {
	log.SetFlags(log.Lmicroseconds)
	flag.Parse()

	if *inputPath == "" {
		flag.Usage()
		log.Fatalln("Error: path to input file is not set.")
	}

	ifd, err = os.Open(*inputPath)
	if err != nil {
		log.Fatalln(err)
	}

	if *outputPath == "._plot" {
		flag.Usage()
		log.Fatal("Error: path to output file is not set.")
	}

	ofd, err = os.OpenFile(*outputPath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0660)
	if err != nil {
		log.Fatalln(err)
	}

	rdwr = bufio.NewReadWriter(bufio.NewReader(ifd), bufio.NewWriter(ofd))
}

func main() {
	defer func() {
		rdwr.Flush()
		ifd.Close()
		ofd.Close()
	}()

outer:
	for err != io.EOF {
		if inline, err = rdwr.ReadBytes('\n'); err != nil && err != io.EOF {
			log.Fatalln(err)
		} else if err == io.EOF {
			break
		}

		for _, skip := range skipCommands {
			if bytes.HasPrefix(inline, skip) {
				continue outer
			}
		}

		if len(inline) <= 1 {
			outline = ""
		} else if inline[3] != 'Z' {
			outline = string(inline)
		} else {
			outline = ""
		}

		if _, err = rdwr.WriteString(outline); err != nil {
			log.Fatalln(err)
		}
	}
}
