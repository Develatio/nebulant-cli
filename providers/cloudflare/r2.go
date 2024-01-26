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

package cloudflare

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/smithy-go"

	"github.com/develatio/nebulant-cli/base"
	"github.com/develatio/nebulant-cli/blueprint"
	hook_providers "github.com/develatio/nebulant-cli/hook/providers"
	"github.com/develatio/nebulant-cli/providers/cloudflare/actors"
)

func ActionValidator(action *blueprint.Action) error {
	if action.Provider != "cloudflare" {
		return nil
	}
	al, exists := actors.ActionFuncMap[action.ActionName]
	if !exists {
		return fmt.Errorf("cloudflare: invalid action name " + action.ActionName)
	}
	if al.N == actors.NextOK && len(action.NextAction.NextKo) > 0 {
		return fmt.Errorf("cloudflare: action " + action.ActionName + " has no KO port")
	}

	if al.N == actors.NextKO && len(action.NextAction.NextOk) > 0 {
		return fmt.Errorf("cloudflare: action " + action.ActionName + " has no OK port")
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
	cfg := p.store.GetPrivateVar("r2AwsConfig")
	if cfg != nil {
		newConf := cfg.(aws.Config).Copy()
		freshStore.SetPrivateVar("r2AwsConfig", newConf)
	}
}

// HandleAction func
func (p *Provider) HandleAction(actx base.IActionContext) (*base.ActionOutput, error) {
	action := actx.GetAction()
	p.Logger.LogDebug("AWS: Received action " + action.ActionName)
	err := p.touchSession()
	if err != nil {
		return nil, err
	}

	if al, exists := actors.ActionFuncMap[action.ActionName]; exists {
		cfg := p.store.GetPrivateVar("r2AwsConfig").(aws.Config)
		p.Logger.LogDebug("Launching provider func")
		return al.F(actors.NewActionContext(cfg, action, p.store, p.Logger))
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
	var aerr smithy.APIError
	if errors.As(aout.Records[0].Error, &aerr) {

		// retry on allowed-retry aws http status codes
		if aerr.ErrorCode() == "RequestFailure" {
			switch aerr.ErrorFault() {
			case smithy.FaultServer:
				phcontext := &hook_providers.ProviderHookContext{
					Logger: p.Logger,
					Store:  p.store,
				}
				return hook_providers.DefaultOnActionErrorHook(phcontext, aout)
			}
		}

		// retry if:
		switch aerr.ErrorCode() {
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

	if p.store.GetPrivateVar("r2AwsConfig") != nil {
		return nil
	}

	p.Logger.LogInfo("Initializing CloudFlare provider...")

	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")
	if len(accountID) <= 3 {
		return fmt.Errorf("please, provide CLOUDFLARE_ACCOUNT_ID")
	}
	accessKeyID := os.Getenv("CLOUDFLARE_ACCESS_KEY_ID")
	if len(accessKeyID) <= 3 {
		return fmt.Errorf("please, provide CLOUDFLARE_ACCESS_KEY_ID")
	}
	accessKeySecret := os.Getenv("CLOUDFLARE_SECRET_ACCESS_KEY")
	if len(accessKeySecret) <= 3 {
		return fmt.Errorf("please, provide CLOUDFLARE_SECRET_ACCESS_KEY")
	}

	// https: //developers.cloudflare.com/r2/examples/aws/aws-sdk-go/
	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID),
		}, nil
	})

	// Init config for r2 (aws s3 lib)
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithEndpointResolverWithOptions(r2Resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, accessKeySecret, "")),
		config.WithRetryMaxAttempts(0),
	)
	if err != nil {
		return err
	}

	// Save session into store, this struct and his values are ephemeral
	p.store.SetPrivateVar("r2AwsConfig", cfg)

	p.Logger.LogInfo("R2: Using access key id: " + accessKeyID[:3] + "..." + accessKeyID[len(accessKeyID)-3:])

	// All credential parameters has been provided, but not validated
	return nil
}
