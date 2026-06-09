package connector

import (
	"fmt"
	"strconv"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/pagination"
)

func formatObjectID(resourceTypeID string, id int64) string {
	return fmt.Sprintf("%c%d", resourceTypeID[0], id)
}

func parseObjectID(id string) (int64, error) {
	return strconv.ParseInt(id[1:], 10, 64)
}

func formatGroupObjectID(id int64) string {
	return fmt.Sprintf("%d", id)
}

// stringFromProfile reads a string value from an AccountInfo profile map, tolerating
// absent or non-string values.
func stringFromProfile(profile map[string]interface{}, key string) string {
	if v, ok := profile[key].(string); ok {
		return v
	}
	return ""
}

func parseGroupObjectID(id string) (int64, error) {
	return strconv.ParseInt(id, 10, 64)
}

func parsePageToken(i string, resourceID *v2.ResourceId) (*pagination.Bag, error) {
	b := &pagination.Bag{}
	err := b.Unmarshal(i)
	if err != nil {
		return nil, err
	}

	if b.Current() == nil {
		b.Push(pagination.PageState{
			ResourceTypeID: resourceID.ResourceType,
			ResourceID:     resourceID.Resource,
		})
	}

	return b, nil
}
