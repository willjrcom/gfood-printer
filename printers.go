package main

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"
)

func getPrinters() ([]string, error) {
	switch runtime.GOOS {
	case "darwin", "linux":
		return getPrintersCUPS()
	case "windows":
		return getPrintersWindows()
	default:
		return nil, fmt.Errorf("sistema operacional não suportado: %s", runtime.GOOS)
	}
}

func getPrintersCUPS() ([]string, error) {
	cmd := exec.Command("lpstat", "-p")

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("erro ao executar lpstat: %v | stderr: %s", err, stderr.String())
	}

	lines := strings.Split(out.String(), "\n")
	printers := []string{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// cobre saídas em pt/eng
		if strings.HasPrefix(line, "printer ") || strings.HasPrefix(line, "impressora ") {
			parts := strings.Fields(line)
			if len(parts) > 1 {
				printers = append(printers, parts[1])
			}
		}
	}
	return printers, nil
}

func getPrintersWindows() ([]string, error) {
	cmd := exec.Command("powershell", "Get-Printer | Select-Object -ExpandProperty Name")

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("erro ao executar PowerShell: %v | stderr: %s", err, stderr.String())
	}

	lines := strings.Split(out.String(), "\n")
	printers := []string{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			printers = append(printers, line)
		}
	}
	return printers, nil
}

// getDefaultPrinter retorna o nome da impressora padrão do sistema
func getDefaultPrinter() (string, error) {
	switch runtime.GOOS {
	case "darwin", "linux":
		// No macOS/Linux com CUPS, usa lpoptions para obter a impressora padrão
		cmd := exec.Command("lpstat", "-d")
		var out bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("erro ao obter impressora padrão: %v | stderr: %s", err, stderr.String())
		}

		output := strings.TrimSpace(out.String())
		// Formato esperado: "system default destination: NomeDaImpressora"
		if strings.HasPrefix(output, "system default destination:") || strings.HasPrefix(output, "destino padrão de sistema:") {
			parts := strings.SplitN(output, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1]), nil
			}
		}
		return "", fmt.Errorf("não foi possível determinar impressora padrão")

	case "windows":
		// No Windows, usa PowerShell para obter a impressora padrão
		cmd := exec.Command("powershell", "Get-Printer | Where-Object {$_.Default -eq $true} | Select-Object -ExpandProperty Name")
		var out bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("erro ao obter impressora padrão: %v | stderr: %s", err, stderr.String())
		}

		printer := strings.TrimSpace(out.String())
		if printer == "" {
			return "", fmt.Errorf("nenhuma impressora padrão encontrada")
		}
		return printer, nil

	default:
		return "", fmt.Errorf("sistema operacional não suportado: %s", runtime.GOOS)
	}
}

func printToOS(printerName, content string) error {
	switch runtime.GOOS {
	case "darwin", "linux":
		var cmd *exec.Cmd

		// Se for "default", usa impressora padrão do sistema
		if printerName == "default" {
			defaultPrinter, err := getDefaultPrinter()
			if err != nil {
				// Se não conseguir obter o nome, usa lp sem -d (imprime na padrão automaticamente)
				log.Printf("Aviso: não foi possível obter nome da impressora padrão (%v), usando lp sem especificar impressora (usará padrão do sistema)", err)
				cmd = exec.Command("lp")
			} else {
				log.Printf("Impressora padrão detectada: %s", defaultPrinter)
				cmd = exec.Command("lp", "-d", defaultPrinter)
			}
		} else {
			// Impressora específica
			cmd = exec.Command("lp", "-d", printerName)
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

	case "windows":
		var printerToUse string

		// Se for "default", obtém a impressora padrão
		if printerName == "default" {
			defaultPrinter, err := getDefaultPrinter()
			if err != nil {
				return fmt.Errorf("não foi possível obter impressora padrão: %v", err)
			}
			printerToUse = defaultPrinter
			log.Printf("Impressora padrão detectada: %s", printerToUse)
		} else {
			printerToUse = printerName
		}

		// Out-Printer lê do stdin
		cmd := exec.Command("powershell", "-Command", "Out-Printer -Name '"+printerToUse+"'")
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
			return fmt.Errorf("erro ao imprimir via Out-Printer: %v | stderr: %s", err, stderr.String())
		}
		return nil

	default:
		return fmt.Errorf("sistema operacional não suportado: %s", runtime.GOOS)
	}
}
