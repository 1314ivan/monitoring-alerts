package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	gosxnotifier "github.com/deckarep/gosx-notifier"
	"github.com/joho/godotenv"
	"github.com/shirou/gopsutil/mem"
)

func getPercentRAM() (int, error) {
	vm, err := mem.VirtualMemory()
	if err != nil {
		return 0, err
	}
	return int(math.Round(vm.UsedPercent)), nil
}

func sendAppleScriptNotification(ram int) {
	later := "Напомнить позже (30мин)"
	dropCache:= "Очистить Кэш"

    script := `set ramUsage to ` + strconv.Itoa(ram) + `
set laterButton to "` + later + `"
set dropCacheButton to "` + dropCache + `"

try
    set userChoice to display dialog "Обнаружено высокое потребление памяти: " & ramUsage & "%" & return & return & ¬
    "Рекомендуется проверить активные процессы." ¬
    with title "⚠️ Мониторинг системы ⚠️" ¬
    buttons {"Мониторинг", dropCacheButton, laterButton} ¬
    default button "Мониторинг" ¬
    cancel button laterButton ¬
    with icon caution
    
    set buttonPressed to button returned of userChoice
    
    if buttonPressed is "Мониторинг" then
        tell application "Activity Monitor" to activate
        return "monitor"
        
    else if buttonPressed is dropCacheButton then
        try
            do shell script "sudo purge" with administrator privileges
            display notification "Кэш очищен" with title "✅ Готово"
            return "cache_cleared"
        on error
            return "cache_failed"
        end try
    end if
    
on error number -128
    -- Пользователь нажал cancel (laterButton) или Esc
    display notification "Напомню через 30 минут" with title "⏰ Мониторинг системы"
    return "remind_later"
end`

	cmd := exec.Command("osascript", "-e", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Ошибка AppleScript: %v\n%s\n", err, output)
		return
	}

	response := strings.TrimSpace(string(output))
	if response == "remind_later" {
		time.Sleep(time.Duration(30) * time.Minute)
	} else if response== "cache_failed"{
		time.Sleep(time.Duration(10) * time.Minute)
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	intervalStr := os.Getenv("INTERVAL")
	interval, err := strconv.Atoi(intervalStr)
	if err != nil {
		panic(err)
	}
	notifAlertStr := os.Getenv("NOTIFICATION_CENTER_ALERT_RAM")
	notifAlert, err := strconv.Atoi(notifAlertStr)
	if err != nil {
		panic(err)
	}
	notifBannerStr := os.Getenv("NOTIFICATION_BANNER_RAM")
	notifBanner, err := strconv.Atoi(notifBannerStr)
	if err != nil {
		panic(err)
	}
	ram, _ := getPercentRAM()
	for {
		if ram > notifBanner {
			sendAppleScriptNotification(ram)
		} else if ram > notifAlert {
			message := "Обнаружено высокое потребление памяти: " + strconv.Itoa(ram) + "%"
			note := gosxnotifier.NewNotification(message)
			note.Title = "⚠️ Мониторинг системы ⚠️"
			note.Push()
		}
		time.Sleep(time.Duration(interval) * time.Second)
	}
}
