package uadmin

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
)

type initialDataRecords []map[string]interface{}

type initialData struct {
	Init   []string                      `json:"init"`
	Data   map[string]initialDataRecords `json:"data"`
	Finish []string                      `json:"finish"`
}

// loadInitialData reads a file named initial_data.json and
// saves its content in the database
func loadInitialData() error {
	buf, err := ioutil.ReadFile("initial_data.json")
	if err != nil {
		return nil
	}

	// Load json daa into struct
	data := initialData{}
	err = json.Unmarshal(buf, &data)
	if err != nil {
		return fmt.Errorf("loadInitialData.Unmarshal: Error parsing json file. %s", err)
	}

	// Execute SQL in Init section
	for _, SQL := range data.Init {
		// Check if this is a uadmin command
		if strings.HasPrefix(SQL, "!") {
			command := strings.Split(SQL, " ")
			switch strings.ToUpper(command[0]) {
			case "!MIGRATE":
				if len(command) < 2 {
					return fmt.Errorf("invalid uadmin command in initial_data.json in init section. %s", SQL)
				}

				// Check if the model name is correct
				if _, ok := models[command[1]]; !ok {
					if len(command) < 2 {
						return fmt.Errorf("model name does not exist in initial_data.json in init section. %s", SQL)
					}
				}
				db.AutoMigrate(models[command[1]])
			}
			continue
		}

		// This is a SQL command
		err = db.Exec(SQL).Error
		if err != nil {
			return fmt.Errorf("loadInitialData.Exec: Error in in Init section (%s). %s", SQL, err)
		}
	}

	// Load data
	for table, records := range data.Data {
		// get modelname from table name
		// For the record:
		//   - Name       :  OrderItem
		//   - DisplayName:  Order Items
		//   - ModelName  :  orderitem
		//   - TableName  :  order_items
		tabeFound := false
		for k, v := range Schema {
			// check if table is a ModelName
			if table == k {
				tabeFound = true
				break
			}
			// check if the table is a database TableName
			// and convert it into modelname
			if v.TableName == table {
				table = k
				tabeFound = true
				break
			}
			// check if the table is a Name
			// and convert it into modelname
			if v.Name == table {
				table = k
				tabeFound = true
				break
			}
		}
		if !tabeFound {
			return fmt.Errorf("loadInitialData: Table not found for (%s)", table)
		}

		// Put records into Model Array
		modelArray, _ := NewModelArray(table, true)
		buf, _ = json.Marshal(records)
		err = json.Unmarshal(buf, modelArray.Interface())
		if err != nil {
			return fmt.Errorf("loadInitialData.Unmarshal: Error parsing Data records in (%s). %s", table, err)
		}

		// Save records
		for i := 0; i < modelArray.Elem().Len(); i++ {
			Get(modelArray.Elem().Index(i).Addr().Interface(), "id = ?", GetID(modelArray.Elem().Index(i)))
			json.Unmarshal(buf, modelArray.Interface())
			err = Save(modelArray.Elem().Index(i).Addr().Interface())
			if err != nil {
				return fmt.Errorf("loadInitialData.Save: Error in %s[%d]. %s", table, i, err)
			}
		}
	}

	// Execute SQL in Finish section
	for _, SQL := range data.Finish {
		// Check if this is a uadmin command
		if strings.HasPrefix(SQL, "!") {
			command := strings.Split(SQL, " ")
			switch strings.ToUpper(command[0]) {
			case "!MIGRATE":
				if len(command) < 2 {
					return fmt.Errorf("invalid uadmin command in initial_data.json in init section. %s", SQL)
				}

				// Check if the model name is correct
				if _, ok := models[command[1]]; !ok {
					if len(command) < 2 {
						return fmt.Errorf("model name does not exist in initial_data.json in init section. %s", SQL)
					}
				}
				db.AutoMigrate(models[command[1]])
			}
			continue
		}

		// This is a SQL command
		err = db.Exec(SQL).Error
		if err != nil {
			return fmt.Errorf("loadInitialData.Exec: Error in in Finish section (%s). %s", SQL, err)
		}
	}

	return nil
}
