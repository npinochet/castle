package game

import (
	"encoding/json"
	"game/comps/stats"
	"game/core"
	"game/entity"
	"game/utils"
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
	X   float64 `json:"x"`
	Y   float64 `json:"y"`
	Exp int     `json:"exp"`
}

type SaveData struct {
	PlayerData PlayerData        `json:"player_data"`
	Pad        utils.ControlPack `json:"keys"`
}

func NewSaveData() *SaveData {
	obj, err := vars.World.Map.FindObjectFromTileID(playerID, "entities")
	if err != nil {
		log.Println("game: error finding player entity:", err)
	}

	return &SaveData{
		PlayerData: PlayerData{X: obj.X, Y: obj.Y},
		Pad:        utils.NewControlPack(),
	}
}

func Save() error {
	var saveData *SaveData
	if len(saveDataCache) != 0 {
		if err := json.Unmarshal(saveDataCache, &saveData); err != nil {
			return err
		}
	} else {
		var err error
		if saveData, err = LoadSave(); err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	updateSaveData(saveData)

	saveFile, err := os.OpenFile(SavePath, os.O_WRONLY|os.O_TRUNC, fileMode) //nolint: nosnakecase
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

func ApplySaveData(sd *SaveData) {
	vars.Player = entity.NewPlayer(sd.PlayerData.X, sd.PlayerData.Y)
	core.Get[*stats.Comp](vars.Player).Exp = sd.PlayerData.Exp
	vars.Pad = sd.Pad
}

func updateSaveData(sd *SaveData) {
	playerStats := core.Get[*stats.Comp](vars.Player)
	sd.PlayerData.X, sd.PlayerData.Y = vars.Player.Position()
	sd.PlayerData.Exp = playerStats.Exp
	sd.Pad = vars.Pad
}
