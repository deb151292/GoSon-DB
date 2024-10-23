package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

// Directory where JSON files will be stored

// Structure of the JSON file (stores any kind of data)
type jSONFile struct {
	Data []interface{} `json:"data"`
}

// Function to create the directory if it doesn't exist
func ensureDir(database string) (string, error) {
	if _, err := os.Stat(database); os.IsNotExist(err) {
		err := os.Mkdir(database, 0755)
		if err != nil {
			return "exist", fmt.Errorf("database already exist")
		}
	} else {
		return "exist", fmt.Errorf("database already exist")
	}
	return "Database Created successfully", nil
}

// Function to create a new JSON file
func createtable(database string, table string) (string, error) {

	msg, err := ensureDir(database)
	if msg == "exist" {
		fmt.Println("Database Found, creating new table...")
	} else if err == nil {
		fmt.Println("Database Not Found, creating Database and table...")

	}

	filename := strings.ToLower(table)

	filePath := filepath.Join(database, filename+".json")

	// Check if the file already exists
	if _, err := os.Stat(filePath); err == nil {
		return "", fmt.Errorf("%s table already exists", filename)
	}

	jsonData := jSONFile{
		Data: []interface{}{},
	}

	fileContent, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		return "", err
	}

	err = os.WriteFile(filePath, fileContent, 0644)
	if err != nil {
		return "", err
	}

	fmt.Printf("%s table created successfully.\n", filename)
	return fmt.Sprintf("File %s table created successfully.\n", filename), nil
}

func loadJsonFile(fileName string, database string) (jSONFile, error) {
	var jsonData jSONFile
	dir, _ := os.Getwd()
	filePath := filepath.Join(dir, "/"+database+"/", fileName+".json")
	file, err := os.Open(filePath)
	if err != nil {
		return jsonData, err
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return jsonData, err
	}

	err = json.Unmarshal(bytes, &jsonData)
	return jsonData, err
}

// Function to populate a struct from a map using reflection
func populateStruct(data map[string]interface{}, dest interface{}) error {
	destVal := reflect.ValueOf(dest).Elem() // Ensure dest is a pointer
	for key, value := range data {
		fieldVal := destVal.FieldByName(key)
		if fieldVal.IsValid() && fieldVal.CanSet() {
			val := reflect.ValueOf(value)
			if fieldVal.Type() == val.Type() {
				fieldVal.Set(val)
			} else if fieldVal.Kind() == reflect.Float64 && val.Kind() == reflect.Float64 {
				fieldVal.Set(val)
			} else if fieldVal.Kind() == reflect.String && val.Kind() == reflect.String {
				fieldVal.SetString(val.String())
			}
		}
	}
	return nil
}

// Function to save JSON data to a file
func saveJsonFile(fileName string, database string, jsonData *jSONFile) error {
	filePath := filepath.Join(database, fileName+".json")
	fileContent, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, fileContent, 0644)
}

// ValidateStruct dynamically validates fields of a struct
func ValidateStruct(s interface{}) error {
	v := reflect.ValueOf(s)

	// Ensure we're working with a struct
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("input must be a struct")
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)           // Field metadata
		fieldValue := v.Field(i).Interface() // Field value
		fieldTag := field.Tag.Get("json")    // JSON tag name
		ifnullTag := field.Tag.Get("ifnull") // ifnull tag

		// Check for "ifnull:notnull" flag
		if ifnullTag == "notnull" && IsZeroValue(fieldValue) {
			return fmt.Errorf(`please provide "%s"`, fieldTag)
		}

		// Perform standard validation dynamically based on field types
		switch field.Type.Kind() {
		case reflect.String:
			if ifnullTag == "notnull" && fieldValue == "" {
				return fmt.Errorf(`violate not null constraint on "%s"`, fieldTag)
			}
		case reflect.Float64, reflect.Int, reflect.Int64, reflect.Float32:
			// Check numeric types for being non-zero or positive
			if ifnullTag == "notnull" && reflect.ValueOf(fieldValue).Float() <= 0 {
				return fmt.Errorf(`please provide a positive value for "%s"`, fieldTag)
			}
		}
	}
	return nil
}

// Helper function to check if a value is a zero value for its type (e.g., "" for string, 0 for int)
func IsZeroValue(value interface{}) bool {
	return reflect.DeepEqual(value, reflect.Zero(reflect.TypeOf(value)).Interface())
}

