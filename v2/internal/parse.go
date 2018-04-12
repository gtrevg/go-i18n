package internal

import (
	"fmt"
	"os"

	"golang.org/x/text/language"
)

// UnmarshalFunc unmarshals data into v.
type UnmarshalFunc func(data []byte, v interface{}) error

// MessageFile represents a parsed message file.
type MessageFile struct {
	Path     string
	Tag      language.Tag
	Format   string
	Messages []*Message
}

// ParseMessageFileBytes returns the messages parsed from file.
func ParseMessageFileBytes(buf []byte, path string, unmarshalFuncs map[string]UnmarshalFunc) (*MessageFile, error) {
	lang, format := parsePath(path)
	tag, err := language.Parse(lang)
	if err != nil {
		return nil, err
	}
	messageFile := &MessageFile{
		Path:   path,
		Tag:    tag,
		Format: format,
	}
	if len(buf) == 0 {
		return messageFile, nil
	}
	var unmarshalFunc UnmarshalFunc
	if unmarshalFuncs != nil {
		unmarshalFunc = unmarshalFuncs[messageFile.Format]
	}
	if unmarshalFunc == nil {
		return nil, fmt.Errorf("no unmarshaler registered for %s", messageFile.Format)
	}
	var raw interface{}
	if err := unmarshalFunc(buf, &raw); err != nil {
		return nil, err
	}
	switch data := raw.(type) {
	case map[string]interface{}:
		messageFile.Messages = make([]*Message, 0, len(data))
		for id, data := range data {
			m, err := NewMessage(data)
			if err != nil {
				return nil, err
			}
			m.ID = id
			messageFile.Messages = append(messageFile.Messages, m)
		}
	case map[interface{}]interface{}:
		messageFile.Messages = make([]*Message, 0, len(data))
		for id, data := range data {
			strid, ok := id.(string)
			if !ok {
				return nil, fmt.Errorf("expected key to be string but got %#v", id)
			}
			m, err := NewMessage(data)
			if err != nil {
				return nil, err
			}
			m.ID = strid
			messageFile.Messages = append(messageFile.Messages, m)
		}
	case []interface{}:
		// Backward compatability for v1 file format.
		messageFile.Messages = make([]*Message, 0, len(data))
		for _, data := range data {
			m, err := NewMessage(data)
			if err != nil {
				return nil, err
			}
			messageFile.Messages = append(messageFile.Messages, m)
		}
	default:
		return nil, fmt.Errorf("unsupported file format %T", raw)
	}
	return messageFile, nil
}

func parsePath(path string) (langTag, format string) {
	formatStartIdx := -1
	for i := len(path) - 1; i >= 0; i-- {
		c := path[i]
		if os.IsPathSeparator(c) {
			if formatStartIdx != -1 {
				langTag = path[i+1 : formatStartIdx]
			}
			return
		}
		if path[i] == '.' {
			if formatStartIdx != -1 {
				langTag = path[i+1 : formatStartIdx]
				return
			}
			if formatStartIdx == -1 {
				format = path[i+1:]
				formatStartIdx = i
			}
		}
	}
	if formatStartIdx != -1 {
		langTag = path[:formatStartIdx]
	}
	return
}
