package common

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
	"io/ioutil"
	"net/http"
)

func GetBody(r *http.Request) []byte {
	buffer := new(bytes.Buffer)
	buffer.ReadFrom(r.Body)
	return buffer.Bytes()
}

func GetBodyResponse(r *http.Response) []byte {
	buffer := new(bytes.Buffer)
	buffer.ReadFrom(r.Body)
	return buffer.Bytes()
}

func ReadXML(b []byte, d interface{}) error {
	return xml.Unmarshal(b, d)
}

func WriteXML(w http.ResponseWriter, d interface{}, withHeader bool) error {
	xmlData, err := xml.Marshal(d)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/xml")
	if withHeader {
		fmt.Fprintf(w, "%s%s", xml.Header, xmlData)
	} else {
		fmt.Fprintf(w, "%s", xmlData)
	}
	return nil
}

func WriteJSON(w http.ResponseWriter, d interface{}) error {
	jsonData, err := json.Marshal(d)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s", jsonData)
	return nil
}

func ReadJSON(b []byte, d interface{}) error {
	return json.Unmarshal(b, d)
}

func UnmarshalResponse(c context.Context, resp *http.Response, value interface{}) error {

	DumpResponse(c, resp)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf(c, "Error while reading body: %v", err)
		return err
	}

	err = json.Unmarshal(body, value)
	if err != nil {
		log.Errorf(c, "Error while decoing JSON: %v", err)
		log.Infof(c, "JSON: %v", B2S(body))
		return err
	}

	return nil
}

func UnmarshalRequest(c context.Context, r *http.Request, value interface{}) error {

	body := GetBody(r)

	err := json.Unmarshal(body, value)
	if err != nil {
		log.Errorf(c, "Error while decoing JSON: %v", err)
		log.Infof(c, "JSON: %v", B2S(body))
		return err
	}

	return nil
}

func Marshal(c context.Context, value interface{}) string {
	data, err := json.Marshal(value)
	if err != nil {
		log.Errorf(c, "[common.Marshal] Error converting json: %v", err)
		return ""
	}
	return B2S(data)
}

func Left(s string, n int) string {
  if len(s)<n {
    return s
  }
  return s[:n]
}

