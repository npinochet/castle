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
	Persistent = false
	SavePath   = "save.json"
	fileMode   = 0666
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
	Opened     []uint            `json:"opened"`
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
	var err error
	if len(saveDataCache) != 0 {
		if err := json.Unmarshal(saveDataCache, &saveData); err != nil {
			return err
		}
	} else {
		if saveData, err = LoadSave(); err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	populateSaveData(saveData)

	if saveDataCache, err = json.MarshalIndent(saveData, "", "	"); err != nil {
		return err
	}

	if Persistent {
		saveFile, err := os.OpenFile(SavePath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, fileMode) //nolint: nosnakecase
		if err != nil {
			return err
		}
		defer saveFile.Close()

		if _, err := saveFile.Write(saveDataCache); err != nil {
			return err
		}
	}

	return nil
}

func LoadSave() (*SaveData, error) {
	if Persistent || len(saveDataCache) == 0 {
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
	for _, opened := range sd.Opened {
		if chest, ok := vars.World.Get(opened).(*entity.Chest); ok {
			chest.Open(false)
		}
		if door, ok := vars.World.Get(opened).(*entity.Door); ok {
			door.Open()
		}
	}
}

func populateSaveData(sd *SaveData) {
	playerStats := core.Get[*stats.Comp](vars.Player)
	sd.PlayerData.X, sd.PlayerData.Y = vars.Player.Position()
	sd.PlayerData.Exp = playerStats.Exp
	sd.Pad = vars.Pad

	for _, e := range vars.World.GetAll() {
		id := vars.World.GetID(e)
		if id == 0 {
			continue
		}
		if chest, ok := e.(*entity.Chest); ok && chest.Opened() {
			sd.Opened = append(sd.Opened, id)
		}
		if door, ok := e.(*entity.Door); ok && door.Opened() {
			sd.Opened = append(sd.Opened, id)
		}
	}
}