// Function to get JSON field names
func GetJSONFieldNames(v interface{}) []string {
	var fieldNames []string
	value := reflect.ValueOf(v)

	// Ensure we are dealing with a struct
	if value.Kind() == reflect.Struct {
		t := value.Type()
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			jsonTag := field.Tag.Get("json")
			if jsonTag != "" {
				// The jsonTag can contain options like "name,omitempty"
				jsonFieldName := jsonTag
				if commaIdx := strings.Index(jsonTag, ","); commaIdx != -1 {
					jsonFieldName = jsonTag[:commaIdx] // Get the part before any comma
				}
				fieldNames = append(fieldNames, jsonFieldName)
			}
		}
	}
	return fieldNames
}
func StringSlicesEqual(arr1, arr2 []string) bool {
	if len(arr1) != len(arr2) {
		return false
	}

	count := make(map[string]int)

	// Count occurrences of each string in the first slice
	for _, str := range arr1 {
		count[str]++
	}

	// Decrease the count for each string in the second slice
	for _, str := range arr2 {
		count[str]--
		if count[str] < 0 {
			return false // Found an extra string in arr2
		}
	}

	return true // All counts should be zero
}

// Function to validate and update struct using reflection
func validateAndUpdateStruct(existingData interface{}, updatedData interface{}) (interface{}, error) {
	var updateValue reflect.Value
	if updatedData != nil {
		var keys []string
		for key := range existingData.(map[string]interface{}) {
			keys = append(keys, key)
		}
		StructKay := GetJSONFieldNames(updatedData)

		updateValue = reflect.ValueOf(updatedData)

		if !StringSlicesEqual(keys, StructKay) {
			return nil, errors.New("update data does not match table fields")
		}
		// Ensure both are structs
		if updateValue.Kind() != reflect.Struct || len(keys) < 1 {
			return nil, errors.New("expected both values to be not empty")
		}
	}

	if updatedData == nil {
		err := ValidateStruct(existingData)
		return nil, err
	}
	var i interface{}
	if updatedData != nil {
		updatedMapData := existingData.(map[string]interface{})
		UpdateMapFields(updatedMapData, updatedData)
		// Assign it to an interface{}
		i = updatedMapData
	}

	return i, nil
}

// Function to update fields in the original map with non-zero fields from the input user struct
func UpdateMapFields(original map[string]interface{}, updates interface{}) {
	// Use reflection to iterate over the fields of the struct
	updatesVal := reflect.ValueOf(updates)

	for i := 0; i < updatesVal.NumField(); i++ {
		// Get the field name based on the struct's JSON tag
		field := updatesVal.Type().Field(i)
		jsonTag := field.Tag.Get("json")

		// Check if the update field is non-zero (non-default)
		if !isZero(updatesVal.Field(i)) {
			// Update the original map's field
			original[jsonTag] = updatesVal.Field(i).Interface()
		}
	}
}

// Helper function to check if a field is zero-valued
func isZero(v reflect.Value) bool {
	return v.Interface() == reflect.Zero(v.Type()).Interface()
}

func ModifyField(v interface{}, fieldName string, newValue interface{}) error {
	value := reflect.ValueOf(v)

	// Check if the value is a pointer to a struct
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	// Ensure we are dealing with a struct
	if value.Kind() != reflect.Struct {
		return fmt.Errorf("expected a struct, got %s", value.Kind())
	}

	// Get the field by name
	field := value.FieldByName(fieldName)
	if !field.IsValid() {
		return fmt.Errorf("no such field: %s", fieldName)
	}

	// Check if the field is settable
	if !field.CanSet() {
		return fmt.Errorf("cannot set field: %s", fieldName)
	}

	// Set the new value, using reflection to convert the type if necessary
	newValueReflect := reflect.ValueOf(newValue)
	if newValueReflect.Type() != field.Type() {
		return fmt.Errorf("provided value type didn't match struct field type")
	}

	field.Set(newValueReflect)
	return nil
}

// ExtractFilterMap extracts non-zero fields from any struct into a map
func extractFilterMap(data interface{}) map[string]interface{} {
	filter := make(map[string]interface{})
	v := reflect.ValueOf(data)

	// Ensure we are dealing with a struct
	if v.Kind() == reflect.Ptr {
		v = v.Elem() // Get the underlying struct if it's a pointer
	}

	if v.Kind() != reflect.Struct {
		return filter // Return empty map if not a struct
	}

	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		fieldValue := v.Field(i)

		// Check if the field is non-zero
		if !reflect.DeepEqual(fieldValue.Interface(), reflect.Zero(fieldValue.Type()).Interface()) {
			// Use the json tag as the key or field name directly
			jsonTag := field.Tag.Get("json")
			if jsonTag != "" {
				filter[jsonTag] = fieldValue.Interface()
			} else {
				// If no tag, use the field name (assuming it's lowercase)
				filter[field.Name] = fieldValue.Interface()
			}
		}
	}

	return filter
}

func ConvertMapsToInterfaces(maps []map[string]interface{}) []interface{} {
	var interfaces []interface{}
	for _, m := range maps {
		interfaces = append(interfaces, m)
	}
	return interfaces
}
