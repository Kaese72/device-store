package models

import (
	devicestoretemplates "github.com/Kaese72/device-store/rest/models"
	"go.mongodb.org/mongo-driver/bson"
)

type MongoGroup struct {
	GroupId        string `bson:"groupId"`
	GroupName      string `bson:"groupName"`
	GroupBridgeKey string `bson:"groupBridgeKey"`
}

func (group MongoGroup) ConvertToUpdate() bson.M {
	return bson.M{
		"$set": map[string]string{
			"groupId":        group.GroupId,
			"groupName":      group.GroupName,
			"groupBridgeKey": string(group.GroupBridgeKey),
		},
		// Not needed, but set anyway because I might want it in the future
		"$currentDate": bson.M{
			"lastSeen": bson.M{"$type": "timestamp"},
		},
	}
}

func (group MongoGroup) ConvertToAPI() devicestoretemplates.Group {
	return devicestoretemplates.Group{
		Identifier: group.GroupId,
		Name:       group.GroupName,
	}
}

func MongoGroupFromAPIModel(group devicestoretemplates.Group, bridgeKey string) MongoGroup {
	return MongoGroup{GroupId: group.Identifier, GroupName: group.Name, GroupBridgeKey: bridgeKey}
}
