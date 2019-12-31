package main

import (
	"encoding/json"
	"fmt"
	"github.com/nlopes/slack"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

const (
	WorkspacesListFile = "./workplaces.json"
)

type WorkplaceInfo struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Domain string `json:"domain"`
	Token  string `json:"token"`
}

type AlfredResponse struct {
	Items []ResponseItem `json:"items"`
}

type ResponseItem struct {
	Uid          string    `json:"uid"`
	Valid        bool      `json:"valid"`
	Title        string    `json:"title"`
	Subtitle     string    `json:"subtitle"`
	Arg          string    `json:"arg"`
	Autocomplete string    `json:"autocomplete"`
	Icon         IconModel `json:"icon"`
	Text         TextModel `json:"text"`
	Mod          ModModel  `json:"mods"`
}

type IconModel struct {
	Type string `json:"type"`
	Path string `json:"path"`
}

type TextModel struct {
	Copy      string `json:"copy"`
	Largetype string `json:"largetype"`
}

type ModModel struct {
	Shift ModItems `json:"shift"`
	Cmd   ModItems `json:"cmd"`
}

type ModItems struct {
	Valid    bool   `json:"valid"`
	Arg      string `json:"arg"`
	Subtitle string `json:"subtitle"`
}

func main() {
	options := os.Args[1:]
	var items []ResponseItem
	if options[0] == "token" {
		registerWrokspace(options[1])
		return
	}

	workspaces, err := LoadWorkspaces()
	if err != nil {
		item := ResponseItem{
			Title:    "Something wrong while loading workspaces",
			Subtitle: err.Error(),
		}
		items = append(items, item)
		res := makeAlfredResponse(items)
		fmt.Println(res)
		return
	}

	var channelsList []ResponseItem
	var usersList []ResponseItem

	for _, workspace := range workspaces {
		api := slack.New(workspace.Token)

		channels := getChannels(api, workspace.ID)
		for _, c := range channels {
			item := ResponseItem{
				Title:    "#" + c.GroupConversation.Name + " - " + workspace.Name,
				Arg:      "slack://channel?team=" + workspace.ID + "&id=" + c.GroupConversation.Conversation.ID,
				Subtitle: c.GroupConversation.Topic.Value,
				Valid:    true,
			}
			channelsList = append(channelsList, item)
		}

		usersImages := ListFiles(workspace.ID + "/images/")
		log.Println(usersImages)
		users := getUsers(api, workspace.ID)
		for _, u := range users {
			hasImage := HasItem(usersImages, u.ID)
			if !hasImage {
				DownloadImage(u.Profile.Image192, workspace.ID+"/images/"+u.ID)
			}

			icon := IconModel{
				Path: workspace.ID + "/images/" + u.ID,
			}
			item := ResponseItem{
				Title:    "@" + u.Profile.DisplayNameNormalized + " - " + workspace.Name,
				Arg:      "slack://user?team=" + u.TeamID + "&id=" + u.ID,
				Subtitle: u.Profile.StatusText,
				Valid:    true,
				Icon:     icon,
			}
			if u.Profile.DisplayNameNormalized == "" {
				item.Title = "@" + u.Name
			}
			usersList = append(usersList, item)
		}

	}

	channelsList = append(channelsList, usersList...)
	res := makeAlfredResponse(channelsList)
	fmt.Println(res)
}

func makeAlfredResponse(items []ResponseItem) string {
	alfredResponse := AlfredResponse{
		Items: items,
	}
	r, err := json.Marshal(alfredResponse)
	if err != nil {
		log.Println(err)
	}
	return string(r)
}

func ListFiles(path string) []string {
	filesInfo, err := ioutil.ReadDir(path)
	if err != nil {
		return nil
	}
	var files []string
	for _, file := range filesInfo {
		files = append(files, file.Name())
	}
	return files
}

func DownloadImage(url string, fileName string) {
	response, err := http.Get(url)
	if err != nil {
		log.Println(err)
	}
	defer response.Body.Close()

	file, err := os.Create(fileName)
	if err != nil {
		log.Println(err)
	}
	defer file.Close()

	io.Copy(file, response.Body)
}

