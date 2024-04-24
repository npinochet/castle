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

var saveFileCache []byte

type PlayerData struct {
	X, Y float64
}

type SaveData struct {
	PlayerData PlayerData
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
	if len(saveFileCache) != 0 {
		if err := json.Unmarshal(saveFileCache, saveData); err != nil {
			return err
		}
	} else {
		var err error
		if saveData, err = LoadSave(); err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	if saveData == nil {
		saveData = NewSaveData()
	} else {
		updateSaveData(saveData)
	}

	saveFile, err := os.OpenFile(SavePath, os.O_WRONLY, fileMode) //nolint: nosnakecase
	if err != nil {
		if saveFile, err = os.Create(SavePath); err != nil {
			return err
		}
	}
	defer saveFile.Close()

	data, err := json.Marshal(saveData)
	if err != nil {
		return err
	}
	saveFileCache = data

	if _, err := saveFile.Write(data); err != nil {
		return err
	}

	return nil
}

func LoadSave() (*SaveData, error) {
	saveFile, err := os.Open(SavePath)
	if err != nil {
		return nil, err
	}
	defer saveFile.Close()

	saveBuffer, err := io.ReadAll(saveFile)
	if err != nil {
		return nil, err
	}
	saveFileCache = saveBuffer

	saveData := &SaveData{}
	if err := json.Unmarshal(saveBuffer, &saveData); err != nil {
		return nil, err
	}

	return saveData, nil
}

func updateSaveData(sd *SaveData) {
	sd.PlayerData.X, sd.PlayerData.Y = vars.Player.Position()
}
