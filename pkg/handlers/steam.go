package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/andygrunwald/vdf"
	"yunion.io/x/log"
)

type SteamOwnedGamesController struct{}

func NewSteamOwnedGamesController() *SteamOwnedGamesController {
	return &SteamOwnedGamesController{}
}

type Game struct {
	AppID      int    `json:"appid"`
	Name       string `json:"name"`
	PlayTime   int    `json:"playtime_forever"`
	PlayTime2W int    `json:"playtime_2weeks,omitempty"`
}

type OwnedGamesResponse struct {
	Response struct {
		GameCount int    `json:"game_count"`
		Games     []Game `json:"games"`
	} `json:"response"`
}

func findSteamID64() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %v", err)
	}

	steamUserDataPath := filepath.Join(homeDir, ".steam", "steam", "userdata")
	entries, err := ioutil.ReadDir(steamUserDataPath)
	if err != nil {
		return "", fmt.Errorf("failed to read steam userdata directory: %v", err)
	}

	var steamIDs []string
	for _, entry := range entries {
		if entry.IsDir() {
			// 检查是否是数字目录（SteamID64）
			if _, err := fmt.Sscanf(entry.Name(), "%d", new(int)); err == nil {
				steamIDs = append(steamIDs, entry.Name())
			}
		}
	}

	if len(steamIDs) == 0 {
		return "", fmt.Errorf("no Steam ID found")
	}
	if len(steamIDs) > 1 {
		return "", fmt.Errorf("multiple Steam IDs found: %v", steamIDs)
	}

	return steamIDs[0], nil
}

func parseLocalConfig(steamID64 string) ([]Game, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %v", err)
	}

	configPath := filepath.Join(homeDir, ".steam", "steam", "userdata", steamID64, "config", "localconfig.vdf")
	f, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open localconfig.vdf: %v", err)
	}
	defer f.Close()

	parser := vdf.NewParser(f)
	data, err := parser.Parse()
	if err != nil {
		return nil, fmt.Errorf("failed to parse vdf: %v", err)
	}
	log.Infof("=== configPath: %s", configPath)

	// 递归进入 Software -> Valve -> Steam -> apps
	configStore := data["UserLocalConfigStore"].(map[string]interface{})
	software, ok := configStore["Software"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no Software section")
	}
	valve, ok := software["Valve"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no Valve section")
	}
	steam, ok := valve["Steam"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no Steam section")
	}
	apps, ok := steam["apps"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no apps section")
	}

	var games []Game
	for appidStr, v := range apps {
		app, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		appid := 0
		fmt.Sscanf(appidStr, "%d", &appid)
		playtime := 0
		playtime2wks := 0
		if pt, ok := app["Playtime"].(string); ok {
			fmt.Sscanf(pt, "%d", &playtime)
		}
		if pt2, ok := app["Playtime2wks"].(string); ok {
			fmt.Sscanf(pt2, "%d", &playtime2wks)
		}
		games = append(games, Game{
			AppID:      appid,
			Name:       "", // localconfig.vdf 里没有名字
			PlayTime:   playtime,
			PlayTime2W: playtime2wks,
		})
	}
	return games, nil
}

func (c *SteamOwnedGamesController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	steamID64, err := findSteamID64()
	if err != nil {
		http.Error(w, "Failed to find Steam ID: "+err.Error(), http.StatusInternalServerError)
		return
	}

	games, err := parseLocalConfig(steamID64)
	if err != nil {
		http.Error(w, "Failed to parse Steam config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := OwnedGamesResponse{
		Response: struct {
			GameCount int    `json:"game_count"`
			Games     []Game `json:"games"`
		}{
			GameCount: len(games),
			Games:     games,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