func HasItem(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func getUsers(api *slack.Client, workspaceID string) []slack.User {
	// Fetch cached users
	users, err := loadCachedUsers(workspaceID)
	if err != nil {
		users, err = api.GetUsers()
		if err != nil {
			log.Println(err)
		}
		usersJson, err := json.Marshal(users)
		if err != nil {
			log.Println(err)
		}
		_ = ioutil.WriteFile(workspaceID+"/users.json", usersJson, 0777)
	}
	return users
}

func loadCachedUsers(workspaceID string) ([]slack.User, error) {
	jsonFile, err := os.Open(workspaceID + "/users.json")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer jsonFile.Close()

	var users []slack.User
	bytes, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	json.Unmarshal(bytes, &users)
	return users, nil
}

func getChannels(api *slack.Client, workspaceID string) []slack.Channel {
	// Fetch cached users
	channels, err := loadCachedChannels(workspaceID)
	if err != nil {
		channels, err = api.GetChannels(true)
		if err != nil {
			log.Println(err)
		}
		channelsJson, err := json.Marshal(channels)
		if err != nil {
			log.Println(err)
		}
		_ = ioutil.WriteFile(workspaceID+"/channels.json", channelsJson, 0777)
	}
	return channels
}

func loadCachedChannels(workspaceID string) ([]slack.Channel, error) {
	jsonFile, err := os.Open(workspaceID + "/channels.json")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer jsonFile.Close()

	var channels []slack.Channel
	bytes, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	json.Unmarshal(bytes, &channels)
	log.Println(channels)
	return channels, nil
}

func getWorkspalce(api *slack.Client) *slack.TeamInfo {
	workspaces, err := loadCachedWorkspace()
	if err != nil {
		workspaces, err = api.GetTeamInfo()
		if err != nil {
			log.Println(err)
		}
	}
	return workspaces
}

func loadCachedWorkspace() (*slack.TeamInfo, error) {
	jsonFile, err := os.Open(WorkspacesListFile)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer jsonFile.Close()

	var workspaces []slack.TeamInfo
	bytes, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	json.Unmarshal(bytes, &workspaces)
	//return workspaces, nil
	return nil, nil
}

func LoadWorkspaces() ([]WorkplaceInfo, error) {
	file, err := os.Open(WorkspacesListFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var workplaces []WorkplaceInfo
	json.Unmarshal(bytes, &workplaces)

	for _, w := range workplaces {
		if _, err := os.Stat(w.ID); os.IsNotExist(err) {
			os.Mkdir(w.ID, 0777)
		}
		if _, err := os.Stat(w.ID + "/images"); os.IsNotExist(err) {
			os.Mkdir(w.ID+"/images", 0777)
		}
	}
	return workplaces, nil
}

func registerWrokspace(token string) {
	var workspaces []WorkplaceInfo

	workspacesFile, err := os.Open(WorkspacesListFile)
	if err != nil {
		log.Println("Failed to load the workspaces.json")
	} else {
		bytes, err := ioutil.ReadAll(workspacesFile)
		if err != nil {
			log.Println("Failed to read bytes")
		}
		json.Unmarshal(bytes, &workspaces)
	}
	defer workspacesFile.Close()

	for _, w := range workspaces {
		if token == w.Token {
			log.Println("Already registered the workspace")
			return
		}
	}

	api := slack.New(token)
	teamInfo, err := api.GetTeamInfo()
	if err != nil {
		log.Println("Failed to get team info : ", err)
	}
	w := WorkplaceInfo{
		ID:     teamInfo.ID,
		Name:   teamInfo.Name,
		Domain: teamInfo.Domain,
		Token:  token,
	}
	workspaces = append(workspaces, w)
	workspacesJson, err := json.Marshal(workspaces)
	if err != nil {
		log.Println("Marshal Error : ", err)
	}
	_ = ioutil.WriteFile(WorkspacesListFile, workspacesJson, 0777)
}
