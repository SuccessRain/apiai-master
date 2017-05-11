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
var session string

func AddSharedFlags(fs *flag.FlagSet) {
	fs.StringVar(&target, "t", "", "required, intent or entity")
	fs.StringVar(&inputFP, "i", "", "required, path to the input file")
	fs.StringVar(&token, "token", "", "required, API.AI token")
	fs.StringVar(&session, "session", "", "required, API.AI session")
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

func convertStrinToUsersays(arr []string) []ai.UserSays{
	var uss []ai.UserSays
	for _, s := range arr {
		var us ai.UserSays
		us = ai.UserSays{
			Data: []ai.UserSaysData{
				ai.UserSaysData{
					Text: s,
				},

			},
		}
		uss = append(uss, us)
	}
	return uss
}

// create an intent
func CreateIntent(client *ai.Client, intentName string, userSays []string,
	action string, lstResponse []string, ) int{
	var count int = 0;
	response, err := client.CreateIntent(ai.IntentObject{
		Name:     intentName,
		Auto:     true,
		Priority: 5,
		UserSays: convertStrinToUsersays(userSays),
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
		fmt.Println(helpMessage)
		os.Exit(1)
	}

	command := os.Args[1]
	if command == "train"{
		trainCmd.Parse(os.Args[2:])
	}else if command == "test" {
		testCmd.Parse(os.Args[2:])
	}else if command == "help" {
		fmt.Println(helpMessage)
		os.Exit(0)
	}

	if target != "intent" && target != "entity" {
		fmt.Println("Error: You must choose intent or entity")
		fmt.Println(helpMessage)
		os.Exit(1)
	}

	if inputFP == "" {
		fmt.Println("Error: Input file is required but empty")
		fmt.Println(helpMessage)
		os.Exit(1)
	}

	if token == "" {
		fmt.Println("Error: Token is required")
		fmt.Println(helpMessage)
		os.Exit(1)
	}
	if session == "" {
		fmt.Println("Error: Session is required")
		fmt.Println(helpMessage)
		os.Exit(1)
	}

	//target = "intent"
	//token := "72e1d493148a4a58826cb1fa3c7adfb4"


	// setup a client
	client := ai.NewClient(session)
	client.Verbose = false

	//CreateIntent(client,"intent_Name",[]string{"C", "A", "B"}, "A", []string{"D","E","F"})
	/*
	qr, _ := client.Query(ai.Query{Query: []string{"VTV Cab",}})
	fmt.Println(qr)

	action := qr.Result.Action
	value := qr.Result.Fulfillment.Speech

	fmt.Print("ACTION:\t"); fmt.Println(action)
	fmt.Print("VALUE:\t"); fmt.Println(value)
	fmt.Print("FULL:\t"); fmt.Println(qr.Result)
	fmt.Print("METADATA:\t"); fmt.Println(qr.Result.Metadata.IntentName)
	*/

	//**********

	if trainCmd.Parsed() {
		if target == "intent"{
			var count int = 0
			obj, err := ReadIntentsFromFile(inputFP)
			if err != nil{
				fmt.Println(err)
			}
			multi := recycleInentName(obj)
			for i := 0; i < len(multi); i++ {
				count += CreateIntent(client,multi[i].intentName,multi[i].values, multi[i].intentName, []string{obj[i].Response,})
			}
			fmt.Printf("Success: %f \n", float64(count) / float64(len(multi)) * 100)
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
			var count int = 0
			obj, err := ReadIntentsFromFile(inputFP)
			if err != nil{
				fmt.Println(err)
			}
			for i := 0; i < len(obj); i++ {
				//fmt.Print("TOKEN:\t"); fmt.Println(token)
				qr, _ := client.Query(ai.Query{Query: []string{obj[i].Name,}},token,session)
				if qr != nil{
					intentName := qr.Result.Metadata.IntentName
					if strings.Compare(obj[i].Intent, intentName) == 0 {
						fmt.Print(i); fmt.Println(".\t Correct!")
						count ++;
					}else if strings.Compare("Default Fallback Intent", intentName) != 0 {
						fmt.Print(i); fmt.Print(".\t No Correct!\t");fmt.Println(intentName)
					}else{
						fmt.Print(i); fmt.Println(".\t No Correct!")
					}
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

type ObjectMutilValue struct{
	intentName string
	values []string
}

func recycleInentName(objs []ObjectFile)  []ObjectMutilValue {

	var omvs []ObjectMutilValue

	data := make(map[string][]string)
	for _, a := range objs {
		samples := data[a.Intent]
		samples = append(samples, a.Name)
		data[a.Intent] = samples
	}

	for key, value := range data {

		var omv ObjectMutilValue
		omv.intentName = key
		omv.values = value
		omvs = append(omvs, omv)
	}

	return omvs

	/*
	var omvs []ObjectMutilValue
	var omvCheck []string
	for _, a := range objs {
		var omv ObjectMutilValue
		fmt.Print("A.INTENT:\t"); fmt.Println(a.Intent)
		fmt.Print("LENGTH OMVCHECK:\t"); fmt.Println(len(omvCheck))
		if containsArray(a.Intent, omvCheck) == false {
			fmt.Println("VAO FALSE")
			omv.intentName = a.Intent
			omv.values = append(omv.values, a.Name)

			omvCheck = append(omvCheck, a.Intent)
			omvs = append(omvs, omv)
		}else{
			fmt.Println("VAO TRUE")
			for _, b := range omvs {
				if b.intentName == a.Intent{
					fmt.Print("VAO TRONG IF:\t"); fmt.Println(a.Name)
					fmt.Print("VAO TRONG IF - LENGTH VLUES 1:\t"); fmt.Println(len(b.values))
					b.values = append(b.values, a.Name)

					fmt.Print("VAO TRONG IF - LENGTH VLUES 2:\t"); fmt.Println(len(b.values))
					break
				}
			}
		}
	}

	fmt.Print("OMVCHECK:\t"); fmt.Println(len(omvCheck))
	fmt.Print("MOT PHAN TU:\t"); fmt.Println(omvCheck[0])

	fmt.Print("AHIHIHIHIHI:\t"); fmt.Println(len(omvs))
	fmt.Print("MOT PHAN TU:\t"); fmt.Println(omvs[0])
	return omvs
	*/
}

const helpMessage string = `
api is CLI tool that helps you train and test API.AI in terminal

Usage: api <command> <option>
Available commands and corresponding options:
	train
	  -t string
	    	required, type of training (intent, entity)
	  -i string
	    	required, path to your input file
	  -token string
	  		required, API.AI token

	test
	  -t string
	    	required, type of training (intent, entity)
	  -i string
	    	required, path to your input file
	  -token string
	  		required, API.AI token

	help
`