package database

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/Kaese72/device-store/internal/config"
	"github.com/Kaese72/device-store/internal/database/models"
	intermediary "github.com/Kaese72/device-store/internal/models/intermediary"
	"github.com/Kaese72/device-store/internal/systemerrors"
	devicestoretemplates "github.com/Kaese72/device-store/rest/models"
	"github.com/Kaese72/huemie-lib/logging"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDBDevicePersistence struct {
	mongoClient *mongo.Client
	dbName      string
}

func (persistence MongoDBDevicePersistence) getDeviceAttributeCollection() *mongo.Collection {
	// TODO Make configurable, at least the database name
	return persistence.mongoClient.Database(persistence.dbName).Collection("deviceAttributes")
}

func (persistence MongoDBDevicePersistence) getDeviceCapabilityCollection() *mongo.Collection {
	// TODO Make configurable, at least the database name
	return persistence.mongoClient.Database(persistence.dbName).Collection("deviceCapabilities")
}

func (persistence MongoDBDevicePersistence) FilterDevices() ([]devicestoretemplates.Device, systemerrors.SystemError) {
	// FIXME Implement capability modification
	attrHandle := persistence.getDeviceAttributeCollection()
	rDeviceAttributes := []models.MongoDeviceAttribute{}
	results, err := attrHandle.Find(context.TODO(), bson.D{})
	if err != nil {
		return nil, systemerrors.WrapSystemError(err, systemerrors.InternalError)
	}
	err = results.All(context.TODO(), &rDeviceAttributes)
	if err != nil {
		logging.Error("Error encountered while decoding devices", map[string]interface{}{"error": err.Error()})
		return nil, systemerrors.WrapSystemError(err, systemerrors.InternalError)
	}
	apiDevices, err := models.CreateAPIDevicesFromAttributes(rDeviceAttributes)
	if err != nil {
		logging.Error("Error encountered while converting mongo devices", map[string]interface{}{"error": err.Error()})
		return nil, systemerrors.WrapSystemError(err, systemerrors.InternalError)
	}

	responseDevices := []devicestoretemplates.Device{}
	for _, device := range apiDevices {
		responseDevices = append(responseDevices, device)
	}
	return responseDevices, nil
}

func (persistence MongoDBDevicePersistence) GetDeviceByIdentifier(identifier string, expandCapabilities bool) (devicestoretemplates.Device, systemerrors.SystemError) {
	attrHandle := persistence.getDeviceAttributeCollection()
	deviceAttributes := []models.MongoDeviceAttribute{}
	// FIXME Deconding here is broken
	cursor, err := attrHandle.Find(context.TODO(), bson.D{primitive.E{Key: "deviceId", Value: identifier}})
	if err != nil {
		logging.Info(err.Error())
		return devicestoretemplates.Device{}, systemerrors.WrapSystemError(err, systemerrors.InternalError)
	}
	err = cursor.All(context.TODO(), &deviceAttributes)
	if err != nil {
		logging.Info(err.Error())
		return devicestoretemplates.Device{}, systemerrors.WrapSystemError(err, systemerrors.InternalError)
	}
	rDevice, err := models.ConvertAttributesToAPIDevice(deviceAttributes)
	if err != nil {
		logging.Info(err.Error())
		return devicestoretemplates.Device{}, systemerrors.WrapSystemError(err, systemerrors.InternalError)
	}
	if expandCapabilities {
		capHandle := persistence.getDeviceCapabilityCollection()
		deviceCapabilities := []models.MongoDeviceCapability{}
		// FIXME Deconding here is broken
		cursor, err := capHandle.Find(context.TODO(), bson.D{primitive.E{Key: "deviceId", Value: identifier}})
		if err != nil {
			logging.Info(err.Error())
			return devicestoretemplates.Device{}, systemerrors.WrapSystemError(err, systemerrors.InternalError)
		}
		err = cursor.All(context.TODO(), &deviceCapabilities)
		if err != nil {
			logging.Info(err.Error())
			return devicestoretemplates.Device{}, systemerrors.WrapSystemError(err, systemerrors.InternalError)
		}
		logging.Info("Found capabilities for device", map[string]interface{}{"identifier": identifier, "nCap": strconv.Itoa(len(deviceCapabilities))})
		deviceCapabilities = models.ReduceToMostRelevantCapabilities(deviceCapabilities)
		logging.Info("Found capabilities for device after deduplication", map[string]interface{}{"identifier": identifier, "nCap": strconv.Itoa(len(deviceCapabilities))})
		rDevice.Capabilities = map[devicestoretemplates.CapabilityKey]devicestoretemplates.Capability{}
		for _, cap := range deviceCapabilities {
			rDevice.Capabilities[devicestoretemplates.CapabilityKey(cap.CapabilityName)] = cap.ConvertToAPICapability()
		}
	}
	return rDevice, nil
}

