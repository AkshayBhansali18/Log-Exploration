package main

import (
	"context"
	"encoding/json"
    "github.com/gin-gonic/gin"
	"net/http"
	"crypto/tls"
	"fmt"
	"flag"
	//"os"

	//"reflect"
	//"io/ioutil"
	"github.com/olivere/elastic/v7"
	//"github.com/elastic/go-elasticsearch/v7"
)

var esClient *elastic.Client
func main() {

	esAddress := flag.String("esaddress", "https://localhost:9200", "ElasticSearch Server Address")
	esCert := flag.String("es-cert","admin-cert","admin-cert file location")
	esKey := flag.String("es-key","admin-key","admin-key file location")
	flag.Parse()
	esClient = InitializeElasticSearchClient(esAddress,esCert,esKey)

	r := gin.Default() //initialise
	//fmt.Println(reflect.Type(esClient))
	//endpoint to get all logs
	r.GET("/", getAllLogs)

	//endpoint to get infrastructure logs
	r.GET("/infra", getInfrastructureLogs)

	//endpoint to get application logs
	r.GET("/app", getApplicationLogs)

	//endpoint to get audit logs
	r.GET("/audit", getAuditLogs)
	//endpoint to filter logs by start and finish time - please enter time in the following format- HH:MM:SS
	r.GET("/filter/:startTime/:finishTime", filterByTime)
	r.Run()
}
func filterByTime(c *gin.Context){
	startTime:= c.Params.ByName("startTime")
	finishTime:= c.Params.ByName("finishTime")
	fmt.Println(startTime)
	fmt.Println(finishTime)
	ctx := context.Background()

	searchResult, err := esClient.Search().  //Get all the logs from all the indexes
		Index("infra-000001","app-000001","audit-000001").
		Pretty(true).
		Do(ctx)
	if err != nil {
		panic("Get error occurred")
	}

	var logs[] string // create a slice of type string to append logs to
	for _, hit := range searchResult.Hits.Hits { //iterate through the logs

		var log map[string]interface{} //to convert logs from JSON to map[string]interface

		err := json.Unmarshal(hit.Source, &log) //convert logs from JSON to map[string]interface
		if err != nil {
			// Deserialization failed
			fmt.Println(err)
		} else {
			for key, value := range log {
				if(key == "@timestamp"){
					str:= fmt.Sprintf("%v",value)
					str=str[11:19]
					if str>= startTime && str<= finishTime{

						validLog := fmt.Sprintf("%v", log) //convert log to string
						validLog = validLog+"\n"
						logs = append(logs, validLog)  //append log to slice logs of type string since it lies between start and end times
					}
				}
			}
		}
	}
	c.JSON(200, gin.H{
		"Logs": logs, //return logs
	})
}
func InitializeElasticSearchClient(esAddress *string,esCert *string,esKey *string) *elastic.Client {

	cert, err := tls.LoadX509KeyPair(*esCert, *esKey)
	if err != nil {
		fmt.Println(err)
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}


	client = &http.Client{ //configure client
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true,
				Certificates: []tls.Certificate{cert},},

		},
	}
	esClient, _ := elastic.NewClient( //initialise esClient
		elastic.SetHttpClient(client),
		elastic.SetURL(*esAddress),
		elastic.SetScheme("https"),
		elastic.SetSniff(false),
	)
	return esClient


}
func getInfrastructureLogs(c *gin.Context) {
	ctx := context.Background()

	searchResult, err := esClient.Search().  //Get all the logs from the infrastructure index
		Index("infra-000001").
		Pretty(true).
		Do(ctx)
	if err != nil {
		panic("Get error occurred")
	}
	var logs[] string
	 logs = getRelevantLogs(searchResult) // create a slice of type string to append logs to

	c.JSON(200, gin.H{
		"Logs": logs, //return logs
	})

}
func getAllLogs(c *gin.Context){
	ctx := context.Background()

	searchResult, err := esClient.Search().  //Get all the logs from all the indexes
		Index("infra-000001","app-000001","audit-000001").
		Pretty(true).
		Do(ctx)
	if err != nil {
		panic("Get error occurred")
	}

	var logs[] string // create a slice of type string to append logs to
	logs = getRelevantLogs(searchResult)

	c.JSON(200, gin.H{
		"Logs": logs, //return logs
	})

}
func getApplicationLogs(c *gin.Context){
	ctx := context.Background()

	searchResult, err := esClient.Search().  //Get all the logs from the application index
		Index("app-000001").
		Pretty(true).
		Do(ctx)
	if err != nil {
		panic("Get error occurred")
	}

	var logs[] string // create a slice of type string to append logs to
	logs = getRelevantLogs(searchResult)
	c.JSON(200, gin.H{
		"Logs": logs, //return logs
	})

}
func getAuditLogs(c *gin.Context){
	ctx := context.Background()

	searchResult, err := esClient.Search().  //Get all the logs from the audit index
		Index("audit-000001").
		Pretty(true).
		Do(ctx)
	if err != nil {
		panic("Get error occurred")
	}
	var logs[] string
	logs = getRelevantLogs(searchResult)

	c.JSON(200, gin.H{
		"Logs": logs, //return logs
	})

}
func getRelevantLogs(searchResult *elastic.SearchResult) []string{
	var logs[] string // create a slice of type string to append logs to
	for _, hit := range searchResult.Hits.Hits { //iterate through the logs

		var log map[string]interface{} //to convert logs from JSON to map[string]interface

		err := json.Unmarshal(hit.Source, &log) //convert logs from JSON to map[string]interface
		if err != nil {
			// Deserialization failed
			fmt.Println(err)
		} else {
			str := fmt.Sprintf("%v", log) //convert log to string
			str = str+"\n"
			logs = append(logs, str) //append log to slice logs of type string
			fmt.Println(log, "\n")
		}
	}
	timeTaken:=fmt.Sprintf("Query took %d milliseconds", searchResult.TookInMillis)
	totalLogsFound:=fmt.Sprintf("Found a total of %d logs", searchResult.TotalHits())
	logs = append(logs,timeTaken)
	logs = append(logs,totalLogsFound)
	return logs
}


