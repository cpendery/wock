package hosts

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"runtime"
	"strings"
	"syscall"

	"golang.org/x/sys/windows"
)

const (
	windowsHostFile = "C:\\Windows\\System32\\drivers\\etc\\hosts"
)

type UpdateResult struct {
}

func UpdateHosts(host string, enable bool) error {
	var hostFilePath string

	switch runtime.GOOS {
	case "windows":
		hostFilePath = windowsHostFile
	}

	file, err := os.Open(hostFilePath)
	if err != nil {
		return fmt.Errorf("unable to open hosts file: %w", err)
	}
	scanner := bufio.NewScanner(file)
	hasWrittenHost := false
	var result bytes.Buffer
	for scanner.Scan() {
		line := scanner.Text()
		cleanedLine := strings.ToLower(strings.TrimSpace(line))
		if strings.Contains(cleanedLine, host) {
			if enable {
				result.WriteString(fmt.Sprintf("127.0.0.1 %s\n", host))
			} else {
				result.WriteString(fmt.Sprintf("# 127.0.0.1 %s\n", host))
			}
			hasWrittenHost = true
		} else {
			result.WriteString(line)
			result.WriteRune('\n')
		}
	}
	if !hasWrittenHost {
		if enable {
			result.WriteString(fmt.Sprintf("127.0.0.1 %s\n", host))
		} else {
			result.WriteString(fmt.Sprintf("# 127.0.0.1 %s\n", host))
		}
	}

	err = file.Close()
	if err != nil {
		return fmt.Errorf("unable to close hosts file on initial read: %w", err)
	}
	file, err = os.Create(hostFilePath)
	if err != nil {
		return fmt.Errorf("unable to open hosts file for truncation: %w", err)
	}
	_, err = file.Write(result.Bytes())
	if err != nil {
		return fmt.Errorf("unable to overwrite hosts file: %w", err)
	}
	err = file.Close()
	if err != nil {
		return fmt.Errorf("unable to close hosts file on overwrite: %w", err)
	}
	return nil
}

func RunMeElevated() {
	verb := "runas"
	exe, _ := os.Executable()
	cwd, _ := os.Getwd()
	args := strings.Join(os.Args[1:], " ")

	verbPtr, _ := syscall.UTF16PtrFromString(verb)
	exePtr, _ := syscall.UTF16PtrFromString(exe)
	cwdPtr, _ := syscall.UTF16PtrFromString(cwd)
	argPtr, _ := syscall.UTF16PtrFromString(args)

	var showCmd int32 = 1 //SW_NORMAL

	err := windows.ShellExecute(0, verbPtr, exePtr, argPtr, cwdPtr, showCmd)
	if err != nil {
		fmt.Println(err)
	}
}

func AmAdmin() bool {
	elevated := windows.GetCurrentProcessToken().IsElevated()
	return elevated
}
