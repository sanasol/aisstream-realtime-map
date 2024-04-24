package main

import (
	"encoding/json"
	"fmt"
	aisstream "github.com/aisstream/ais-message-models/golang/aisStream"
	"github.com/gorilla/websocket"
	"log"
	"os"
	"time"
)

type ship struct {
	Lat  float64 `json:"lat"`
	Lon  float64 `json:"lon"`
	name string
}

func main() {

	url := "wss://stream.aisstream.io/v0/stream"
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatalln(err)
	}
	defer ws.Close()

	subMsg := aisstream.SubscriptionMessage{
		APIKey:        "0594aa81c8264b71d5060c3fb31161285d42ac62",
		BoundingBoxes: [][][]float64{{{41.58668835697237, 126.98547363281251}, {44.12308489306967, 136.47766113281253}}},
		//FiltersShipMMSI: []string{""},
	}

	subMsgBytes, _ := json.Marshal(subMsg)
	if err := ws.WriteMessage(websocket.TextMessage, subMsgBytes); err != nil {
		log.Fatalln(err)
	}

	ships := make(map[string]ship)
	ticker := time.NewTicker(5 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				file, _ := json.MarshalIndent(ships, "", " ")

				_ = os.WriteFile("ships.json", file, 0644)
				fmt.Printf("file saved %d\n", len(ships))

			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	for {
		_, p, err := ws.ReadMessage()
		if err != nil {
			log.Fatalln(err)
		}
		var packet aisstream.AisStreamMessage

		err = json.Unmarshal(p, &packet)
		if err != nil {
			log.Fatalln(err)
		}

		var shipName string
		// field may or may not be populated
		if packetShipName, ok := packet.MetaData["ShipName"]; ok {
			shipName = packetShipName.(string)
		}

		switch packet.MessageType {
		case aisstream.POSITION_REPORT:
			var positionReport aisstream.PositionReport
			positionReport = *packet.Message.PositionReport
			ships[shipName] = ship{
				Lat:  positionReport.Latitude,
				Lon:  positionReport.Longitude,
				name: shipName,
			}
		}
	}
}
