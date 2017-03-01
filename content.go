package confluence

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type Storage struct {
	Value          string `json:"value"`
	Representation string `json:"representation"`
}

type Body struct {
	Storage Storage `json:"storage"`
}

type Version struct {
	Number int `json:"number"`
}

type Ancestor struct {
	Id string `json:"id,omitempty"`
}

type Space struct {
	Key string `json:"key,omitempty"`
}

type Content struct {
	Id          string     `json:"id,omitempty"`
	Type        string     `json:"type,omitempty"`
	Status      string     `json:"status,omitempty"`
	Title       string     `json:"title,omitempty"`
	Body        Body       `json:"body,omitempty"`
	Version     Version    `json:"version,omitempty"`
	Ancestors   []Ancestor `json:"ancestors,omitempty"`
	Space       Space      `json:"space,omitempty"`
	LabelPrefix string     `json:"prefix,omitempty"`
	LabelName   string     `json:"name,omitempty"`
	Attachments []string
}

func (w *Wiki) contentEndpoint(contentID string) (*url.URL, error) {
	return url.ParseRequestURI(w.endPoint.String() + "/content/" + contentID)
}

func (w *Wiki) DeleteContent(contentID string) error {
	contentEndPoint, err := w.contentEndpoint(contentID)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("DELETE", contentEndPoint.String(), nil)
	if err != nil {
		return err
	}

	_, err = w.sendRequest(req)
	if err != nil {
		return err
	}
	return nil
}

func (w *Wiki) GetContent(contentID string, expand []string) (*Content, error) {
	contentEndPoint, err := w.contentEndpoint(contentID)
	if err != nil {
		return nil, err
	}
	data := url.Values{}
	data.Set("expand", strings.Join(expand, ","))
	contentEndPoint.RawQuery = data.Encode()

	req, err := http.NewRequest("GET", contentEndPoint.String(), nil)
	if err != nil {
		return nil, err
	}

	res, err := w.sendRequest(req)
	if err != nil {
		return nil, err
	}

	var content Content
	err = json.Unmarshal(res, &content)
	if err != nil {
		return nil, err
	}

	return &content, nil
}

func (w *Wiki) UpdateContent(content *Content) (*Content, error) {
	jsonbody, err := json.Marshal(content)
	if err != nil {
		return nil, err
	}

	contentEndPoint, err := w.contentEndpoint(content.Id)
	req, err := http.NewRequest("PUT", contentEndPoint.String(), strings.NewReader(string(jsonbody)))
	req.Header.Add("Content-Type", "application/json")

	res, err := w.sendRequest(req)
	if err != nil {
		return nil, err
	}

	var newContent Content
	err = json.Unmarshal(res, &newContent)
	if err != nil {
		return nil, err
	}

	return &newContent, nil
}

func (w *Wiki) CreateContent(content *Content) (*Content, error) {
	jsonbody, err := json.Marshal(content)
	if err != nil {
		return nil, err
	}

	contentEndPoint, err := w.contentEndpoint("")
	req, err := http.NewRequest("POST", contentEndPoint.String(), strings.NewReader(string(jsonbody)))
	req.Header.Add("Content-Type", "application/json")

	res, err := w.sendRequest(req)
	if err != nil {
		return nil, err
	}

	var newContent Content
	err = json.Unmarshal(res, &newContent)
	if err != nil {
		return nil, err
	}

	return &newContent, nil
}

func (w *Wiki) AddLabel(content *Content) (*Content, error) {
	jsonbody, err := json.Marshal(content)
	if err != nil {
		return nil, err
	}

	contentEndPoint, err := w.contentEndpoint(content.Id + "/label")
	req, err := http.NewRequest("POST", contentEndPoint.String(), strings.NewReader(string(jsonbody)))
	req.Header.Add("Content-Type", "application/json")

	res, err := w.sendRequest(req)
	if err != nil {
		return nil, err
	}

	var newContent Content
	err = json.Unmarshal(res, &newContent)
	if err != nil {
		return nil, err
	}

	return &newContent, nil
}

func (w *Wiki) AddAttachments(content *Content) (*Content, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for _, file := range content.Attachments {
		f, err := os.Open(file)
		if err != nil {
			continue
		}
		defer f.Close()

		part, err := writer.CreateFormFile("file", filepath.Base(file))
		if err != nil {
			return nil, err
		}

		_, err = io.Copy(part, f)
		if err != nil {
			return nil, err
		}
	}

	err := writer.Close()
	if err != nil {
		return nil, err
	}

	contentEndPoint, err := w.contentEndpoint(content.Id + "/child/attachment")
	req, err := http.NewRequest("POST", contentEndPoint.String(), body)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	req.Header.Add("X-Atlassian-Token", "no-check")

	res, err := w.sendRequest(req)
	if err != nil {
		return nil, err
	}

	var newContent Content
	err = json.Unmarshal(res, &newContent)
	if err != nil {
		return nil, err
	}

	return &newContent, nil
}
