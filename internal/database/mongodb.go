package database

import (
	"context"
	"strconv"

	"github.com/pkg/errors"

	"github.com/Kaese72/device-store/internal/config"
	"github.com/Kaese72/device-store/internal/database/models"
	"github.com/Kaese72/device-store/internal/logging"
	"github.com/Kaese72/device-store/internal/models/intermediaries"
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

func (persistence MongoDBDevicePersistence) getDeviceCollection() *mongo.Collection {
	// TODO Make configurable, at least the database name
	return persistence.mongoClient.Database(persistence.dbName).Collection("devices")
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

func (persistence MongoDBDevicePersistence) purge() error {
	logging.Error("Purging mongodb database", context.TODO())
	result, err := persistence.getDeviceCollection().DeleteMany(context.TODO(), bson.D{})
	if err != nil {
		logging.Error("Failed to purge devices", context.TODO())
		return err
	}
	logging.Info("Successfully deleted devices", context.TODO(), map[string]interface{}{"n": result.DeletedCount})

	result, err = persistence.getDeviceAttributeCollection().DeleteMany(context.TODO(), bson.D{})
	if err != nil {
		logging.Error("Failed to purge devices", context.TODO())
		return err
	}
	logging.Info("Successfully deleted devices", context.TODO(), map[string]interface{}{"n": result.DeletedCount})

	result, err = persistence.getDeviceCapabilityCollection().DeleteMany(context.TODO(), bson.D{})
	if err != nil {
		logging.Error("Failed to purge devices", context.TODO())
		return err
	}
	logging.Info("Successfully deleted devices", context.TODO(), map[string]interface{}{"n": result.DeletedCount})

	result, err = persistence.getGroupCapabilityCollection().DeleteMany(context.TODO(), bson.D{})
	if err != nil {
		logging.Error("Failed to purge devices", context.TODO())
		return err
	}
	logging.Info("Successfully deleted devices", context.TODO(), map[string]interface{}{"n": result.DeletedCount})

	result, err = persistence.getGroupCollection().DeleteMany(context.TODO(), bson.D{})
	if err != nil {
		logging.Error("Failed to purge devices", context.TODO())
		return err
	}
	logging.Info("Successfully deleted devices", context.TODO(), map[string]interface{}{"n": result.DeletedCount})

	return nil
}

func (persistence MongoDBDevicePersistence) GetDevices(ctx context.Context) ([]intermediaries.DeviceIntermediary, error) {
	// FIXME Implement capability modification
	devicesHandle := persistence.getDeviceCollection()
	rDevices := []models.MongoDevice{}
	results, err := devicesHandle.Find(ctx, bson.D{})
	if err != nil {
		return nil, liberrors.NewApiError(liberrors.InternalError, err)
	}
	err = results.All(ctx, &rDevices)
	if err != nil {
		logging.Error("Error encountered while decoding devices", ctx, map[string]interface{}{"error": err.Error()})
		return nil, liberrors.NewApiError(liberrors.InternalError, err)
	}

	responseDevices := []intermediaries.DeviceIntermediary{}
	for _, device := range rDevices {
		responseDevices = append(responseDevices, intermediaries.DeviceIntermediary{
			BridgeIdentifier:      device.BridgeIdentifier,
			BridgeKey:             device.BridgeKey,
			DeviceStoreIdentifier: device.DeviceStoreIdentifier,
		})
	}
	return responseDevices, nil
}

func (persistence MongoDBDevicePersistence) GetStoreDevice(identifier string, expandCapabilities bool, ctx context.Context) (intermediaries.DeviceIntermediary, error) {
	deviceHandle := persistence.getDeviceCollection()
	mDevice := models.MongoDevice{}
	// FIXME Deconding here is broken
	err := deviceHandle.FindOne(ctx, models.UniqueDeviceStoreQuery(identifier)).Decode(&mDevice)
	if err != nil {
		logging.Info(err.Error(), ctx)
		return intermediaries.DeviceIntermediary{}, liberrors.NewApiError(liberrors.InternalError, err)
	}
	// FIXME Move conversion to other location
	return intermediaries.DeviceIntermediary{
		BridgeIdentifier:      mDevice.BridgeIdentifier,
		BridgeKey:             mDevice.BridgeKey,
		DeviceStoreIdentifier: mDevice.DeviceStoreIdentifier,
	}, nil
}

func (persistence MongoDBDevicePersistence) GetDeviceAttributes(storeDeviceIdentifier string, ctx context.Context) ([]intermediaries.AttributeIntermediary, error) {
	attrHandle := persistence.getDeviceAttributeCollection()
	deviceAttributes := []models.MongoDeviceAttribute{}
	// FIXME Deconding here is broken
	cursor, err := attrHandle.Find(ctx, bson.D{primitive.E{Key: "storeDeviceIdentifier", Value: storeDeviceIdentifier}})
	if err != nil {
		logging.Info(err.Error(), ctx)
		return nil, liberrors.NewApiError(liberrors.InternalError, err)
	}
	err = cursor.All(ctx, &deviceAttributes)
	if err != nil {
		logging.Info(err.Error(), ctx)
		return nil, liberrors.NewApiError(liberrors.InternalError, err)
	}
	results := []intermediaries.AttributeIntermediary{}
	for _, attribute := range deviceAttributes {
		results = append(results, intermediaries.AttributeIntermediary{
			DeviceStoreIdentifier: storeDeviceIdentifier,
			Name:                  attribute.Name,
			State: intermediaries.AttributeStateIntermediary{
				Boolean: attribute.State.Boolean,
				Numeric: attribute.State.Numeric,
				Text:    attribute.State.Text,
			},
		})
	}
	return results, nil
}

func (persistence MongoDBDevicePersistence) updateDeviceAttribute(attribute intermediaries.AttributeIntermediary, ctx context.Context) error {
	attrHandle := persistence.getDeviceAttributeCollection()
	mongoAttribute := models.MongoDeviceAttribute{
		DeviceStoreIdentifier: attribute.DeviceStoreIdentifier,
		Name:                  attribute.Name,
		State: models.AttributeState{
			Boolean: attribute.State.Boolean,
			Text:    attribute.State.Text,
			Numeric: attribute.State.Numeric,
		},
	}
	logging.Info("Updating attribute", ctx, map[string]interface{}{"device": attribute.DeviceStoreIdentifier, "name": attribute.Name})
	err := attrHandle.FindOneAndUpdate(ctx, mongoAttribute.UniqueQuery(), mongoAttribute.ConvertToUpdate(), options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)).Err()
	return err
}

func (persistence MongoDBDevicePersistence) updateDevice(device intermediaries.DeviceIntermediary, ctx context.Context) (models.MongoDevice, error) {
	deviceHandle := persistence.getDeviceCollection()
	mongoDevice := models.MongoDevice{
		BridgeIdentifier: device.BridgeIdentifier,
		BridgeKey:        device.BridgeKey,
		// DeviceStoreIdentifier: device.DeviceStoreIdentifier, // Should not be set
	}
	logging.Info("Updating device", ctx, map[string]interface{}{"device": device.DeviceStoreIdentifier})
	rDevice := models.MongoDevice{}
	err := deviceHandle.FindOneAndUpdate(ctx, mongoDevice.UniqueBridgeQuery(), mongoDevice.ConvertToUpdate(), options.FindOneAndUpdate().SetReturnDocument(options.After).SetUpsert(true)).Decode(&rDevice)
	return rDevice, err
}

func (persistence MongoDBDevicePersistence) UpdateDevice(apiDevice devicestoretemplates.Device, bridgeKey string, ctx context.Context) error {
	if len(apiDevice.Identifier) == 0 {
		return liberrors.NewApiError(liberrors.NotFound, errors.New("can not update device without bridge device identifier"))
	}
	if len(apiDevice.BridgeKey) == 0 {
		return liberrors.NewApiError(liberrors.NotFound, errors.New("can not update device without bridge key"))
	}
	// # Update device
	mongoDevice, err := persistence.updateDevice(intermediaries.DeviceIntermediary{
		BridgeIdentifier: apiDevice.Identifier,
		BridgeKey:        apiDevice.BridgeKey,
	}, ctx)
	if err != nil {
		return liberrors.NewApiError(liberrors.NotFound, errors.New("Failed to update device entity"))
	}
	if len(mongoDevice.DeviceStoreIdentifier) == 0 {
		return liberrors.NewApiError(liberrors.NotFound, errors.New("Could not fetch store identifier"))
	}
	// # Update attributes
	for attributeKey, attributeState := range apiDevice.Attributes {
		err := persistence.updateDeviceAttribute(intermediaries.AttributeIntermediary{
			DeviceStoreIdentifier: mongoDevice.DeviceStoreIdentifier,
			Name:                  string(attributeKey),
			State: intermediaries.AttributeStateIntermediary{
				Boolean: attributeState.Boolean,
				Numeric: attributeState.Numeric,
				Text:    attributeState.Text,
			},
		}, ctx)
		if err != nil {
			return liberrors.NewApiError(liberrors.NotFound, errors.New("Could not update attribute"))
		}
	}

	// # Update capabilities
	mongoCapabilities := models.ExtractCapabilityModelsFromAPIDeviceModel(apiDevice, mongoDevice.DeviceStoreIdentifier, bridgeKey)
	capHandle := persistence.getDeviceCapabilityCollection()
	for _, capability := range mongoCapabilities {
		logging.Info("Updating capability", ctx, map[string]interface{}{"DeviceStoreIdentifier": capability.StoreDeviceIdentifier, "name": capability.Name})
		err := capHandle.FindOneAndUpdate(ctx, capability.UniqueQuery(), capability.ConvertToUpdate(), options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)).Err()
		if err != nil {
			return liberrors.NewApiError(liberrors.NotFound, errors.New("Could not update capability"))
		}
	}
	return nil
}

