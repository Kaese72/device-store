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
	db, err := apmsql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=UTC", conf.User, conf.Password, conf.Host, conf.Port, conf.Database))
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
		"in": func(value string) (string, []string, error) {
			return inClause("id", value)
		},
	},
}

// inClause builds a "column IN (?,?,...)" fragment from a comma-separated
// list of integer values, e.g. "1,2,3". Whitespace around each value is
// ignored; an empty list or any non-integer value is a 400 error.
func inClause(column string, value string) (string, []string, error) {
	rawValues := strings.Split(value, ",")
	values := make([]string, 0, len(rawValues))
	for _, rawValue := range rawValues {
		trimmed := strings.TrimSpace(rawValue)
		if !regexp.MustCompile(`^\d+$`).MatchString(trimmed) {
			return "", nil, huma.Error400BadRequest(fmt.Sprintf("%s filter must be a comma-separated list of integer values", column))
		}
		values = append(values, trimmed)
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(values)), ",")
	return fmt.Sprintf("%s IN (%s)", column, placeholders), values, nil
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
	Name          string    `json:"name"`
	Updated       time.Time `json:"updated"`
	ArgumentSpecs []struct {
		Name    string `json:"name"`
		Boolean *struct {
			Default *bool `json:"default,omitempty"`
		} `json:"boolean,omitempty"`
		Numeric *struct {
			Min     float32  `json:"min"`
			Max     float32  `json:"max"`
			Default *float32 `json:"default,omitempty"`
		} `json:"numeric,omitempty"`
		Text *struct {
			Default *string `json:"default,omitempty"`
			Min     *int    `json:"min,omitempty"`
			Max     *int    `json:"max,omitempty"`
		} `json:"text,omitempty"`
	} `json:"argument-specs"`
}

func (i GetDevicesCapabilityIntermediate) toRest() restmodels.DeviceCapability {
	return restmodels.DeviceCapability{
		Name:    i.Name,
		Updated: i.Updated,
		ArgumentSpecs: func() []restmodels.ArgumentSpec {
			var specs []restmodels.ArgumentSpec
			for _, argSpec := range i.ArgumentSpecs {
				spec := restmodels.ArgumentSpec{
					Name: argSpec.Name,
				}
				if argSpec.Boolean != nil {
					spec.Boolean = &restmodels.BooleanArgumentSpec{
						Default: argSpec.Boolean.Default,
					}
				}
				if argSpec.Numeric != nil {
					spec.Numeric = &restmodels.NumericArgumentSpec{
						Min:     argSpec.Numeric.Min,
						Max:     argSpec.Numeric.Max,
						Default: argSpec.Numeric.Default,
					}
				}
				if argSpec.Text != nil {
					spec.Text = &restmodels.TextArgumentSpec{
						Default: argSpec.Text.Default,
						Min:     argSpec.Text.Min,
						Max:     argSpec.Text.Max,
					}
				}
				specs = append(specs, spec)
			}
			return specs
		}(),
	}
}

type GetDevicesAttributeIntermediate struct {
	Name         string    `json:"name"`
	BooleanValue *float32  `json:"boolean"`
	NumericValue *float32  `json:"numeric"`
	TextValue    *string   `json:"text"`
	Updated      time.Time `json:"updated"`
}

