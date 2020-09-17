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

// ActionCreateVolume const
const ActionCreateVolume = "create_volume"

// ActionDeleteVolume const
const ActionDeleteVolume = "delete_volume"

// ActionFindVolumes const
const ActionFindVolumes = "find_volumes"

// ActionFindOneVolume const
const ActionFindOneVolume = "findone_volume"

// ActionAttachVolume const
const ActionAttachVolume = "attach_volume"

// ActionDetachVolume const
const ActionDetachVolume = "detach_volume"

// ActionFindInstances const
const ActionFindInstances = "find_instances"

// ActionFindOneInstance const
const ActionFindOneInstance = "findone_instance"

// ActionRunInstance const
const ActionRunInstance = "run_instance"

// ActionDeleteInstance const
const ActionDeleteInstance = "delete_instance"

// ActionStopInstance const
const ActionStopInstance = "stop_instance"

// ActionStartInstance const
const ActionStartInstance = "start_instance"

// ActionFindImages const
const ActionFindImages = "find_images"

// ActionFindOneImage const
const ActionFindOneImage = "findone_image"

// ActionFindNetworkInterfaces const
const ActionFindNetworkInterfaces = "find_ifaces"

// ActionFindOneNetworkInterface const
const ActionFindOneNetworkInterface = "findone_iface"

// ActionDeleteNetworkInterface func
const ActionDeleteNetworkInterface = "delete_iface"

// ActionCreateDatabase const
const ActionCreateDatabase = "create_db"

// ActionDeleteDatabase const
const ActionDeleteDatabase = "delete_db"

// ActionFindDatabases const
const ActionFindDatabases = "find_databases"

// ActionFindOneDatabase const
const ActionFindOneDatabase = "findone_database"

// ActionCreateDatabaseSnapshot const
const ActionCreateDatabaseSnapshot = "database_snapshot"

// ActionRestoreDatabaseSnapshot const
const ActionRestoreDatabaseSnapshot = "restore_snapshot"

// ActionAllocateAddress const
const ActionAllocateAddress = "allocate_address"

// ActionReleaseAddress const
const ActionReleaseAddress = "release_address"

// ActionFindAddresses const
const ActionFindAddresses = "find_addresses"

// ActionFindOneAddress const
const ActionFindOneAddress = "findone_address"

// ActionAttachAddress const
const ActionAttachAddress = "attach_address"

// ActionDetachAddress const
const ActionDetachAddress = "detach_address"

// ActionFindVpcs const
const ActionFindVpcs = "find_vpcs"

// ActionFindOneVpc const
const ActionFindOneVpc = "findone_vpc"

// ActionDeleteVpc const
const ActionDeleteVpc = "delete_vpc"

// ActionFindSubnets const
const ActionFindSubnets = "find_subnets"

// ActionFindOneSubnet const
const ActionFindOneSubnet = "findone_subnet"

// ActionDeleteSubnet const
const ActionDeleteSubnet = "delete_subnet"

// ActionSetRegion const
const ActionSetRegion = "set_region"

// ActionFindSecurityGroups const
const ActionFindSecurityGroups = "find_securitygroups"

// ActionFindOneSecurityGroup const
const ActionFindOneSecurityGroup = "findone_securitygroup"

// ActionDeleteSecurityGroup const
const ActionDeleteSecurityGroup = "delete_securitygroup"

// ActionFindKeyPairs const
const ActionFindKeyPairs = "find_keypairs"

// ActionFindOneKeyPair const
const ActionFindOneKeyPair = "findone_keypair"

// ActionDeleteKeyPair const
const ActionDeleteKeyPair = "delete_keypair"

type ec2Client func() ec2iface.EC2API

// ActionContext struct
type ActionContext struct {
	AwsSess      *session.Session
	Action       *blueprint.Action
	Store        base.IStore
	Logger       base.ILogger
	NewEC2Client ec2Client
}

