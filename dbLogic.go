package goSon

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
	"sync"
)

type DatabaseInfo struct {
	Database string `json:"database"`
}

var (
	dbLock = sync.Mutex{}
)

// Function to check the goSon database connection
func Connect_GoSon() string {
	return "Welcome to GoSon Database, You are ready to connect"
}

func (db DatabaseInfo) CreateDatabase() (string, error) {
	dbLock.Lock()
	defer dbLock.Unlock()
	msg, err := ensureDir(db.Database)
	if err != nil {
		return "", err
	}
	return msg, nil
}

func (db DatabaseInfo) CreateTable(table string) (string, error) {
	dbLock.Lock()
	defer dbLock.Unlock()
	return createtable(db.Database, table)
}

// Function to add data to an existing JSON file
func (db DatabaseInfo) AddData(Table string, newData interface{}) error {
	dbLock.Lock()
	defer dbLock.Unlock()
	fileName := strings.ToLower(Table)
	// Validate new data before adding
	if err := ValidateStruct(newData); err != nil {
		return err
	}
	// Use reflection to update the struct
	if _, err := validateAndUpdateStruct(newData, nil); err != nil {
		return err
	}

	jsonData, err := loadJsonFile(fileName, db.Database)
	if err != nil {
		log.Println("Inserting First Data set")
	}
	index := 0
	if len(jsonData.Data) != 0 {
		index = len(jsonData.Data) - 1
	}

	if index == 0 {
		// Add the new data to the slice
		jsonData.Data = append(jsonData.Data, newData)
	} else {
		// Insert the new data in the middle of the slice
		jsonData.Data = append(jsonData.Data[:index+1], newData)
	}
	err = saveJsonFile(fileName, db.Database, &jsonData)

	if err != nil {
		return err
	}

	fmt.Printf("Data added to %s successfully.\n", fileName)
	return nil
}

// Function to update data in a JSON file by index (uses reflection to update fields)
func (db DatabaseInfo) UpdateData(Table string, index int, updatedData interface{}) error {
	dbLock.Lock()
	defer dbLock.Unlock()
	fileName := strings.ToLower(Table)
	jsonData, err := loadJsonFile(fileName, db.Database)
	if err != nil {
		return err
	}
	log.Println(jsonData.Data)
	if index < 0 || index >= len(jsonData.Data) {
		return fmt.Errorf("index out of bounds in file %s", fileName)
	}
	var updatedStruct interface{}
	var ValidateErr error
	// Use reflection to update the struct
	if updatedStruct, ValidateErr = validateAndUpdateStruct(jsonData.Data[index], updatedData); err != nil {
		return ValidateErr
	}
	jsonData.Data[index] = updatedStruct
	SaveErr := saveJsonFile(fileName, db.Database, &jsonData)
	if SaveErr != nil {
		return SaveErr
	}

	fmt.Printf("Data updated in %s successfully.\n", fileName)
	return nil
}

// Function to delete data from JSON file by index
func (db DatabaseInfo) DeleteData(Table string, index int) error {
	dbLock.Lock()
	defer dbLock.Unlock()

	fileName := strings.ToLower(Table)

	jsonData, err := loadJsonFile(fileName, db.Database)
	if err != nil {
		return err
	}

	if index < 0 || index >= len(jsonData.Data) {
		return fmt.Errorf("index out of bounds in file %s", fileName)
	}

	// Remove data at the given index
	jsonData.Data = append(jsonData.Data[:index], jsonData.Data[index+1:]...)
	err = saveJsonFile(fileName, db.Database, &jsonData)
	if err != nil {
		return err
	}

	fmt.Printf("Data deleted from %s table successfully.\n", fileName)
	return nil
}

/* FilterByFields filters a data based on specified filter criteria | filterOption == "" to retrieve all records
 */
func (db DatabaseInfo) FindMany(Table string, filterOptions interface{}) ([]byte, error) {
	dbLock.Lock()
	defer dbLock.Unlock()
	var results []map[string]interface{}

	// Extract filter criteria from the filterOptions and convert interface{} to map[string]interface{}

	fileName := strings.ToLower(Table)

	jsonData, err := loadJsonFile(fileName, db.Database)
	if err != nil {
		return nil, errors.New("Error: failed to load table data.")
	}
	if !IsZeroValue(filterOptions) {
		data, _ := json.Marshal(jsonData.Data)
		return data, nil
	}
	filters := extractFilterMap(filterOptions)

	// Iterate through the data and check if each item matches the filter criteria
	for _, item := range jsonData.Data {
		match := true
		it := item.(map[string]interface{})

		for filterKey, filterValue := range filters {
			if value, exists := it[filterKey]; !exists || !reflect.DeepEqual(value, filterValue) {
				match = false
			}
		}
		if match {
			results = append(results, it)
		}
	}
	if len(results) == 0 {
		return nil, errors.New("Error: No data found matching the filter criteria.")
	}
	byteData, _ := json.Marshal(ConvertMapsToInterfaces(results))
	return byteData, nil
}

/*
FindOne filters a data based on specified filter criteria |
if filterOption == "" means unfiltered -> first record is returned |
(it is recomended to use for data fetched through id field or any unique field)
*/
func (db DatabaseInfo) FindOne(Table string, filterOptions interface{}) ([]byte, error) {
	dbLock.Lock()
	defer dbLock.Unlock()

	var result interface{}

	// Extract filter criteria from the filterOptions and convert interface{} to map[string]interface{}

	fileName := strings.ToLower(Table)

	jsonData, err := loadJsonFile(fileName, db.Database)
	if err != nil {
		return nil, errors.New("Error: failed to load table data.")
	}
	if IsZeroValue(filterOptions) {
		data, _ := json.Marshal(jsonData.Data[0])
		return data, nil
	}
	filters := extractFilterMap(filterOptions)

	// Iterate through the data and check if each item matches the filter criteria
	for _, item := range jsonData.Data {
		match := true
		it := item.(map[string]interface{})

		for filterKey, filterValue := range filters {
			if value, exists := it[filterKey]; !exists || !reflect.DeepEqual(value, filterValue) {
				match = false
			}
		}
		if match {
			result = item
			break
		}
	}
	if result == nil {
		return nil, errors.New("Error: No data found matching the filter criteria.")
	}
	byteData, _ := json.Marshal(result)
	return byteData, nil
}
