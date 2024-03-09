package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/manifoldco/promptui"
)

func configure() {
	usernamePrompt := promptui.Prompt{
		Label: "Username",
	}

	username, err := usernamePrompt.Run()
	if err != nil {
		panic(err)
	}

	passwordPrompt := promptui.Prompt{
		Label: "Password",
		Mask:  '*',
	}

	password, err := passwordPrompt.Run()
	if err != nil {
		panic(err)
	}

	hostPrompt := promptui.Prompt{
		Label:   "Authentication Host",
		Default: "https://puter.com",
	}

	host, err := hostPrompt.Run()
	if err != nil {
		panic(err)
	}

	hostAPIPrompt := promptui.Prompt{
		Label:   "API Host",
		Default: "https://api.puter.com",
	}

	hostAPI, err := hostAPIPrompt.Run()
	if err != nil {
		panic(err)
	}

	mountpointPrompt := promptui.Prompt{
		Label:   "Mountpoint",
		Default: "/tmp/mnt",
	}

	mountpoint, err := mountpointPrompt.Run()
	if err != nil {
		panic(err)
	}

	// Get token from server
	payload := map[string]string{
		"username": username,
		"password": password,
	}

	jsonStr, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest(
		"POST",
		host+"/login",
		bytes.NewBuffer(jsonStr),
	)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		// Print response
		body, _ := io.ReadAll(resp.Body)
		fmt.Println(string(body))

		panic(fmt.Errorf("unexpected status: %d", resp.StatusCode))
	}

	// Save token
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	responseData := map[string]interface{}{}
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		panic(err)
	}

	configToWrite := map[string]interface{}{
		"mountpoint": mountpoint,
		"url":        hostAPI,
		"token":      responseData["token"],
	}

	// Write config
	configDir, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}

	configFile, err := os.Create(configDir + "/puterfuse/config.json")
	if err != nil {
		panic(err)
	}
	defer configFile.Close()

	configJSON, err := json.Marshal(configToWrite)
	if err != nil {
		panic(err)
	}

	_, err = configFile.Write(configJSON)
	if err != nil {
		panic(err)
	}
}