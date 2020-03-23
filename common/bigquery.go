package common

import (
	"errors"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	bigquery "google.golang.org/api/bigquery/v2"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
	"net/http"
	"time"
)

func GetBQServiceAccountClient(c context.Context) (*bigquery.Service, error) {

	serviceAccountClient := &http.Client{
		Transport: &oauth2.Transport{
			Source: google.AppEngineTokenSource(c,
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/bigquery"),
			Base: &urlfetch.Transport{
				Context: c,
			},
		},
	}
	return bigquery.New(serviceAccountClient)
}

func CreateTableInBigQuery(c context.Context, newTable *bigquery.Table) error {

	if newTable == nil {
		return errors.New("No newTable defined for CreateTableInBigQuery")
	}

	if newTable.TableReference == nil {
		return errors.New("No newTable.TableReference defined for CreateTableInBigQuery")
	}

	if newTable.Schema == nil {
		return errors.New("No newTable.Schema defined for CreateTableInBigQuery")
	}

	bqServiceAccountService, err := GetBQServiceAccountClient(c)
	if err != nil {
		log.Errorf(c, "Error getting BigQuery Service: %v", err)
		return err
	}

	err = bigquery.
		NewTablesService(bqServiceAccountService).
		Delete(
		newTable.TableReference.ProjectId,
		newTable.TableReference.DatasetId,
		newTable.TableReference.TableId).
		Do()
	if err != nil {
		log.Warningf(c, "There was an error while trying to delete old snapshot table: %v", err)
	}

	_, err = bigquery.
		NewTablesService(bqServiceAccountService).
		Insert(
		newTable.TableReference.ProjectId,
		newTable.TableReference.DatasetId,
		newTable).
		Do()

	return err
}

func StreamDataInBigquery(c context.Context, projectId, datasetId, tableId string, req *bigquery.TableDataInsertAllRequest) error {

	if req == nil {
		return errors.New("No req defined for StreamDataInBigquery")
	}

	bqServiceAccountService, err := GetBQServiceAccountClient(c)
	if err != nil {
		log.Errorf(c, "Error getting BigQuery Service: %v", err)
		return err
	}

	resp, err := bigquery.
		NewTabledataService(bqServiceAccountService).
		InsertAll(projectId, datasetId, tableId, req).
		Do()
	if err != nil {
		log.Warningf(c, "Error streaming data to Big Query, trying again in 10 seconds: %v", err)
		time.Sleep(time.Second * 10)
		resp, err = bigquery.
			NewTabledataService(bqServiceAccountService).
			InsertAll(projectId, datasetId, tableId, req).
			Do()
		if err != nil {
			log.Errorf(c, "Error again streaming data to Big Query: %v", err)
			return err
		} else {
			log.Infof(c, "2nd try was successful")
		}
	}

	isError := false
	for i, insertError := range resp.InsertErrors {
		if insertError != nil {
			for j, e := range insertError.Errors {
				if (e.DebugInfo != "") || (e.Message != "") || (e.Reason != "") {
					log.Errorf(c, "BigQuery error %v: %v at %v/%v", e.Reason, e.Message, i, j)
					isError = true
				}
			}
		}
	}

	if isError {
		return errors.New("There was an error streaming data to Big Query")
	}

	return nil

}
