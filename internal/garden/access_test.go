package garden

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGardenAccessCapabilities(t *testing.T) {
	cases := map[string]struct {
		access GardenAccess
		want   GardenPermissions
	}{
		"owner":             {GardenAccess{Owner: true}, allPermissions},
		"manage":            {GardenAccess{Permissions: GardenPermissions{GardenPermissionManage}}, allPermissions},
		"harvest":           {GardenAccess{Permissions: GardenPermissions{GardenPermissionHarvest}}, GardenPermissions{GardenPermissionView, GardenPermissionHarvest}},
		"plant":             {GardenAccess{Permissions: GardenPermissions{GardenPermissionPlant}}, GardenPermissions{GardenPermissionView, GardenPermissionPlant}},
		"view":              {GardenAccess{Permissions: GardenPermissions{GardenPermissionView}}, GardenPermissions{GardenPermissionView}},
		"plant and harvest": {GardenAccess{Permissions: GardenPermissions{GardenPermissionPlant, GardenPermissionHarvest}}, GardenPermissions{GardenPermissionView, GardenPermissionPlant, GardenPermissionHarvest}},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.access.Capabilities())
			assert.Equal(t, tc.access.Owner, tc.access.Can(ownerOnly))
		})
	}
}
