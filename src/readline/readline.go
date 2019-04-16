package readline

/*
// IMPORTANT: choose one
#cgo LDFLAGS: -ledit
//#cgo LDFLAGS: -lreadline // NOTE: libreadline is GPL

// free()
#include <stdlib.h>
// readline()
#include <stdio.h> // FILE *
#include <readline/readline.h>
// add_history()
#include <readline/history.h>
*/
import "C"

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"unsafe"
)

var histFile = ".mal-history"
var historyPath string

func init() {
	historyPath = filepath.Join(os.Getenv("HOME"), histFile)
	loadHistory(historyPath)
}

func loadHistory(filename string) error {
	content, err := ioutil.ReadFile(historyPath)
	if err != nil {
		return err
	}

	for _, addLine := range strings.Split(string(content), "\n") {
		if addLine == "" {
			continue
		}
		cAddLine := C.CString(addLine)
		C.add_history(cAddLine)
		C.free(unsafe.Pointer(cAddLine))
	}

	return nil
}

// Readline will read in a single line of text with the provided prompt
func Readline(prompt string) (string, error) {
	cPrompt := C.CString(prompt)
	defer C.free(unsafe.Pointer(cPrompt))
	cLine := C.readline(cPrompt)
	defer C.free(unsafe.Pointer(cLine))
	if cLine == nil {
		return "", errors.New("C.readline call failed")
	}
	C.add_history(cLine)
	line := C.GoString(cLine)
	// append to file
	f, e := os.OpenFile(historyPath, os.O_APPEND|os.O_WRONLY, 0600)
	if e == nil {
		defer f.Close()
		_, e = f.WriteString(line + "\n")
		if e != nil {
			fmt.Printf("error writing to history")
		}
	}
	return line, nil
}
