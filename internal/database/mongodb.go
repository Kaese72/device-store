package database

import (
	"context"
	"strconv"

	"github.com/pkg/errors"

	"github.com/Kaese72/device-store/internal/config"
	"github.com/Kaese72/device-store/internal/database/models"
	"github.com/Kaese72/device-store/internal/logging"
	intermediary "github.com/Kaese72/device-store/internal/models/intermediary"
	devicestoretemplates "github.com/Kaese72/device-store/rest/models"
	"github.com/Kaese72/huemie-lib/liberrors"
	"go.elastic.co/apm/module/apmmongo/v2"
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

func (persistence MongoDBDevicePersistence) getGroupCapabilityCollection() *mongo.Collection {
	// TODO Make configurable, at least the database name
	return persistence.mongoClient.Database(persistence.dbName).Collection("bridgeGroupCapabilities")
}

func (persistence MongoDBDevicePersistence) getGroupCollection() *mongo.Collection {
	// TODO Make configurable, at least the database name
	return persistence.mongoClient.Database(persistence.dbName).Collection("bridgeGroups")
}

func (persistence MongoDBDevicePersistence) FilterDevices(ctx context.Context) ([]devicestoretemplates.Device, error) {
	// FIXME Implement capability modification
	attrHandle := persistence.getDeviceAttributeCollection()
	rDeviceAttributes := []models.MongoDeviceAttribute{}
	results, err := attrHandle.Find(ctx, bson.D{})
	if err != nil {
		return nil, liberrors.NewApiError(liberrors.InternalError, err)
	}
	err = results.All(ctx, &rDeviceAttributes)
	if err != nil {
		logging.Error("Error encountered while decoding devices", ctx, map[string]interface{}{"error": err.Error()})
		return nil, liberrors.NewApiError(liberrors.InternalError, err)
	}
	apiDevices, err := models.CreateAPIDevicesFromAttributes(rDeviceAttributes)
	if err != nil {
		logging.Error("Error encountered while converting mongo devices", ctx, map[string]interface{}{"error": err.Error()})
		return nil, liberrors.NewApiError(liberrors.InternalError, err)
	}

	responseDevices := []devicestoretemplates.Device{}
	for _, device := range apiDevices {
		responseDevices = append(responseDevices, device)
	}
	return responseDevices, nil
}

func (persistence MongoDBDevicePersistence) GetDeviceByIdentifier(identifier string, expandCapabilities bool, ctx context.Context) (devicestoretemplates.Device, error) {
	attrHandle := persistence.getDeviceAttributeCollection()
	deviceAttributes := []models.MongoDeviceAttribute{}
	// FIXME Deconding here is broken
	cursor, err := attrHandle.Find(ctx, bson.D{primitive.E{Key: "deviceId", Value: identifier}})
	if err != nil {
		logging.Info(err.Error(), ctx)
		return devicestoretemplates.Device{}, liberrors.NewApiError(liberrors.InternalError, err)
	}
	err = cursor.All(ctx, &deviceAttributes)
	if err != nil {
		logging.Info(err.Error(), ctx)
		return devicestoretemplates.Device{}, liberrors.NewApiError(liberrors.InternalError, err)
	}
	rDevice, err := models.ConvertAttributesToAPIDevice(deviceAttributes)
	if err != nil {
		logging.Info(err.Error(), ctx)
		return devicestoretemplates.Device{}, liberrors.NewApiError(liberrors.InternalError, err)
	}
	if expandCapabilities {
		capHandle := persistence.getDeviceCapabilityCollection()
		deviceCapabilities := []models.MongoDeviceCapability{}
		// FIXME Deconding here is broken
		cursor, err := capHandle.Find(ctx, bson.D{primitive.E{Key: "deviceId", Value: identifier}})
		if err != nil {
			logging.Info(err.Error(), ctx)
			return devicestoretemplates.Device{}, liberrors.NewApiError(liberrors.InternalError, err)
		}
		err = cursor.All(ctx, &deviceCapabilities)
		if err != nil {
			logging.Info(err.Error(), ctx)
			return devicestoretemplates.Device{}, liberrors.NewApiError(liberrors.InternalError, err)
		}
		logging.Info("Found capabilities for device", ctx, map[string]interface{}{"identifier": identifier, "nCap": strconv.Itoa(len(deviceCapabilities))})
		deviceCapabilities = models.ReduceToMostRelevantCapabilities(deviceCapabilities)
		logging.Info("Found capabilities for device after deduplication", ctx, map[string]interface{}{"identifier": identifier, "nCap": strconv.Itoa(len(deviceCapabilities))})
		rDevice.Capabilities = map[devicestoretemplates.CapabilityKey]devicestoretemplates.Capability{}
		for _, cap := range deviceCapabilities {
			rDevice.Capabilities[devicestoretemplates.CapabilityKey(cap.CapabilityName)] = cap.ConvertToAPICapability()
		}
	}
	return rDevice, nil
}

func (persistence MongoDBDevicePersistence) UpdateDeviceAttributes(apiDevice devicestoretemplates.Device, returnResult bool, ctx context.Context) (devicestoretemplates.Device, error) {
	if len(apiDevice.Identifier) == 0 {
		return devicestoretemplates.Device{}, liberrors.NewApiError(liberrors.NotFound, errors.New("can not update device without ID"))
	}
	attrHandle := persistence.getDeviceAttributeCollection()
	//FIXME updates
	attributeUpdates := models.ExtractAttributeModelsFromAPIDeviceModel(apiDevice)
	for _, attributeUpdate := range attributeUpdates {
		logging.Info("Updating attribute", ctx, map[string]interface{}{"deviceId": attributeUpdate.DeviceId, "attributeName": attributeUpdate.AttributeName})
		attrHandle.FindOneAndUpdate(ctx, bson.D{primitive.E{Key: "deviceId", Value: attributeUpdate.DeviceId}, primitive.E{Key: "attributeName", Value: attributeUpdate.AttributeName}}, attributeUpdate.ConvertToUpdate(), options.FindOneAndUpdate().SetUpsert(true))
		// FIXME Determine if todo
		// FIXME Error handling
		// FIXME verify upserts
	}
	if returnResult {
		return persistence.GetDeviceByIdentifier(apiDevice.Identifier, true, ctx)
	} else {
		return devicestoretemplates.Device{}, nil
	}
}

func (persistence MongoDBDevicePersistence) UpdateDeviceAttributesAndCapabilities(apiDevice devicestoretemplates.Device, sourceBridge string, ctx context.Context) (devicestoretemplates.Device, error) {
	_, err := persistence.UpdateDeviceAttributes(apiDevice, false, ctx)
	if err != nil {
		logging.Info(err.Error(), ctx)
		return devicestoretemplates.Device{}, liberrors.NewApiError(liberrors.InternalError, err)
	}
	mongoCapabilities := models.ExtractCapabilityModelsFromAPIDeviceModel(apiDevice, sourceBridge)
	capHandle := persistence.getDeviceCapabilityCollection()
	for _, capability := range mongoCapabilities {
		logging.Info("Updating attribute", ctx, map[string]interface{}{"deviceId": capability.DeviceId, "capabilityName": capability.CapabilityName})
		capHandle.FindOneAndUpdate(ctx, bson.D{primitive.E{Key: "deviceId", Value: capability.DeviceId}, primitive.E{Key: "capabilityName", Value: capability.CapabilityName}}, capability.ConvertToUpdate(), options.FindOneAndUpdate().SetUpsert(true))
		// FIXME Determine if todo
		// FIXME Error handling
		// FIXME verify upserts
	}
	return persistence.GetDeviceByIdentifier(apiDevice.Identifier, true, ctx)
}

func (persistence MongoDBDevicePersistence) GetCapability(deviceId string, capName string, ctx context.Context) (intermediary.CapabilityIntermediary, error) {
	logging.Info("Fetching capability", ctx, map[string]interface{}{"deviceId": deviceId, "capabilityName": capName})
	capHandle := persistence.getDeviceCapabilityCollection()
	rCapabilities := []models.MongoDeviceCapability{}
	cursor, err := capHandle.Find(ctx, bson.D{primitive.E{Key: "deviceId", Value: deviceId}, primitive.E{Key: "capabilityName", Value: capName}}, options.Find().SetLimit(1), options.Find().SetSort(bson.D{{Key: "lastSeen", Value: -1}}))
	if err != nil {
		logging.Info(err.Error(), ctx)
		return intermediary.CapabilityIntermediary{}, liberrors.NewApiError(liberrors.InternalError, err)
	}
	err = cursor.All(ctx, &rCapabilities)
	if err != nil {
		logging.Info(err.Error(), ctx)
		return intermediary.CapabilityIntermediary{}, liberrors.NewApiError(liberrors.InternalError, err)
	}
	if len(rCapabilities) != 1 {
		return intermediary.CapabilityIntermediary{}, liberrors.NewApiError(liberrors.NotFound, errors.New("could not find device capability"))
	}
	return intermediary.CapabilityIntermediary{
		DeviceId:            rCapabilities[0].DeviceId,
		CapabilityName:      rCapabilities[0].CapabilityName,
		CapabilityBridgeKey: rCapabilities[0].CapabilityBridgeKey,
		LastSeen:            rCapabilities[0].LastSeen,
	}, nil
}

func (persistence MongoDBDevicePersistence) FilterGroups(ctx context.Context) ([]devicestoretemplates.Group, error) {
	// FIXME Implement capability modification
	gHandle := persistence.getGroupCollection()
	dbGroups := []models.MongoGroup{}
	results, err := gHandle.Find(ctx, bson.D{})
	if err != nil {
		return nil, liberrors.NewApiError(liberrors.InternalError, err)
	}
	err = results.All(ctx, &dbGroups)
	if err != nil {
		logging.Error("Error encountered while decoding devices", ctx, map[string]interface{}{"error": err.Error()})
		return nil, liberrors.NewApiError(liberrors.InternalError, err)
	}
	rGroups := []devicestoretemplates.Group{}
	for _, model := range dbGroups {
		rGroups = append(rGroups, model.ConvertToAPI())
	}
	return rGroups, nil
}

func (persistence MongoDBDevicePersistence) GetGroupByIdentifier(groupId string, expandCapabilities bool, ctx context.Context) (devicestoretemplates.Group, error) {
	groupHandle := persistence.getGroupCollection()
	rGroups := []models.MongoGroup{}
	cursor, err := groupHandle.Find(ctx, bson.D{primitive.E{Key: "groupId", Value: groupId}}, options.Find().SetLimit(1), options.Find().SetSort(bson.D{{Key: "lastSeen", Value: -1}}))
	if err != nil {
		logging.Error(err.Error(), ctx)
		return devicestoretemplates.Group{}, liberrors.NewApiError(liberrors.InternalError, err)
	}
	err = cursor.All(ctx, &rGroups)
	if err != nil {
		logging.Error(err.Error(), ctx)
		return devicestoretemplates.Group{}, liberrors.NewApiError(liberrors.InternalError, err)
	}
	if len(rGroups) < 1 {
		logging.Info(err.Error(), ctx)
		return devicestoretemplates.Group{}, liberrors.NewApiError(liberrors.NotFound, err)
	}
	rGroup := devicestoretemplates.Group{
		Identifier: rGroups[0].GroupId,
		Name:       rGroups[0].GroupName,
	}
	if !expandCapabilities {
		return rGroup, nil
	}
	gCapHandle := persistence.getGroupCapabilityCollection()
	cursor, err = gCapHandle.Find(ctx, bson.D{primitive.E{Key: "groupId", Value: groupId}}, options.Find(), options.Find().SetSort(bson.D{{Key: "lastSeen", Value: -1}}))
	if err != nil {
		logging.Error(err.Error(), ctx)
		return devicestoretemplates.Group{}, liberrors.NewApiError(liberrors.InternalError, err)
	}
	rCapabilities := []models.MongoGroupCapability{}
	err = cursor.All(ctx, &rCapabilities)
	if err != nil {
		logging.Error(err.Error(), ctx)
		return devicestoretemplates.Group{}, liberrors.NewApiError(liberrors.InternalError, err)
	}
	rGroup.Capabilities = map[devicestoretemplates.CapabilityKey]devicestoretemplates.Capability{}
	for _, capability := range rCapabilities {
		rGroup.Capabilities[devicestoretemplates.CapabilityKey(capability.CapabilityName)] = devicestoretemplates.Capability{
			LastSeen: capability.LastSeen,
		}
	}
	return rGroup, nil
}

func (persistence MongoDBDevicePersistence) GetGroupCapability(groupId string, capName string, ctx context.Context) (intermediary.GroupCapabilityIntermediary, error) {
	gCapHandle := persistence.getGroupCapabilityCollection()
	gCaps := []intermediary.GroupCapabilityIntermediary{}
	cursor, err := gCapHandle.Find(ctx, bson.D{primitive.E{Key: "groupId", Value: groupId}, primitive.E{Key: "capabilityName", Value: capName}}, options.Find(), options.Find().SetSort(bson.D{{Key: "lastSeen", Value: -1}}))
	if err != nil {
		logging.Error(err.Error(), ctx)
		return intermediary.GroupCapabilityIntermediary{}, liberrors.NewApiError(liberrors.InternalError, err)
	}
	err = cursor.All(ctx, &gCaps)
	if err != nil {
		logging.Error(err.Error(), ctx)
		return intermediary.GroupCapabilityIntermediary{}, liberrors.NewApiError(liberrors.InternalError, err)
	}
	if len(gCaps) < 1 {
		logging.Info(err.Error(), ctx)
		return intermediary.GroupCapabilityIntermediary{}, liberrors.NewApiError(liberrors.NotFound, err)
	}
	return gCaps[0], nil
}

func (persistence MongoDBDevicePersistence) updateGroupCapability(capability models.MongoGroupCapability, ctx context.Context) error {
	if capability.GroupId == "" {
		return liberrors.NewApiError(liberrors.UserError, errors.New("must supply group id when updating group capabilities"))
	}
	if capability.CapabilityName == "" {
		return liberrors.NewApiError(liberrors.UserError, errors.New("must supply capability name when updating group capabilities"))
	}
	gCapHandle := persistence.getGroupCapabilityCollection()
	gCapHandle.FindOneAndUpdate(ctx, bson.D{primitive.E{Key: "groupId", Value: capability.GroupId}, primitive.E{Key: "capabilityName", Value: capability.CapabilityName}}, capability.ConvertToUpdate(), options.FindOneAndUpdate().SetUpsert(true))
	// FIXME error handling
	return nil
}

func (persistence MongoDBDevicePersistence) UpdateGroup(group devicestoretemplates.Group, sourceBridge string, ctx context.Context) (devicestoretemplates.Group, error) {
	if group.Identifier == "" {
		return devicestoretemplates.Group{}, liberrors.NewApiError(liberrors.NotFound, errors.New("can not update device without ID"))
	}
	gHandle := persistence.getGroupCollection()
	dbGroup := models.MongoGroupFromAPIModel(group, sourceBridge)
	gHandle.FindOneAndUpdate(ctx, bson.D{primitive.E{Key: "groupId", Value: group.Identifier}}, dbGroup.ConvertToUpdate(), options.FindOneAndUpdate().SetUpsert(true))
	for _, capability := range models.ExtractGroupCapabilityFromAPI(group, sourceBridge) {
		if err := persistence.updateGroupCapability(capability, ctx); err != nil {
			return devicestoretemplates.Group{}, err
		}
	}
	return devicestoretemplates.Group{}, nil
}

func NewMongoDBDevicePersistence(conf config.MongoDBConfig) (DevicePersistenceDB, error) {
	mongoClient, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(conf.ConnectionString).SetMonitor(apmmongo.CommandMonitor()))
	if err != nil {
		return nil, liberrors.NewApiError(liberrors.InternalError, err)
	}

	if err := mongoClient.Ping(context.TODO(), nil); err != nil {
		return nil, liberrors.NewApiError(liberrors.InternalError, err)
	}

	// FIXME Verify connection successful

	return MongoDBDevicePersistence{
		mongoClient: mongoClient,
		dbName:      conf.DbName,
	}, nil
}
