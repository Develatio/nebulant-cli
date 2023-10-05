// Nebulant
// Copyright (C) 2020  Develatio Technologies S.L.

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package actors

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/blueprint"
)

type ec2Client func() ec2iface.EC2API

// ActionContext struct
type ActionContext struct {
	Rehearsal    bool
	AwsSess      *session.Session
	Action       *blueprint.Action
	Store        base.IStore
	Logger       base.ILogger
	NewEC2Client ec2Client
}

var NewActionContext = func(awsSess *session.Session, action *blueprint.Action, store base.IStore, logger base.ILogger) *ActionContext {
	l := logger.Duplicate()
	l.SetActionID(action.ActionID)
	return &ActionContext{
		AwsSess: awsSess,
		Action:  action,
		Store:   store,
		Logger:  l,
		NewEC2Client: func() ec2iface.EC2API {
			return ec2.New(awsSess)
		},
	}
}

// ActionFunc func
type ActionFunc func(ctx *ActionContext) (*base.ActionOutput, error)

type NextType int

const (
	// NextOKKO const 0
	NextOKKO NextType = iota
	// NextOK const 1
	NextOK
	// NextKO const
	NextKO
)

type ActionLayout struct {
	F ActionFunc
	N NextType
}

// ActionFuncMap map
var ActionFuncMap map[string]*ActionLayout = map[string]*ActionLayout{
	"attach_volume":  {F: AttachVolume, N: NextOKKO},
	"create_volume":  {F: CreateVolume, N: NextOKKO},
	"delete_volume":  {F: DeleteVolume, N: NextOKKO},
	"find_volumes":   {F: FindVolumes, N: NextOKKO},
	"findone_volume": {F: FindOneVolume, N: NextOKKO},
	"detach_volume":  {F: DetachVolume, N: NextOKKO},

	"run_instance":     {F: RunInstance, N: NextOKKO},
	"delete_instance":  {F: DeleteInstance, N: NextOKKO},
	"find_instances":   {F: FindInstances, N: NextOKKO},
	"findone_instance": {F: FindOneInstance, N: NextOKKO},
	"stop_instance":    {F: StopInstance, N: NextOKKO},
	"start_instance":   {F: StartInstance, N: NextOKKO},

	"find_images":   {F: FindImages, N: NextOKKO},
	"findone_image": {F: FindOneImage, N: NextOKKO},

	"find_ifaces":   {F: FindNetworkInterfaces, N: NextOKKO},
	"findone_iface": {F: FindNetworkInterface, N: NextOKKO},
	"delete_iface":  {F: DeleteNetworkInterface, N: NextOKKO},

	"find_databases":    {F: FindDatabases, N: NextOKKO},
	"findone_database":  {F: FindOneDatabase, N: NextOKKO},
	"create_db":         {F: CreateDatabase, N: NextOKKO},
	"delete_db":         {F: DeleteDatabase, N: NextOKKO},
	"database_snapshot": {F: CreateDatabase, N: NextOKKO},
	"restore_snapshot":  {F: RestoreSnapshotDatabase, N: NextOKKO},

	"allocate_address": {F: AllocateAddress, N: NextOKKO},
	"find_addresses":   {F: FindAddresses, N: NextOKKO},
	"findone_address":  {F: FindOneAddress, N: NextOKKO},

	"attach_address":  {F: AttachAddress, N: NextOKKO},
	"release_address": {F: ReleaseAddress, N: NextOKKO},
	"detach_address":  {F: DetachAddress, N: NextOKKO},

	"set_region": {F: SetRegion, N: NextOKKO},

	"find_vpcs":   {F: FindVpcs, N: NextOKKO},
	"findone_vpc": {F: FindOneVpc, N: NextOKKO},
	"delete_vpc":  {F: DeleteVpc, N: NextOKKO},

	"find_subnets":   {F: FindSubnets, N: NextOKKO},
	"findone_subnet": {F: FindOneSubnet, N: NextOKKO},
	"delete_subnet":  {F: DeleteSubnet, N: NextOKKO},

	"find_securitygroups":   {F: FindSecurityGroups, N: NextOKKO},
	"findone_securitygroup": {F: FindOneSecurityGroup, N: NextOKKO},
	"delete_securitygroup":  {F: DeleteSecurityGroup, N: NextOKKO},

	"find_keypairs":   {F: FindKeyPairs, N: NextOKKO},
	"findone_keypair": {F: FindOneKeyPair, N: NextOKKO},
	"delete_keypair":  {F: DeleteKeyPair, N: NextOKKO},
}
