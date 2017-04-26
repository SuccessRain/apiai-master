package main

import (
	ai "github.com/SuccessRain/apiai-fsoft-sdk"
	"fmt"
	"os"
	"bufio"
	"strings"
	"flag"
)

var target string
var inputFP string
var token string

func AddSharedFlags(fs *flag.FlagSet) {
	fs.StringVar(&target, "t", "", "required, intent or entity")
	fs.StringVar(&inputFP, "i", "", "required, path to the input file")
	fs.StringVar(&token, "token", "", "required, API.AI token")
}

type IntentUtterance struct {
	Intent string `json:"intent"`
	Utterance string `json:"utterance"`
}

type ObjectFile struct{
	Intent string
	Name string
	Response string
}

func ReadIntentsFromFile(inputFP string) ([]ObjectFile, error) {
	input, err := os.Open(inputFP)
	if err != nil {
		return nil, err
	}
	defer input.Close()

	var context []ObjectFile
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		var con ObjectFile
		text := strings.Replace(scanner.Text(), `"`, ``, -1)
		tokens := strings.SplitN(text, ",", 3)
		//fmt.Print("TOKEN: \t"); fmt.Println(tokens)
		con.Intent, con.Name, con.Response = strings.TrimSpace(tokens[0]), strings.TrimSpace(tokens[1]), strings.TrimSpace(tokens[2])
		context = append(context, con)
		//fmt.Print("CONTEXT: \t"); fmt.Println(con.Intent); fmt.Println(con.Name); fmt.Println(con.Response)
	}

	return context, nil
}

var responseAll ai.ApiResponse

// create an intent
func CreateIntent(client *ai.Client, intentName string, userSays string,
	action string, lstResponse []string, ) int{

	var count int = 0;
	response, err := client.CreateIntent(ai.IntentObject{
		Name:     intentName,
		Auto:     true,
		Priority: 5,
		UserSays: []ai.UserSays{
			ai.UserSays{
				Data: []ai.UserSaysData{
					ai.UserSaysData{
						Text: userSays,
					},

				},
			},
		},
		Responses: []ai.IntentResponse{
			ai.IntentResponse{
				Action:        action,
				Speech: lstResponse,
			},
		},
	});

	if err == nil {
		//fmt.Printf(">>> created intent = %+v\n", response)
		fmt.Println("Success: Created intent!")
		count ++
		responseAll = response;
	} else {
		fmt.Printf("*** error: %s\n", err)
	}
	return  count;
}

// update an intent
func updateIntent(client *ai.Client, response ai.ApiResponse, intentName string, userSays string, action string,
	lstResponse []string){
	if response, err := client.UpdateIntent(response.Id, ai.IntentObject{
		Name:     intentName,
		Auto:     true,
		UserSays: []ai.UserSays{
			ai.UserSays{
				Data: []ai.UserSaysData{
					ai.UserSaysData{
						Text: userSays,
					},
				},
			},
		},
		Responses: []ai.IntentResponse{
			ai.IntentResponse{
				Action:        action,
				ResetContexts: false,
				Speech: lstResponse,
			},
		},
	}); err == nil {
		fmt.Printf(">>> updated intent = %+v\n", response)
	} else {
		fmt.Printf("*** error: %s\n", err)
	}
}

// delete an intent
func deleteIntent(client *ai.Client, response ai.ApiResponse)  {
	if response, err := client.DeleteIntent(response.Id); err == nil {
		fmt.Printf(">>> deleted intent = %+v\n", response)
	} else {
		fmt.Printf("*** error: %s\n", err)
	}
}

//create entity
func createEntity(client *ai.Client, entityName string, value string, synomyms []string,) int {
	response, err := client.CreateEntity(ai.EntityObject{
		Name: entityName,
		Entries: []ai.EntityEntryObject{
			ai.EntityEntryObject{
				Value: value,
				Synonyms: synomyms,
			},
		},
	});

	if err == nil {
		//fmt.Printf(">>> created entity = %+v\n", response)
		fmt.Println("Success: Created entity!")
		responseAll = response
		return 1
	} else {
		fmt.Printf("*** error: %s\n", err)
	}
	return 0
}

