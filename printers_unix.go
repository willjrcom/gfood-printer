//go:build !windows
// +build !windows

package main

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
)

func printToOS(printerName, content string) error {
	var cmd *exec.Cmd

	// Se for "default", usa impressora padrão do sistema
	if printerName == "default" || printerName == "" {
		defaultPrinter, err := getDefaultPrinter()
		if err != nil {
			// Se não conseguir obter o nome, usa lp -o raw sem -d (imprime na padrão automaticamente em modo raw)
			log.Printf("Aviso: não foi possível obter nome da impressora padrão (%v), usando lp -o raw sem especificar impressora (usará padrão do sistema)", err)
			cmd = exec.Command("lp", "-o", "raw")
		} else {
			log.Printf("Impressora padrão detectada: %s (Raw mode)", defaultPrinter)
			cmd = exec.Command("lp", "-d", defaultPrinter, "-o", "raw")
		}
	} else {
		// Impressora específica em modo raw
		log.Printf("Imprimindo em [%s] (Raw mode)", printerName)
		cmd = exec.Command("lp", "-d", printerName, "-o", "raw")
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	if _, err := stdin.Write([]byte(content)); err != nil {
		stdin.Close()
		return err
	}
	stdin.Close()
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("erro ao imprimir via lp: %v | stderr: %s", err, stderr.String())
	}
	return nil
}
