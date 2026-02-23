//go:build windows
// +build windows

package main

import (
	"fmt"
	"log"
	"syscall"
	"unsafe"
)

func printToOS(printerName, content string) error {
	var printerToUse string

	// Se for "default", obtém a impressora padrão
	if printerName == "default" {
		defaultPrinter, err := getDefaultPrinter()
		if err != nil {
			return fmt.Errorf("não foi possível obter impressora padrão: %v", err)
		}
		printerToUse = defaultPrinter
		log.Printf("Impressora padrão detectada: %s (Raw mode)", printerToUse)
	} else {
		printerToUse = printerName
		log.Printf("Imprimindo em [%s] (Raw mode Windows)", printerToUse)
	}

	return printRawWindows(printerToUse, content)
}

// printRawWindows usa a API winspool.drv para enviar bytes brutos para a impressora
func printRawWindows(printerName, content string) error {
	winspool := syscall.NewLazyDLL("winspool.drv")
	openPrinter := winspool.NewProc("OpenPrinterW")
	startDocPrinter := winspool.NewProc("StartDocPrinterW")
	startPagePrinter := winspool.NewProc("StartPagePrinter")
	writePrinter := winspool.NewProc("WritePrinter")
	endPagePrinter := winspool.NewProc("EndPagePrinter")
	endDocPrinter := winspool.NewProc("EndDocPrinter")
	closePrinter := winspool.NewProc("ClosePrinter")

	var hPrinter uintptr
	printerNamePtr, _ := syscall.UTF16PtrFromString(printerName)
	ret, _, err := openPrinter.Call(uintptr(unsafe.Pointer(printerNamePtr)), uintptr(unsafe.Pointer(&hPrinter)), 0)
	if ret == 0 {
		return fmt.Errorf("falha ao abrir impressora [%s]: %v", printerName, err)
	}
	defer closePrinter.Call(hPrinter)

	type DocInfo1 struct {
		DocName    *uint16
		OutputFile *uint16
		Datatype   *uint16
	}

	docNamePtr, _ := syscall.UTF16PtrFromString("GFood Print Job")
	dataTypePtr, _ := syscall.UTF16PtrFromString("RAW")
	di1 := DocInfo1{
		DocName:    docNamePtr,
		OutputFile: nil,
		Datatype:   dataTypePtr,
	}

	ret, _, err = startDocPrinter.Call(hPrinter, 1, uintptr(unsafe.Pointer(&di1)))
	if ret == 0 {
		return fmt.Errorf("falha ao iniciar documento: %v", err)
	}
	defer endDocPrinter.Call(hPrinter)

	ret, _, err = startPagePrinter.Call(hPrinter)
	if ret == 0 {
		return fmt.Errorf("falha ao iniciar página: %v", err)
	}
	defer endPagePrinter.Call(hPrinter)

	var written uint32
	contentBytes := []byte(content)
	ret, _, err = writePrinter.Call(hPrinter, uintptr(unsafe.Pointer(&contentBytes[0])), uintptr(len(contentBytes)), uintptr(unsafe.Pointer(&written)))
	if ret == 0 {
		return fmt.Errorf("falha ao escrever na impressora: %v", err)
	}

	return nil
}
