package main

import (
	"context"
	"fmt"

	"net/http"
	"os"
	"time"

	"ha-images/pkg/clients"
	"ha-images/pkg/entitiesmap"
	"ha-images/pkg/imgbuilder"

	retry "github.com/avast/retry-go"

	ha "github.com/mkelcik/go-ha-client"
)

const (
	fontSize       = 32
	paddingBetween = 32 / 2
)

func main() {
	var (
		haToken     = os.Getenv("HA_TOKEN")
		haHost      = os.Getenv("HA_HOST")
		remoteUrl   = os.Getenv("REMOTE_URL")
		entitiesMap = os.Getenv("ENTITIES_MAP")
	)

	groups, err := entitiesmap.Read(entitiesMap)
	if err != nil {
		panic(err)
	}

	client := ha.NewClient(ha.ClientConfig{Token: haToken, Host: haHost}, &http.Client{
		Timeout: 30 * time.Second,
	})

	for indx, grp := range groups {
		sensors := grp.Sensors

		img := imgbuilder.NewImgBuilder(240, 240)

		offset := fontSize

		for _, sensor := range sensors {
			stateStr := "Unknown"
			state, err := client.GetStateForEntity(context.Background(), sensor.EntityId)
			if err == nil {
				stateStr = state.State
			}

			img.AddLabel(fmt.Sprintf("%s: %s", sensor.Name, stateStr), fontSize, 0, offset)

			offset += fontSize + paddingBetween
		}

		b, err := img.Generate()
		if err != nil {
			panic(err)
		}

		cl := clients.NewFileUploader(remoteUrl)

		err = retry.Do(
			func() error {
				err := cl.Upload(clients.FileContent{
					Filename: fmt.Sprintf("%d.jpg", indx+1),
					Filetype: "image/jpeg",
					Data:     b,
				})
				if err != nil {
					fmt.Printf("err=%v\n", err)
				}
				return err
			},
			retry.Attempts(10),
			retry.Delay(5*time.Second),
		)

		if err != nil {
			panic(err)
		}
	}
}
