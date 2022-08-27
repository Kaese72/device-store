package database

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Kaese72/device-store/config"
	"github.com/Kaese72/device-store/database/models"
	intermediary "github.com/Kaese72/device-store/models"
	"github.com/Kaese72/sdup-lib/devicestoretemplates"
	"github.com/Kaese72/sdup-lib/logging"
	"github.com/Kaese72/sdup-lib/sduptemplates"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDBDevicePersistence struct {
	mongoClient *mongo.Client
}

func (persistence MongoDBDevicePersistence) getDeviceAttributeCollection() *mongo.Collection {
	// TODO Make configurable, at least the database name
	return persistence.mongoClient.Database("huemie").Collection("deviceAttributes")
}

func (persistence MongoDBDevicePersistence) getDeviceCapabilityCollection() *mongo.Collection {
	// TODO Make configurable, at least the database name
	return persistence.mongoClient.Database("huemie").Collection("deviceCapabilities")
}

func (persistence MongoDBDevicePersistence) getBridgeCollection() *mongo.Collection {
	// TODO Make configurable, at least the database name
	return persistence.mongoClient.Database("huemie").Collection("bridges")
}

func (persistence MongoDBDevicePersistence) FilterDevices() ([]devicestoretemplates.Device, error) {
	// FIXME Implement capability modification
	attrHandle := persistence.getDeviceAttributeCollection()
	rDeviceAttributes := []models.MongoDeviceAttribute{}
	results, err := attrHandle.Find(context.TODO(), bson.D{})
	if err != nil {
		return nil, err
	}
	err = results.All(context.TODO(), &rDeviceAttributes)
	if err != nil {
		logging.Error("Error encountered while decoding devices", map[string]string{"error": err.Error()})
		return nil, UnknownError(err)
	}
	apiDevices, err := models.CreateAPIDevicesFromAttributes(rDeviceAttributes)
	if err != nil {
		logging.Error("Error encountered while converting mongo devices", map[string]string{"error": err.Error()})
		return nil, UnknownError(err)
	}

	responseDevices := []devicestoretemplates.Device{}
	for _, device := range apiDevices {
		responseDevices = append(responseDevices, device)
	}
	return responseDevices, nil
}

func (persistence MongoDBDevicePersistence) GetDeviceByIdentifier(identifier string, expandCapabilities bool) (devicestoretemplates.Device, error) {
	attrHandle := persistence.getDeviceAttributeCollection()
	deviceAttributes := []models.MongoDeviceAttribute{}
	// FIXME Deconding here is broken
	cursor, err := attrHandle.Find(context.TODO(), bson.D{primitive.E{Key: "deviceId", Value: identifier}})
	if err != nil {
		logging.Info(err.Error())
		return devicestoretemplates.Device{}, UnknownError(err)
	}
	err = cursor.All(context.TODO(), &deviceAttributes)
	if err != nil {
		logging.Info(err.Error())
		return devicestoretemplates.Device{}, UnknownError(err)
	}
	rDevice, err := models.ConvertAttributesToAPIDevice(deviceAttributes)
	if err != nil {
		logging.Info(err.Error())
		return devicestoretemplates.Device{}, UnknownError(err)
	}
	if expandCapabilities {
		capHandle := persistence.getDeviceCapabilityCollection()
		deviceCapabilities := []models.MongoDeviceCapability{}
		// FIXME Deconding here is broken
		cursor, err := capHandle.Find(context.TODO(), bson.D{primitive.E{Key: "deviceId", Value: identifier}})
		if err != nil {
			logging.Info(err.Error())
			return devicestoretemplates.Device{}, UnknownError(err)
		}
		err = cursor.All(context.TODO(), &deviceCapabilities)
		if err != nil {
			logging.Info(err.Error())
			return devicestoretemplates.Device{}, UnknownError(err)
		}
		logging.Info("Found capabilities for device", map[string]string{"identifier": identifier, "nCap": strconv.Itoa(len(deviceCapabilities))})
		deviceCapabilities = models.ReduceToMostRelevantCapabilities(deviceCapabilities)
		logging.Info("Found capabilities for device after deduplication", map[string]string{"identifier": identifier, "nCap": strconv.Itoa(len(deviceCapabilities))})
		rDevice.Capabilities = map[sduptemplates.CapabilityKey]devicestoretemplates.Capability{}
		for _, cap := range deviceCapabilities {
			rDevice.Capabilities[sduptemplates.CapabilityKey(cap.CapabilityName)] = cap.ConvertToAPICapability()
		}
	}
	return rDevice, nil
}

