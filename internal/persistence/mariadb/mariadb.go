package mariadb

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/Kaese72/device-store/ingestmodels"
	"github.com/Kaese72/device-store/internal/config"
	"github.com/Kaese72/device-store/internal/logging"
	"github.com/Kaese72/device-store/internal/persistence/intermediaries"
	"github.com/Kaese72/device-store/restmodels"
	"github.com/danielgtaylor/huma/v2"
	"go.elastic.co/apm/module/apmsql"
)

type mariadbPersistence struct {
	db *sql.DB
}

func NewMariadbPersistence(conf config.DatabaseConfig) (mariadbPersistence, error) {
	db, err := apmsql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", conf.User, conf.Password, conf.Host, conf.Port, conf.Database))
	if err != nil {
		logging.Fatal(err.Error(), context.Background())
		return mariadbPersistence{}, err
	}
	return mariadbPersistence{
		db: db,
	}, nil
}

// deviceFilters defines what filters are available for the devices model
var deviceFilters = map[string]map[string]func(string) (string, []string, error){
	"bridge-identifier": {
		"eq": func(value string) (string, []string, error) {
			return "bridgeIdentifier = ?", []string{value}, nil
		},
	},
	"id": {
		"eq": func(value string) (string, []string, error) {
			if regexp.MustCompile(`^\d+$`).MatchString(value) {
				return "id = ?", []string{value}, nil
			}
			return "", nil, huma.Error400BadRequest("id filter must be an integer value")
		},
	},
}

