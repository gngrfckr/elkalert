package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type Signature struct {
	Name   string
	Signed bool
}

//Template структура шаблона сообщеий

//Destination ...
type Destination struct {
	To     string `json:"to"`
	Email  string `json:"email"`
	Chatid string `json:"chatid"`
	Botid  string `json:"botid"`
}

//Alert ...
type Alert struct {
	Template    string `json:"template"`
	Destination Destination
	Message     string
}

type Operator interface {
	Send(Alert) bool
}

type Operation struct {
	Operator Operator
}

func (o *Operation) Send(Alert Alert) bool {
	return o.Operator.Send(Alert)
}

type TelegramSend struct{}

func (TelegramSend) Send(Alert Alert) bool {
	client := &http.Client{}
	readyMessage := map[string]string{
		"text":       Alert.Message,
		"chat_id":    Alert.Destination.Chatid,
		"parse_mode": "html",
	}
	apiurl := "https://api.telegram.org/bot" + Alert.Destination.Botid + "/sendMessage"
	b, err := json.Marshal(readyMessage)
	if err != nil {
		panic(err)
	}
	json := []byte(b)
	req, err := http.NewRequest("POST", apiurl, bytes.NewBuffer(json))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	log.Println(resp)
	return true
}

//GetValue ...
func GetValue(key string, s interface{}) (v interface{}, err error) {
	var (
		i  int64
		ok bool
	)
	switch s.(type) {
	case map[string]interface{}:
		if v, ok = s.(map[string]interface{})[key]; !ok {
			err = fmt.Errorf("Key not present. [Key:%s]", key)
		}
	case []interface{}:
		if i, err = strconv.ParseInt(key, 10, 64); err == nil {
			array := s.([]interface{})
			if int(i) < len(array) {
				v = array[i]
			} else {
				err = fmt.Errorf("Index out of bounds. [Index:%d] [Array:%v]", i, array)
			}
		}
	}
	return v, err
}

func resolveTemplate(Alert Alert, Hit map[string]interface{}) Alert {
	var err error
	var newMessage string
	newMessage = Alert.Template
	r := regexp.MustCompile(`\{\{(.*?)\}\}`)

	for _, arr := range r.FindAllStringSubmatch(Alert.Template, -1) {
		var value interface{} = Hit
		path := arr[1]
		keys := strings.Split(path, ".")
		for _, key := range keys {
			if value, err = GetValue(key, value); err != nil {
				break
			}
		}
		if err == nil {
			valueStr := fmt.Sprintf("%v", value)
			newMessage = strings.Replace(newMessage, arr[0], valueStr, -1)
		}

	}
	Alert.Message = newMessage
	return Alert
}

//SendAlert method
func SendAlert(Alert Alert, Hit map[string]interface{}) {
	resolvedAlert := resolveTemplate(Alert, Hit)
	if resolvedAlert.Destination.To == "telegram" {
		telegram := Operation{TelegramSend{}}
		telegram.Send(resolvedAlert)
	}

}