var NewActionContext = func(awsSess *session.Session, action *blueprint.Action, store base.IStore, logger base.ILogger) *ActionContext {
	return &ActionContext{
		AwsSess: awsSess,
		Action:  action,
		Store:   store,
		Logger:  logger,
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
	ActionAttachVolume:  {F: AttachVolume, N: NextOKKO},
	ActionCreateVolume:  {F: CreateVolume, N: NextOKKO},
	ActionDeleteVolume:  {F: DeleteVolume, N: NextOKKO},
	ActionFindVolumes:   {F: FindVolumes, N: NextOKKO},
	ActionFindOneVolume: {F: FindOneVolume, N: NextOKKO},
	ActionDetachVolume:  {F: DetachVolume, N: NextOKKO},

	ActionRunInstance:     {F: RunInstance, N: NextOKKO},
	ActionDeleteInstance:  {F: DeleteInstance, N: NextOKKO},
	ActionFindInstances:   {F: FindInstances, N: NextOKKO},
	ActionFindOneInstance: {F: FindOneInstance, N: NextOKKO},
	ActionStopInstance:    {F: StopInstance, N: NextOKKO},
	ActionStartInstance:   {F: StartInstance, N: NextOKKO},

	ActionFindImages:   {F: FindImages, N: NextOKKO},
	ActionFindOneImage: {F: FindOneImage, N: NextOKKO},

	ActionFindNetworkInterfaces:   {F: FindNetworkInterfaces, N: NextOKKO},
	ActionFindOneNetworkInterface: {F: FindNetworkInterface, N: NextOKKO},
	ActionDeleteNetworkInterface:  {F: DeleteNetworkInterface, N: NextOKKO},

	ActionFindDatabases:           {F: FindDatabases, N: NextOKKO},
	ActionFindOneDatabase:         {F: FindOneDatabase, N: NextOKKO},
	ActionCreateDatabase:          {F: CreateDatabase, N: NextOKKO},
	ActionDeleteDatabase:          {F: DeleteDatabase, N: NextOKKO},
	ActionCreateDatabaseSnapshot:  {F: CreateDatabase, N: NextOKKO},
	ActionRestoreDatabaseSnapshot: {F: RestoreSnapshotDatabase, N: NextOKKO},

	ActionAllocateAddress: {F: AllocateAddress, N: NextOKKO},
	ActionFindAddresses:   {F: FindAddresses, N: NextOKKO},
	ActionFindOneAddress:  {F: FindOneAddress, N: NextOKKO},

	ActionAttachAddress:  {F: AttachAddress, N: NextOKKO},
	ActionReleaseAddress: {F: ReleaseAddress, N: NextOKKO},
	ActionDetachAddress:  {F: DetachAddress, N: NextOKKO},

	ActionSetRegion: {F: SetRegion, N: NextOKKO},

	ActionFindVpcs:   {F: FindVpcs, N: NextOKKO},
	ActionFindOneVpc: {F: FindOneVpc, N: NextOKKO},
	ActionDeleteVpc:  {F: DeleteVpc, N: NextOKKO},

	ActionFindSubnets:   {F: FindSubnets, N: NextOKKO},
	ActionFindOneSubnet: {F: FindOneSubnet, N: NextOKKO},
	ActionDeleteSubnet:  {F: DeleteSubnet, N: NextOKKO},

	ActionFindSecurityGroups:   {F: FindSecurityGroups, N: NextOKKO},
	ActionFindOneSecurityGroup: {F: FindOneSecurityGroup, N: NextOKKO},
	ActionDeleteSecurityGroup:  {F: DeleteSecurityGroup, N: NextOKKO},

	ActionFindKeyPairs:   {F: FindKeyPairs, N: NextOKKO},
	ActionFindOneKeyPair: {F: FindOneKeyPair, N: NextOKKO},
	ActionDeleteKeyPair:  {F: DeleteKeyPair, N: NextOKKO},
}
