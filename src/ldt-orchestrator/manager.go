package ldtorchestrator

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Process struct {
	Pid     int
	Name    string
	started time.Time
}

type Manager struct {
	monitor   *Monitor
	discovery *DiscoveryConfig
}

func NewManager(config, ldt_list_path string) *Manager {
	manager := &Manager{
		monitor:   NewMonitor(ldt_list_path),
		discovery: NewConfig(config),
	}

	if err := manager.monitor.DeserializeLDTs(); err != nil {
		log.Fatal(err)
	}

	return manager
}

func (manager *Manager) GetAvailableLDTs() string {
	manager.discovery.DiscoverLDTs()
	return manager.discovery.supportedLDTs.String()
}

func (manager *Manager) GetURLFromLDTByID(id int) string {
	url, err := manager.discovery.GetUrlFromLDT(id)
	if err != nil {
		log.Fatal(err)
	}
	return url
}

func (manager *Manager) DownloadLDTArchive(address string) error {
	url, _ := url.Parse(address)
	filename := strings.Split(url.Path, "/")[6]

	file, err := os.Create("resources/" + filename)
	if err != nil {
		return err
	}
	defer file.Close()

	response, err := http.Get(address)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}

	log.Printf("Downloaded LDT Archive: %s\n", file.Name())
	return nil
}

func (manager *Manager) DownloadLDT(url string) (string, error) {
	file, err := os.Create("./resources/child_webserver")
	if err != nil {
		return "", err
	}
	defer file.Close()

	response, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		return "", err
	}

	if err := os.Chmod(file.Name(), 0755); err != nil {
		log.Fatalf("Could not set executable flag: %v", err)
	}

	log.Printf("Downloaded LDT: %s\n", file.Name())
	return file.Name(), nil
}

func (manager *Manager) StartLDT(name string) (*Process, error) {
	cmd := exec.Command("./" + name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		log.Fatal("Could not start LDT\n")
		return nil, err
	}

	fmt.Printf("Started LDT with PID %d\n", cmd.Process.Pid)
	return &Process{
		Pid:     cmd.Process.Pid,
		Name:    name,
		started: time.Now(),
	}, nil

}

func (manager *Manager) StopLDT(pid int, graceful bool) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		log.Fatal(err)
		return err
	}
	if graceful == true {
		err = proc.Signal(os.Interrupt)
	} else {
		err = proc.Kill()
	}

	if err != nil {
		log.Fatalf("Unable to stop LDT \t graceful? %t", graceful)
		return err
	}
	return nil
}

func (manager *Manager) shutdown() {
	if err := manager.monitor.SerializeLDTs(); err != nil {
		log.Fatal(err)
	}
}
