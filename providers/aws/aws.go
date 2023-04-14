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

package aws

import (
	"fmt"
	"net"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/blueprint"
	hook_providers "github.com/develatio/nebulant-cli/hook/providers"
	"github.com/develatio/nebulant-cli/providers/aws/actors"
)

func ActionValidator(action *blueprint.Action) error {
	if action.Provider != "aws" {
		return nil
	}
	al, exists := actors.ActionFuncMap[action.ActionName]
	if !exists {
		return fmt.Errorf("aws: invalid action name " + action.ActionName)
	}
	if al.N == actors.NextOK && len(action.NextAction.NextKo) > 0 {
		return fmt.Errorf("generic: action " + action.ActionName + " has no KO port")
	}

	if al.N == actors.NextKO && len(action.NextAction.NextOk) > 0 {
		return fmt.Errorf("generic: action " + action.ActionName + " has no OK port")
	}

	ac := &actors.ActionContext{
		Rehearsal: true,
		Action:    action,
	}
	_, err := al.F(ac)
	if err != nil {
		return err
	}

	return nil
}

// New var
var New base.ProviderInitFunc = func(store base.IStore) (base.IProvider, error) {
	prov := &Provider{
		store:  store,
		Logger: store.GetLogger(),
	}
	return prov, nil
}

// Provider struct
type Provider struct {
	store  base.IStore
	Logger base.ILogger
	mu     sync.Mutex
}

// DumpPrivateVars func
func (p *Provider) DumpPrivateVars(freshStore base.IStore) {
	sess := p.store.GetPrivateVar("awsSess")
	if sess != nil {
		newSess := sess.(*session.Session).Copy()
		freshStore.SetPrivateVar("awsSess", newSess)
	}
}

// HandleAction func
func (p *Provider) HandleAction(action *blueprint.Action) (*base.ActionOutput, error) {
	p.Logger.LogDebug("AWS: Received action " + action.ActionName)
	err := p.touchSession()
	if err != nil {
		return nil, err
	}

	if al, exists := actors.ActionFuncMap[action.ActionName]; exists {
		sess := p.store.GetPrivateVar("awsSess").(*session.Session)
		return al.F(actors.NewActionContext(sess, action, p.store, p.Logger))
	}
	return nil, fmt.Errorf("AWS: Unknown action: " + action.ActionName)
}

