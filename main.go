package main

import (
	"github.com/NVIDIA/go-nvml/pkg/nvml"
	"github.com/getlantern/systray"
	"github.com/ssimunic/gosensors"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"
	_ "embed"
)

//go:embed icon.png
var icon []byte

var cpuTemp string
var gpuTemp string
var altText string

func onReady() {
	systray.SetTitle("WatchBar")
	systray.SetIcon(icon)
	mQuitOrig := systray.AddMenuItem("Quit", "Quit the whole app")
	go func() {
		<-mQuitOrig.ClickedCh
		systray.Quit()
		os.Exit(0)
	}()
	systray.AddSeparator()
	go func() {
		for {
			systray.SetTitle("CPU: " + cpuTemp + gpuTemp )
			systray.SetTooltip(altText)
			time.Sleep(500 * time.Millisecond)
		}
	}()
}

func onExit(){

}

func getGpuTemp(){
	ret := nvml.Init()
	if ret != nvml.SUCCESS {
		log.Fatalf("Unable to initialize NVML: %v", nvml.ErrorString(ret))
	}
	defer func() {
		ret := nvml.Shutdown()
		if ret != nvml.SUCCESS {
			log.Fatalf("Unable to shutdown NVML: %v", nvml.ErrorString(ret))
		}
	}()
	count, ret := nvml.DeviceGetCount()
	if ret != nvml.SUCCESS {
		log.Fatalf("Unable to get device count: %v", nvml.ErrorString(ret))
	}
	buildString := ""
	for i := 0; i < count; i++ {

		device, ret := nvml.DeviceGetHandleByIndex(i)
		if ret != nvml.SUCCESS {
			log.Fatalf("Unable to get device at index %d: %v", i, nvml.ErrorString(ret))
		}
		temp, ret := device.GetTemperature(0)
		if ret != nvml.SUCCESS {
			log.Fatalf("Unable to get Tempature of device at index %d: %v", i, nvml.ErrorString(ret))
		}
		buildString = buildString + "| GPU" + strconv.Itoa(i) + ": " + strconv.Itoa(int(temp)) + "°C "
	}
	gpuTemp = buildString
	time.Sleep(1 * time.Second)
}

func main() {
	args := os.Args
if len(args) == 1{
		cwd, _ := os.Getwd()
		args := append(os.Args, "--detached")
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = cwd
		cmd.Start()
		println("Starting in background")
		cmd.Process.Release()
		os.Exit(0)
}
	go systray.Run(onReady, onExit)
	for {
		sensors, _ := gosensors.NewFromSystem()
		getGpuTemp()
		for chip := range sensors.Chips {
			for key, value := range sensors.Chips[chip] {
				if key == "PECI Agent 0 Calibration" {
					cpuTemp = value
				}
			}
		}
	}
}