func main()  {

	trainCmd := flag.NewFlagSet("train", flag.ExitOnError)
	AddSharedFlags(trainCmd)

	testCmd := flag.NewFlagSet("test", flag.ExitOnError)
	AddSharedFlags(testCmd)

	if len(os.Args) < 2 {
		fmt.Println("Error: Input is not enough")
		os.Exit(1)
	}

	command := os.Args[1]
	if command == "train"{
		trainCmd.Parse(os.Args[2:])
	}else if command == "test" {
		testCmd.Parse(os.Args[2:])
	}

	if target != "intent" && target != "entity" {
		fmt.Println("Error: You must choose intent or entity")
		//fmt.Println(helpMessage)
		os.Exit(1)
	}

	if inputFP == "" {
		fmt.Println("Error: Input file is required but empty")
		//fmt.Println(helpMessage)
		os.Exit(1)
	}

	if token == "" {
		fmt.Println("Error: Token is required")
		//fmt.Println(helpMessage)
		os.Exit(1)
	}

	token := "72e1d493148a4a58826cb1fa3c7adfb4"

	// setup a client
	client := ai.NewClient(token)
	client.Verbose = false

	/*
	obj, err := ReadIntentsFromFile(inputFP)
	fmt.Print("OBJECT: \t"); fmt.Println(obj)
	fmt.Print("ERROR: \t"); fmt.Println(err)
	*/

	if trainCmd.Parsed() {
		if target == "intent"{
			var count int = 0
			obj, err := ReadIntentsFromFile(inputFP)
			if err != nil{
				fmt.Println(err)
			}
			for i := 0; i < len(obj); i++ {
				count += CreateIntent(client,obj[i].Intent,obj[i].Name, obj[i].Name, []string{obj[i].Response,})
			}
			fmt.Printf("Success: %f \n", float64(count) / float64(len(obj)) * 100)
		}else if target == "entity"{
			var count int = 0
			obj, err := ReadIntentsFromFile(inputFP)
			if err != nil{
				fmt.Println(err)
			}
			for i := 0; i < len(obj); i++ {
				count += createEntity(client,obj[i].Intent,obj[i].Name, []string{obj[i].Response,})
			}
			fmt.Printf("Success: %f \n", float64(count) / float64(len(obj)) * 100)
		}
	}

	if testCmd.Parsed() {
		if target == "intent"{
			count := 0
			obj, err := ReadIntentsFromFile(inputFP)
			if err != nil{
				fmt.Println(err)
			}

			intents, err := client.AllIntents();
			if  err != nil {
				fmt.Println(err)
			}
			for i := 0; i < len(obj); i++ {
				check := true
				for _, intent := range intents {
					// get an intent
					if response, err := client.Intent(intent.Id); err == nil {
						//fmt.Print("Convert 1:\t"); fmt.Print(convertUsaystoString(response.UserSays));fmt.Print("\tName 1:\t");fmt.Println(obj[i].Name)
						if containsArray(obj[i].Name, convertUsaystoString(response.UserSays)){
							fmt.Println("Correct!")
							count ++
							check = false

							/*
							fmt.Print("Convert 2:\t"); fmt.Print(convertIntentResponsetoString(response.Responses));fmt.Print("\tResponse 2:\t");fmt.Println(obj[i].Response)
							fmt.Print("RESPONSE:\t"); fmt.Println(response.Responses)
							if containsArray(obj[i].Response, convertIntentResponsetoString(response.Responses)){
								fmt.Println("Correct!")
								count ++
							}else{
								fmt.Println("No Correct 2!")
							}
							*/
						}
						//fmt.Print("RESPONE:\t"); fmt.Println(response)
						//fmt.Printf(">>> intent = %+v\n", response)
					} else {
						fmt.Printf("*** error: %s\n", err)
					}
				}
				if check{
					fmt.Println("No Correct!")
				}
			}
			fmt.Printf("Success: %f \n", float64(count) / float64(len(obj)) * 100)
		}else if target == "entity"{
			count := 0
			obj, err := ReadIntentsFromFile(inputFP)
			if err != nil{
				fmt.Println(err)
			}

			entitys, err := client.AllEntities();
			if  err != nil {
				fmt.Println(err)
			}
			for i := 0; i < len(obj); i++ {
				check := true
				for _, entity := range entitys {
					// get an intent
					if response, err := client.Entity(entity.Id); err == nil {
						//fmt.Print("\t obj[i].Name:\t");	fmt.Print(obj[i].Name)
						//fmt.Print("\t convertEntityEntryObjecttoString(response.Entries):\t");	fmt.Println(convertEntityEntryObjecttoString(response.Entries))
						if containsArray(obj[i].Name, convertEntityEntryObjecttoString(response.Entries)){
							fmt.Println("Correct!")
							count ++
							check = false
						}
					} else {
						fmt.Printf("*** error: %s\n", err)
					}
				}
				if check{
					fmt.Println("No Correct!")
				}
			}
			fmt.Printf("Success: %f \n", float64(count) / float64(len(obj)) * 100)
		}
	}
}

func containsArray(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func convertUsaystoString(usay []ai.UserSays) []string{
	var result []string
	for _, b := range usay {
		for _, c := range b.Data {
			result = append(result, c.Text)
		}
	}
	return result
}

func convertIntentResponsetoString(inte []ai.IntentResponse) []string{
	var result []string
	for _, b := range inte {
		for _, c := range b.Speech {
			result = append(result, c)
		}
	}
	return result
}

func convertEntityEntryObjecttoString(ent []ai.EntityEntryObject) []string{
	var result []string
	for _, b := range ent {
		result = append(result, b.Value)
	}
	return result
}