// OnActionErrorHook func
func (p *Provider) OnActionErrorHook(aout *base.ActionOutput) ([]*blueprint.Action, error) {

	// retry on net err, skip others
	if _, ok := aout.Records[0].Error.(net.Error); ok {
		phcontext := &hook_providers.ProviderHookContext{
			Logger: p.Logger,
			Store:  p.store,
		}
		return hook_providers.DefaultOnActionErrorHook(phcontext, aout)
	}

	// retry on some aws api errs
	if aerr, ok := aout.Records[0].Error.(awserr.Error); ok {

		// retry on allowed-retry aws http status codes
		if reqErr, ok := aout.Records[0].Error.(awserr.Error).(awserr.RequestFailure); ok {
			// p.Logger.LogDebug(fmt.Sprintf("AWS: Retry Hook. AWS Service errCode:%v - StatusCode:%v - RequestID:%v", aerr.Code(), reqErr.StatusCode(), reqErr.RequestID()))
			switch reqErr.StatusCode() {
			case 418:
				p.Logger.LogDebug("Tea time 🫖")
			case 429, 502, 503, 504:
				phcontext := &hook_providers.ProviderHookContext{
					Logger: p.Logger,
					Store:  p.store,
				}
				return hook_providers.DefaultOnActionErrorHook(phcontext, aout)
			}
		}

		// retry if:
		switch aerr.Code() {
		case
			// https://docs.aws.amazon.com/AWSEC2/latest/APIReference/errors-overview.html
			//
			// ---- Common client error codes
			// AuthFailure
			// Blocked
			// DryRunOperation
			// DryRun
			// IdempotentParameterMismatch
			// IncompleteSignature
			// InvalidAction
			// InvalidCharacter
			// InvalidClientTokenId
			// InvalidPaginationToken
			// InvalidParameter
			// InvalidParameterCombination
			// InvalidParameterValue
			// InvalidQueryParameter
			// "MalformedQueryString",
			// MissingAction
			// MissingAuthenticationToken
			// MissingParameter
			// OptInRequired
			// PendingVerification
			// RequestExpired
			// UnauthorizedOperation
			// DecodeAuthorizationMessage
			// UnknownParameter
			// UnsupportedInstanceAttribute
			// UnsupportedOperation
			// UnsupportedProtocol
			// ValidationError
			// ---- Client error codes for specific actions
			// ActiveVpcPeeringConnectionPerVpcLimitExceeded
			// AddressLimitExceeded
			// AsnConflict
			// AttachmentLimitExceeded
			// BootForVolumeTypeUnsupported
			// BundlingInProgress
			// CannotDelete
			// ClientVpnAuthorizationRuleLimitExceeded
			// ClientVpnCertificateRevocationListLimitExceeded
			// ClientVpnEndpointAssociationExists
			// ClientVpnEndpointLimitExceeded
			// ClientVpnRouteLimitExceeded
			// ClientVpnTerminateConnectionsLimitExceeded
			// CidrConflict
			// ConcurrentCreateImageNoRebootLimitExceeded
			// ConcurrentSnapshotLimitExceeded
			// ConcurrentTagAccess
			// CreditSpecificationUpdateInProgress
			// CustomerGatewayLimitExceeded
			// CustomerKeyHasBeenRevoked
			// DeleteConversionTaskError
			// DefaultSubnetAlreadyExistsInAvailabilityZone
			// DefaultVpcAlreadyExists
			// DefaultVpcDoesNotExist
			// DependencyViolation
			// DisallowedForDedicatedTenancyNetwork
			// DiskImageSizeTooLarge
			// DuplicateSubnetsInSameZone
			// EIPMigratedToVpc
			// EncryptedVolumesNotSupported
			// ExistingVpcEndpointConnections
			// FleetNotInModifiableState
			// FlowLogAlreadyExists
			// FlowLogsLimitExceeded
			// FilterLimitExceeded
			// Gateway.NotAttached
			// HostAlreadyCoveredByReservation
			// HostLimitExceeded
			// IdempotentInstanceTerminated
			// InaccessibleStorageLocation
			// IncorrectInstanceState
			// IncorrectModificationState
			// IncorrectState
			// IncompatibleHostRequirements
			// InstanceAlreadyLinked
			// InstanceCreditSpecification.NotSupported
			// InstanceLimitExceeded
			// InsufficientCapacityOnHost
			// InsufficientFreeAddressesInSubnet
			// InsufficientReservedInstancesCapacity
			// InterfaceInUseByTrafficMirrorSession
			// InterfaceInUseByTrafficMirrorTarget
			// InternetGatewayLimitExceeded
			// InvalidAddress.Locked
			// "InvalidAddress.Malformed",
			// InvalidAddress.NotFound
			// InvalidAddressID.NotFound
			// InvalidAffinity
			// InvalidAllocationID.NotFound
			// InvalidAMIAttributeItemValue
			// "InvalidAMIID.Malformed",
			// InvalidAMIID.NotFound
			// InvalidAMIID.Unavailable
			// InvalidAMIName.Duplicate
			// "InvalidAMIName.Malformed",
			// InvalidAssociationID.NotFound
			// InvalidAttachment.NotFound
			// InvalidAttachmentID.NotFound
			// InvalidAutoPlacement
			// InvalidAvailabilityZone
			// InvalidBlockDeviceMapping
			// InvalidBundleID.NotFound
			// InvalidCidr.InUse
			// InvalidClientToken
			// InvalidClientVpnAssociationIdNotFound
			// InvalidClientVpnConnection.IdNotFound
			// InvalidClientVpnConnection.UserNotFound
			// InvalidClientVpnDuplicateAssociationException
			// InvalidClientVpnDuplicateAuthorizationRule
			// InvalidClientVpnDuplicateRoute
			// InvalidClientVpnEndpointAuthorizationRuleNotFound
			// InvalidClientVpnRouteNotFound
			// InvalidClientVpnSubnetId.DifferentAccount
			// InvalidClientVpnSubnetId.DuplicateAz
			// InvalidClientVpnSubnetId.NotFound
			// InvalidClientVpnSubnetId.OverlappingCidr
			// InvalidClientVpnActiveAssociationNotFound
			// InvalidClientVpnEndpointId.NotFound
			// InvalidConversionTaskId
			// "InvalidConversionTaskId.Malformed",
			// "InvalidCpuCredits.Malformed",
			// InvalidCustomerGateway.DuplicateIpAddress
			// "InvalidCustomerGatewayId.Malformed",
			// InvalidCustomerGatewayID.NotFound
			// InvalidCustomerGatewayState
			// InvalidDevice.InUse
			// InvalidDhcpOptionID.NotFound
			// InvalidDhcpOptionsID.NotFound
			// "InvalidDhcpOptionsId.Malformed",
			// InvalidExportTaskID.NotFound
			// InvalidFilter
			// InvalidFlowLogId.NotFound
			// InvalidFormat
			// "InvalidFpgaImageID.Malformed",
			// InvalidFpgaImageID.NotFound
			// InvalidGatewayID.NotFound
			// InvalidGroup.Duplicate
			// "InvalidGroupId.Malformed",
			// InvalidGroup.InUse
			// InvalidGroup.NotFound
			// InvalidGroup.Reserved
			// InvalidHostConfiguration
			// InvalidHostId
			// "InvalidHostID.Malformed",
			// "InvalidHostId.Malformed",
			// InvalidHostID.NotFound
			// InvalidHostId.NotFound
			// "InvalidHostReservationId.Malformed",
			// "InvalidHostReservationOfferingId.Malformed",
			// InvalidHostState
			// "InvalidIamInstanceProfileArn.Malformed",
			// InvalidID
			// InvalidInput
			// InvalidInstanceAttributeValue
			// InvalidInstanceCreditSpecification.DuplicateInstanceId
			// InvalidInstanceFamily
			// InvalidInstanceID
			// "InvalidInstanceID.Malformed",
			// InvalidInstanceID.NotFound
			// InvalidInstanceID.NotLinkable
			// InvalidInstanceState
			// InvalidInstanceType
			// InvalidInterface.IpAddressLimitExceeded
			// "InvalidInternetGatewayId.Malformed",
			// InvalidInternetGatewayID.NotFound
			// InvalidIPAddress.InUse
			// "InvalidKernelId.Malformed",
			// InvalidKey.Format
			// InvalidKeyPair.Duplicate
			// InvalidKeyPair.Format
			// InvalidKeyPair.NotFound
			// "InvalidCapacityReservationIdMalformedException",
			// InvalidCapacityReservationIdNotFoundException
			// "InvalidLaunchTemplateId.Malformed",
			// InvalidLaunchTemplateId.NotFound
			// InvalidLaunchTemplateId.VersionNotFound
			// InvalidLaunchTemplateName.AlreadyExistsException
			// "InvalidLaunchTemplateName.MalformedException",
			// InvalidLaunchTemplateName.NotFoundException
			// InvalidManifest
			// InvalidMaxResults
			// InvalidNatGatewayID.NotFound
			// InvalidNetworkAclEntry.NotFound
			// "InvalidNetworkAclId.Malformed",
			// InvalidNetworkAclID.NotFound
			// "InvalidNetworkLoadBalancerArn.Malformed",
			// InvalidNetworkLoadBalancerArn.NotFound
			// "InvalidNetworkInterfaceAttachmentId.Malformed",
			// InvalidNetworkInterface.InUse
			// "InvalidNetworkInterfaceId.Malformed",
			// InvalidNetworkInterfaceID.NotFound
			// InvalidNextToken
			// InvalidOption.Conflict
			// InvalidPermission.Duplicate
			// "InvalidPermission.Malformed",
			// InvalidPermission.NotFound
			// InvalidPlacementGroup.Duplicate
			// InvalidPlacementGroup.InUse
			// InvalidPlacementGroup.Unknown
			// InvalidPolicyDocument
			// "InvalidPrefixListId.Malformed",
			// InvalidPrefixListId.NotFound
			// InvalidProductInfo
			// InvalidPurchaseToken.Expired
			// "InvalidPurchaseToken.Malformed",
			// InvalidQuantity
			// "InvalidRamDiskId.Malformed",
			// InvalidRegion
			// InvalidRequest
			// "InvalidReservationID.Malformed",
			// InvalidReservationID.NotFound
			// InvalidReservedInstancesId
			// InvalidReservedInstancesOfferingId
			// InvalidResourceType.Unknown
			// InvalidRoute.InvalidState
			// "InvalidRoute.Malformed",
			// InvalidRoute.NotFound
			// "InvalidRouteTableId.Malformed",
			// InvalidRouteTableID.NotFound
			// InvalidScheduledInstance
			// "InvalidSecurityGroupId.Malformed",
			// InvalidSecurityGroupID.NotFound
			// InvalidSecurity.RequestHasExpired
			// InvalidServiceName
			// "InvalidSnapshotID.Malformed",
			// InvalidSnapshot.InUse
			// InvalidSnapshot.NotFound
			// InvalidSpotDatafeed.NotFound
			// InvalidSpotFleetRequestConfig
			// "InvalidSpotFleetRequestId.Malformed",
			// InvalidSpotFleetRequestId.NotFound
			// "InvalidSpotInstanceRequestID.Malformed",
			// InvalidSpotInstanceRequestID.NotFound
			// InvalidState
			// InvalidStateTransition
			// InvalidSubnet
			// InvalidSubnet.Conflict
			// "InvalidSubnetID.Malformed",
			// InvalidSubnetID.NotFound
			// InvalidSubnetId.NotFound
			// InvalidSubnet.Range
			// "InvalidTagKey.Malformed",
			// InvalidTargetArn.Unknown
			// InvalidTenancy
			// InvalidTime
			// InvalidTrafficMirrorFilterNotFound
			// InvalidTrafficMirrorFilterRuleNotFound
			// InvalidTrafficMirrorSessionNotFound
			// InvalidTrafficMirrorTargetNoFound
			// "InvalidUserID.Malformed",
			// InvalidVolumeID.Duplicate
			// "InvalidVolumeID.Malformed",
			// InvalidVolumeID.ZoneMismatch
			// InvalidVolume.NotFound
			// InvalidVolume.ZoneMismatch
			// "InvalidVpcEndpointId.Malformed",
			// InvalidVpcEndpoint.NotFound
			// InvalidVpcEndpointId.NotFound
			// InvalidVpcEndpointService.NotFound
			// InvalidVpcEndpointServiceId.NotFound
			// InvalidVpcEndpointType
			// "InvalidVpcID.Malformed",
			// InvalidVpcID.NotFound
			// "InvalidVpcPeeringConnectionId.Malformed",
			// InvalidVpcPeeringConnectionID.NotFound
			// InvalidVpcPeeringConnectionState.DnsHostnamesDisabled
			// InvalidVpcRange
			// InvalidVpcState
			// InvalidVpnConnectionID
			// InvalidVpnConnectionID.NotFound
			// InvalidVpnConnection.InvalidState
			// InvalidVpnConnection.InvalidType
			// InvalidVpnGatewayAttachment.NotFound
			// InvalidVpnGatewayID.NotFound
			// InvalidVpnGatewayState
			// InvalidZone.NotFound
			// KeyPairLimitExceeded
			// LegacySecurityGroup
			// LimitPriceExceeded
			// LogDestinationNotFoundException
			// LogDestinationPermissionIssue
			// MaxConfigLimitExceededException
			// MaxIOPSLimitExceeded
			// MaxScheduledInstanceCapacityExceeded
			// MaxSpotFleetRequestCountExceeded
			// MaxSpotInstanceCountExceeded
			// MaxTemplateLimitExceeded
			// MaxTemplateVersionLimitExceeded
			// MissingInput
			// NatGatewayLimitExceeded
			// "NatGatewayMalformed",
			// NatGatewayNotFound
			// NetworkAclEntryAlreadyExists
			// NetworkAclEntryLimitExceeded
			// NetworkAclLimitExceeded
			// NetworkInterfaceLimitExceeded
			// NetworkInterfaceNotFoundException
			// NetworkInterfaceNotSupportedException
			// NetworkLoadBalancerNotFoundException
			// NlbInUseByTrafficMirrorTargetException
			// NonEBSInstance
			// NoSuchVersion
			// NotExportable
			// OperationNotPermitted
			// OutstandingVpcPeeringConnectionLimitExceeded
			// PendingSnapshotLimitExceeded
			// PendingVpcPeeringConnectionLimitExceeded
			// PlacementGroupLimitExceeded
			// PrivateIpAddressLimitExceeded
			// RequestResourceCountExceeded
			// ReservationCapacityExceeded
			// ReservedInstancesCountExceeded
			// ReservedInstancesLimitExceeded
			// ReservedInstancesUnavailable
			// Resource.AlreadyAssigned
			// Resource.AlreadyAssociated
			// ResourceCountExceeded
			// ResourceCountLimitExceeded
			// ResourceLimitExceeded
			// RouteAlreadyExists
			// RouteLimitExceeded
			// RouteTableLimitExceeded
			// RulesPerSecurityGroupLimitExceeded
			// ScheduledInstanceLimitExceeded
			// ScheduledInstanceParameterMismatch
			// ScheduledInstanceSlotNotOpen
			// ScheduledInstanceSlotUnavailable
			// SecurityGroupLimitExceeded
			// SecurityGroupsPerInstanceLimitExceeded
			// SecurityGroupsPerInterfaceLimitExceeded
			// SignatureDoesNotMatch
			// SnapshotCopyUnsupported.InterRegion
			// SnapshotCreationPerVolumeRateExceeded
			// SnapshotLimitExceeded
			// SubnetLimitExceeded
			// TagLimitExceeded
			// TargetCapacityLimitExceededException
			// TrafficMirrorFilterInUse
			// TrafficMirrorSessionsPerInterfaceLimitExceeded
			// TrafficMirrorSessionsPerTargetLimitExceeded
			// TrafficMirrorSourcesPerTargetLimitExceeded
			// TrafficMirrorTargetInUseException
			// TrafficMirrorFilterLimitExceeded
			// TrafficMirrorFilterRuleLimitExceeded
			// TrafficMirrorSessionLimitExceeded
			// TrafficMirrorFilterRuleAlreadyExists
			// UnavailableHostRequirements
			// UnknownPrincipalType.Unsupported
			// UnknownVolumeType
			// Unsupported
			// UnsupportedException
			// UnsupportedHibernationConfiguration
			// UnsupportedHostConfiguration
			// UnsupportedInstanceTypeOnHost
			// UnsupportedTenancy
			// UpdateLimitExceeded
			// VolumeInUse
			// VolumeIOPSLimit
			// VolumeLimitExceeded
			// VolumeModificationSizeLimitExceeded
			// VolumeTypeNotAvailableInZone
			// VpcCidrConflict
			// VPCIdNotSpecified
			// VpcEndpointLimitExceeded
			// VpcLimitExceeded
			// VpcPeeringConnectionAlreadyExists
			// VpcPeeringConnectionsPerVpcLimitExceeded
			// VPCResourceNotSpecified
			// VpnConnectionLimitExceeded
			// VpnGatewayAttachmentLimitExceeded
			// VpnGatewayLimitExceeded
			// ZonesMismatched
			// ---- Server error codes
			"InsufficientAddressCapacity",
			"InsufficientCapacity",
			"InsufficientInstanceCapacity",
			"InsufficientHostCapacity",
			"InsufficientReservedInstanceCapacity",
			"InsufficientVolumeCapacity",
			"InternalError",
			"InternalFailure",
			"RequestLimitExceeded",
			"ServiceUnavailable",
			"Unavailable",
			"hey!, hi dev! :)":
			phcontext := &hook_providers.ProviderHookContext{
				Logger: p.Logger,
				Store:  p.store,
			}
			return hook_providers.DefaultOnActionErrorHook(phcontext, aout)
		}
	}
	return nil, nil
}

func (p *Provider) touchSession() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.store.GetPrivateVar("awsSess") != nil {
		return nil
	}

	p.Logger.LogInfo("Initializing AWS session...")

	// Init session
	// NewSessionWithOptions + SharedConfigState + SharedConfigEnable = use
	// credentials and config from ~/.aws/config and ~/.aws/credentials
	sess, serr := session.NewSessionWithOptions(session.Options{
		Config:            *aws.NewConfig().WithMaxRetries(0),
		SharedConfigState: session.SharedConfigEnable,
	})
	if serr != nil {
		return &base.ProviderAuthError{Err: serr}
	}

	// Save session into store, this struct and his values are ephemeral
	p.store.SetPrivateVar("awsSess", sess)

	// Check that the credentials have been provided. Here its validity
	// is not checked, only its existence.
	credentials, err := sess.Config.Credentials.Get()
	if err != nil {
		return &base.ProviderAuthError{Err: err}
	}

	p.Logger.LogInfo("AWS: Using access key id: " + credentials.AccessKeyID[:3] + "..." + credentials.AccessKeyID[len(credentials.AccessKeyID)-3:])

	// All credential parameters has been provided, but not validated
	return nil
}