func (persistence MongoDBDevicePersistence) UpdateDeviceAttributes(apiDevice devicestoretemplates.Device, returnResult bool) (devicestoretemplates.Device, error) {
	if len(apiDevice.Identifier) == 0 {
		return devicestoretemplates.Device{}, NotFound(errors.New("can not update device without ID"))
	}
	attrHandle := persistence.getDeviceAttributeCollection()
	//FIXME updates
	attributeUpdates := models.ExtractAttributeModelsFromAPIDeviceModel(apiDevice)
	for _, attributeUpdate := range attributeUpdates {
		logging.Info("Updating attribute", map[string]string{"deviceId": attributeUpdate.DeviceId, "attributeName": attributeUpdate.AttributeName})
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

func (persistence MongoDBDevicePersistence) UpdateDeviceAttributesAndCapabilities(apiDevice devicestoretemplates.Device, sourceBridge devicestoretemplates.BridgeKey) (devicestoretemplates.Device, error) {
	_, err := persistence.UpdateDeviceAttributes(apiDevice, false)
	if err != nil {
		logging.Info(err.Error())
		return devicestoretemplates.Device{}, UnknownError(err)
	}
	bridgeHandle := persistence.getBridgeCollection()
	relevantBridge := models.Bridge{}
	err = bridgeHandle.FindOne(context.TODO(), bson.D{primitive.E{Key: "identifier", Value: sourceBridge}}).Decode(&relevantBridge)
	if err != nil {
		logging.Info(err.Error())
		return devicestoretemplates.Device{}, UnknownError(err)
	}
	mongoCapabilities := models.ExtractCapabilityModelsFromAPIDeviceModel(apiDevice, relevantBridge)
	capHandle := persistence.getDeviceCapabilityCollection()
	for _, capability := range mongoCapabilities {
		logging.Info("Updating attribute", map[string]string{"deviceId": capability.DeviceId, "capabilityName": capability.CapabilityName})
		capHandle.FindOneAndUpdate(context.TODO(), bson.D{primitive.E{Key: "deviceId", Value: capability.DeviceId}, primitive.E{Key: "capabilityName", Value: capability.CapabilityName}}, capability.ConvertToUpdate(), options.FindOneAndUpdate().SetUpsert(true))
		// FIXME Determine if todo
		// FIXME Error handling
		// FIXME verify upserts
	}
	return persistence.GetDeviceByIdentifier(apiDevice.Identifier, true)
}

func (persistence MongoDBDevicePersistence) GetCapability(deviceId string, capName string) (intermediary.CapabilityIntermediary, error) {
	logging.Info("Fetching capability", map[string]string{"deviceId": deviceId, "capabilityName": capName})
	capHandle := persistence.getDeviceCapabilityCollection()
	rCapabilities := []models.MongoDeviceCapability{}
	cursor, err := capHandle.Find(context.TODO(), bson.D{primitive.E{Key: "deviceId", Value: deviceId}, primitive.E{Key: "capabilityName", Value: capName}}, options.Find().SetLimit(1), options.Find().SetSort(bson.D{{Key: "lastSeen", Value: -1}}))
	if err != nil {
		logging.Info(err.Error())
		return intermediary.CapabilityIntermediary{}, UnknownError(err)
	}
	err = cursor.All(context.TODO(), &rCapabilities)
	if err != nil {
		logging.Info(err.Error())
		return intermediary.CapabilityIntermediary{}, UnknownError(err)
	}
	if len(rCapabilities) != 1 {
		return intermediary.CapabilityIntermediary{}, NotFound(err)
	}
	return intermediary.CapabilityIntermediary{
		DeviceId:            rCapabilities[0].DeviceId,
		CapabilityName:      rCapabilities[0].CapabilityName,
		CapabilityBridgeKey: rCapabilities[0].CapabilityBridgeKey,
		CapabilityBridgeURI: rCapabilities[0].CapabilityBridgeURI,
		LastSeen:            rCapabilities[0].LastSeen,
	}, nil
}

func (persistence MongoDBDevicePersistence) TriggerCapability(deviceId string, capName string, capArgs devicestoretemplates.CapabilityArgs) error {
	cap, err := persistence.GetCapability(deviceId, capName)
	if err != nil {
		return err
	}
	jsonEncoded, err := json.Marshal(capArgs)
	if err != nil {
		return err
	}
	capUri := fmt.Sprintf("%s/devices/%s/capabilities/%s", cap.CapabilityBridgeURI, deviceId, capName)
	resp, err := http.Post(capUri, "application/json", bytes.NewBuffer(jsonEncoded))
	if err != nil {
		// FIXME What if there is interesting debug information in the response?
		// We should log it or incorporate it in the response message or something
		return err
	}
	// It is the callers responsibility to Close the body reader
	// But there should not be anything of interest here at the moment
	defer resp.Body.Close()
	return nil
}

func (persistence MongoDBDevicePersistence) EnrollBridge(apiBridge devicestoretemplates.Bridge) (devicestoretemplates.Bridge, error) {
	if len(apiBridge.Identifier) == 0 {
		// FIXME Allow allocation of bridge key
		apiBridge.Identifier = devicestoretemplates.BridgeKey(uuid.New().String())
	}
	if len(apiBridge.URI) == 0 {
		// FIXME Validate URI
		return devicestoretemplates.Bridge{}, errors.New("URI may not be empty")
	}
	// FIXME Run healthcheck multiple times
	if err := apiBridge.HealthCheck(); err != nil {
		return devicestoretemplates.Bridge{}, errors.New("health check failed")
	}
	dbBridge := models.Bridge{
		Identifier: apiBridge.Identifier,
		URI:        apiBridge.URI,
	}
	handle := persistence.getBridgeCollection()
	updateResults, err := handle.UpdateOne(context.TODO(), bson.D{primitive.E{Key: "identifier", Value: dbBridge.Identifier}}, bson.M{"$setOnInsert": dbBridge}, options.Update().SetUpsert(true))
	if err != nil {
		return devicestoretemplates.Bridge{}, err
	}
	if updateResults.UpsertedCount == 0 {
		return devicestoretemplates.Bridge{}, errors.New("bridge Identifier already exists")
	}

	return devicestoretemplates.Bridge{
		Identifier: dbBridge.Identifier,
		URI:        dbBridge.URI,
	}, nil
}

func (persistence MongoDBDevicePersistence) ForgetBridge(bridgeKey devicestoretemplates.BridgeKey) error {
	handle := persistence.getBridgeCollection()
	result, err := handle.DeleteMany(context.TODO(), bson.D{primitive.E{Key: "identifier", Value: bridgeKey}})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return NotFound(errors.New("could not find Bridge"))
	}
	capHandle := persistence.getDeviceCapabilityCollection()
	result, err = capHandle.DeleteMany(context.TODO(), bson.D{primitive.E{Key: "capabilityBridgeKey", Value: bridgeKey}})
	if err != nil {
		return err
	}
	logging.Info("Capabilities removed as a consequence of removing bridge", map[string]string{"amount": strconv.Itoa(int(result.DeletedCount))})
	return nil
}

func (persistence MongoDBDevicePersistence) ListBridges() ([]devicestoretemplates.Bridge, error) {
	handle := persistence.getBridgeCollection()
	dbBridges := []models.Bridge{}
	cursor, err := handle.Find(context.TODO(), bson.D{})
	if err != nil {
		return nil, err
	}
	err = cursor.All(context.TODO(), &dbBridges)
	if err != nil {
		return nil, err
	}
	apiBridges := []devicestoretemplates.Bridge{}
	for _, bridge := range dbBridges {
		apiBridges = append(apiBridges, bridge.ConvertToAPIBridge())
	}
	return apiBridges, nil
}

func NewMongoDBDevicePersistence(conf config.MongoDBConfig) (DevicePersistenceDB, error) {
	mongoClient, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(conf.ConnectionString))
	if err != nil {
		return nil, err
	}

	if err := mongoClient.Ping(context.TODO(), nil); err != nil {
		return nil, err
	}

	// FIXME Verify connection successful

	return MongoDBDevicePersistence{
		mongoClient: mongoClient,
	}, nil
}