func (persistence MongoDBDevicePersistence) UpdateDeviceAttributes(apiDevice devicestoretemplates.Device, returnResult bool) (devicestoretemplates.Device, systemerrors.SystemError) {
	if len(apiDevice.Identifier) == 0 {
		return devicestoretemplates.Device{}, systemerrors.WrapSystemError(errors.New("can not update device without ID"), systemerrors.NotFound)
	}
	attrHandle := persistence.getDeviceAttributeCollection()
	//FIXME updates
	attributeUpdates := models.ExtractAttributeModelsFromAPIDeviceModel(apiDevice)
	for _, attributeUpdate := range attributeUpdates {
		logging.Info("Updating attribute", map[string]interface{}{"deviceId": attributeUpdate.DeviceId, "attributeName": attributeUpdate.AttributeName})
		attrHandle.FindOneAndUpdate(context.TODO(), bson.D{primitive.E{Key: "deviceId", Value: attributeUpdate.DeviceId}, primitive.E{Key: "attributeName", Value: attributeUpdate.AttributeName}}, attributeUpdate.ConvertToUpdate(), options.FindOneAndUpdate().SetUpsert(true))
		// FIXME Determine if todo
		// FIXME Error handling
		// FIXME verify upserts
	}
	if returnResult {
		return persistence.GetDeviceByIdentifier(apiDevice.Identifier, true)
	} else {
		return devicestoretemplates.Device{}, nil
	}
}

func (persistence MongoDBDevicePersistence) UpdateDeviceAttributesAndCapabilities(apiDevice devicestoretemplates.Device, sourceBridge string) (devicestoretemplates.Device, systemerrors.SystemError) {
	_, err := persistence.UpdateDeviceAttributes(apiDevice, false)
	if err != nil {
		logging.Info(err.Error())
		return devicestoretemplates.Device{}, systemerrors.WrapSystemError(err, systemerrors.InternalError)
	}
	mongoCapabilities := models.ExtractCapabilityModelsFromAPIDeviceModel(apiDevice, sourceBridge)
	capHandle := persistence.getDeviceCapabilityCollection()
	for _, capability := range mongoCapabilities {
		logging.Info("Updating attribute", map[string]interface{}{"deviceId": capability.DeviceId, "capabilityName": capability.CapabilityName})
		capHandle.FindOneAndUpdate(context.TODO(), bson.D{primitive.E{Key: "deviceId", Value: capability.DeviceId}, primitive.E{Key: "capabilityName", Value: capability.CapabilityName}}, capability.ConvertToUpdate(), options.FindOneAndUpdate().SetUpsert(true))
		// FIXME Determine if todo
		// FIXME Error handling
		// FIXME verify upserts
	}
	return persistence.GetDeviceByIdentifier(apiDevice.Identifier, true)
}

func (persistence MongoDBDevicePersistence) GetCapability(deviceId string, capName string) (intermediary.CapabilityIntermediary, systemerrors.SystemError) {
	logging.Info("Fetching capability", map[string]interface{}{"deviceId": deviceId, "capabilityName": capName})
	capHandle := persistence.getDeviceCapabilityCollection()
	rCapabilities := []models.MongoDeviceCapability{}
	cursor, err := capHandle.Find(context.TODO(), bson.D{primitive.E{Key: "deviceId", Value: deviceId}, primitive.E{Key: "capabilityName", Value: capName}}, options.Find().SetLimit(1), options.Find().SetSort(bson.D{{Key: "lastSeen", Value: -1}}))
	if err != nil {
		logging.Info(err.Error())
		return intermediary.CapabilityIntermediary{}, systemerrors.WrapSystemError(err, systemerrors.InternalError)
	}
	err = cursor.All(context.TODO(), &rCapabilities)
	if err != nil {
		logging.Info(err.Error())
		return intermediary.CapabilityIntermediary{}, systemerrors.WrapSystemError(err, systemerrors.InternalError)
	}
	if len(rCapabilities) != 1 {
		return intermediary.CapabilityIntermediary{}, systemerrors.WrapSystemError(fmt.Errorf("Unexpected amount of capabilities found, %d != 1", len(rCapabilities)), systemerrors.NotFound)
	}
	return intermediary.CapabilityIntermediary{
		DeviceId:            rCapabilities[0].DeviceId,
		CapabilityName:      rCapabilities[0].CapabilityName,
		CapabilityBridgeKey: rCapabilities[0].CapabilityBridgeKey,
		LastSeen:            rCapabilities[0].LastSeen,
	}, nil
}

func NewMongoDBDevicePersistence(conf config.MongoDBConfig) (DevicePersistenceDB, systemerrors.SystemError) {
	mongoClient, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(conf.ConnectionString))
	if err != nil {
		return nil, systemerrors.WrapSystemError(err, systemerrors.InternalError)
	}

	if err := mongoClient.Ping(context.TODO(), nil); err != nil {
		return nil, systemerrors.WrapSystemError(err, systemerrors.InternalError)
	}

	// FIXME Verify connection successful

	return MongoDBDevicePersistence{
		mongoClient: mongoClient,
		dbName:      conf.DbName,
	}, nil
}
