package mariadb

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/Kaese72/device-store/internal/config"
	"github.com/Kaese72/device-store/internal/logging"
	"github.com/Kaese72/device-store/internal/persistence/intermediaries"
	"github.com/Kaese72/huemie-lib/liberrors"
	"github.com/georgysavva/scany/sqlscan"
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

func (persistence mariadbPersistence) GetDevices(ctx context.Context) ([]intermediaries.DeviceIntermediary, error) {
	fields := []string{
		"id",
		"bridgeIdentifier",
		"bridgeKey",
		"(SELECT COALESCE(JSON_ARRAYAGG(JSON_OBJECT(\"name\", name, \"boolean\", booleanValue, \"numeric\", numericValue, \"text\", textValue)), JSON_ARRAY()) FROM deviceAttributes WHERE deviceAttributes.deviceId = devices.id) as attributes",
		"(SELECT COALESCE(JSON_ARRAYAGG(JSON_OBJECT(\"name\", name)), JSON_ARRAY()) FROM deviceCapabilities WHERE deviceId = devices.id) as capabilities",
	}
	query := `SELECT ` + strings.Join(fields, ",") + ` FROM devices`
	devices := []intermediaries.DeviceIntermediary{}
	err := sqlscan.Select(ctx, persistence.db, &devices, query)
	return devices, err
}

type idList []struct {
	ID int `db:"id"`
}

func (persistence mariadbPersistence) PostDevice(ctx context.Context, device intermediaries.DeviceIntermediary) error {
	foundIds := idList{}
	err := sqlscan.Select(ctx, persistence.db, &foundIds, `SELECT id FROM devices WHERE bridgeIdentifier = ? AND bridgeKey = ?`, device.BridgeIdentifier, device.BridgeKey)
	if err != nil {
		return err
	}
	var deviceId int
	if len(foundIds) == 0 {
		createdIdsList := idList{}
		result, err := persistence.db.QueryContext(ctx, `INSERT INTO devices (bridgeIdentifier, bridgeKey) VALUES (?, ?) RETURNING id`, device.BridgeIdentifier, device.BridgeKey)
		if err != nil {
			return err
		}
		err = sqlscan.ScanAll(&createdIdsList, result)
		if err != nil {
			return err
		}
		deviceId = createdIdsList[0].ID
	} else {
		deviceId = foundIds[0].ID
	}
	for _, attribute := range device.Attributes {
		_, err = persistence.db.ExecContext(ctx, `INSERT INTO deviceAttributes (deviceId, name, booleanValue, numericValue, textValue) VALUES (?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE booleanValue=VALUES(booleanValue),numericValue=VALUES(numericValue),textValue=VALUES(textValue)`, deviceId, attribute.Name, attribute.Boolean(), attribute.Numeric, attribute.Text)
		if err != nil {
			return err
		}
	}
	for _, capability := range device.Capabilities {
		_, err = persistence.db.ExecContext(ctx, `INSERT IGNORE INTO deviceCapabilities (deviceId, name) VALUES (?, ?)`, deviceId, capability.Name)
		if err != nil {
			return err
		}
	}
	return nil
}

func (persistence mariadbPersistence) GetCapabilityForActivation(ctx context.Context, storeIdentifier int, capabilityName string) (intermediaries.CapabilityIntermediaryActivation, error) {
	capabilities := []intermediaries.CapabilityIntermediaryActivation{}
	err := sqlscan.Select(ctx, persistence.db, &capabilities, `SELECT bridgeIdentifier, name, bridgeKey FROM deviceCapabilities INNER JOIN devices on deviceCapabilities.deviceId = devices.id WHERE deviceId = ? AND name = ?`, storeIdentifier, capabilityName)
	if err != nil {
		return intermediaries.CapabilityIntermediaryActivation{}, err
	}
	if len(capabilities) == 0 {
		return intermediaries.CapabilityIntermediaryActivation{}, liberrors.NewApiError(liberrors.NotFound, fmt.Errorf("capability %s not found for device %d", capabilityName, storeIdentifier))
	}
	return capabilities[0], err
}
