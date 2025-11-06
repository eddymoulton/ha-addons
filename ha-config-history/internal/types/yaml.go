package types

import (
	"crypto/sha1"
	"encoding/base64"

	"gopkg.in/yaml.v3"
)

func hashByteSlice(bytes []byte) string {
	hasher := sha1.New()
	hasher.Write(bytes)
	sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	return sha
}

func GetYamlNodeValue(yamlNode *yaml.Node, key string) string {
	for i := 0; i < len(yamlNode.Content)-1; i += 2 {
		keyNode := yamlNode.Content[i]
		valueNode := yamlNode.Content[i+1]
		if keyNode.Value == key {
			return valueNode.Value
		}
	}
	return "unknown"
}