func (persistence MongoDBDevicePersistence) GetCapability(deviceId string, capName string, ctx context.Context) (intermediaries.CapabilityIntermediary, error) {
	logging.Info("Fetching capability", ctx, map[string]interface{}{"deviceId": deviceId, "capabilityName": capName})
	capHandle := persistence.getDeviceCapabilityCollection()
	rCapability := models.MongoDeviceCapability{}
	err := capHandle.FindOne(ctx, models.MongoDeviceCapabilityUniqueQuery(deviceId, capName), options.FindOne().SetSort(bson.D{{Key: "lastSeen", Value: -1}})).Decode(&rCapability)
	if err != nil {
		logging.Info(err.Error(), ctx)
		return intermediaries.CapabilityIntermediary{}, liberrors.NewApiError(liberrors.InternalError, err)
	}
	return intermediaries.CapabilityIntermediary{
		StoreDeviceIdentifier: rCapability.StoreDeviceIdentifier,
		Name:                  rCapability.Name,
		BridgeKey:             rCapability.BridgeKey,
		LastSeen:              rCapability.LastSeen,
	}, nil
}

func (persistence MongoDBDevicePersistence) GetDeviceCapabilities(deviceId string, ctx context.Context) ([]intermediaries.CapabilityIntermediary, error) {
	capHandle := persistence.getDeviceCapabilityCollection()
	deviceCapabilities := []models.MongoDeviceCapability{}
	// FIXME Deconding here is broken
	cursor, err := capHandle.Find(ctx, models.MongoDeviceCapabilityUniqueQuery(deviceId, ""))
	if err != nil {
		logging.Info(err.Error(), ctx)
		return nil, liberrors.NewApiError(liberrors.InternalError, err)
	}
	err = cursor.All(ctx, &deviceCapabilities)
	if err != nil {
		logging.Info(err.Error(), ctx)
		return nil, liberrors.NewApiError(liberrors.InternalError, err)
	}
	logging.Info("Found capabilities for device", ctx, map[string]interface{}{"identifier": deviceId, "nCap": strconv.Itoa(len(deviceCapabilities))})

	rCapabilities := []intermediaries.CapabilityIntermediary{}
	for _, capability := range deviceCapabilities {
		rCapabilities = append(rCapabilities, intermediaries.CapabilityIntermediary{
			StoreDeviceIdentifier:  capability.StoreDeviceIdentifier,
			BridgeDeviceIdentifier: capability.BridgeDeviceIdentifier,
			BridgeKey:              capability.BridgeKey,
			Name:                   capability.Name,
			LastSeen:               capability.LastSeen,
		})
	}
	return rCapabilities, nil
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

func (persistence MongoDBDevicePersistence) GetGroupCapability(groupId string, capName string, ctx context.Context) (intermediaries.GroupCapabilityIntermediary, error) {
	gCapHandle := persistence.getGroupCapabilityCollection()
	gCaps := []models.MongoGroupCapability{}
	cursor, err := gCapHandle.Find(ctx, bson.D{primitive.E{Key: "groupId", Value: groupId}, primitive.E{Key: "capabilityName", Value: capName}}, options.Find(), options.Find().SetSort(bson.D{{Key: "lastSeen", Value: -1}}))
	if err != nil {
		logging.Error(err.Error(), ctx)
		return intermediaries.GroupCapabilityIntermediary{}, liberrors.NewApiError(liberrors.InternalError, err)
	}
	err = cursor.All(ctx, &gCaps)
	if err != nil {
		logging.Error(err.Error(), ctx)
		return intermediaries.GroupCapabilityIntermediary{}, liberrors.NewApiError(liberrors.InternalError, err)
	}
	if len(gCaps) < 1 {
		logging.Info(err.Error(), ctx)
		return intermediaries.GroupCapabilityIntermediary{}, liberrors.NewApiError(liberrors.NotFound, err)
	}
	return intermediaries.GroupCapabilityIntermediary{
		GroupId:             gCaps[0].GroupId,
		CapabilityName:      gCaps[0].CapabilityName,
		CapabilityBridgeKey: gCaps[0].GroupBridgeKey,
		LastSeen:            gCaps[0].LastSeen,
	}, nil
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

func NewMongoDBDevicePersistence(conf config.MongoDBConfig) (MongoDBDevicePersistence, error) {
	mongoClient, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(conf.ConnectionString).SetMonitor(apmmongo.CommandMonitor()))
	if err != nil {
		return MongoDBDevicePersistence{}, liberrors.NewApiError(liberrors.InternalError, err)
	}

	if err := mongoClient.Ping(context.TODO(), nil); err != nil {
		return MongoDBDevicePersistence{}, liberrors.NewApiError(liberrors.InternalError, err)
	}

	// FIXME Verify connection successful

	return MongoDBDevicePersistence{
		mongoClient: mongoClient,
		dbName:      conf.DbName,
	}, nil
}
