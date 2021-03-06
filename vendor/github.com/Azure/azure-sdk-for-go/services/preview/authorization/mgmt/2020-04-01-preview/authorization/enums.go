package authorization

// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.
//
// Code generated by Microsoft (R) AutoRest Code Generator.
// Changes may cause incorrect behavior and will be lost if the code is regenerated.

// PrincipalType enumerates the values for principal type.
type PrincipalType string

const (
	// Application ...
	Application PrincipalType = "Application"
	// DirectoryObjectOrGroup ...
	DirectoryObjectOrGroup PrincipalType = "DirectoryObjectOrGroup"
	// DirectoryRoleTemplate ...
	DirectoryRoleTemplate PrincipalType = "DirectoryRoleTemplate"
	// Everyone ...
	Everyone PrincipalType = "Everyone"
	// ForeignGroup ...
	ForeignGroup PrincipalType = "ForeignGroup"
	// Group ...
	Group PrincipalType = "Group"
	// MSI ...
	MSI PrincipalType = "MSI"
	// ServicePrincipal ...
	ServicePrincipal PrincipalType = "ServicePrincipal"
	// Unknown ...
	Unknown PrincipalType = "Unknown"
	// User ...
	User PrincipalType = "User"
)

// PossiblePrincipalTypeValues returns an array of possible values for the PrincipalType const type.
func PossiblePrincipalTypeValues() []PrincipalType {
	return []PrincipalType{Application, DirectoryObjectOrGroup, DirectoryRoleTemplate, Everyone, ForeignGroup, Group, MSI, ServicePrincipal, Unknown, User}
}
