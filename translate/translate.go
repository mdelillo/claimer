package translate

import (
	"fmt"
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v2"
)

var translations map[interface{}]interface{}

type TArgs map[string]string

func LoadTranslationFile(path string) error {
	translations = nil

	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %s", path)
	}

	if err := yaml.Unmarshal(contents, &translations); err != nil {
		return fmt.Errorf("failed to parse YAML: %s", contents)
	}

	return nil
}

func LoadTranslations(contents string) error {
	if err := yaml.Unmarshal([]byte(contents), &translations); err != nil {
		return fmt.Errorf("failed to parse YAML: %s", contents)
	}

	return nil
}

func T(yamlPath string, vars TArgs) string {
	splitPath := strings.Split(yamlPath, ".")
	innerTranslations := translations
	translatedString := ""
	for i, key := range splitPath {
		value, ok := innerTranslations[key]
		if !ok {
			return yamlPath
		}

		stringValue, ok := value.(string)
		if ok && i == (len(splitPath)-1) {
			translatedString = stringValue
			break
		}

		mapValue, ok := innerTranslations[key].(map[interface{}]interface{})
		if !ok {
			return yamlPath
		}

		innerTranslations = mapValue
	}

	for name, value := range vars {
		translatedString = strings.Replace(translatedString, fmt.Sprintf("{{.%s}}", name), value, -1)
	}

	return translatedString
}