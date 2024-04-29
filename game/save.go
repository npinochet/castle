package game

import (
	"encoding/json"
	"game/vars"
	"io"
	"log"
	"os"
)

const (
	SavePath = "save.json"
	fileMode = 0666
)

var saveDataCache []byte

type PlayerData struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type SaveData struct {
	PlayerData PlayerData `json:"player_data"`
}

func NewSaveData() *SaveData {
	obj, err := vars.World.Map.FindObjectFromTileID(playerID, "entities")
	if err != nil {
		log.Println("game: error finding player entity:", err)
	}

	return &SaveData{PlayerData: PlayerData{X: obj.X, Y: obj.Y}}
}

func Save() error {
	var saveData *SaveData
	if len(saveDataCache) != 0 {
		if err := json.Unmarshal(saveDataCache, saveData); err != nil {
			return err
		}
	} else {
		var err error
		if saveData, err = LoadSave(); err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	updateSaveData(saveData)

	saveFile, err := os.OpenFile(SavePath, os.O_WRONLY, fileMode) //nolint: nosnakecase
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		if saveFile, err = os.Create(SavePath); err != nil {
			return err
		}
	}
	defer saveFile.Close()

	if saveDataCache, err = json.Marshal(saveData); err != nil {
		return err
	}
	if _, err := saveFile.Write(saveDataCache); err != nil {
		return err
	}

	return nil
}

func LoadSave() (*SaveData, error) {
	saveFile, err := os.Open(SavePath)
	if err != nil {
		if os.IsNotExist(err) {
			return NewSaveData(), nil
		}

		return nil, err
	}
	defer saveFile.Close()

	if saveDataCache, err = io.ReadAll(saveFile); err != nil {
		return nil, err
	}

	var saveData *SaveData
	if err := json.Unmarshal(saveDataCache, &saveData); err != nil {
		return nil, err
	}

	return saveData, nil
}

func updateSaveData(sd *SaveData) {
	sd.PlayerData.X, sd.PlayerData.Y = vars.Player.Position()
}