func (i GetDevicesAttributeIntermediate) toRest() restmodels.Attribute {
	return restmodels.Attribute{
		Name:    i.Name,
		Updated: i.Updated,
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

type queryAble interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// paginationClause returns the SQL "LIMIT ? OFFSET ?" fragment and its arguments
// for the given pagination. A zero Limit means unbounded, in which case no
// clause is applied.
func paginationClause(pagination restmodels.Pagination) (string, []any) {
	if pagination.Limit <= 0 {
		return "", nil
	}
	offset := pagination.Offset
	if offset < 0 {
		offset = 0
	}
	return " LIMIT ? OFFSET ?", []any{pagination.Limit, offset}
}

// countRows executes "SELECT COUNT(*) FROM <table> [WHERE <whereClause>]" and
// returns the total number of matching rows, ignoring pagination.
func countRows(ctx context.Context, tx queryAble, table string, whereClause string, variables []any) (int, error) {
	query := `SELECT COUNT(*) FROM ` + table
	if whereClause != "" {
		query += " WHERE " + whereClause
	}
	var total int
	row := tx.QueryRowContext(ctx, query, variables...)
	if err := row.Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (persistence mariadbPersistence) GetDevices(ctx context.Context, filters []restmodels.Filter, pagination restmodels.Pagination) ([]restmodels.Device, int, error) {
	fields := []string{
		"id",
		"bridgeIdentifier",
		"adapterId",
		"updated",
		"(SELECT COALESCE(JSON_ARRAYAGG(JSON_OBJECT(\"name\", name, \"boolean\", booleanValue, \"numeric\", numericValue, \"text\", textValue, \"updated\", DATE_FORMAT(updated, '%Y-%m-%dT%H:%i:%sZ'))), JSON_ARRAY()) FROM deviceAttributes WHERE deviceAttributes.deviceId = devices.id) as attributes",
		"(SELECT COALESCE(JSON_ARRAYAGG(JSON_OBJECT(\"name\", name, \"argument-specs\", argumentJsonSchema, \"updated\", DATE_FORMAT(updated, '%Y-%m-%dT%H:%i:%sZ'))), JSON_ARRAY()) FROM deviceCapabilities WHERE deviceId = devices.id) as capabilities",
		"(SELECT COALESCE(JSON_ARRAYAGG(groupId), JSON_ARRAY()) FROM groupDevices WHERE deviceId = devices.id) as groupIds",
		"(SELECT COALESCE(JSON_ARRAYAGG(JSON_OBJECT(\"name\", name)), JSON_ARRAY()) FROM deviceTriggers WHERE deviceTriggers.deviceId = devices.id) as triggers",
	}
	query := `SELECT ` + strings.Join(fields, ",") + ` FROM devices`
	attributeFilters, otherFilters := intermediaries.SplitAttributeFilters(filters)
	queryFragments, variables, err := intermediaries.TranslateFiltersToQueryFragments(otherFilters, deviceFilters)
	if err != nil {
		return nil, 0, err
	}
	attributeFragments, attributeValues, err := intermediaries.TranslateAttributeFiltersToQueryFragments(attributeFilters, "devices.id")
	if err != nil {
		return nil, 0, err
	}
	queryFragments = append(queryFragments, attributeFragments...)
	variables = append(variables, attributeValues...)
	whereClause := strings.Join(queryFragments, " AND ")
	if whereClause != "" {
		query += " WHERE " + whereClause
	}
	total, err := countRows(ctx, persistence.db, "devices", whereClause, variables)
	if err != nil {
		return nil, 0, err
	}
	query += " ORDER BY id ASC"
	limitClause, limitArgs := paginationClause(pagination)
	query += limitClause
	var retDevices []restmodels.Device
	rows, err := persistence.db.Query(query, append(append([]any{}, variables...), limitArgs...)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	for rows.Next() {
		var device restmodels.Device
		var capabilitiesBytes []byte
		var attributesBytes []byte
		var triggerBytes []byte
		var groupIdsBytes []byte
		err = rows.Scan(&device.ID, &device.BridgeIdentifier, &device.AdapterId, &device.Updated, &attributesBytes, &capabilitiesBytes, &groupIdsBytes, &triggerBytes)
		if err != nil {
			return nil, 0, err
		}
		// Attributes
		var attributeIntermediates []GetDevicesAttributeIntermediate
		err = json.Unmarshal(attributesBytes, &attributeIntermediates)
		if err != nil {
			return nil, 0, err
		}
		for _, attribute := range attributeIntermediates {
			device.Attributes = append(device.Attributes, attribute.toRest())
		}
		// Capabilities
		var capabilityIntermediates []GetDevicesCapabilityIntermediate
		err = json.Unmarshal(capabilitiesBytes, &capabilityIntermediates)
		if err != nil {
			return nil, 0, err
		}
		device.Capabilities = []restmodels.DeviceCapability{}
		for _, capability := range capabilityIntermediates {
			device.Capabilities = append(device.Capabilities, capability.toRest())
		}
		// Group IDs
		err = json.Unmarshal(groupIdsBytes, &device.GroupIds)
		if err != nil {
			return nil, 0, err
		}
		// Append device to result list
		retDevices = append(retDevices, device)
	}
	return retDevices, total, rows.Err()
}

func (persistence mariadbPersistence) DeleteGroup(ctx context.Context, storeIdentifier int) error {
	result, err := persistence.db.ExecContext(ctx, `DELETE FROM groups WHERE id = ?`, storeIdentifier)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return huma.Error404NotFound(fmt.Sprintf("group %d not found", storeIdentifier))
	}
	return nil
}

func (persistence mariadbPersistence) DeleteDevice(ctx context.Context, storeIdentifier int) error {
	result, err := persistence.db.ExecContext(ctx, `DELETE FROM devices WHERE id = ?`, storeIdentifier)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return huma.Error404NotFound(fmt.Sprintf("device %d not found", storeIdentifier))
	}
	return nil
}

// GetAttributeAudits
func (persistence mariadbPersistence) GetAttributeAudits(ctx context.Context, filters []restmodels.Filter, pagination restmodels.Pagination) ([]restmodels.AttributeAudit, int, error) {
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
		return nil, 0, err
	}
	whereClause := strings.Join(queryFragments, " AND ")
	if whereClause != "" {
		query += " WHERE " + whereClause
	}
	total, err := countRows(ctx, persistence.db, "deviceAttributeAudit", whereClause, variables)
	if err != nil {
		return nil, 0, err
	}
	query += " ORDER BY id ASC"
	limitClause, limitArgs := paginationClause(pagination)
	query += limitClause
	retAudits := []restmodels.AttributeAudit{}
	rows, err := persistence.db.Query(query, append(append([]any{}, variables...), limitArgs...)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	for rows.Next() {
		var audit restmodels.AttributeAudit
		err = rows.Scan(&audit.ID, &audit.DeviceID, &audit.Name, &audit.Timestamp, &audit.OldBooleanValue, &audit.OldNumericValue, &audit.OldTextValue, &audit.NewBooleanValue, &audit.NewNumericValue, &audit.NewTextValue)
		if err != nil {
			return nil, 0, err
		}
		retAudits = append(retAudits, audit)
	}
	return retAudits, total, rows.Err()
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

func getAttributeUpdated(ctx context.Context, tx queryAble, deviceId int, attributeName string) (time.Time, error) {
	var updated time.Time
	row := tx.QueryRowContext(ctx, `SELECT updated FROM deviceAttributes WHERE deviceId = ? AND name = ?`, deviceId, attributeName)
	if err := row.Scan(&updated); err != nil {
		return time.Time{}, err
	}
	return updated, nil
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
	row := tx.QueryRowContext(ctx, `SELECT id FROM devices WHERE bridgeIdentifier = ? AND adapterId = ?`, device.BridgeIdentifier, device.AdapterId)
	err = row.Scan(&foundId)
	if err != nil && err != sql.ErrNoRows {
		return 0, nil, err
	}

	var deviceId int
	var presentAttributes map[string]dbAttribute = make(map[string]dbAttribute)
	if foundId == 0 {
		rows := tx.QueryRowContext(ctx, `INSERT INTO devices (bridgeIdentifier, adapterId, updated) VALUES (?, ?, NOW()) RETURNING id`, device.BridgeIdentifier, device.AdapterId)
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
				_, err = tx.ExecContext(ctx, `UPDATE deviceAttributes SET booleanValue=?, numericValue=?, textValue=?, updated=NOW() WHERE deviceId=? AND name=?`, toDbBoolean(attribute.Boolean), attribute.Numeric, attribute.Text, deviceId, attribute.Name)
				if err != nil {
					return 0, nil, err
				}
				updated, err := getAttributeUpdated(ctx, tx, deviceId, attribute.Name)
				if err != nil {
					return 0, nil, err
				}
				updatedAttribute := attribute
				updatedAttribute.Updated = updated
				updatedAttributes = append(updatedAttributes, updatedAttribute)
			}
			// If equal, do nothing ...
		} else {
			// If the attribute is not present, insert it
			_, err = tx.ExecContext(ctx, `INSERT INTO deviceAttributes (deviceId, name, booleanValue, numericValue, textValue, updated) VALUES (?, ?, ?, ?, ?, NOW())`, deviceId, attribute.Name, toDbBoolean(attribute.Boolean), attribute.Numeric, attribute.Text)
			if err != nil {
				return 0, nil, err
			}
			updated, err := getAttributeUpdated(ctx, tx, deviceId, attribute.Name)
			if err != nil {
				return 0, nil, err
			}
			updatedAttribute := attribute
			updatedAttribute.Updated = updated
			updatedAttributes = append(updatedAttributes, updatedAttribute)
		}
	}
	deviceWasUpdated := len(updatedAttributes) > 0
	for _, capability := range device.Capabilities {
		// JSON encode ArgumentsJsonSchema so it can be saved in the database
		argumentsJsonSchema, err := json.Marshal(capability.ArgumentSpecs)
		if err != nil {
			return 0, nil, err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO deviceCapabilities (deviceId, name, argumentJsonSchema, updated) VALUES (?, ?, ?, NOW()) ON DUPLICATE KEY UPDATE argumentJsonSchema = VALUES(argumentJsonSchema), updated = NOW()`, deviceId, capability.Name, argumentsJsonSchema)
		if err != nil {
			return 0, nil, err
		}
		deviceWasUpdated = true
	}
	if deviceWasUpdated {
		_, err = tx.ExecContext(ctx, `UPDATE devices SET updated = GREATEST(updated, NOW()) WHERE id = ?`, deviceId)
		if err != nil {
			return 0, nil, err
		}
	}
	return deviceId, updatedAttributes, tx.Commit()
}

func (persistence mariadbPersistence) GetDeviceCapabilityForActivation(ctx context.Context, storeIdentifier int, capabilityName string) (intermediaries.DeviceCapabilityIntermediaryActivation, error) {
	capability := intermediaries.DeviceCapabilityIntermediaryActivation{}
	row := persistence.db.QueryRowContext(ctx, `SELECT bridgeIdentifier, name, adapterId FROM deviceCapabilities INNER JOIN devices on deviceCapabilities.deviceId = devices.id WHERE deviceId = ? AND name = ?`, storeIdentifier, capabilityName)
	err := row.Scan(&capability.BridgeIdentifier, &capability.Name, &capability.AdapterId)
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
	"adapter-id": {
		"eq": func(value string) (string, []string, error) {
			return "adapterId = ?", []string{value}, nil
		},
	},
}

func (persistence mariadbPersistence) GetGroups(ctx context.Context, filters []restmodels.Filter, pagination restmodels.Pagination) ([]restmodels.Group, int, error) {
	return getGroupsTx(ctx, filters, pagination, persistence.db)
}

func getGroupsTx(ctx context.Context, filters []restmodels.Filter, pagination restmodels.Pagination, tx queryAble) ([]restmodels.Group, int, error) {
	fields := []string{
		"id",
		"bridgeIdentifier",
		"adapterId",
		"name",
		"updated",
		"(SELECT COALESCE(JSON_ARRAYAGG(JSON_OBJECT(\"name\", name, \"argument-specs\", argumentJsonSchema, \"updated\", DATE_FORMAT(updated, '%Y-%m-%dT%H:%i:%sZ'))), JSON_ARRAY()) FROM groupCapabilities WHERE groupId = groups.id) as capabilities",
		"(SELECT COALESCE(JSON_ARRAYAGG(deviceId), JSON_ARRAY()) FROM groupDevices WHERE groupId = groups.id) as deviceIds",
	}
	query := `SELECT ` + strings.Join(fields, ",") + ` FROM groups`
	queryFragments, variables, err := intermediaries.TranslateFiltersToQueryFragments(filters, groupFilters)
	if err != nil {
		return nil, 0, err
	}
	whereClause := strings.Join(queryFragments, " AND ")
	if whereClause != "" {
		query += " WHERE " + whereClause
	}
	total, err := countRows(ctx, tx, "groups", whereClause, variables)
	if err != nil {
		return nil, 0, err
	}
	query += " ORDER BY id ASC"
	limitClause, limitArgs := paginationClause(pagination)
	query += limitClause
	var groups []restmodels.Group
	rows, err := tx.QueryContext(ctx, query, append(append([]any{}, variables...), limitArgs...)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	for rows.Next() {
		var group restmodels.Group
		var capabilitiesBytes []byte
		var deviceIdsBytes []byte
		err = rows.Scan(&group.ID, &group.BridgeIdentifier, &group.AdapterId, &group.Name, &group.Updated, &capabilitiesBytes, &deviceIdsBytes)
		if err != nil {
			return nil, 0, err
		}
		err = json.Unmarshal(capabilitiesBytes, &group.Capabilities)
		if err != nil {
			return nil, 0, err
		}
		err = json.Unmarshal(deviceIdsBytes, &group.DeviceIds)
		if err != nil {
			return nil, 0, err
		}
		groups = append(groups, group)
	}
	return groups, total, rows.Err()
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
	foundGroups, _, err := getGroupsTx(ctx, []restmodels.Filter{
		{
			Key:      "bridge-identifier",
			Operator: "eq",
			Value:    group.BridgeIdentifier,
		},
		{
			Key:      "adapter-id",
			Operator: "eq",
			Value:    fmt.Sprintf("%d", group.AdapterId),
		},
	}, restmodels.Pagination{}, tx)
	if err != nil {
		return err
	}
	var groupId int
	deviceIdsInGroup := []int{}
	if len(foundGroups) == 0 {
		result := tx.QueryRowContext(ctx, `INSERT INTO groups (bridgeIdentifier, adapterId, name) VALUES (?, ?, ?) RETURNING id`, group.BridgeIdentifier, group.AdapterId, group.Name)
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
		argumentsJsonSchema, err := json.Marshal(capability.ArgumentSpecs)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO groupCapabilities (groupId, name, argumentJsonSchema, updated) VALUES (?, ?, ?, NOW()) ON DUPLICATE KEY UPDATE argumentJsonSchema = VALUES(argumentJsonSchema), updated = NOW()`, groupId, capability.Name, argumentsJsonSchema)
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

func (persistence mariadbPersistence) WriteCapabilityTriggerAudit(ctx context.Context, deviceId int, capabilityName string, success bool, errorMessage *string, arguments string) error {
	_, err := persistence.db.ExecContext(ctx,
		`INSERT INTO deviceCapabilityTriggerAudit (deviceId, name, success, errorMessage, arguments) VALUES (?, ?, ?, ?, ?)`,
		deviceId, capabilityName, success, errorMessage, arguments,
	)
	return err
}

func (persistence mariadbPersistence) GetCapabilityTriggerAudits(ctx context.Context, deviceId int, pagination restmodels.Pagination) ([]restmodels.CapabilityTriggerAudit, int, error) {
	if pagination.Limit <= 0 {
		pagination.Limit = 50
	}
	total, err := countRows(ctx, persistence.db, "deviceCapabilityTriggerAudit", "deviceId = ?", []any{deviceId})
	if err != nil {
		return nil, 0, err
	}
	limitClause, limitArgs := paginationClause(pagination)
	rows, err := persistence.db.QueryContext(ctx,
		`SELECT id, deviceId, name, success, errorMessage, timestamp, arguments FROM deviceCapabilityTriggerAudit WHERE deviceId = ? ORDER BY timestamp DESC`+limitClause,
		append([]any{deviceId}, limitArgs...)...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var audits []restmodels.CapabilityTriggerAudit
	for rows.Next() {
		var audit restmodels.CapabilityTriggerAudit
		if err := rows.Scan(&audit.ID, &audit.DeviceID, &audit.Name, &audit.Success, &audit.ErrorMessage, &audit.Timestamp, &audit.Arguments); err != nil {
			return nil, 0, err
		}
		audits = append(audits, audit)
	}
	return audits, total, rows.Err()
}

func (persistence mariadbPersistence) WriteGroupCapabilityTriggerAudit(ctx context.Context, groupId int, capabilityName string, success bool, errorMessage *string, arguments string) error {
	_, err := persistence.db.ExecContext(ctx,
		`INSERT INTO groupCapabilityTriggerAudit (groupId, name, success, errorMessage, arguments) VALUES (?, ?, ?, ?, ?)`,
		groupId, capabilityName, success, errorMessage, arguments,
	)
	return err
}

func (persistence mariadbPersistence) GetGroupCapabilityTriggerAudits(ctx context.Context, groupId int, pagination restmodels.Pagination) ([]restmodels.GroupCapabilityTriggerAudit, int, error) {
	if pagination.Limit <= 0 {
		pagination.Limit = 50
	}
	total, err := countRows(ctx, persistence.db, "groupCapabilityTriggerAudit", "groupId = ?", []any{groupId})
	if err != nil {
		return nil, 0, err
	}
	limitClause, limitArgs := paginationClause(pagination)
	rows, err := persistence.db.QueryContext(ctx,
		`SELECT id, groupId, name, success, errorMessage, timestamp, arguments FROM groupCapabilityTriggerAudit WHERE groupId = ? ORDER BY timestamp DESC`+limitClause,
		append([]any{groupId}, limitArgs...)...,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var audits []restmodels.GroupCapabilityTriggerAudit
	for rows.Next() {
		var audit restmodels.GroupCapabilityTriggerAudit
		if err := rows.Scan(&audit.ID, &audit.GroupID, &audit.Name, &audit.Success, &audit.ErrorMessage, &audit.Timestamp, &audit.Arguments); err != nil {
			return nil, 0, err
		}
		audits = append(audits, audit)
	}
	return audits, total, rows.Err()
}

func (persistence mariadbPersistence) GetGroupCapabilityForActivation(ctx context.Context, storeIdentifier int, capabilityName string) (intermediaries.GroupCapabilityIntermediaryActivation, error) {
	capability := intermediaries.GroupCapabilityIntermediaryActivation{}
	row := persistence.db.QueryRowContext(ctx, `SELECT bridgeIdentifier, groupCapabilities.name, adapterId FROM groupCapabilities INNER JOIN groups on groupCapabilities.groupId = groups.id WHERE groupId = ? AND groupCapabilities.name = ?`, storeIdentifier, capabilityName)
	err := row.Scan(&capability.BridgeIdentifier, &capability.Name, &capability.AdapterId)
	if err != nil {
		if err != sql.ErrNoRows {
			return intermediaries.GroupCapabilityIntermediaryActivation{}, err
		}
		return intermediaries.GroupCapabilityIntermediaryActivation{}, huma.Error404NotFound(fmt.Sprintf("capability %s not found for group %d", capabilityName, storeIdentifier))
	}
	return capability, err
}