// validateTimestamp validates that the valis is on the format 'YYYY-MM-DD' or 'YYYY-MM-DD HH:MM:SS'
func validateTimestamp(value string) error {
	if regexp.MustCompile(`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$`).MatchString(value) {
		_, err := time.Parse("2006-01-02 15:04:05", value)
		if err == nil {
			return nil
		}
		return huma.Error400BadRequest(fmt.Sprintf("timestamp parsing error %s", err.Error()))
	}
	if regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`).MatchString(value) {
		_, err := time.Parse("2006-01-02", value)
		if err == nil {
			return nil
		}
		return huma.Error400BadRequest(fmt.Sprintf("timestamp parsing error %s", err.Error()))
	}
	return huma.Error400BadRequest("timestamp filter must be on the format 'YYYY-MM-DD' or 'YYYY-MM-DD HH:MM:SS'")
}

// deviceAttributeAuditFilters defines what filters are available for the deviceAttributeAudit model
var deviceAttributeAuditFilters = map[string]map[string]func(string) (string, []string, error){
	"deviceId": {
		"eq": func(value string) (string, []string, error) {
			if regexp.MustCompile(`^\d+$`).MatchString(value) {
				return "deviceId = ?", []string{value}, nil
			}
			return "", nil, huma.Error400BadRequest("deviceId filter must be an integer value")
		},
	},
	"name": {
		"eq": func(value string) (string, []string, error) {
			return "name = ?", []string{value}, nil
		},
	},
	"timestamp": {
		"eq": func(value string) (string, []string, error) {
			return "timestamp = ?", []string{value}, validateTimestamp(value)
		},
		"gt": func(value string) (string, []string, error) {
			return "timestamp > ?", []string{value}, validateTimestamp(value)
		},
		"lt": func(value string) (string, []string, error) {
			return "timestamp < ?", []string{value}, validateTimestamp(value)
		},
	},
	// We intentionally leave out old and new values for now
}

type GetDevicesCapabilityIntermediate struct {
	Name string `json:"name"`
}

func (i GetDevicesCapabilityIntermediate) toRest() restmodels.DeviceCapability {
	return restmodels.DeviceCapability{
		Name: i.Name,
	}
}

type GetDevicesAttributeIntermediate struct {
	Name         string   `json:"name"`
	BooleanValue *float32 `json:"boolean"`
	NumericValue *float32 `json:"numeric"`
	TextValue    *string  `json:"text"`
}

func (i GetDevicesAttributeIntermediate) toRest() restmodels.Attribute {
	return restmodels.Attribute{
		Name: i.Name,
		Boolean: func() *bool {
			if i.BooleanValue == nil {
				return nil
			}
			return &[]bool{*i.BooleanValue == 1}[0]
		}(),
		Numeric: i.NumericValue,
		Text:    i.TextValue,
	}
}

type TriggerIntermediate struct {
	Name string `json:"name"`
}

func (i TriggerIntermediate) toRest() restmodels.Trigger {
	return restmodels.Trigger{
		Name: i.Name,
	}
}

type queryAble interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func (persistence mariadbPersistence) GetDevices(ctx context.Context, filters []restmodels.Filter) ([]restmodels.Device, error) {
	fields := []string{
		"id",
		"bridgeIdentifier",
		"bridgeKey",
		"(SELECT COALESCE(JSON_ARRAYAGG(JSON_OBJECT(\"name\", name, \"boolean\", booleanValue, \"numeric\", numericValue, \"text\", textValue)), JSON_ARRAY()) FROM deviceAttributes WHERE deviceAttributes.deviceId = devices.id) as attributes",
		"(SELECT COALESCE(JSON_ARRAYAGG(JSON_OBJECT(\"name\", name)), JSON_ARRAY()) FROM deviceCapabilities WHERE deviceId = devices.id) as capabilities",
		"(SELECT COALESCE(JSON_ARRAYAGG(groupId), JSON_ARRAY()) FROM groupDevices WHERE deviceId = devices.id) as groupIds",
		"(SELECT COALESCE(JSON_ARRAYAGG(JSON_OBJECT(\"name\", name)), JSON_ARRAY()) FROM deviceTriggers WHERE deviceTriggers.deviceId = devices.id) as triggers",
	}
	query := `SELECT ` + strings.Join(fields, ",") + ` FROM devices`
	queryFragments, variables, err := intermediaries.TranslateFiltersToQueryFragments(filters, deviceFilters)
	if err != nil {
		return nil, err
	}
	if len(queryFragments) > 0 {
		query += " WHERE "
		query += strings.Join(queryFragments, " AND ")
	}
	var retDevices []restmodels.Device
	rows, err := persistence.db.Query(query, variables...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var device restmodels.Device
		var capabilitiesBytes []byte
		var attributesBytes []byte
		var triggerBytes []byte
		var groupIdsBytes []byte
		err = rows.Scan(&device.ID, &device.BridgeIdentifier, &device.BridgeKey, &attributesBytes, &capabilitiesBytes, &groupIdsBytes, &triggerBytes)
		if err != nil {
			return nil, err
		}
		// Attributes
		var attributeIntermediates []GetDevicesAttributeIntermediate
		err = json.Unmarshal(attributesBytes, &attributeIntermediates)
		if err != nil {
			return nil, err
		}
		for _, attribute := range attributeIntermediates {
			device.Attributes = append(device.Attributes, attribute.toRest())
		}
		// Capabilities
		var capabilityIntermediates []GetDevicesCapabilityIntermediate
		err = json.Unmarshal(capabilitiesBytes, &capabilityIntermediates)
		if err != nil {
			return nil, err
		}
		device.Capabilities = []restmodels.DeviceCapability{}
		for _, capability := range capabilityIntermediates {
			device.Capabilities = append(device.Capabilities, capability.toRest())
		}
		// Group IDs
		err = json.Unmarshal(groupIdsBytes, &device.GroupIds)
		if err != nil {
			return nil, err
		}
		// Triggers
		var triggerIntermediates []TriggerIntermediate
		err = json.Unmarshal(triggerBytes, &triggerIntermediates)
		if err != nil {
			return nil, err
		}
		device.Triggers = []restmodels.Trigger{}
		for _, trigger := range triggerIntermediates {
			device.Triggers = append(device.Triggers, trigger.toRest())
		}
		// Append device to result list
		retDevices = append(retDevices, device)
	}
	return retDevices, rows.Err()
}

// GetAttributeAudits
func (persistence mariadbPersistence) GetAttributeAudits(ctx context.Context, filters []restmodels.Filter) ([]restmodels.AttributeAudit, error) {
	fields := []string{
		"id",
		"deviceId",
		"name",
		"timestamp",
		"oldBooleanValue",
		"oldNumericValue",
		"oldTextValue",
		"newBooleanValue",
		"newNumericValue",
		"newTextValue",
	}
	query := `SELECT ` + strings.Join(fields, ",") + ` FROM deviceAttributeAudit`
	queryFragments, variables, err := intermediaries.TranslateFiltersToQueryFragments(filters, deviceAttributeAuditFilters)
	if err != nil {
		return nil, err
	}
	if len(queryFragments) > 0 {
		query += " WHERE "
		query += strings.Join(queryFragments, " AND ")
	}
	retAudits := []restmodels.AttributeAudit{}
	rows, err := persistence.db.Query(query, variables...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var audit restmodels.AttributeAudit
		var timestampStr string
		err = rows.Scan(&audit.ID, &audit.DeviceID, &audit.Name, &timestampStr, &audit.OldBooleanValue, &audit.OldNumericValue, &audit.OldTextValue, &audit.NewBooleanValue, &audit.NewNumericValue, &audit.NewTextValue)
		if err != nil {
			return nil, err
		}
		// Parse timestamp string to time.Time
		audit.Timestamp, err = time.Parse("2006-01-02 15:04:05", timestampStr)
		if err != nil {
			return nil, err
		}
		retAudits = append(retAudits, audit)
	}
	return retAudits, rows.Err()
}

func toDbBoolean(value *bool) *float32 {
	if value == nil {
		return nil
	}
	if *value {
		return &[]float32{1}[0]
	}
	return &[]float32{0}[0]
}

type dbAttribute struct {
	Name         string
	BooleanValue *float32
	NumericValue *float32
	TextValue    *string
}

// Equal checks whether two dbAttributes are equal.
func (a dbAttribute) EqualRest(other ingestmodels.IngestAttribute) bool {
	if a.Name != other.Name {
		return false
	}
	dbBoolean := toDbBoolean(other.Boolean)
	// Not equal if one is nil and the other is not or if they are not equal
	// Boolean
	if (dbBoolean == nil) != (a.BooleanValue == nil) {
		return false
	}
	if dbBoolean != nil && a.BooleanValue != nil && *dbBoolean != *a.BooleanValue {
		return false
	}
	// Numeric
	if (other.Numeric == nil) != (a.NumericValue == nil) {
		return false
	}
	// Numeric comparison is even more tricky because float point numbers have rounding errors
	// and such, so we compare if the numbers are more than two decimal points apart from each other
	// In the database the type is DECIMAL(10,4), so we have some leeway
	if other.Numeric != nil && a.NumericValue != nil && math.Abs(float64(*other.Numeric)-float64(*a.NumericValue)) > 0.001 {
		return false
	}
	// Text
	if (other.Text == nil) != (a.TextValue == nil) {
		return false
	}
	if other.Text != nil && a.TextValue != nil && *other.Text != *a.TextValue {
		return false
	}
	// All checks passed, they are equal
	return true
}

func (persistence mariadbPersistence) PostDevice(ctx context.Context, device ingestmodels.IngestDevice) (int, []ingestmodels.IngestAttribute, error) {
	var foundId int
	tx, err := persistence.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, nil, err
	}

	defer tx.Rollback()
	row := tx.QueryRowContext(ctx, `SELECT id FROM devices WHERE bridgeIdentifier = ? AND bridgeKey = ?`, device.BridgeIdentifier, device.BridgeKey)
	err = row.Scan(&foundId)
	if err != nil && err != sql.ErrNoRows {
		return 0, nil, err
	}

	var deviceId int
	var presentAttributes map[string]dbAttribute = make(map[string]dbAttribute)
	if foundId == 0 {
		rows := tx.QueryRowContext(ctx, `INSERT INTO devices (bridgeIdentifier, bridgeKey) VALUES (?, ?) RETURNING id`, device.BridgeIdentifier, device.BridgeKey)
		err := rows.Scan(&deviceId)
		if err != nil {
			return 0, nil, err
		}
	} else {
		deviceId = foundId
		// Find already present attributes
		rows, err := tx.QueryContext(ctx, `SELECT name, booleanValue, numericValue, textValue FROM deviceAttributes WHERE deviceId = ?`, deviceId)
		if err != nil {
			return 0, nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var presentAttribute dbAttribute
			err = rows.Scan(&presentAttribute.Name, &presentAttribute.BooleanValue, &presentAttribute.NumericValue, &presentAttribute.TextValue)
			if err != nil {
				return 0, nil, err
			}
			presentAttributes[presentAttribute.Name] = presentAttribute
		}
	}
	var updatedAttributes []ingestmodels.IngestAttribute
	for _, attribute := range device.Attributes {
		// If the attributes is already present and is different, update it record for event updates later
		if presentAttribute, ok := presentAttributes[attribute.Name]; ok {
			if !presentAttribute.EqualRest(attribute) {
				_, err = tx.ExecContext(ctx, `UPDATE deviceAttributes SET booleanValue=?, numericValue=?, textValue=? WHERE deviceId=? AND name=?`, toDbBoolean(attribute.Boolean), attribute.Numeric, attribute.Text, deviceId, attribute.Name)
				if err != nil {
					return 0, nil, err
				}
				updatedAttributes = append(updatedAttributes, attribute)
			}
			// If equal, do nothing ...
		} else {
			// If the attribute is not present, insert it
			_, err = tx.ExecContext(ctx, `INSERT INTO deviceAttributes (deviceId, name, booleanValue, numericValue, textValue) VALUES (?, ?, ?, ?, ?)`, deviceId, attribute.Name, toDbBoolean(attribute.Boolean), attribute.Numeric, attribute.Text)
			if err != nil {
				return 0, nil, err
			}
			updatedAttributes = append(updatedAttributes, attribute)
		}
	}
	for _, capability := range device.Capabilities {
		// JSON encode ArgumentsJsonSchema so it can be saved in the database
		argumentsJsonSchema, err := json.Marshal(capability.ArgumentSpecs)
		if err != nil {
			return 0, nil, err
		}
		_, err = tx.ExecContext(ctx, `INSERT IGNORE INTO deviceCapabilities (deviceId, name, argumentJsonSchema) VALUES (?, ?, ?)`, deviceId, capability.Name, argumentsJsonSchema)
		if err != nil {
			return 0, nil, err
		}
	}
	for _, trigger := range device.Triggers {
		_, err = tx.ExecContext(ctx, `INSERT IGNORE INTO deviceTriggers (deviceId, name) VALUES (?, ?)`, deviceId, trigger.Name)
		if err != nil {
			return 0, nil, err
		}
	}
	return deviceId, updatedAttributes, tx.Commit()
}

func (persistence mariadbPersistence) GetDeviceCapabilityForActivation(ctx context.Context, storeIdentifier int, capabilityName string) (intermediaries.DeviceCapabilityIntermediaryActivation, error) {
	capability := intermediaries.DeviceCapabilityIntermediaryActivation{}
	row := persistence.db.QueryRowContext(ctx, `SELECT bridgeIdentifier, name, bridgeKey FROM deviceCapabilities INNER JOIN devices on deviceCapabilities.deviceId = devices.id WHERE deviceId = ? AND name = ?`, storeIdentifier, capabilityName)
	err := row.Scan(&capability.BridgeIdentifier, &capability.Name, &capability.BridgeKey)
	if err != nil {
		if err != sql.ErrNoRows {
			return intermediaries.DeviceCapabilityIntermediaryActivation{}, err
		}
		return intermediaries.DeviceCapabilityIntermediaryActivation{}, huma.Error404NotFound(fmt.Sprintf("capability %s not found for device %d", capabilityName, storeIdentifier))
	}
	return capability, nil
}

var groupFilters = map[string]map[string]func(string) (string, []string, error){
	"bridge-identifier": {
		"eq": func(value string) (string, []string, error) {
			return "bridgeIdentifier = ?", []string{value}, nil
		},
	},
	"id": {
		"eq": func(value string) (string, []string, error) {
			return "id = ?", []string{value}, nil
		},
	},
	"bridge-key": {
		"eq": func(value string) (string, []string, error) {
			return "bridgeKey = ?", []string{value}, nil
		},
	},
}

func (persistence mariadbPersistence) GetGroups(ctx context.Context, filters []restmodels.Filter) ([]restmodels.Group, error) {
	return getGroupsTx(ctx, filters, persistence.db)
}

func getGroupsTx(ctx context.Context, filters []restmodels.Filter, tx queryAble) ([]restmodels.Group, error) {
	fields := []string{
		"id",
		"bridgeIdentifier",
		"bridgeKey",
		"name",
		"(SELECT COALESCE(JSON_ARRAYAGG(JSON_OBJECT(\"name\", name)), JSON_ARRAY()) FROM groupCapabilities WHERE groupId = groups.id) as capabilities",
		"(SELECT COALESCE(JSON_ARRAYAGG(deviceId), JSON_ARRAY()) FROM groupDevices WHERE groupId = groups.id) as deviceIds",
	}
	query := `SELECT ` + strings.Join(fields, ",") + ` FROM groups`
	queryFragments, variables, err := intermediaries.TranslateFiltersToQueryFragments(filters, groupFilters)
	if err != nil {
		return nil, err
	}
	if len(queryFragments) > 0 {
		query += " WHERE "
		query += strings.Join(queryFragments, " AND ")
	}
	var groups []restmodels.Group
	rows, err := tx.QueryContext(ctx, query, variables...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var group restmodels.Group
		var capabilitiesBytes []byte
		var deviceIdsBytes []byte
		err = rows.Scan(&group.ID, &group.BridgeIdentifier, &group.BridgeKey, &group.Name, &capabilitiesBytes, &deviceIdsBytes)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(capabilitiesBytes, &group.Capabilities)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(deviceIdsBytes, &group.DeviceIds)
		if err != nil {
			return nil, err
		}
		groups = append(groups, group)
	}
	return groups, rows.Err()
}

func (persistence mariadbPersistence) PostGroup(ctx context.Context, group ingestmodels.IngestGroup) error {
	tx, err := persistence.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	err = postGroupTx(ctx, group, tx)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func postGroupTx(ctx context.Context, group ingestmodels.IngestGroup, tx queryAble) error {
	foundGroups, err := getGroupsTx(ctx, []restmodels.Filter{{
		Key:      "bridge-identifier",
		Operator: "eq",
		Value:    group.BridgeIdentifier,
	},
		{
			Key:      "bridge-key",
			Operator: "eq",
			Value:    group.BridgeKey,
		},
	}, tx)
	if err != nil {
		return err
	}
	var groupId int
	deviceIdsInGroup := []int{}
	if len(foundGroups) == 0 {
		result := tx.QueryRowContext(ctx, `INSERT INTO groups (bridgeIdentifier, bridgeKey, name) VALUES (?, ?, ?) RETURNING id`, group.BridgeIdentifier, group.BridgeKey, group.Name)
		if result == nil {
			return fmt.Errorf("failed to insert group: QueryRowContext returned nil")
		}
		err := result.Scan(&groupId)
		if err != nil {
			return err
		}
	} else {
		groupId = foundGroups[0].ID
		deviceIdsInGroup = foundGroups[0].DeviceIds
		_, err := tx.ExecContext(ctx, `UPDATE groups SET name = ? WHERE id = ?`, group.Name, groupId)
		if err != nil {
			return err
		}
	}
	// Update capabilities
	for _, capability := range group.Capabilities {
		_, err = tx.ExecContext(ctx, `INSERT IGNORE INTO groupCapabilities (groupId, name) VALUES (?, ?)`, groupId, capability.Name)
		if err != nil {
			return err
		}
	}
	// Update deviceIds
	// // Add missing deviceIds
	for _, deviceId := range group.DeviceIds {
		if slices.Contains(deviceIdsInGroup, deviceId) {
			continue
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO groupDevices (groupId, deviceId) VALUES (?, ?)`, groupId, deviceId)
		if err != nil {
			return err
		}
	}
	// // Remove deviceIds that are not in the new list
	for _, deviceId := range deviceIdsInGroup {
		if !slices.Contains(group.DeviceIds, deviceId) {
			_, err = tx.ExecContext(ctx, `DELETE FROM groupDevices WHERE groupId = ? AND deviceId = ?`, groupId, deviceId)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (persistence mariadbPersistence) GetGroupCapabilityForActivation(ctx context.Context, storeIdentifier int, capabilityName string) (intermediaries.GroupCapabilityIntermediaryActivation, error) {
	capability := intermediaries.GroupCapabilityIntermediaryActivation{}
	row := persistence.db.QueryRowContext(ctx, `SELECT bridgeIdentifier, groupCapabilities.name, bridgeKey FROM groupCapabilities INNER JOIN groups on groupCapabilities.groupId = groups.id WHERE groupId = ? AND groupCapabilities.name = ?`, storeIdentifier, capabilityName)
	err := row.Scan(&capability.BridgeIdentifier, &capability.Name, &capability.BridgeKey)
	if err != nil {
		if err != sql.ErrNoRows {
			return intermediaries.GroupCapabilityIntermediaryActivation{}, err
		}
		return intermediaries.GroupCapabilityIntermediaryActivation{}, huma.Error404NotFound(fmt.Sprintf("capability %s not found for group %d", capabilityName, storeIdentifier))
	}
	return capability, err
}
