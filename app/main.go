package main

import (
	"fmt"
	"os"
	"os/signal"
	"petrol/build"
	"petrol/infrastructure"
	"petrol/infrastructure/stations"
	"sync"
	"syscall"
	"time"
)

var wogUrls = []string{
	"https://api.wog.ua/fuel_stations/886",
	"https://api.wog.ua/fuel_stations/1094",
	"https://api.wog.ua/fuel_stations/808",
	"https://api.wog.ua/fuel_stations/1096",
}

func main() {
	fmt.Println("Version:\t", build.Version)
	fmt.Println("Time:\t", build.Time)
	fmt.Println("User:\t", build.User)
	fmt.Println("Hash:\t", build.Hash)

	handler, err := os.OpenFile(os.Getenv("PETROL_STORAGE_PATH"), os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}

	defer handler.Close()

	shutdownCh := makeShutdownCh()
	var telegramBotToken infrastructure.TelegramToken

	telegramBotToken = infrastructure.TelegramToken(os.Getenv("PETROL_TELEGRAM_BOT_TOKEN"))
	telegramBot, err := infrastructure.NewTelegramBot(telegramBotToken, os.Getenv("PETROL_TELEGRAM_WEBHOOK_URL"))

	if err != nil {
		panic(err)
	}

	mu := &sync.Mutex{}
	allowedWOGPetrolTypes := map[string]struct{}{"М95": struct{}{}, "А95": struct{}{}}
	httpClient := infrastructure.NewInternalHttpClient()
	logger := infrastructure.NewLogger()
	storage := infrastructure.NewBinaryStorage(logger, handler)
	petrolRepository := infrastructure.NewPetrolBinaryRepository(storage)
	petrolInfo, err := petrolRepository.ReadAll()

	if err != nil {
		panic(err)
	}

	logger.Info("Start")
	petrolInfoCh := make(chan *infrastructure.PetrolStationInfo, 2)
	for _, wogUrl := range wogUrls {
		go func(url string) {
			wog := stations.NewWog(url, allowedWOGPetrolTypes, httpClient, logger)
			wog.Operate(petrolInfoCh, 60*time.Second)
		}(wogUrl)
	}

	go telegramBot.Start()

	telegramBot.Inform(fmt.Sprintf("Version: \t%s\n Time: \t%s\n User: \t%s\n Hash: \t%s\n", build.Version, build.Time, build.User, build.Hash))

CLOSE:
	for {
		select {
		case <-shutdownCh:
			logger.Info("shutdown signal...")

			petrolRepository.SaveAll(petrolInfo)

			break CLOSE
		case info := <-petrolInfoCh:
			infrastructure.ExecutePipeline(
				infrastructure.Job(func(in, out chan interface{}) {
					out <- info
				}),
				infrastructure.Job(func(in, out chan interface{}) {
					for data := range in {
						petrol := data.(*infrastructure.PetrolStationInfo)
						_, exists := petrolInfo[petrol.Id]
						if !exists || petrolInfo[petrol.Id].State != petrol.State {
							out <- petrol
						}
					}
				}),
				infrastructure.Job(func(in, out chan interface{}) {
					for data := range in {
						petrol := data.(*infrastructure.PetrolStationInfo)
						mu.Lock()
						petrolInfo[petrol.Id] = petrol
						mu.Unlock()
						out <- data
					}
				}),
				infrastructure.Job(func(in, out chan interface{}) {
					for data := range in {
						petrol := data.(*infrastructure.PetrolStationInfo)
						inform(telegramBot, petrol)
					}
				},
				))
		}
	}

	logger.Info("Finish")

	os.Exit(0)
}

func inform(telegramBot *infrastructure.TelegramBot, data *infrastructure.PetrolStationInfo) {
	telegramBot.Inform(fmt.Sprintf(
		"%s \t %s \n %s\n https://www.google.com/maps/search/?api=1&query=%f,%f",
		data.PetrolType,
		data.State,
		data.Address,
		data.Coordinates.Lat,
		data.Coordinates.Lon,
	))
}

func makeShutdownCh() <-chan struct{} {
	resultCh := make(chan struct{})
	signalCh := make(chan os.Signal, 4)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		for {
			<-signalCh
			resultCh <- struct{}{}
		}
	}()

	return resultCh
}
