### 

# Simple JSON Database Implementation

A lightweight JSON-based database implementation in Go that provides basic CRUD operations and filtering capabilities.

## Features

- Create databases and tables
- Add, update, and delete records
- Filter data using flexible criteria
- Simple and straightforward API
- JSON-based storage

## Installation

```go
// Import the package (replace with actual import path)
go get github.com/deb151292/GoSon-DB
```
### Struct for the table should be defined with appropriate fields

```go
/*value of tag should be in double quote like  > json:"abc" |
 in tag the key value should be separated with ":" without blank spaces |
 ifnull:"notnull"`->> this tag is used to mark a field as nullable |
 null for nullable fields , notnull for non-nullable fields */

type Admin struct {
	Name  string  `json:"name" ifnull:"notnull"` 
	Age   float64 `json:"age" ifnull:"null"`
	Email string  `json:"email" ifnull:"null"`
}

```

```go 
//This will show if GoSon db is ready with a greeting msg
fmt.Println(Connect_GoSon())
```
## Usage
### Create a new instance of the database
```go
//Call predefined DatabaseInfo struct from the Package
	db := DatabaseInfo{Database: "db"}
```

### Creating a Database

```go

//Create new database
msg, err := db.CreateDatabase("db")
if err != nil {
    log.Println("Error:", err)
}
log.Println(msg)
```

### Creating a Table

```go
//Create new table
msg, err := db.CreateTable("db", "Admin")
if err != nil {
    fmt.Println("Error:", err)
}
fmt.Println(msg)
```

Note: Table names are automatically converted to lowercase.

### Adding Data

```go
// Define your struct
type Admin struct {
    Name  string `json:"name" ifnull:"null"`
    Age   int `json:"age" ifnull:"null"`
    Email string `json:"email" ifnull:"null"`
}

// Create new record
newUser := Admin{
    Name:  "John Doe",
    Age:   30,
    Email: "john@example.com",
}

// Add to database
err := db.AddData("db", "Admin", newUser)
if err != nil {
    fmt.Println("Error:", err)
}
```

### Updating Data

```go
//filter for Update record should be the same struct that is used to create new record
//User can provide all/any one field from the struct while updating to update specific field

updatedUser := Admin{
    Name:  "xyz",
    Age:   31,
    Email: "xyzc@update.com",
}

// Update record with ID 0
err := db.UpdateData("db", "Admin", 0, updatedUser)
if err != nil {
    fmt.Println("Error:", err)
}
```

### Deleting Data

```go
// Delete record with ID 0
err := db.DeleteData("db", "admin", 0)
if err != nil {
    fmt.Println("Error:", err)
}
```

### Filtering Data

#### Find Many Records

```go
// Create a filter using an anonymous struct
filter := struct {
    Name  string  `json:"name"`
    Age   float64 `json:"age"`
    Email string  `json:"email"`
}{Age: 30}

var results []Admin
filteredData, err := db.FindMany("db", "Admin", filter)
if err != nil {
    log.Println(err)
} else {
    json.Unmarshal(filteredData, &results)
    log.Println(results)
}
```

#### Find One Record

```go
filters := struct {
    Name  string  `json:"name"`
    Age   float64 `json:"age"`
    Email string  `json:"email"`
}{Age: 30}

oneData, err := db.FindOne("db", "Admin", filters)
if err != nil {
    log.Println(err)
} else {
    log.Println(string(oneData))
}
```

## Important Notes

1. Table names are automatically converted to lowercase for consistency
2. For no filter in Find operations, pass an empty string ("") instead of nil
3. All operations perform validation before executing
4. The database uses JSON files for storage
5. IDs are automatically managed by the system

## Error Handling

All functions return appropriate error messages when:
- Database or table doesn't exist
- Invalid data is provided
- File operations fail
- Data validation fails

## Best Practices

1. Always check for errors after operations
2. Use proper struct definitions matching your table structure
3. Implement proper error handling in your application
4. Back up your database files regularly
5. Use appropriate data types in your structs

## Limitations

1. JSON-based storage might not be suitable for very large datasets
2. No built-in indexing
3. No concurrent write operations support
4. Limited query capabilities compared to full-featured databases

